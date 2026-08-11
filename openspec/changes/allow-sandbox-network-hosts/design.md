## Context

See `proposal.md` for motivation. Docker Sandboxes supports local allow rules
scoped by sandbox name through `sbx policy allow network --sandbox`, and removes
them by sandbox and resource through `sbx policy rm network --sandbox
--resource`. Organisation-managed allow policy takes precedence over local
allow rules.

The lifecycle application package already owns create, recreate, and destroy
coordination. Its current recreation path calls sandbox removal directly rather
than reusing the destroy runner, while the declaration adapter's teardown loader
intentionally resolves only identity.

## Goals / Non-Goals

**Goals:**

- Preserve declaration order from YAML through lifecycle and Docker requests.
- Give recreation and destroy one removal-and-network-cleanup implementation.
- Keep subprocess construction in the outbound adapter and orchestration in the
  lifecycle application package.
- Preserve teardown's ability to ignore unrelated invalid agent and kit fields.

**Non-Goals:**

- Reconcile changes when entering an existing sandbox without `--recreate`.
- Predict whether organisation governance will activate a local allow rule.
- Implement Docker's network-resource parser in sbxflow.
- Track or protect manual edits that overlap resources declared as sbxflow-owned.

## Decisions

### Model network resources explicitly at the declaration boundary

Add `Network` to the declaration sandbox model with ordered `AllowedHosts`.
Although Docker accepts resources broader than bare hosts, `allowedHosts`
matches the public issue language and Docker's user-facing host terminology.
The schema checks strings for presence and uniqueness; Docker remains the source
of truth for wildcard, URL, IP, and port syntax.

Parsing the syntax in sbxflow was rejected because it would duplicate a
versioned Docker contract and risk rejecting resources supported by the pinned
CLI.

### Add narrow network-rule capabilities to the Sandbox port

Represent add and remove requests as structured port values containing sandbox
name and ordered resources. The outbound adapter translates them into one
`sbx policy allow network --sandbox` invocation and resource-specific
`sbx policy rm network --sandbox --resource` invocations. Captured execution is
used because these operations do not require an attached terminal.

Generating a custom kit internally was rejected because it would introduce
temporary filesystem artifacts and local-kit trust solely to reach a native
Docker policy interface.

### Apply the rule only while creating or recreating

For a missing sandbox, `up` adds the scoped rule before `sbx run` so selected
kits and the initial agent session receive access. An existing sandbox remains
untouched unless `--recreate` is requested, consistent with the established
no-reconciliation contract.

Applying on every `up` was rejected because additive policy calls cannot remove
hosts deleted from the declaration and could accumulate stale rules.

### Share removal orchestration inside the lifecycle package

Introduce a lifecycle removal component used by both the destroy runner and the
recreation branch. It first performs Docker's existing guarded sandbox removal
and only after success removes each currently declared network resource. A
cleanup failure reports that removal already occurred and returns failure.

This remains in `internal/application/lifecycle`: creating a new application
package would violate the no-peer-application-dependency rule, while placing
external-operation sequencing in the domain would misclassify orchestration as
a reusable business rule. The target loader will be extended narrowly to retain
valid declared network resources while continuing to ignore unrelated agent and
kit configuration.

Removing rules before the sandbox was rejected because a cancelled or failed
sandbox removal would unexpectedly revoke access from a surviving sandbox.

## Risks / Trade-offs

- **[Creation fails after the allow rule is added]** → Attempt compensating
  removal and report both failures if compensation also fails.
- **[Cleanup fails after permanent sandbox removal]** → Return a precise partial
  completion error containing the sandbox and resource so the user can retry the
  documented `sbx policy rm` command.
- **[A user manually changes an overlapping resource]** → Document declared
  resources as sbxflow-owned and remove them according to the declaration.
- **[Organisation governance makes the local rule inactive]** → Do not claim
  effective reachability; document that organisation policy remains stricter.
- **[A removed host is no longer present in the current declaration]** → The
  normal no-reconciliation contract means users must recreate while the old
  declaration still identifies resources to clean. Manual drift is outside the
  ownership guarantee.

## Migration Plan

The field is optional, so existing version 1 declarations retain their current
behavior. Rollback consists of removing `sandbox.network` and recreating the
sandbox after removing any previously declared scoped resources with Docker's
policy command.

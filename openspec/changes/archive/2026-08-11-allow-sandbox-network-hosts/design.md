## Context

See `proposal.md` for motivation. Docker Sandboxes supports local allow rules
scoped by sandbox name through `sbx policy allow network --sandbox`, and removes
them by sandbox and resource through `sbx policy rm network --sandbox
--resource`. Organisation-managed allow policy takes precedence over local
allow rules.

Three properties of that CLI (verified against sbx v0.35.0) constrain the design:

- `sbx policy allow network --sandbox NAME` fails with `sandbox "NAME" not found`
  unless the sandbox already exists, and neither `sbx run` nor `sbx create`
  accepts local policy at creation time. A rule therefore cannot precede the
  sandbox it is scoped to.
- `RESOURCES` is a single comma-separated positional argument. A second
  positional is rejected as a mistaken sandbox name.
- `sbx policy rm network` fails when the sandbox has no scoped policy
  (`no scoped policy found`) and when any named resource is not in that policy
  (`remove-resource: rule not found`), rejecting the whole request in the latter
  case rather than removing the resources it did find.

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
`allowedHosts` matches the public issue language and Docker's user-facing host
terminology. The schema checks presence, uniqueness, and that each entry is a
host, domain, wildcard subdomain, IP literal, or `**`, each with an optional
`:port` suffix.

Deferring resource syntax entirely to Docker was rejected. Docker's policy
authorizer matches requests by host and port, so `sbx policy allow network`
accepts `https://example.com` and stores it verbatim while it never matches any
request — an entry that reads as applied and silently is not. Validating the
accepted forms turns that into a validation error instead. The trade-off is that
sbxflow now mirrors part of a versioned Docker contract and could reject a form a
later CLI supports; the pattern is deliberately permissive about label characters
to limit that exposure.

### Add narrow network-rule capabilities to the Sandbox port

Keep `NetworkPolicy` a narrow interface in the existing `ports/sandbox` package
rather than folding its two methods into `Remover` and the create capability.
`sbx policy` is a distinct command family from `sbx run`, `sbx create`, and
`sbx rm`, and consumers already compose the interfaces they need, so the cost of
a separate interface is one embedded line per consumer.

Represent add and remove requests as structured port values containing sandbox
name and resources. The outbound adapter joins allow resources into the single
comma-separated argument Docker requires, and removes resources one invocation at
a time so that one already-absent resource cannot block cleanup of the rest.
Removal is idempotent at the adapter boundary: the adapter owns the CLI contract,
including its diagnostics, so it recognises both absence diagnostics and reports
success. Captured execution is used because these operations do not require an
attached terminal.

Generating a custom kit internally was rejected because it would introduce
temporary filesystem artifacts and local-kit trust solely to reach a native
Docker policy interface.

### Split creation from attachment so the rule can precede the agent

Docker will not scope a rule to a sandbox that does not exist yet, so `up` splits
what was one `sbx run` into `sbx create` followed by `sbx run --name`, and applies
the rule between them. Add a `Creator` port for the provisioning half; `Runner`
becomes attachment-only and its request drops the creation inputs and the
`Exists` discriminator.

This gives the rule force before the agent starts, but not before kit
provisioning, which happens inside `sbx create`. Kits declare their own
sandbox-scoped policy, which Docker applies as part of creation, so kit
provisioning does not depend on `allowedHosts`.

Because the sandbox exists by the time a rule can be rejected, compensation is
inverted relative to a single `sbx run`: a rejected rule removes the sandbox that
was just created, so `up` neither creates nor enters a sandbox lacking its
declared access.

An existing sandbox remains untouched unless `--recreate` is requested,
consistent with the established no-reconciliation contract. Applying on every
`up` was rejected because additive policy calls cannot remove hosts deleted from
the declaration and could accumulate stale rules.

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

- **[The allow rule is rejected after the sandbox is created]** → Remove the
  just-created sandbox and report both failures if that removal also fails.
- **[Attachment fails after the rule is applied]** → Leave the sandbox and its
  rule in place; the sandbox now exists, so a later `up` attaches to it without
  reapplying anything.
- **[Cleanup fails after permanent sandbox removal]** → Return a precise partial
  completion error containing the sandbox and resource so the user can retry the
  documented `sbx policy rm` command. Absence is not such a failure.
- **[`sbx create` diverges from `sbx run`'s implicit creation]** → Both accept the
  same `--name`, `--kit`, agent, and workspace inputs, and Docker documents
  `sbx run --name` as the way to attach after `sbx create`.
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

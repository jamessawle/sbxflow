## Context

See `proposal.md` for the motivation. Repository discovery is already isolated
in `config.Discover`, but `config.Load` validates the entire document against the
published schema before returning a typed configuration. The validation runner
then links selections, resolves local paths, and invokes Docker for selected
local kits. `up` needs all of those results; stopping and future removal need
only a safely interpreted declaration identity.

The lifecycle package already has exact-name discovery through
`sbx ls --quiet`, a captured subprocess runner for bounded inspection, and an
attached subprocess runner that preserves Docker's terminal output and process
result. The new path should reuse those boundaries without weakening `up` or
the standalone `validate` command.

## Goals / Non-Goals

**Goals:**

- Give teardown lifecycle operations one reusable resolver for the nearest
  declaration path and exact sandbox name.
- Preserve the YAML safety properties shared with complete loading while
  validating only the version and sandbox name consumed by teardown.
- Keep `down` idempotent when the named sandbox is absent or already stopped.
- Keep repository and Docker orchestration injectable and testable without a
  real sandbox.

**Non-Goals:**

- Relax or replace complete validation for `validate` or `up`.
- Interpret future configuration versions using version 1 assumptions.
- Inspect or reconcile an existing sandbox's agent, workspace, kits, or trust.
- Implement `destroy`, status reporting, lifecycle flags, or arbitrary sandbox
  names.
- Guarantee convergence in the presence of concurrent external sandbox
  creation, removal, or restart.

## Decisions

### Add a target-specific loader over a shared safe document decoder

Configuration loading will separate safe single-document YAML decoding from
the validation applied to the resulting JSON-compatible document. Both complete
loading and target loading will use the shared decoder, retaining malformed
YAML, duplicate-key, multi-document, and YAML-to-JSON conversion rejection.

The target loader will decode only `version` and `sandbox.name`, require version
1, require the nested objects and string fields to have the expected types, and
apply the same non-empty name constraint as the published version 1 schema.
Other fields remain uninterpreted and cannot trigger linking, path resolution,
kit validation, or trust derivation.

Calling the complete loader and selectively ignoring later failures was
rejected because schema validation would still reject unrelated fields before a
target could be returned. Parsing the name with an independent YAML path query
was rejected because it would create different ambiguity and document-safety
rules. Ignoring the version was rejected because a future schema may assign
different meaning to the same field path.

### Resolve repository targets behind a lifecycle interface

A repository target resolver in the lifecycle package will combine nearest-file
discovery, file reading, and target loading, and return the declaration path and
exact sandbox name. `down` will depend on a narrow resolver interface so command
tests can supply targets or failures directly. A future `destroy` runner can use
the same interface without depending on `down` or duplicating configuration
parsing.

Putting this orchestration into the complete validation report was rejected
because it would imply target-only operations require linked configuration and
local-kit results. Adding a user-facing skip-validation mode was rejected
because target resolution is itself the correct validation boundary, not a
bypass.

### List exact names before stopping

After resolving the target and locating `sbx`, `down` will reuse the existing
quiet name listing and whole-line comparison. An absent exact name is already in
the requested down state, so the command returns success without invoking
`sbx stop`. A present name is passed as the sole sandbox operand to
`sbx stop`.

Listing first makes the absent case explicitly idempotent without classifying
Docker error text. Directly calling `sbx stop` and interpreting a not-found
diagnostic was rejected because that diagnostic is not a stable machine-readable
contract. Inspecting JSON status was rejected because status is unnecessary for
the operation and would couple sbxflow to a larger output schema; Docker's stop
operation handles an already stopped sandbox.

The list-then-stop sequence has a race, as does the existing list-then-run path
for `up`. Docker remains authoritative: a failure after an existence result is
reported rather than retried or converted to success.

### Attach the stop subprocess to the user streams

The stop request will use the existing no-shell attached subprocess boundary so
Docker's standard output and diagnostics remain visible once, normal
cancellation reaches the process, and sbxflow does not impose an artificial
stop timeout. No standard input or kit-trust environment overrides are needed.
Failures already rendered by the attached Docker process will use the existing
error-suppression pattern so Cobra does not append a duplicate diagnostic.

Using the captured lookup runner for the state-changing stop was considered,
but it would buffer observable output and inherit the short inspection timeout.
A new subprocess abstraction was rejected because the existing attached runner
already has the required behavior despite having been introduced for an
interactive agent session.

### Keep the CLI surface symmetrical with up

`down` will be a Cobra command that obtains the current working directory,
accepts no positional arguments or command-specific flags, and delegates to an
injected runner. Help execution and argument rejection occur before repository
discovery or Docker invocation. Root construction will accept a down runner in
the same style as the existing doctor, validate, and up dependencies.

The CLI will not print `Configuration valid` because target resolution does not
claim complete validation. Docker's successful stop output remains the
operation result; an absent sandbox may complete silently.

## Risks / Trade-offs

- **Invalid unrelated configuration can coexist with a successful `down`** →
  Document that `down` validates identity only and keep `validate` as the
  complete configuration check.
- **Target constraints can drift from the published schema** → Share document
  decoding and cover version/name acceptance and rejection with tests derived
  from the schema's current constraints.
- **Listing and stopping are not atomic** → Preserve Docker's subsequent error
  rather than hiding races or adding retries with surprising effects.
- **An absent sandbox produces no Docker confirmation** → Specify successful
  no-op behavior and cover it at both lifecycle and CLI boundaries.
- **Stopping ends active agent and service processes** → Keep `down` explicit,
  document its effect, and preserve all persisted sandbox state; do not add a
  confirmation intended for destructive removal.

## Migration Plan

Add the target loader and resolver without changing existing complete loading,
then register `down` and update documentation. Rollback consists of removing the
new command and resolver; the configuration format and existing commands require
no data or state migration.

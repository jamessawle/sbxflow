## Why

`up` currently enters an existing sandbox even when the repository declaration
has changed, forcing users to run a separate destructive command before they can
rebuild it from current configuration. A deliberate recreation option makes
that refresh a single, explicit lifecycle operation.

## What Changes

- Add a `--recreate` flag to `sbxflow up` without adding a shorthand or
  positional arguments.
- Preserve the normal create-and-enter flow when the declared sandbox does not
  exist.
- When the exact declared sandbox exists, force-remove it without confirmation,
  then create and enter a replacement from the validated current declaration.
- Stop without creating or entering a sandbox if forced removal fails.
- Document the destructive, state-resetting behavior and cover the CLI and
  lifecycle branches with tests.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `cli-interface`: Allow and document the `up --recreate` option while
  continuing to reject positional arguments and unrelated flags.
- `sandbox-lifecycle`: Define validation, exact-name removal, failure handling,
  and create-and-enter behavior for requested recreation.

## Impact

- Changes the public `up` CLI contract and README lifecycle documentation.
- Extends the inbound CLI-to-application request with the recreation choice.
- Extends the existing lifecycle workflow to compose sandbox lookup, forced
  removal, and creation through the existing Sandbox ports and adapter.
- Adds CLI, application workflow, wiring, and executable-level regression
  coverage; no configuration schema or external dependency changes are
  required.

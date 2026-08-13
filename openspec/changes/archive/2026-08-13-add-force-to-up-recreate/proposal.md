## Why

`sbxflow up --recreate` cannot currently approve destructive recreation non-interactively when the declared sandbox is running. Automation and already-informed users need an explicit confirmation bypass without making ordinary `up` destructive.

## What Changes

- Add a long-only `--force` option to `sbxflow up` that is valid only with `--recreate`.
- Make `sbxflow up --recreate --force` skip sbxflow's running-sandbox confirmation while retaining the existing exact-name force-removal, cleanup, replacement, and entry behavior.
- Reject `sbxflow up --force` as an invalid invocation before validation or lifecycle work begins.
- Preserve the existing interactive confirmation behavior for `sbxflow up --recreate` without `--force`.
- Document that forced recreation permanently removes persisted sandbox state and can terminate other attached sessions.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `cli-interface`: Add and constrain the `up --force` option, its help text, and its interaction with running-sandbox confirmation.
- `sandbox-lifecycle`: Allow explicitly forced recreation to bypass running-sandbox confirmation while preserving the existing recreation lifecycle and safety boundaries.

## Impact

- Public CLI: `sbxflow up` gains a long-only `--force` flag requiring `--recreate`.
- Application lifecycle: up options and recreation approval logic gain an explicit force choice and defensive invalid-combination handling.
- Tests: CLI adapter, lifecycle workflow, and executable-level coverage changes.
- Documentation and specifications: README plus the `cli-interface` and `sandbox-lifecycle` capabilities.
- No new packages, dependencies, configuration fields, or architecture relationships are required.

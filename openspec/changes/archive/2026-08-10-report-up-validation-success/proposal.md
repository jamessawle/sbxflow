## Why

`sbxflow up` currently performs complete repository validation silently when it
succeeds, so users cannot tell that startup has passed configuration checks
before Docker inspection and agent attachment begin. A concise success message
makes that phase boundary visible and distinguishes later Docker failures from
configuration failures.

## What Changes

- Report the successfully validated declaration path during `sbxflow up`.
- Write the status to standard error after complete validation and before any
  Docker sandbox inspection or lifecycle action.
- Keep the standalone `sbxflow validate` success report unchanged.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `cli-interface`: Define the successful startup-validation status emitted by
  the interactive `up` command.

## Impact

- Affects `up` lifecycle orchestration and its CLI and executable-level tests.
- Updates README documentation for the observable `up` output.
- Adds no dependencies and does not change validation or Docker lifecycle
  semantics.

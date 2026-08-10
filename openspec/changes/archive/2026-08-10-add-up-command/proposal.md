## Why

sbxflow can validate a repository declaration but cannot yet use it to enter the
declared Docker Sandbox. Adding a focused interactive `up` command provides the
first useful lifecycle path while relying on Docker Sandboxes for creation,
restart, and agent attachment behavior.

## What Changes

- Add an interactive `sbxflow up` command that discovers and validates the
  nearest repository declaration before invoking Docker Sandboxes.
- Create and enter the declared sandbox, including its ordered kits and
  process-local trust settings, when its name does not exist.
- Enter an existing named sandbox through `sbx run`, allowing Docker Sandboxes
  to attach immediately when running or restart it when stopped.
- Preserve the user's terminal streams, signals, and the `sbx run` exit result
  for the interactive session.
- Explicitly defer inspection, drift detection, kit reconciliation, automatic
  recreation, and agent-argument passthrough for existing sandboxes.
- Update the documented lifecycle scope to describe `up` as an interactive
  entry command rather than an implementation of desired-state reconciliation.

## Capabilities

### New Capabilities

- `sandbox-lifecycle`: Defines repository-aware sandbox creation, restart, and
  interactive agent entry through `sbxflow up`.

### Modified Capabilities

- `cli-interface`: Makes `up` discoverable and defines its contextual help and
  argument boundary.

## Impact

- Affects the Cobra command tree, repository validation integration, linked kit
  execution references, subprocess execution, and command-level tests.
- Adds quiet sandbox-name discovery and an interactive `sbx run`
  subprocess contract against the supported Docker Sandboxes version range.
- Requires process-local Docker kit trust environment variables without
  modifying global Docker Sandboxes settings.
- Updates README lifecycle documentation; `down`, `destroy`, and configuration
  reconciliation remain unimplemented.

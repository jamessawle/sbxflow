## Why

`sbxflow up --recreate` currently force-removes an existing sandbox without
checking whether it is running, so it can unexpectedly terminate an agent
session attached from another terminal. Recreation should require deliberate
confirmation when that specific risk is present while remaining convenient for
stopped sandboxes.

## What Changes

- Inspect the exact declared sandbox's lifecycle state when recreation is
  requested.
- Recreate absent and stopped sandboxes through their existing paths without an
  additional prompt.
- Warn and request confirmation before force-removing a running sandbox,
  defaulting to cancellation.
- Fail without removal when state cannot be determined, confirmation input is
  unavailable, or the user declines.
- Keep forced removal after confirmation so an approved recreation can replace
  a sandbox with an active session.
- Update CLI help and lifecycle documentation to describe the conditional
  confirmation and its effect on other attached terminals.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `sandbox-lifecycle`: Make recreation state-aware and require confirmation
  before replacing a running sandbox.
- `cli-interface`: Define the running-sandbox warning, confirmation interaction,
  cancellation behavior, and command-stream requirements.

## Impact

- Changes the externally observable behavior of `sbxflow up --recreate` for a
  running sandbox.
- Extends the Sandbox port and Docker Sandboxes adapter with narrow exact-name
  lifecycle-state inspection.
- Adds confirmation coordination across the lifecycle application and CLI
  adapter without changing the repository configuration schema or architecture
  dependency matrix.
- Requires tests for state parsing, branch ordering, confirmation responses,
  non-interactive input, inspection failures, and end-to-end CLI streams.

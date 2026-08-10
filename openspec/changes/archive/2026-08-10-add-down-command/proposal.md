## Why

Stopping an existing sandbox currently depends on a validation pipeline whose
agent, kit, local-path, and trust checks are irrelevant to that operation. A
smaller declaration-identity boundary will let lifecycle teardown operations
reliably target the repository's named sandbox even when creation configuration
has become invalid or unavailable.

## What Changes

- Add a reusable lifecycle-target resolution path that discovers the nearest
  `sbxflow.yaml`, safely parses its supported version and non-empty
  `sandbox.name`, and does not validate unrelated agent or kit configuration.
- Add `sbxflow down` to stop the declared sandbox while preserving its persisted
  state.
- Make `down` a repository-aware, argument-free command that treats an absent
  declared sandbox as already down and delegates stopping an existing sandbox
  to Docker Sandboxes.
- Preserve complete validation for `up` and `validate`; the smaller route is
  designed for reuse by a future `destroy` implementation without adding that
  command in this change.
- Update lifecycle documentation and CLI help to expose the new behavior and
  distinguish identity resolution from complete configuration validation.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `configuration-validation`: Define the smaller, reusable declaration-identity
  resolution boundary used by teardown lifecycle operations.
- `sandbox-lifecycle`: Define repository-aware, state-preserving sandbox stop
  behavior through `down`.
- `cli-interface`: Make `down` discoverable and define its argument and help
  contract.

## Impact

- Affects configuration document loading, lifecycle orchestration, Cobra command
  construction, subprocess error handling, and their unit tests.
- Adds `sbx ls --quiet` and `sbx stop <declared-name>` execution to the `down`
  path without adding new dependencies or changing the configuration schema.
- Updates README lifecycle documentation; `destroy` and configuration
  reconciliation remain unimplemented.

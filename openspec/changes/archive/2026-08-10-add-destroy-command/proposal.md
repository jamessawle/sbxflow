## Why

sbxflow can create, enter, and stop a repository's declared sandbox, but users
must leave the repository-aware workflow to remove it permanently. Adding the
planned `destroy` command completes the initial lifecycle while retaining exact
targeting and an explicit confirmation boundary for destructive state removal.

## What Changes

- Add `sbxflow destroy` to remove the exact sandbox named by the nearest
  repository declaration and all of its persisted Docker Sandbox state.
- Reuse lifecycle-target resolution so destruction depends only on a safely
  interpreted configuration version and `sandbox.name`, not complete agent or
  kit validation.
- Make destruction idempotent when the declared sandbox is absent and protect
  similarly named sandboxes through exact-name matching.
- Delegate removal and its default confirmation prompt to `sbx rm`, preserving
  Docker's attached terminal output, diagnostics, cancellation, and exit status.
- Add `--force`/`-f` to skip confirmation and permit Docker to remove a sandbox
  with an active session; continue rejecting positional sandbox names and all
  other command-specific flags.
- Update help and lifecycle documentation to distinguish permanent destruction
  from the state-preserving `down` command.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `cli-interface`: Make `destroy` discoverable and define its argument, force
  flag, help, and input-forwarding contract.
- `sandbox-lifecycle`: Define repository-aware, exact-name, confirmed and forced
  permanent sandbox removal behavior.

## Impact

- Affects lifecycle orchestration, Cobra command construction and dependency
  injection, subprocess stream handling, executable-level fixtures, and their
  tests.
- Adds `sbx rm <declared-name>` and `sbx rm --force <declared-name>` execution
  without changing the configuration schema or adding dependencies.
- Updates the README to expose `destroy` as available and explain its
  irreversible effect on sandbox state while preserving the host repository.

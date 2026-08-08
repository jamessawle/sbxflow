## Why

Users need a quick, repository-independent way to determine whether Docker
Sandboxes is healthy and safely configured before relying on sbxflow lifecycle
commands. Docker's own diagnostics cover runtime health, while sbxflow can add a
small compatibility and configuration-posture layer without duplicating those
checks.

## What Changes

- Add an `sbxflow doctor` command that can run from any directory without
  loading `sbxflow.yaml`.
- Check that `sbx` is installed at a version compatible with sbxflow.
- Run Docker's machine-readable diagnostics and report only their aggregate
  pass, warning, failure, and skipped totals, directing users to `sbx diagnose`
  for details.
- Check that the global Docker Sandboxes network policy is initialized.
- Warn when the effective global `kit.allowedSources` setting contains the
  unrestricted `"*"` entry, while respecting the setting's reported source.
- Implement the checks as independent strategies with consistent results,
  dependency-aware skipping, and non-mutating remediation guidance.
- Remove the root help-rendering workaround that exists only to expose `help`
  when it is the command tree's sole subcommand; Cobra's normal rendering can
  expose both `help` and `doctor` once the real command is registered.
- Update CLI help and documentation to describe the available command and its
  system-level scope.

## Capabilities

### New Capabilities

- `environment-diagnostics`: Defines system-level Docker Sandboxes
  compatibility, diagnostic summary, and configuration-posture checks.

### Modified Capabilities

- `cli-interface`: Makes the `doctor` command discoverable through the existing
  command tree and help interface.

## Impact

- Extends the internal CLI command tree and command-level/process-level tests.
- Adds internal abstractions for executing `sbx` and running diagnostic
  strategies, without adding configuration loading or changing lifecycle
  behavior.
- Depends on machine-readable `sbx version`, `sbx diagnose`, `sbx policy`, and
  `sbx settings` command output within an explicitly supported CLI range.
- Revises the README's planned doctor description, which currently refers to
  declarations and local paths, to reflect repository-independent system
  diagnostics.

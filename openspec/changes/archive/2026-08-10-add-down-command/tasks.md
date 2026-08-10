## 1. Lifecycle Target Loading

- [x] 1.1 Refactor configuration loading around a shared single-document YAML
      decoder while preserving all existing complete-validation behavior and
      tests.
- [x] 1.2 Add and test target-only loading for supported version 1 and a
      non-empty `sandbox.name`, including malformed, ambiguous, unsupported,
      missing, and incorrectly typed identity cases.
- [x] 1.3 Prove with tests that target-only loading accepts invalid or
      unavailable agent and kit configuration without linking selections,
      resolving local paths, or invoking Docker kit validation.

## 2. Repository Target Resolution

- [x] 2.1 Add an injectable lifecycle target resolver that discovers and reads
      the nearest declaration and returns its path and exact target name.
- [x] 2.2 Cover nearest-ancestor selection, absent and unreadable declarations,
      identity-loading failures, and successful target resolution with focused
      tests suitable for later `destroy` reuse.

## 3. Down Lifecycle Execution

- [x] 3.1 Add a down runner that resolves the target, locates `sbx`, and reuses
      exact quiet-name lookup without invoking complete configuration or
      local-kit validation.
- [x] 3.2 Implement and test successful no-op behavior for absent names and
      attached `sbx stop <exact-name>` execution for existing sandboxes,
      including similar names and already-stopped behavior.
- [x] 3.3 Preserve list and stop diagnostics, subprocess results, output streams,
      cancellation, and single-rendered-error behavior in failure tests.

## 4. CLI and Documentation

- [x] 4.1 Register an injected, argument-free `down` Cobra command with root and
      contextual help coverage and parse-error tests that prove no lifecycle
      work occurs.
- [x] 4.2 Update README command, lifecycle, validation-boundary, persistence, and
      planned-command documentation while keeping `destroy` explicitly
      unimplemented.

## 5. Validation

- [x] 5.1 Run `mise run fmt` and inspect the resulting Go and Markdown changes.
- [x] 5.2 Run focused Go tests for configuration, lifecycle, and CLI packages,
      then run `mise run validate` and resolve every failure.

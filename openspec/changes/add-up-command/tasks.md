## 1. Lifecycle Inputs

- [x] 1.1 Add lifecycle planning that converts validated linked selections and
      local-kit results into ordered Docker `--kit` execution references.
- [x] 1.2 Encode the derived remote-source allowlist and local-kit permission as
      process-local Docker Sandboxes environment overrides while preserving the
      remaining host environment.
- [x] 1.3 Add unit coverage for mixed Git, OCI, and local selections, declaration
      order, canonical local paths, and exact trust values.

## 2. Docker Execution Boundaries

- [x] 2.1 Add bounded sandbox-name discovery through `sbx ls --quiet`, including
      exact-line matching and actionable lookup, invocation, and non-zero-exit
      failures.
- [x] 2.2 Add a separate no-timeout interactive subprocess runner that inherits
      stdin, stdout, stderr, signals, and controlled environment overrides without
      using a shell.
- [x] 2.3 Test name-list parsing and failures, plus interactive stdin, both output
      streams, environment replacement, and successful and unsuccessful process
      results with controlled subprocess fixtures.

## 3. Up Lifecycle Orchestration

- [x] 3.1 Implement the `up` runner so complete repository validation gates all
      sandbox-name discovery and agent execution, reusing cohesive validation error
      reporting.
- [x] 3.2 Implement the missing-name path as interactive `sbx run` with the
      declared name, agent, declaration-directory workspace, ordered kit arguments,
      and derived trust environment.
- [x] 3.3 Implement the existing-name path as interactive `sbx run <agent>
--name <name>` without workspace or kit creation arguments.
- [x] 3.4 Preserve Docker diagnostics and failure status without appending a
      redundant wrapper error after an attached subprocess has already written to
      standard error.
- [x] 3.5 Add orchestration tests proving validation failure performs no
      lifecycle lookup, lookup failure performs no run, and missing and existing
      names select the exact required argument shapes.

## 4. CLI Integration

- [x] 4.1 Register `up` in the Cobra command tree with no positional arguments or
      command-specific flags and with contextual create-or-enter help.
- [x] 4.2 Update root and command tests for `up` discovery, both help forms,
      invalid arguments, unsupported flags, and dependency injection.
- [x] 4.3 Add executable-level tests with a fake `sbx` proving nested declaration
      discovery, initial creation arguments and trust environment, existing-name
      attachment, terminal stream preservation, and exit status propagation.

## 5. Documentation and Validation

- [x] 5.1 Update README purpose, available-command, and lifecycle sections to
      document interactive `up`, creation and restart behavior, and the explicit
      absence of existing-sandbox reconciliation.
- [x] 5.2 Run `mise run fmt` and review all Markdown and Go formatting changes.
- [x] 5.3 Run `mise run validate` and resolve all specification, vet, test, and
      build failures.

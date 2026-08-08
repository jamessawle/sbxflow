## 1. Go Project Setup

- [x] 1.1 Add a Go module, pin a supported Go toolchain with Mise, and record
      Cobra at a compatible version.
- [x] 1.2 Extend the existing Mise formatting and aggregate validation tasks to
      format, build, vet, and test the Go application.

## 2. CLI Foundation

- [x] 2.1 Add a build-metadata package with a deterministic development version
      and linker-injectable release version and commit values.
- [x] 2.2 Implement a fresh root-command constructor with the sbxflow name,
      description, no-argument help behavior, explicit help command, and argument
      validation.
- [x] 2.3 Add the thin `cmd/sbxflow` process entry point and execute the root
      command with version metadata and completions disabled.
- [x] 2.4 Ensure Cobra routes successful help/version output to standard
      output and actionable command errors to standard error with non-zero status.

## 3. CLI Verification

- [x] 3.1 Add command-level tests for no-argument, `help`, `-h`, and `--help`
      behavior and for the expected initial help content.
- [x] 3.2 Add tests proving `-v` and `--version` return the same non-empty build
      identity while `version`, completion, man-page, lifecycle, and other unknown
      commands remain unavailable.
- [x] 3.3 Add built-executable coverage for exit statuses and stdout/stderr
      separation without relying on ANSI styling snapshots.

## 4. Documentation and Validation

- [x] 4.1 Update the README to distinguish the available foundational CLI from
      planned lifecycle commands and document help and version invocations.
- [x] 4.2 Update contributor guidance for Go setup, build, test, formatting, and
      validation through Mise.
- [x] 4.3 Run repository formatting and the complete Mise validation task, then
      address all reported issues.

## 5. Presentation Revision

- [x] 5.1 Remove Fang in favour of direct Cobra execution, update the semantic
      and executable tests for compact output, and rerun complete validation.

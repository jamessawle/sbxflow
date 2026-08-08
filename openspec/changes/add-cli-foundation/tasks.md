## 1. Go Project Setup

- [ ] 1.1 Add a Go module, pin a supported Go toolchain with Mise, and record
      Cobra and Fang at compatible versions.
- [ ] 1.2 Add Mise tasks for formatting, building, vetting, and testing the Go
      application, and include the non-mutating checks in aggregate validation.

## 2. CLI Foundation

- [ ] 2.1 Add a build-metadata package with a deterministic development version
      and linker-injectable release version and commit values.
- [ ] 2.2 Implement a fresh root-command constructor with the sbxflow name,
      description, no-argument help behavior, explicit help command, and argument
      validation.
- [ ] 2.3 Add the thin `cmd/sbxflow` process entry point and execute the root
      command through Fang with version metadata, completions disabled, and man
      pages disabled.
- [ ] 2.4 Ensure Fang and Cobra route successful help/version output to standard
      output and actionable command errors to standard error with non-zero status.

## 3. CLI Verification

- [ ] 3.1 Add command-level tests for no-argument, `help`, `-h`, and `--help`
      behavior and for the expected initial help content.
- [ ] 3.2 Add tests proving `-v` and `--version` return the same non-empty build
      identity while `version`, completion, man-page, lifecycle, and other unknown
      commands remain unavailable.
- [ ] 3.3 Add built-executable coverage for exit statuses and stdout/stderr
      separation without relying on ANSI styling snapshots.

## 4. Documentation and Validation

- [ ] 4.1 Update the README to distinguish the available foundational CLI from
      planned lifecycle commands and document help and version invocations.
- [ ] 4.2 Update contributor guidance for Go setup, build, test, formatting, and
      validation through Mise.
- [ ] 4.3 Run repository formatting and the complete Mise validation task, then
      address all reported issues.

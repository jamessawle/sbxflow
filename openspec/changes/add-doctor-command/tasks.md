## 1. Diagnostic Foundation

- [ ] 1.1 Add the context-aware subprocess runner that preserves stdout,
      stderr, and exit status separately and supports bounded command timeouts.
- [ ] 1.2 Define doctor result statuses, required-versus-advisory grading,
      strategy prerequisites, and the deterministic sequential runner.
- [ ] 1.3 Add unit tests for ordered execution, independent continuation,
      prerequisite skipping, warning-only success, and required-failure status.

## 2. Docker Sandboxes Strategies

- [ ] 2.1 Establish and document the initial supported `sbx` version range from
      the required command contracts, then implement executable lookup, version
      parsing, compatibility grading, and boundary tests.
- [ ] 2.2 Implement the Docker diagnostic summary strategy using only the
      versioned envelope and aggregate counts, including fixtures for unknown
      fields, diagnostic failures with valid JSON, malformed output, and
      unsupported envelope versions.
- [ ] 2.3 Implement the network-policy strategy with fixtures for initialized
      local policy, active organisation policy, absent policy, and unavailable
      machine-readable policy state without relying on individual diagnostic
      results.
- [ ] 2.4 Implement the `kit.allowedSources` strategy with source-aware guidance
      and fixtures for restricted prefixes, wildcard mixed with prefixes, local
      overrides, environment values, remotely managed values, and unreadable
      settings.
- [ ] 2.5 Verify all strategies are inspection-only and invoke no setting,
      policy, daemon, or setup mutation commands.

## 3. CLI Integration

- [ ] 3.1 Add the `doctor` Cobra command, inject the production doctor runner
      through command construction, and render ordered plain-text results using the
      existing stream conventions.
- [ ] 3.2 Register `doctor` with the root command, remove the obsolete
      sole-subcommand `HelpFunc` workaround, and preserve the existing help,
      version, error, completion, and man-page behavior.
- [ ] 3.3 Add command tests for root and contextual help, repository-independent
      execution, deterministic output, warning-only success, required failures,
      and skipped checks.
- [ ] 3.4 Add a built-executable test covering doctor stdout, stderr, and process
      exit status with a controlled fake `sbx` executable environment.

## 4. Documentation and Validation

- [ ] 4.1 Update README.md to move `doctor` into the available CLI, describe its
      system-level checks and non-mutating behavior, and remove declaration and
      local-path diagnostics from its scope.
- [ ] 4.2 Run `mise run fmt` and review all Markdown and Go formatting changes.
- [ ] 4.3 Run `mise run validate` and resolve all specification, lint, vet, test,
      and build failures.

## 1. Lifecycle Port Model

- [ ] 1.1 Add a backend-neutral environment value to the sandbox port for the
      name, agent, workspace, ordered kits, and derived trust, and use it in create,
      run, and command requests.
- [ ] 1.2 Update lifecycle orchestration and its fakes to build and pass the
      shared environment value without changing creation, network, initialization,
      or entry ordering.

## 2. Private Environment Rendering

- [ ] 2.1 Add schema-version-1 environment serialization in the outbound SBX
      adapter for full lifecycle definitions and name-only removal definitions.
- [ ] 2.2 Render each definition as an owner-only `.sbxenv.yaml` in an
      owner-only temporary directory outside the mounted workspace, and guarantee
      cleanup after success, failure, and cancellation.
- [ ] 2.3 Add adapter tests for ordered document fields, identity-only removal,
      permissions, workspace isolation, invalid temporary placement, and cleanup.

## 3. Environment Command Backend

- [ ] 3.1 Replace sandbox creation and interactive entry invocations with
      `sbx env create` and `sbx env run`, preserving streams and applying derived
      trust only to their subprocesses.
- [ ] 3.2 Replace initialization execution with `sbx env exec`, preserving the
      declared workspace default, detached stdin, forwarded output, literal
      argument vectors, and process-scoped trust.
- [ ] 3.3 Replace removal and rollback with `sbx env rm`, preserving exact-name
      targeting, force behavior, interactive safeguards, diagnostics, and
      identity-only teardown.
- [ ] 3.4 Extend adapter and executable tests to cover exact command arguments,
      stream behavior, trust isolation, existing-sandbox entry, and removal with
      unrelated invalid declaration fields.

## 4. Lifecycle Cleanup Integration

- [ ] 4.1 Remove explicit per-resource network cleanup from the shared removal
      path and let successful environment removal own sandbox-scoped resource
      cleanup.
- [ ] 4.2 Preserve exact-name state inspection before destruction and
      recreation so absent teardown remains silent and running-session safeguards
      remain effective.
- [ ] 4.3 Update lifecycle tests for destroy, forced recreation, network-policy
      rollback, initialization rollback, primary-plus-cleanup errors, and mixed
      kit/standalone network rules.

## 5. Compatibility and Documentation

- [ ] 5.1 Change doctor compatibility boundaries and tests to accept SBX 0.39.0
      inclusive through 0.40.0 exclusive.
- [ ] 5.2 Update the repository's `jamessawle/sbx-kits` references to
      `v2026.08.03` and confirm no incompatible `v2026.08.02` pin remains.
- [ ] 5.3 Update README lifecycle and compatibility documentation to describe
      the unchanged `sbxflow.yaml` interface, SBX 0.39.x requirement, and internal
      environment backend without presenting `.sbxenv.yaml` as user configuration.

## 6. Verification

- [ ] 6.1 Run focused Go unit and executable tests for the sandbox adapter,
      lifecycle application, and doctor boundaries.
- [ ] 6.2 Run repository formatting, linting, architecture, and full validation
      commands documented in `CONTRIBUTING.md`.
- [ ] 6.3 Repeat the disposable SBX 0.39 full-kit smoke test with
      `jamessawle/sbx-kits@v2026.08.03`, standalone allowed hosts, initialization,
      entry, failure rollback, and final environment removal, then verify no test
      sandbox or policy remains.

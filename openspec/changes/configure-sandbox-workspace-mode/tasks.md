## 1. Declaration Contract and Domain Policy

- [ ] 1.1 Extend the published JSON Schema with the non-empty strict `sandbox.workspace` object and `direct|clone` mode enum, then add schema/load tests covering explicit modes, omission, an empty object, unknown properties, and unsupported values.
- [ ] 1.2 Add presence-aware workspace mode types to the declaration port and configuration aliases, preserve them through decoding and linking, and verify tests distinguish omission from explicit `direct` while retaining explicit `clone`.
- [ ] 1.3 Normalize omitted mode to effective `direct` in the configuration domain and produce the ordered advisory warning without invalidating the resolution; verify domain tests cover omitted, explicit, and warning-plus-later-error results.

## 2. Validation Warning Reporting

- [ ] 2.1 Extend resolution and validation reports with warnings distinct from errors, preserve them through external local-kit validation, and verify `Valid()` and exit eligibility depend only on errors.
- [ ] 2.2 Update validation CLI rendering to send warnings to standard error while preserving the successful report and zero exit status, and verify warning-only, explicit-mode, error-only, and mixed warning/error output.
- [ ] 2.3 Render validation warnings from `up` before lifecycle inspection or mutation and continue only when the report has no errors; verify application and CLI tests cover warning-only continuation, explicit-mode silence, and mixed-result blocking.

## 3. Workspace-Aware Lifecycle Provisioning

- [ ] 3.1 Carry the effective backend-neutral workspace mode alongside the declaration-directory path through lifecycle planning and the sandbox port, and verify plan tests cover omitted, explicit direct, and clone inputs without changing kit, trust, network, or hook ordering.
- [ ] 3.2 Render direct workspaces using Docker's scalar path and clone workspaces using the `{path, clone: true}` environment form, and verify adapter tests cover exact YAML, cleanup, create, run, exec, and removal-document compatibility.
- [ ] 3.3 Verify lifecycle and executable integration tests cover clone creation and recreation, direct compatibility, Docker-authoritative clone failures, existing-sandbox non-reconciliation, and initialization/rollback behavior in both modes.

## 4. User Guidance and Destructive Semantics

- [ ] 4.1 Update generic recreate and destroy messaging to cover persisted work stored only inside the sandbox without requiring mode-specific teardown parsing, and verify CLI help and confirmation tests assert the expanded disclosure.
- [ ] 4.2 Update README configuration and lifecycle guidance to make explicit mode the normal form, recommend `clone` for new declarations, explain current omitted-mode compatibility and the planned pre-1.0 switch, and document host visibility, Git transfer, hooks, destroy, and recreation consequences.
- [ ] 4.3 Update first-party examples and example documentation so normal examples explicitly choose `direct` or `clone` while at least one compatibility fixture verifies omission still behaves as direct, then verify schema and executable tests consume the revised declarations.

## 5. Validation

- [ ] 5.1 Run `mise run fmt` and verify all Go and Markdown artifacts are formatted.
- [ ] 5.2 Run focused package tests for declaration, configuration, validation, lifecycle, sandbox rendering, CLI, schema, and executable behavior and resolve any regressions.
- [ ] 5.3 Run `mise run test:architecture` and `mise run validate`, verifying the architecture matrix remains unchanged and the complete repository validation passes.

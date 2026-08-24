## 1. Declaration Contract

- [x] 1.1 Add strict `sandbox.hooks.initialize` command objects to the
      version 1 JSON Schema and declaration DTOs, with decoding tests for
      omission, order preservation, empty vectors or arguments, and unknown
      fields.
- [x] 1.2 Carry defensive copies of validated command vectors through the domain
      configuration result and lifecycle plan, with tests that validation never
      executes commands and teardown projections remain command-free.

## 2. Sandbox Command Execution

- [x] 2.1 Add a narrow command-execution request and capability to the Sandbox
      port for sandbox name, workspace, argument vector, and attached output
      streams.
- [x] 2.2 Implement Docker Sandbox exec invocation in the outbound adapter, with
      tests for exact sandbox and working-directory targeting, literal argument
      preservation, output streams, and unsuccessful process results.

## 3. Initialization Lifecycle

- [x] 3.1 Add initialization to `up` after creation and network-policy
      application but before agent attachment for missing and recreated
      sandboxes, while existing running and stopped sandboxes attach without
      hook execution or reconciliation.
- [x] 3.2 Add lifecycle tests for ordered execution, explicit-shell and literal
      arguments, output forwarding, non-interactive input, cancellation,
      fail-fast behavior, actionable command context, skipped attachment,
      sandbox-and-resource rollback, combined cleanup failures, and retry after
      partial host-workspace effects.
- [x] 3.3 Add executable-level CLI coverage proving malformed hooks fail before
      Docker state access, creation and recreation execute hooks before
      attachment, existing-sandbox `up` skips them, and `validate`, `down`, and
      `destroy` never execute them.

## 4. Public Documentation and Validation

- [x] 4.1 Document `sandbox.hooks.initialize`, creation-only non-interactive
      execution, recreation for changed hooks, explicit shell usage, ordering,
      output, cancellation, rollback, host-workspace caveats, and retry behavior
      in the README and examples documentation.
- [x] 4.2 Add a generic Node project example with bounded `language-node-npm`
      mount readiness plus repository bootstrap commands, keeping the example
      valid against the published schema.
- [x] 4.3 Format the repository and run focused declaration, configuration,
      lifecycle, adapter, and executable tests followed by the architecture test,
      strict OpenSpec validation, and full repository validation.

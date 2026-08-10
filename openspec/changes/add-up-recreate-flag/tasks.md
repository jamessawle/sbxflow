## 1. Lifecycle Orchestration

- [ ] 1.1 Add a lifecycle-owned up options value carrying the recreation choice,
      extend `UpRunner` with the existing Sandbox remover capability, and update its
      callers and test doubles to the new request contract.
- [ ] 1.2 Implement the validated lookup branches so an absent sandbox follows
      create-and-enter, an existing sandbox follows enter-existing by default, and
      an existing sandbox with recreation is force-removed before a create-and-enter
      request built from the current plan.
- [ ] 1.3 Add lifecycle tests for ordinary behavior, absent-sandbox recreation,
      exact-name forced removal with attached streams, operation ordering, lookup
      failure, removal failure without execution, and replacement execution
      failure.

## 2. CLI and Executable Wiring

- [ ] 2.1 Register the long-form-only `up --recreate` flag, describe its
      destructive behavior in command help, and pass the selected option into the
      lifecycle workflow while preserving existing validation and attached-process
      error rendering.
- [ ] 2.2 Update CLI tests for help discovery, default and recreate option
      injection, rejection of positional arguments, `-r`, and unrelated flags, and
      no workflow invocation during help or parse failures.
- [ ] 2.3 Extend the entrypoint's narrowed up dependency with the Sandbox remover
      capability and update root or executable-level tests for the changed runner
      contract and advertised interface.

## 3. Documentation

- [ ] 3.1 Update README `up` usage and lifecycle guidance to explain normal
      behavior, exact-name forced recreation, permanent sandbox-state loss,
      validation and failure ordering, the lack of a shorthand, and recovery with a
      later ordinary `up` if replacement creation fails.

## 4. Validation

- [ ] 4.1 Run `mise run fmt` and review the resulting Markdown and Go formatting.
- [ ] 4.2 Run `mise run validate` to verify strict OpenSpec checks, architecture
      boundaries, vet, unit and executable tests, and builds.

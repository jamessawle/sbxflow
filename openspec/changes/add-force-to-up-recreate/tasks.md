## 1. Lifecycle Contract

- [ ] 1.1 Add the explicit force choice to up lifecycle options and reject force without recreation before configuration validation or sandbox operations.
- [ ] 1.2 Bypass the confirmer for forced running recreation while preserving confirmation for unforced running recreation and the existing stopped and absent paths.
- [ ] 1.3 Extend lifecycle tests to cover invalid options, forced running recreation, stopped and absent forced recreation, and preservation of interactive cancellation and failure behavior.

## 2. CLI Interface

- [ ] 2.1 Add the long-only `up --force` flag, require `--recreate` before invoking the runner, and pass the explicit force choice into lifecycle options.
- [ ] 2.2 Update `up` help text to explain the flag relationship, confirmation bypass, persisted-state loss, and risk to attached sessions.
- [ ] 2.3 Extend CLI adapter tests for help, option propagation, unsupported `-f`, and force-without-recreate rejection without runner invocation.

## 3. Executable Behavior and Documentation

- [ ] 3.1 Add executable-level coverage proving forced running recreation skips confirmation and retains exact-name removal, cleanup, creation, and entry behavior.
- [ ] 3.2 Add executable-level coverage proving `up --force` fails before declaration validation or Docker Sandbox interaction and that unforced recreation remains interactive.
- [ ] 3.3 Update the README command examples and safety guidance for `up --recreate --force`.

## 4. Validation

- [ ] 4.1 Run `mise run fmt` and review the formatted implementation, documentation, and OpenSpec artifacts.
- [ ] 4.2 Run `mise run validate` and resolve any CLI, lifecycle, architecture, documentation, or OpenSpec failures.

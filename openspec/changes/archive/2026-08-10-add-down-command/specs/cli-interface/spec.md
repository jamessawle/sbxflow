## ADDED Requirements

### Requirement: Down command is discoverable

The CLI SHALL register `down` as an available command and provide contextual
help that describes its repository-aware, state-preserving stop behavior.

#### Scenario: Root help lists down

- **WHEN** a user displays root help
- **THEN** the available commands include `down`

#### Scenario: Down help is requested

- **WHEN** a user invokes `sbxflow down --help` or `sbxflow help down`
- **THEN** the CLI displays `down` usage and its stop-without-removal purpose
- **AND** exits successfully without discovering a declaration or invoking
  Docker Sandboxes

### Requirement: Down accepts no arguments

The `down` command SHALL accept no positional arguments or command-specific
flags other than help in its initial interface.

#### Scenario: Positional argument is supplied

- **WHEN** a user invokes `sbxflow down` with a positional argument
- **THEN** the CLI reports the invalid invocation on standard error
- **AND** exits with a non-zero status without invoking Docker Sandboxes

#### Scenario: Unsupported flag is supplied

- **WHEN** a user invokes `sbxflow down` with an unsupported flag
- **THEN** the CLI reports the unknown flag on standard error
- **AND** exits with a non-zero status without invoking Docker Sandboxes

## ADDED Requirements

### Requirement: Up command is discoverable

The CLI SHALL register `up` as an available command and provide contextual help
that describes its repository-aware interactive sandbox purpose.

#### Scenario: Root help lists up

- **WHEN** a user displays root help
- **THEN** the available commands include `up`

#### Scenario: Up help is requested

- **WHEN** a user invokes `sbxflow up --help` or `sbxflow help up`
- **THEN** the CLI displays `up` usage and its create-or-enter purpose
- **AND** exits successfully without discovering a declaration or invoking
  Docker Sandboxes

### Requirement: Up accepts no arguments

The `up` command SHALL accept no positional arguments or command-specific flags
other than help in its initial interface.

#### Scenario: Positional argument is supplied

- **WHEN** a user invokes `sbxflow up` with a positional argument
- **THEN** the CLI reports the invalid invocation on standard error
- **AND** exits with a non-zero status without invoking Docker Sandboxes

#### Scenario: Unsupported flag is supplied

- **WHEN** a user invokes `sbxflow up` with an unsupported flag
- **THEN** the CLI reports the unknown flag on standard error
- **AND** exits with a non-zero status without invoking Docker Sandboxes

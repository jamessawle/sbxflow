## ADDED Requirements

### Requirement: Initialization hooks are strictly declared

The `validate` command SHALL accept an optional `sandbox.hooks.initialize` array
whose entries each contain a `command`
argument vector of one or more non-empty strings. It SHALL preserve both command
order and argument order exactly, reject unknown lifecycle or command-entry
fields, and validate the declaration without executing any command.

#### Scenario: Ordered commands are declared

- **WHEN** a version 1 declaration contains multiple valid
  `sandbox.hooks.initialize` entries
- **THEN** structural and semantic validation succeed
- **AND** each command and its arguments remain in declaration order
- **AND** no declared command is executed

#### Scenario: Lifecycle configuration is omitted

- **WHEN** a version 1 declaration omits `sandbox.hooks` or declares no
  initialization commands
- **THEN** validation succeeds subject to the existing declaration requirements

#### Scenario: Command vector is empty

- **WHEN** an `initialize` entry has an absent or empty `command` vector
- **THEN** structural validation identifies the malformed entry
- **AND** exits with a non-zero status without inspecting or mutating Docker state

#### Scenario: Command argument is empty

- **WHEN** an `initialize` command contains an empty string argument
- **THEN** structural validation identifies the malformed argument
- **AND** exits with a non-zero status without inspecting or mutating Docker state

#### Scenario: Hooks declaration contains an unknown field

- **WHEN** `sandbox.hooks` or one of its `initialize` entries contains an
  unknown field
- **THEN** structural validation reports the unknown field
- **AND** exits with a non-zero status without inspecting or mutating Docker state

#### Scenario: Shell syntax is declared as plain text

- **WHEN** a command argument contains shell metacharacters without an explicit
  shell executable and evaluation option in the same command vector
- **THEN** validation retains that argument literally rather than interpreting it
  through a host shell

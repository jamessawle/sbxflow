# cli-interface Specification

## Purpose

Defines the stable root command-line behavior through which users discover and
identify sbxflow and its available capabilities.

## Requirements

### Requirement: Root help is available

The CLI SHALL display root help and exit successfully when invoked without
arguments, through the `help` command, or with the `-h` or `--help` flag.

#### Scenario: Invocation without arguments

- **WHEN** a user invokes `sbxflow` without arguments
- **THEN** the CLI prints root help to standard output and exits successfully

#### Scenario: Help command

- **WHEN** a user invokes `sbxflow help`
- **THEN** the CLI prints root help to standard output and exits successfully

#### Scenario: Help flag

- **WHEN** a user invokes `sbxflow -h` or `sbxflow --help`
- **THEN** the CLI prints root help to standard output and exits successfully

### Requirement: Help describes the available interface

Root help SHALL identify sbxflow, summarize its purpose, show its usage, and
describe only commands and flags that are available in the executable.

#### Scenario: Inspecting root help

- **WHEN** a user displays root help
- **THEN** the output identifies sbxflow and describes its registered commands
  and flags
- **AND** the output does not advertise commands or flags that are unavailable

#### Scenario: Optional framework commands remain unavailable

- **WHEN** a user displays root help
- **THEN** completion and man-page commands are not advertised

### Requirement: Version is available through flags

The CLI SHALL print the same non-empty build identity and exit successfully for
both `-v` and `--version`, and SHALL NOT expose a `version` subcommand.

#### Scenario: Short version flag

- **WHEN** a user invokes `sbxflow -v`
- **THEN** the CLI prints the sbxflow build identity to standard output and
  exits successfully

#### Scenario: Long version flag

- **WHEN** a user invokes `sbxflow --version`
- **THEN** the CLI prints the same build identity as `sbxflow -v` and exits
  successfully

#### Scenario: Development build identity

- **WHEN** sbxflow is built without release version metadata
- **THEN** its version output explicitly identifies the build as a development
  build rather than reporting an empty version

#### Scenario: Version subcommand is not supported

- **WHEN** a user invokes `sbxflow version`
- **THEN** the CLI treats `version` as an unknown command and exits with a
  non-zero status

### Requirement: Command errors are actionable

The CLI SHALL write command parsing and dispatch errors to standard error,
return a non-zero status, and keep the error understandable without depending
on terminal styling.

#### Scenario: Unknown command

- **WHEN** a user invokes an unknown command
- **THEN** the CLI writes an error identifying the unknown command to standard
  error and exits with a non-zero status

#### Scenario: Output without terminal styling

- **WHEN** styled output is unavailable or disabled
- **THEN** help and error text retain all information needed to understand the
  invocation and result

### Requirement: Doctor command is discoverable

The CLI SHALL register `doctor` as an available command and provide contextual
help that describes its system-level diagnostic purpose.

#### Scenario: Root help lists doctor

- **WHEN** a user displays root help
- **THEN** the available commands include `doctor`

#### Scenario: Doctor help is requested

- **WHEN** a user invokes `sbxflow doctor --help` or `sbxflow help doctor`
- **THEN** the CLI displays doctor usage and its diagnostic purpose
- **AND** exits successfully

### Requirement: Validate command is discoverable

The CLI SHALL register `validate` as an available command and provide contextual
help that describes its repository configuration validation purpose.

#### Scenario: Root help lists validate

- **WHEN** a user displays root help
- **THEN** the available commands include `validate`

#### Scenario: Validate help is requested

- **WHEN** a user invokes `sbxflow validate --help` or
  `sbxflow help validate`
- **THEN** the CLI displays validate usage and its configuration validation
  purpose
- **AND** exits successfully

### Requirement: Up reports successful configuration validation

After the complete repository validation pipeline succeeds, the `up` command
SHALL write `Configuration valid: <declaration path>` followed by a newline to
standard error before inspecting or changing Docker Sandbox state.

#### Scenario: Up validation succeeds

- **WHEN** `sbxflow up` completes validation of the nearest declaration
  successfully
- **THEN** the command writes `Configuration valid: <declaration path>` to
  standard error
- **AND** writes the status before attempting to list, create, restart, or enter
  a sandbox

#### Scenario: Up validation fails

- **WHEN** `sbxflow up` cannot complete validation successfully
- **THEN** the command does not write the successful configuration-validation
  status
- **AND** reports the validation failure according to the existing error
  contract

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

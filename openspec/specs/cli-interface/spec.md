# cli-interface Specification

## Purpose

Defines the stable root command-line behavior through which users discover and
identify sbxflow and its available capabilities.

## Requirements

### Requirement: Public versions follow Semantic Versioning

Published sbxflow versions SHALL follow Semantic Versioning. Before v1.0.0, a
minor release MAY make incompatible changes to the CLI or configuration format,
while patch releases SHALL NOT intentionally introduce incompatible changes.

#### Scenario: Pre-1.0 minor release

- **WHEN** an incompatible CLI or configuration change is released before
  v1.0.0
- **THEN** the release increments the minor version
- **AND** its release notes identify the incompatible change

#### Scenario: Pre-1.0 patch release

- **WHEN** a patch release is published before v1.0.0
- **THEN** it does not intentionally introduce an incompatible CLI or
  configuration change

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

### Requirement: Up recreation option is discoverable

The `up` command help SHALL describe `--recreate` as a destructive option that
force-removes an existing declared sandbox before creating and entering its
replacement, and SHALL disclose that recreating a running sandbox requires
confirmation because it can terminate other attached sessions.

#### Scenario: Up recreation help is requested

- **WHEN** a user invokes `sbxflow up --help` or `sbxflow help up`
- **THEN** the CLI displays the `--recreate` option, its destructive
  force-recreation behavior, and the running-sandbox confirmation behavior
- **AND** exits successfully without discovering a declaration or invoking
  Docker Sandboxes

### Requirement: Running sandbox recreation uses the command streams

When recreation targets a running sandbox, the CLI SHALL write a warning and
confirmation prompt to the command's error stream and read the response from
the command's input stream. The warning SHALL identify the sandbox, state that
recreation permanently removes its persisted state, and state that other
attached terminal sessions can be terminated. Only an explicit affirmative
response SHALL authorize recreation; an empty, negative, malformed, or
unavailable response SHALL cancel it.

#### Scenario: User confirms running sandbox recreation

- **WHEN** `sbxflow up --recreate` detects a running sandbox and the user enters
  an affirmative response through standard input
- **THEN** the warning and prompt are observable on standard error
- **AND** recreation proceeds

#### Scenario: User does not affirm running sandbox recreation

- **WHEN** `sbxflow up --recreate` detects a running sandbox and the response is
  empty, negative, or malformed
- **THEN** the CLI reports cancellation without force-removing the sandbox
- **AND** exits with a non-zero status

#### Scenario: Confirmation input is unavailable

- **WHEN** `sbxflow up --recreate` detects a running sandbox but cannot read a
  response from the command input stream
- **THEN** the CLI reports that confirmation could not be obtained
- **AND** exits with a non-zero status without force-removing the sandbox

### Requirement: Up accepts no arguments

The `up` command SHALL accept no positional arguments. It SHALL accept
`--recreate` as its only command-specific flag other than help, SHALL NOT
provide a shorthand for that flag, and SHALL reject all other command-specific
flags.

#### Scenario: Up is invoked without options

- **WHEN** a user invokes `sbxflow up`
- **THEN** the CLI begins the normal create-or-enter flow for the repository's
  declared sandbox

#### Scenario: Recreate option is supplied

- **WHEN** a user invokes `sbxflow up --recreate`
- **THEN** the CLI begins the recreate-if-present flow for the repository's
  declared sandbox

#### Scenario: Positional argument is supplied

- **WHEN** a user invokes `sbxflow up` with a positional argument
- **THEN** the CLI reports the invalid invocation on standard error
- **AND** exits with a non-zero status without invoking Docker Sandboxes

#### Scenario: Unsupported flag is supplied

- **WHEN** a user invokes `sbxflow up` with an unsupported flag, including
  `-r`
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

### Requirement: Destroy command is discoverable

The CLI SHALL register `destroy` as an available command and provide contextual
help that describes its repository-aware, irreversible removal behavior and
its distinction from `down`.

#### Scenario: Root help lists destroy

- **WHEN** a user displays root help
- **THEN** the available commands include `destroy`

#### Scenario: Destroy help is requested

- **WHEN** a user invokes `sbxflow destroy --help` or `sbxflow help destroy`
- **THEN** the CLI displays `destroy` usage, its permanent state-removal effect,
  and the behavior of `--force`
- **AND** exits successfully without discovering a declaration or invoking
  Docker Sandboxes

### Requirement: Destroy accepts only the force option

The `destroy` command SHALL accept no positional arguments. It SHALL accept
`--force` and `-f` as equivalent command-specific flags and SHALL reject other
command-specific flags, including flags that select a sandbox name or all
sandboxes.

#### Scenario: Destroy is invoked without options

- **WHEN** a user invokes `sbxflow destroy`
- **THEN** the CLI begins destruction for the repository's declared sandbox
  with Docker's confirmation enabled

#### Scenario: Long force option is supplied

- **WHEN** a user invokes `sbxflow destroy --force`
- **THEN** the CLI begins forced destruction for the repository's declared
  sandbox

#### Scenario: Short force option is supplied

- **WHEN** a user invokes `sbxflow destroy -f`
- **THEN** the CLI begins the same forced destruction as `--force`

#### Scenario: Positional argument is supplied

- **WHEN** a user invokes `sbxflow destroy other-sandbox`
- **THEN** the CLI reports the invalid invocation on standard error
- **AND** exits with a non-zero status without invoking Docker Sandboxes

#### Scenario: Unsupported selection flag is supplied

- **WHEN** a user invokes `sbxflow destroy --all` or `sbxflow destroy --name other`
- **THEN** the CLI reports the unknown flag on standard error
- **AND** exits with a non-zero status without invoking Docker Sandboxes

### Requirement: Destroy confirmation uses the command streams

When destruction requires confirmation, the CLI SHALL connect Docker's removal
process to the command's standard input, standard output, and standard error so
the user can observe and answer the confirmation prompt.

#### Scenario: Docker asks for confirmation

- **WHEN** `sbxflow destroy` invokes Docker's confirmed removal
- **THEN** the user can read the prompt and respond through the same terminal
  streams used to invoke sbxflow

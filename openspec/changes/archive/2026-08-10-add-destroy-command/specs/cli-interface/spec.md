## ADDED Requirements

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

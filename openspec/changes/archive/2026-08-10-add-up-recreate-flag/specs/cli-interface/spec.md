## ADDED Requirements

### Requirement: Up recreation option is discoverable

The `up` command help SHALL describe `--recreate` as a destructive option that
force-removes an existing declared sandbox before creating and entering its
replacement.

#### Scenario: Up recreation help is requested

- **WHEN** a user invokes `sbxflow up --help` or `sbxflow help up`
- **THEN** the CLI displays the `--recreate` option and its force-recreation
  behavior
- **AND** exits successfully without discovering a declaration or invoking
  Docker Sandboxes

## MODIFIED Requirements

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

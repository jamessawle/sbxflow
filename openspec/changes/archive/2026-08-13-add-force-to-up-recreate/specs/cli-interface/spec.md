## MODIFIED Requirements

### Requirement: Up recreation option is discoverable

The `up` command help SHALL describe `--recreate` as a destructive option that force-removes an existing declared sandbox before creating and entering its replacement. It SHALL describe long-only `--force` as valid only with `--recreate` and disclose that it bypasses running-sandbox confirmation even though recreation permanently removes persisted state and can terminate other attached sessions.

#### Scenario: Up recreation help is requested

- **WHEN** a user invokes `sbxflow up --help` or `sbxflow help up`
- **THEN** the CLI displays the `--recreate` and `--force` options
- **AND** explains their destructive recreation, flag-combination, and running-sandbox confirmation behavior
- **AND** exits successfully without discovering a declaration or invoking Docker Sandboxes

### Requirement: Running sandbox recreation uses the command streams

When unforced recreation targets a running sandbox, the CLI SHALL write a warning and confirmation prompt to the command's error stream and read the response from the command's input stream. The warning SHALL identify the sandbox, state that recreation permanently removes its persisted state, and state that other attached terminal sessions can be terminated. A newline SHALL terminate a response containing zero or more bytes, while EOF SHALL terminate a complete response only after one or more bytes. Only a complete, explicit affirmative response SHALL authorize recreation; an empty, negative, malformed, or unavailable response SHALL cancel it. Immediate EOF and non-EOF read failures SHALL be treated as unavailable input rather than consent. When forced recreation targets a running sandbox, the CLI SHALL bypass this confirmation interaction.

#### Scenario: User confirms running sandbox recreation

- **WHEN** `sbxflow up --recreate` detects a running sandbox and the user enters an affirmative response through standard input terminated by a newline or EOF
- **THEN** the warning and prompt are observable on standard error
- **AND** recreation proceeds

#### Scenario: User does not affirm running sandbox recreation

- **WHEN** `sbxflow up --recreate` detects a running sandbox and receives an empty newline-terminated response or a negative or malformed response terminated by a newline or EOF
- **THEN** the CLI reports cancellation without force-removing the sandbox
- **AND** exits with a non-zero status

#### Scenario: Confirmation input is unavailable

- **WHEN** `sbxflow up --recreate` detects a running sandbox but receives EOF before any response bytes or encounters another read failure
- **THEN** the CLI reports that confirmation could not be obtained
- **AND** exits with a non-zero status without force-removing the sandbox

#### Scenario: Forced recreation bypasses confirmation streams

- **WHEN** `sbxflow up --recreate --force` detects a running sandbox
- **THEN** recreation proceeds without reading a confirmation response or writing a confirmation prompt

### Requirement: Up accepts no arguments

The `up` command SHALL accept no positional arguments. It SHALL accept `--recreate` and long-only `--force` as its command-specific flags other than help, SHALL NOT provide a shorthand for either flag, and SHALL reject all other command-specific flags. `--force` SHALL require `--recreate`; the CLI SHALL reject `up --force` before configuration validation or lifecycle work begins rather than ignoring it or making an ordinary `up` destructive.

#### Scenario: Up is invoked without options

- **WHEN** a user invokes `sbxflow up`
- **THEN** the CLI begins the normal create-or-enter flow for the repository's declared sandbox

#### Scenario: Recreate option is supplied

- **WHEN** a user invokes `sbxflow up --recreate`
- **THEN** the CLI begins the recreate-if-present flow for the repository's declared sandbox

#### Scenario: Force is supplied with recreate

- **WHEN** a user invokes `sbxflow up --recreate --force`
- **THEN** the CLI begins the forced recreate-if-present flow for the repository's declared sandbox

#### Scenario: Force is supplied without recreate

- **WHEN** a user invokes `sbxflow up --force`
- **THEN** the CLI reports that `--force` requires `--recreate` on standard error
- **AND** exits with a non-zero status without validating configuration or invoking lifecycle operations

#### Scenario: Positional argument is supplied

- **WHEN** a user invokes `sbxflow up` with a positional argument
- **THEN** the CLI reports the invalid invocation on standard error
- **AND** exits with a non-zero status without invoking Docker Sandboxes

#### Scenario: Unsupported flag is supplied

- **WHEN** a user invokes `sbxflow up` with an unsupported flag, including `-r` or `-f`
- **THEN** the CLI reports the invalid invocation on standard error
- **AND** exits with a non-zero status without invoking Docker Sandboxes

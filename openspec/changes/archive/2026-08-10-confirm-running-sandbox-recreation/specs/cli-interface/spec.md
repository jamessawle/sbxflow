## MODIFIED Requirements

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

## ADDED Requirements

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

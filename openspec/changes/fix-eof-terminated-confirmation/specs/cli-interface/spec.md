## MODIFIED Requirements

### Requirement: Running sandbox recreation uses the command streams

When recreation targets a running sandbox, the CLI SHALL write a warning and
confirmation prompt to the command's error stream and read the response from
the command's input stream. The warning SHALL identify the sandbox, state that
recreation permanently removes its persisted state, and state that other
attached terminal sessions can be terminated. A newline or EOF after one or
more response bytes SHALL terminate a complete response. Only a complete,
explicit affirmative response SHALL authorize recreation; an empty, negative,
malformed, or unavailable response SHALL cancel it. EOF without any response
bytes and non-EOF read failures SHALL be treated as unavailable input rather
than consent.

#### Scenario: User confirms running sandbox recreation

- **WHEN** `sbxflow up --recreate` detects a running sandbox and the user enters
  an affirmative response through standard input terminated by a newline or EOF
- **THEN** the warning and prompt are observable on standard error
- **AND** recreation proceeds

#### Scenario: User does not affirm running sandbox recreation

- **WHEN** `sbxflow up --recreate` detects a running sandbox and a newline- or
  EOF-terminated response is empty, negative, or malformed
- **THEN** the CLI reports cancellation without force-removing the sandbox
- **AND** exits with a non-zero status

#### Scenario: Confirmation input is unavailable

- **WHEN** `sbxflow up --recreate` detects a running sandbox but receives EOF
  before any response bytes or encounters another read failure
- **THEN** the CLI reports that confirmation could not be obtained
- **AND** exits with a non-zero status without force-removing the sandbox

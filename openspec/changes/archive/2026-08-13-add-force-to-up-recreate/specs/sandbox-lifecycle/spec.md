## MODIFIED Requirements

### Requirement: Up can replace an existing sandbox from the current declaration

When recreation is requested, `up` SHALL inspect the lifecycle state of the exact declared sandbox before removal. An absent sandbox SHALL follow the ordinary creation path, and a stopped sandbox SHALL be force-removed without additional confirmation. A running sandbox SHALL be force-removed only after the user explicitly confirms a warning that recreation can terminate other attached sessions or explicitly requests forced recreation. The removal SHALL remain limited to the exact declared name and SHALL permanently discard the removed sandbox's persisted state without deleting files from the repository's host workspace.

If lifecycle state cannot be determined, or unforced confirmation cannot be obtained or is declined, `up` SHALL stop without removing, creating, entering, or otherwise changing the sandbox. A force request without recreation SHALL be rejected before configuration validation, sandbox inspection, or any other lifecycle work.

#### Scenario: Existing sandbox is recreated

- **WHEN** the exact declared sandbox exists, a user invokes `sbxflow up --recreate`, and any required running-sandbox confirmation is granted
- **THEN** `up` invokes `sbx rm --force` with that sandbox name
- **AND** Docker does not ask for an additional removal confirmation
- **AND** after successful removal, `up` creates and enters a sandbox with the declared name, agent, repository workspace, selected kits, and derived trust

#### Scenario: Running sandbox recreation is confirmed

- **WHEN** the exact declared sandbox is running, recreation is requested without force, and the user confirms the warning
- **THEN** `up` invokes `sbx rm --force` with that sandbox name
- **AND** after successful removal, `up` creates and enters a sandbox with the declared name, agent, repository workspace, selected kits, and derived trust

#### Scenario: Running sandbox recreation is forced

- **WHEN** the exact declared sandbox is running and forced recreation is requested
- **THEN** `up` does not request interactive confirmation
- **AND** invokes `sbx rm --force` with the exact declared sandbox name
- **AND** after successful removal, creates and enters its replacement from the current declaration

#### Scenario: Running sandbox recreation is declined

- **WHEN** the exact declared sandbox is running, unforced recreation is requested, and the user declines or provides no affirmative response
- **THEN** `up` cancels recreation without removing, creating, or entering a sandbox
- **AND** exits with a non-zero status

#### Scenario: Running sandbox confirmation is unavailable

- **WHEN** the exact declared sandbox is running, unforced recreation is requested, and confirmation input cannot be obtained
- **THEN** `up` reports that recreation was not confirmed
- **AND** does not remove, create, or enter a sandbox
- **AND** exits with a non-zero status

#### Scenario: Existing stopped sandbox is recreated

- **WHEN** the exact declared sandbox exists in a stopped state and recreation is requested with or without force
- **THEN** `up` force-removes it without an additional confirmation prompt
- **AND** creates and enters its replacement from the current declaration

#### Scenario: Missing sandbox follows creation path

- **WHEN** the exact declared sandbox is absent and recreation is requested with or without force
- **THEN** `up` creates and enters the sandbox through the ordinary missing-sandbox path

#### Scenario: Force without recreation is rejected

- **WHEN** lifecycle is requested with force enabled and recreation disabled
- **THEN** `up` reports the invalid option combination before configuration validation or sandbox inspection
- **AND** does not remove, create, or enter a sandbox

#### Scenario: Recreation state inspection fails

- **WHEN** Docker Sandboxes cannot report a recognized lifecycle state for the exact declared sandbox
- **THEN** `up` reports the inspection failure
- **AND** does not remove, create, or enter a sandbox
- **AND** exits with a non-zero status

#### Scenario: Recreation removal fails

- **WHEN** Docker Sandboxes cannot force-remove an approved, forced, or stopped declared sandbox
- **THEN** its diagnostic output remains visible to the user
- **AND** `up` does not create or enter a sandbox
- **AND** exits with a non-zero status

#### Scenario: Recreated sandbox starts with new state

- **WHEN** forced removal succeeds and `up` creates the replacement sandbox
- **THEN** state from the removed sandbox is unavailable in the replacement
- **AND** files in the repository's host workspace remain intact

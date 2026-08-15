## ADDED Requirements

### Requirement: Machine-readable lifecycle inspection isolates Docker update notices

When `up` inspects the exact declared sandbox through Docker Sandboxes'
machine-readable listing, it SHALL accept one complete, valid listing followed
only by Docker Sandboxes' known update-available notice. The notice SHALL NOT
change the decoded sandbox state or prevent the selected lifecycle path from
continuing. Inspection MUST continue to reject leading contamination,
malformed or incomplete JSON, invalid UTF-8 within the listing, multiple JSON
documents, and trailing output that is not the known update notice.

#### Scenario: Update notice follows a valid listing

- **WHEN** Docker Sandboxes successfully returns one valid machine-readable
  sandbox listing followed by its update-available notice
- **THEN** `up` derives the declared sandbox state from the listing
- **AND** continues through the lifecycle path selected by that state

#### Scenario: Listing is malformed before the update notice

- **WHEN** Docker Sandboxes returns malformed or incomplete machine-readable
  sandbox data with or without an update notice
- **THEN** `up` reports the state-inspection failure
- **AND** does not remove, create, or enter a sandbox

#### Scenario: Output is contaminated before the listing

- **WHEN** non-whitespace output precedes the machine-readable sandbox listing
- **THEN** `up` reports the state-inspection failure
- **AND** does not remove, create, or enter a sandbox

#### Scenario: Listing contains invalid UTF-8

- **WHEN** the machine-readable sandbox listing contains invalid UTF-8 bytes
- **THEN** `up` reports the state-inspection failure rather than deriving state
  from silently substituted replacement characters
- **AND** does not remove, create, or enter a sandbox

#### Scenario: Multiple machine-readable documents are returned

- **WHEN** Docker Sandboxes returns more than one JSON document
- **THEN** `up` reports the ambiguous state-inspection output
- **AND** does not remove, create, or enter a sandbox

#### Scenario: Unknown output follows the listing

- **WHEN** a valid machine-readable sandbox listing is followed by output other
  than whitespace or Docker Sandboxes' known update-available notice
- **THEN** `up` reports the unexpected state-inspection output
- **AND** does not remove, create, or enter a sandbox

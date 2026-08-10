## ADDED Requirements

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

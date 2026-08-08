## ADDED Requirements

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

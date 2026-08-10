## ADDED Requirements

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

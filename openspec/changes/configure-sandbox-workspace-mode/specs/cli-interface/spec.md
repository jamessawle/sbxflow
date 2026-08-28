## ADDED Requirements

### Requirement: Configuration warnings are visible without preventing valid commands

The `validate` and `up` commands SHALL render every advisory warning produced by repository validation on standard error. A warning-only validation result SHALL remain successful: `validate` SHALL exit with a zero status, and `up` SHALL continue its selected lifecycle operation after displaying the successful-validation status and warnings. Validation errors SHALL continue to prevent `up` lifecycle work regardless of whether warnings are also present.

#### Scenario: Validate reports an omitted workspace mode

- **WHEN** `sbxflow validate` validates an otherwise valid declaration whose workspace mode is omitted
- **THEN** the command writes the omitted-mode warning to standard error
- **AND** exits successfully

#### Scenario: Up reports an omitted workspace mode

- **WHEN** `sbxflow up` validates an otherwise valid declaration whose workspace mode is omitted
- **THEN** the command writes the successful-validation status and omitted-mode warning to standard error before lifecycle work
- **AND** continues the create, recreate, or enter path

#### Scenario: Explicit mode has no workspace warning

- **WHEN** `validate` or `up` processes a valid declaration with an explicit `direct` or `clone` mode
- **THEN** the command does not display an omitted-mode warning

#### Scenario: Validation has warnings and errors

- **WHEN** `up` receives a validation result containing both warnings and errors
- **THEN** the command may display both kinds of diagnostics
- **AND** does not inspect, remove, create, restart, or enter a sandbox
- **AND** exits with a non-zero status

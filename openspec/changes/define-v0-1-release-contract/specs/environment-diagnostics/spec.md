## ADDED Requirements

### Requirement: Releases declare Docker Sandboxes compatibility

Each sbxflow release SHALL document the bounded `sbx` version range accepted by its compatibility check, and SHALL NOT infer operating-system or architecture support promises from the environment used to build or test sbxflow.

#### Scenario: User reviews release compatibility

- **WHEN** a user reviews the documentation for a released sbxflow version
- **THEN** the supported lower-inclusive and upper-exclusive `sbx` version bounds are stated
- **AND** those bounds agree with the range enforced by `doctor`

#### Scenario: Build environment changes

- **WHEN** maintainers change the environment used to build or test sbxflow without changing its `sbx` compatibility
- **THEN** the documented compatibility contract remains the bounded `sbx` version range

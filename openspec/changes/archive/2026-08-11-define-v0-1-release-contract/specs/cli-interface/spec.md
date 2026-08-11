## ADDED Requirements

### Requirement: Public versions follow Semantic Versioning

Published sbxflow versions SHALL follow Semantic Versioning. Before v1.0.0, a minor release MAY make incompatible changes to the CLI or configuration format, while patch releases SHALL NOT intentionally introduce incompatible changes.

#### Scenario: Pre-1.0 minor release

- **WHEN** an incompatible CLI or configuration change is released before v1.0.0
- **THEN** the release increments the minor version
- **AND** its release notes identify the incompatible change

#### Scenario: Pre-1.0 patch release

- **WHEN** a patch release is published before v1.0.0
- **THEN** it does not intentionally introduce an incompatible CLI or configuration change

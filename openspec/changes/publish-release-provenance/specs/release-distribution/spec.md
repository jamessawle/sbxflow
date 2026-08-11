## ADDED Requirements

### Requirement: Published release artifacts carry verifiable provenance

Every release archive and checksum manifest published by a successful release
MUST have GitHub-verifiable build provenance that identifies the sbxflow
repository and its trusted release workflow as the source.

#### Scenario: Consumer verifies a release archive

- **WHEN** a consumer downloads an archive from a release published after
  provenance support is enabled
- **THEN** GitHub verifies that the archive was produced by the sbxflow
  repository's release workflow

#### Scenario: Consumer verifies the checksum manifest

- **WHEN** a consumer downloads the checksum manifest from a release published
  after provenance support is enabled
- **THEN** GitHub verifies that the manifest was produced by the sbxflow
  repository's release workflow

#### Scenario: Provenance publication fails

- **WHEN** any required archive or checksum manifest cannot be attested
- **THEN** the release workflow fails and does not report the release as
  successfully completed

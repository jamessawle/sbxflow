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

#### Scenario: Provenance is discoverable from the release

- **WHEN** a release is published with build provenance
- **THEN** its GitHub release assets include an in-toto JSONL provenance bundle
  discoverable by the OpenSSF Scorecard Signed-Releases check

#### Scenario: Provenance publication fails

- **WHEN** any required archive or checksum manifest cannot be attested or the
  generated provenance bundle cannot be attached to the release
- **THEN** the release workflow fails and does not report the release as
  successfully completed

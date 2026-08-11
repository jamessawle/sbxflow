## ADDED Requirements

### Requirement: Release artifacts can be prepared without publication

The release process MUST provide a snapshot operation that builds the complete supported operating-system and architecture matrix from a clean checkout without creating or updating a GitHub release, package-manager distribution, or repository tag.

#### Scenario: Maintainer prepares a snapshot

- **WHEN** a maintainer runs the documented snapshot operation from a clean checkout
- **THEN** the process produces every configured release artifact locally without publishing or updating an external distribution

### Requirement: Snapshot artifacts have a verifiable identity

Every executable produced by the snapshot process MUST run on its target platform and report the expected snapshot version and source commit.

#### Scenario: Supported snapshot executable is exercised

- **WHEN** automation invokes a snapshot executable on its target platform with `--version`
- **THEN** the output identifies the snapshot version and the source commit used to build it

### Requirement: Release artifact packaging is stable and verifiable

The release process MUST document and consistently produce archive names and formats for each supported target, MUST include the target executable, license, and release metadata in each archive, and MUST produce a checksum manifest covering the distributable archives.

#### Scenario: Maintainer inspects snapshot output

- **WHEN** a snapshot completes successfully
- **THEN** every supported target has the documented archive name, format, and contents
- **AND** the checksum manifest verifies every distributable archive

# release-distribution Specification

## Purpose

Defines the evidence and artifacts required to publish an identifiable, reproducible sbxflow release from a version tag.

## Requirements

### Requirement: Releases originate from Semantic Version tags

The release process MUST publish only commits identified by a tag whose name is a valid Semantic Version prefixed with `v`.

#### Scenario: Valid release tag

- **WHEN** a `v<major>.<minor>.<patch>` tag triggers the release process
- **THEN** the process uses that tag's version and commit as the release identity

#### Scenario: Invalid release tag

- **WHEN** a tag does not contain a valid `v`-prefixed Semantic Version
- **THEN** the process does not publish a release

### Requirement: Release publication is gated by automated evidence

The release process MUST validate the repository, build the configured release artifacts with version and commit metadata, exercise each artifact's version and help interfaces, and generate checksums before publication.

#### Scenario: Every automated check succeeds

- **WHEN** repository validation, artifact builds, interface checks, and checksum generation all succeed
- **THEN** the process publishes the release and its artifacts

#### Scenario: An automated check fails

- **WHEN** any required validation, build, interface check, or checksum step fails
- **THEN** the process does not publish the release

### Requirement: Published artifacts identify their source

Every published executable MUST report the release version and source commit from which it was built.

#### Scenario: User inspects a release executable

- **WHEN** a user invokes a published executable with `--version`
- **THEN** its output identifies the release version and source commit

### Requirement: Release approval retains minimal human evidence

The v0.1.0 release checklist MUST require confirmation that the changes which will form the generated release notes accurately identify relevant compatibility or breaking changes and that a smoke test passed against an `sbx` version within the supported range.

#### Scenario: Maintainer approves the v0.1.0 tag

- **WHEN** a maintainer is ready to create the release tag
- **THEN** the maintainer confirms the planned release-note content and supported-`sbx` smoke-test evidence before creating the tag

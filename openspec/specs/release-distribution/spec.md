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

The release process MUST validate the repository and test the source before GoReleaser builds versioned executables, generates checksums, creates the GitHub release, and updates the Homebrew tap.

#### Scenario: Every automated check succeeds

- **WHEN** repository validation and source tests succeed for a valid release tag
- **THEN** GoReleaser publishes the release artifacts and updates `jamessawle/homebrew-tap`

#### Scenario: An automated check fails

- **WHEN** repository validation, source tests, build, packaging, GitHub publication, or Homebrew tap publication fails
- **THEN** the workflow reports failure and does not claim a successful release

### Requirement: Tagged releases are available through Homebrew

Every successfully published release MUST update the sbxflow cask in `jamessawle/homebrew-tap` using the versioned release artifacts.

#### Scenario: User installs a published release with Homebrew

- **WHEN** a user installs `jamessawle/tap/sbxflow` after a successful tagged release
- **THEN** Homebrew installs the corresponding published sbxflow version

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

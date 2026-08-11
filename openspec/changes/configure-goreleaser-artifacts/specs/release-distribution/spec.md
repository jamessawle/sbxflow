## MODIFIED Requirements

### Requirement: Release publication is gated by automated evidence

The release process MUST validate the repository and test the source before GoReleaser builds versioned executables, generates checksums, creates the GitHub release, and updates the Homebrew tap.

#### Scenario: Every automated check succeeds

- **WHEN** repository validation and source tests succeed for a valid release tag
- **THEN** GoReleaser publishes the release artifacts and updates `jamessawle/homebrew-tap`

#### Scenario: An automated check fails

- **WHEN** repository validation, source tests, build, packaging, GitHub publication, or Homebrew tap publication fails
- **THEN** the workflow reports failure and does not claim a successful release

## ADDED Requirements

### Requirement: Tagged releases are available through Homebrew

Every successfully published release MUST update the sbxflow cask in `jamessawle/homebrew-tap` using the versioned release artifacts.

#### Scenario: User installs a published release with Homebrew

- **WHEN** a user installs `jamessawle/tap/sbxflow` after a successful tagged release
- **THEN** Homebrew installs the corresponding published sbxflow version

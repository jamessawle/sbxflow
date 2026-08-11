## MODIFIED Requirements

### Requirement: Tagged releases are available through Homebrew

Every successfully published release MUST update the sbxflow cask in
`jamessawle/homebrew-tap` using the versioned release artifacts, and the first
commit that release pushes to the tap MUST pass the tap's checks without a
corrective commit.

#### Scenario: User installs a published release with Homebrew

- **WHEN** a user installs `jamessawle/tap/sbxflow` after a successful tagged
  release
- **THEN** Homebrew installs the corresponding published sbxflow version

#### Scenario: Tap update needs no follow-up correction

- **WHEN** a release publishes its Homebrew update to `jamessawle/homebrew-tap`
- **THEN** the first commit pushed to the tap branch passes the tap's checks and
  no further commit is required to correct it

#### Scenario: Homebrew evaluates the cask for every platform it supports

- **WHEN** Homebrew evaluates the published cask for each operating system and
  architecture it supports
- **THEN** the cask resolves a published archive and its checksum for each of
  them

## ADDED Requirements

### Requirement: Installed macOS executables run without manual intervention

The published Homebrew package MUST leave the installed sbxflow executable
runnable on macOS without the user clearing its quarantine attribute.

#### Scenario: User runs sbxflow after a clean Homebrew installation

- **WHEN** a user installs sbxflow through Homebrew on a machine that has never
  run it
- **THEN** the executable runs, rather than being refused by macOS as damaged or
  from an unidentified developer

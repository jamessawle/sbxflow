## MODIFIED Requirements

### Requirement: Tagged releases are available through Homebrew

Every successfully published release MUST update the sbxflow formula in
`jamessawle/homebrew-tap` using the versioned release artifacts, and the first
commit that release pushes to the tap MUST pass the tap's syntax checks without
a corrective commit.

#### Scenario: User installs a published release with Homebrew

- **WHEN** a user installs `jamessawle/tap/sbxflow` after a successful tagged
  release
- **THEN** Homebrew installs the corresponding published sbxflow version

#### Scenario: Tap update needs no follow-up correction

- **WHEN** a release publishes its Homebrew update to `jamessawle/homebrew-tap`
- **THEN** the first commit pushed to the tap branch passes the tap's syntax
  checks and no further commit is required to correct it

#### Scenario: Homebrew evaluates the formula for every platform it supports

- **WHEN** Homebrew evaluates the published formula for each operating system
  and architecture it supports
- **THEN** the formula resolves a published archive and its checksum for each of
  them

## ADDED Requirements

### Requirement: Homebrew packaging is validated before a tag is pushed

The repository MUST provide checks that validate the Homebrew package GoReleaser
generates from the release configuration, rather than a package produced by the
check itself.

#### Scenario: Maintainer validates a release candidate

- **WHEN** a maintainer runs the repository's release validation before tagging
- **THEN** the check inspects the Homebrew package written by GoReleaser and
  fails if its structure or the published archive set is not as expected

#### Scenario: Maintainer checks the package against Homebrew itself

- **WHEN** a maintainer runs the repository's tap syntax check before tagging
- **THEN** the generated package is checked by the same Homebrew commands the
  tap's own test-bot job runs, and failures are reported before a tag exists

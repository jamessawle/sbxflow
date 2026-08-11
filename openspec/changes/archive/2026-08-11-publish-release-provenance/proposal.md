## Why

Published sbxflow archives and checksums can be integrity-checked against one
another, but consumers cannot currently verify that they were produced by the
repository's trusted release workflow. GitHub-hosted build provenance will
close that supply-chain gap for future releases and remediate the OpenSSF
Scorecard Signed-Releases finding tracked by issue #45.

## What Changes

- Generate GitHub artifact attestations for each release archive and the
  checksum manifest after GoReleaser publishes them.
- Attach the generated in-toto bundle to each GitHub release so provenance is
  discoverable by the OpenSSF Scorecard Signed-Releases check.
- Grant the release job only the additional GitHub permissions required to
  issue build provenance.
- Document how consumers and maintainers verify release artifacts with the
  GitHub CLI, alongside the existing checksum verification.
- Treat missing or unverifiable provenance as an incomplete release and cover
  its recovery in the release runbook.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `release-distribution`: Require future published release archives and the
  checksum manifest to carry verifiable provenance from the sbxflow release
  workflow.

## Impact

The change affects the tag-triggered GitHub Actions release workflow, its
release assets, job-level permissions and pinned actions, the
release-distribution contract, and consumer and maintainer verification
documentation. It does not rewrite existing releases or change archive
contents, supported platforms, GoReleaser packaging, or Homebrew publication.

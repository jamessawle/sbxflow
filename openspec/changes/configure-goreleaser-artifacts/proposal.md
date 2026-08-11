## Why

The release workflow currently assembles platform archives independently in GitHub Actions, so maintainers cannot reproduce and inspect the complete release artifact set locally before publishing a version. A pinned GoReleaser snapshot workflow will make artifact identity, naming, contents, and checksums explicit and testable without creating a GitHub release.

## What Changes

- Define the supported Linux, macOS, and Windows release matrix in a GoReleaser configuration.
- Embed snapshot version and source-commit metadata in every executable and exercise each supported artifact's version interface.
- Produce stable platform archives containing the executable, license, and release metadata, plus a verifiable checksum manifest.
- Pin GoReleaser through Mise and add a non-publishing snapshot task suitable for local use and CI.
- Document the artifact names, archive contents, checksum verification, and snapshot workflow.
- Align release automation with the same GoReleaser configuration so snapshot evidence and published artifacts do not drift.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `release-distribution`: Require a reproducible, non-publishing snapshot process that defines and verifies the supported release artifacts before tag-driven publication.

## Impact

The change affects the release-distribution contract, Mise tool and task configuration, GoReleaser configuration, release documentation, and GitHub Actions release or validation workflows. It adds GoReleaser as a pinned development and automation tool but does not change the CLI commands, configuration schema, package architecture, or supported runtime behavior. SBOM generation and its additional tooling are deferred until there is a concrete supply-chain requirement.

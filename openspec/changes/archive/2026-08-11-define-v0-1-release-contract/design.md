## Context

The CLI already accepts linker-injected version and commit values, and `doctor` already enforces `sbx >= 0.35.0, < 0.38.0`. CI currently validates repository changes but there is no license, release automation, or central release policy. See `proposal.md` for motivation and the delta specs for the release contract.

## Goals / Non-Goals

**Goals:**

- Make a release an auditable extension of the existing validation pipeline.
- Keep user-facing policy short and place maintainer procedure with contributor documentation.
- Ensure every uploaded executable proves its tag and commit identity before publication.

**Non-Goals:**

- Claim support for particular operating systems or architectures.
- Promise compatibility beyond Semantic Versioning.
- Add package signing, provenance attestations, an installer, or a new release dependency for v0.1.0.
- Turn absent prospective features into a limitations catalogue.

## Decisions

### Use the MIT License

Add the standard MIT text with the repository owner's copyright. MIT provides the intended permissive terms with less policy and notice machinery than Apache-2.0. Apache-2.0 was considered for its explicit patent grant, but there is no current requirement for that additional contract.

### Keep policy close to its audience

The README will state only the user-relevant release facts: MIT licensing, Semantic Versioning, the supported `sbx` range, and the existing-sandbox reconciliation behavior beside `up`. The release checklist will live in `CONTRIBUTING.md`, where maintainers already find validation and pull-request procedure. A separate limitations or support-matrix document would add navigation without adding a meaningful contract.

### Extend GitHub Actions without a release framework

A release workflow triggered by `v*` tags will first reject tags that are not canonical Semantic Versions and run `mise run validate`. It will build a small configured matrix of ordinary Go release artifacts, inject the tag version and `github.sha` through the existing build-info variables, and verify `--version` and `--help` before checksumming and publishing them.

The configured artifact matrix is distribution coverage, not a platform support statement. Keeping the matrix in the workflow makes it explicit and easy to adjust without changing the compatibility contract. Shell and native Go tooling are sufficient, so GoReleaser or another release dependency is unnecessary at this stage.

### Publish only after all artifacts are gathered

Matrix jobs will upload temporary workflow artifacts. A final publication job, dependent on every build job, will gather them, create one checksum manifest, and create the GitHub release. This prevents a failed matrix entry from producing a partially published release. GitHub-generated release notes will be used so maintainers review content before creating the tag rather than maintaining a second changelog format.

### Keep a short pre-tag checklist

`CONTRIBUTING.md` will require maintainers to run or observe repository validation, review the changes that will form the generated release notes for compatibility and breaking changes, and record a smoke test using an `sbx` version inside the supported range. All build identity, artifact exercise, checksum, and publication checks remain automated after tagging.

## Risks / Trade-offs

- **A tag is published before the human checks occur** → Document the checks as pre-tag requirements and protect release-tag creation through maintainer practice; the automated workflow still refuses invalid artifacts.
- **Cross-built executables cannot all run on one runner** → Use native runners for the configured artifact matrix so each uploaded executable is exercised before publication.
- **Generated release notes omit important compatibility context** → Make release-note accuracy an explicit pre-tag review item and require incompatible changes to be identified under the SemVer contract.
- **The documented `sbx` range drifts from code** → Reuse the exact range already covered by doctor tests and include consistency in release review rather than duplicating a machine-readable value in workflow configuration.

## Migration Plan

Merge the license, documentation, and workflow before creating v0.1.0. Run the pre-tag checklist, then create and push the v0.1.0 tag. If the workflow fails, fix the cause on a new commit and issue a new version tag; do not move a published release tag.

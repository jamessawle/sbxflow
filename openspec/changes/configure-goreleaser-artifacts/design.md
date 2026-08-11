## Context

See `proposal.md` for motivation and `specs/release-distribution/spec.md` for the observable contract. The current tag workflow builds four targets in separate native GitHub Actions jobs: Linux amd64, macOS amd64, macOS arm64, and Windows amd64. Each job manually embeds linker metadata and assembles an archive, while the publish job separately creates checksums and a GitHub release.

The CLI already exposes linker-set version and commit values. Mise is the repository's toolchain authority, GitHub Actions are expected to use pinned actions and minimal permissions, and release archives need to remain inspectable with standard platform tools.

## Goals / Non-Goals

**Goals:**

- Make one declarative configuration authoritative for target selection, linker metadata, archive formats, included metadata, and checksums.
- Give contributors one Mise task that builds the entire snapshot artifact set without network publication.
- Exercise each executable on its native target platform in automation rather than treating a successful cross-compilation as sufficient evidence.
- Keep snapshot and tag-release assembly on the same configuration path.

**Non-Goals:**

- Adding targets beyond Linux amd64, macOS amd64, macOS arm64, and Windows amd64.
- Publishing snapshots, updating a Homebrew tap, signing artifacts, generating provenance attestations, or changing tag approval policy.
- Generating SBOMs or adding a separate software-composition-analysis tool without a concrete supply-chain requirement.
- Claiming byte-for-byte reproducibility across different GoReleaser, Go, or host-tool versions; the pinned toolchain and `-trimpath` define the supported build environment.
- Changing the CLI's version-output format or its package architecture.

## Decisions

### Use GoReleaser v2 as the artifact manifest and builder

Add a version-2 GoReleaser configuration at the repository root and pin the exact GoReleaser release in `mise.toml`. The configuration will define one `sbxflow` build with `CGO_ENABLED=0`, `-trimpath`, the existing version and commit linker variables, and the four-target allow-list. Linux and macOS outputs use `tar.gz`; Windows uses `zip`.

This replaces duplicated shell assembly with a locally runnable release definition. Keeping the existing matrix as independent build logic was considered, but it cannot give a clean checkout one command that constructs the complete set and encourages snapshot and publication behavior to drift.

### Give archives an explicit, versioned naming contract

Archives will follow `sbxflow_<version>_<os>_<arch>.<format>`, with GoReleaser's normalized architecture names and `tar.gz` except for Windows `zip`. Each archive will contain the platform executable at its root together with `LICENSE`, `README.md`, and a generated release metadata file that records at least project name, version, commit, target OS, and target architecture. GoReleaser's checksum output will use a fixed `checksums.txt` name and SHA-256.

The metadata file will be generated through GoReleaser's templated archive files or an equally declarative pre-build mechanism and must not depend on a dirty working-tree edit. A single unversioned archive name was considered, but including the version avoids collisions when snapshots or releases are retained together.

### Separate complete snapshot construction from native execution evidence

A Mise task will run GoReleaser in snapshot and clean modes, producing the full matrix in `dist/` without any publish step. Snapshot automation will retain that output and use native runners to unpack the matching archive, verify its documented contents and checksum, invoke `--version`, and invoke `--help`. The expected snapshot version and full source commit will come from GoReleaser metadata rather than being reimplemented independently in shell.

Cross-compiled binaries cannot all be executed on one runner. QEMU or Wine was considered, but native GitHub-hosted runners better match the existing evidence and avoid adding emulation behavior to the release contract.

### Defer SBOM generation

Do not configure GoReleaser's SBOM pipeline in this change. GoReleaser delegates SBOM generation to an additional external tool such as Syft, so enabling it would add another pinned dependency, CI installation step, artifact class, and verification obligation.

Issue #19 makes SBOMs conditional on clean support. Archives and SHA-256 checksums satisfy the current release-evidence need without that extra machinery. SBOMs can be introduced separately if signing, provenance, vulnerability scanning, or downstream policy creates a concrete requirement.

### Use the same configuration for tag publication

Refactor the tag workflow to validate the tag and repository, run the configured GoReleaser build, exercise artifacts on native runners, and publish only after all evidence succeeds. Publication will remain an explicit final GitHub Actions step with `contents: write`; snapshot workflows and local tasks will never run GoReleaser's release mode or receive write permissions.

Using GoReleaser's GitHub publisher directly was considered, but retaining a visibly separate publication boundary makes the snapshot non-publication guarantee and least-privilege permissions easier to audit.

## Risks / Trade-offs

- [GoReleaser template behavior changes between versions] → Pin the exact tool version and validate configuration plus expected filenames and contents in automation.
- [Snapshot versions derived from repository state are mistaken for release versions] → Read the produced identity from GoReleaser metadata and document snapshot artifacts as disposable, non-published outputs.
- [Metadata generation accidentally dirties the checkout] → Generate only beneath GoReleaser's output or temporary directories and test snapshot creation from a clean checkout with a clean-tree assertion afterward.
- [Native exercise jobs duplicate small amounts of archive handling] → Keep target facts in one matrix and use GoReleaser metadata as the source of expected identity; do not duplicate compilation or linker flags.
- [Deferring SBOMs limits machine-readable dependency evidence] → Revisit SBOM generation as a separate change when downstream consumers or supply-chain policy require it.
- [Refactoring release automation could publish incomplete artifacts] → Preserve publication as the final job, require all native exercise jobs, and initially verify the snapshot path before relying on the tag path.

## Migration Plan

1. Add the pinned GoReleaser tool, configuration, metadata template, snapshot task, and artifact documentation.
2. Add automated configuration, snapshot, archive-content, checksum, clean-tree, and native executable checks.
3. Replace manual tag artifact assembly with the same GoReleaser configuration while retaining tag validation and a separately permissioned publication job.
4. Run a snapshot from a clean checkout and retain its artifact manifest as review evidence; do not create a tag or GitHub release.
5. If the tag workflow later fails, restore the prior manual workflow while keeping the local snapshot configuration available for diagnosis; no published tag is moved or overwritten.

## Context

The current tag workflow validates the repository, builds four targets in separate jobs, assembles archives with shell commands, generates checksums, and creates a GitHub release. The project also needs each release published to `jamessawle/homebrew-tap`.

## Goals / Non-Goals

**Goals:**

- Replace custom release assembly with a small GoReleaser v2 configuration.
- Preserve the four existing targets and linker-provided version identity.
- Create the GitHub release and update the Homebrew tap from the tag workflow.
- Keep GoReleaser out of the local Mise toolchain.

**Non-Goals:**

- Adding snapshot workflows, custom artifact-verification scripts, generated per-archive metadata, signing, provenance, or SBOMs.
- Supporting local release publication.
- Adding release targets beyond Linux amd64, macOS amd64, macOS arm64, and Windows amd64.

## Decisions

### Use GoReleaser only for tagged releases

Add a root GoReleaser v2 configuration and invoke a pinned GoReleaser GitHub Action after the existing tag and repository validation. The action builds static, trimmed executables, injects the existing version and commit variables, packages platform archives, generates SHA-256 checksums, and creates the GitHub release.

GoReleaser is a release automation dependency, so it is pinned in the workflow rather than installed through Mise. There is no local snapshot or release task.

### Publish a Homebrew cask

Configure GoReleaser's Homebrew cask publisher for `jamessawle/homebrew-tap`. The workflow supplies a dedicated repository secret to write to the tap; the ordinary GitHub token remains scoped to publishing the sbxflow release.

The cask installs the `sbxflow` executable from the published archives and uses the repository description, homepage, and MIT license metadata.

## Risks / Trade-offs

- [The tap credential is absent or invalid] → GoReleaser fails the release job visibly; repository setup documents the required secret.
- [GoReleaser behavior changes] → Pin both the GitHub Action commit and the exact GoReleaser version.
- [Publication partially succeeds before a tap update fails] → Treat the workflow as failed and repair the tap entry; do not move an existing release tag.

## Migration Plan

1. Add and validate the GoReleaser configuration.
2. Replace manual build and publish jobs with one GoReleaser release job after validation.
3. Configure the tap credential and document Homebrew installation.
4. Validate the workflow and configuration without creating a tag.

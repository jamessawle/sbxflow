## Why

The release workflow manually builds and packages each platform artifact and does not update the project's Homebrew tap. GoReleaser can replace that custom release assembly with a small declarative configuration and publish the corresponding Homebrew cask.

## What Changes

- Define the existing Linux, macOS, and Windows release targets in GoReleaser.
- Preserve version and commit linker metadata, platform archives, and checksums.
- Replace manual artifact assembly and GitHub release creation with a pinned GoReleaser GitHub Action.
- Publish the release cask to `jamessawle/homebrew-tap`.
- Document the Homebrew installation command and release credential requirement.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `release-distribution`: Publish tagged releases through GoReleaser and update the Homebrew tap after repository validation succeeds.

## Impact

The change affects the release-distribution contract, the tag release workflow, contributor release guidance, and a root GoReleaser configuration. GoReleaser is pinned only in GitHub Actions and is not added to Mise or the local contributor workflow.

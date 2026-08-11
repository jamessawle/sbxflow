## 1. GoReleaser configuration

- [x] 1.1 Add a GoReleaser v2 configuration for the four existing targets, static trimmed builds, existing linker metadata, platform archives, and SHA-256 checksums.
- [x] 1.2 Configure the Homebrew cask publisher for `jamessawle/homebrew-tap`.

## 2. Release automation

- [x] 2.1 Replace manual artifact assembly and publication with a pinned GoReleaser GitHub Action while preserving tag and repository validation.
- [x] 2.2 Give only the release job the GitHub and Homebrew tap credentials required for publication.

## 3. Documentation and validation

- [x] 3.1 Document Homebrew installation and the release workflow's tap credential requirement.
- [x] 3.2 Run formatting, OpenSpec validation, GoReleaser configuration validation, actionlint, and `mise run validate`.
- [x] 3.3 Review the final workflow permissions and GoReleaser artifact list.

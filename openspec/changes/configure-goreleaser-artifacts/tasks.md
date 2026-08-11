## 1. Release toolchain and configuration

- [ ] 1.1 Select and pin a compatible GoReleaser v2 version in `mise.toml`.
- [ ] 1.2 Add the GoReleaser configuration for the four supported targets, existing linker metadata, trimmed static builds, platform archive formats, and SHA-256 checksums.
- [ ] 1.3 Add declarative release metadata generation and include the executable, `LICENSE`, `README.md`, and target metadata in every archive without modifying tracked files.
- [ ] 1.4 Add Mise tasks that validate the GoReleaser configuration and create a clean, non-publishing snapshot in `dist/`.

## 2. Artifact verification

- [ ] 2.1 Add a portable verification script or task that reads GoReleaser metadata and checks the exact target set, archive names, formats, required contents, and SHA-256 checksum manifest.
- [ ] 2.2 Add native-platform verification that extracts the selected artifact and checks both `--version` for the expected snapshot version and commit and `--help` for successful execution.
- [ ] 2.3 Prove the snapshot task succeeds from a clean checkout and leaves the tracked working tree unchanged while performing no GitHub release, package-manager, or tag update.

## 3. Automation and publication alignment

- [ ] 3.1 Add non-publishing CI coverage that builds the complete GoReleaser snapshot and uploads its output for verification.
- [ ] 3.2 Exercise Linux amd64, macOS amd64, macOS arm64, and Windows amd64 snapshot executables on matching native runners and require every target to pass.
- [ ] 3.3 Refactor the tag release workflow to consume artifacts built from the same GoReleaser configuration while preserving tag validation, repository validation, least-privilege permissions, and publication only after all artifact evidence succeeds.

## 4. Documentation and validation

- [ ] 4.1 Document the snapshot command, supported matrix, stable archive naming and contents, snapshot identity, and checksum verification in the contributor and release guidance.
- [ ] 4.2 Run Markdown and Go formatting, GoReleaser configuration validation, the clean-checkout snapshot test, and `mise run validate`.
- [ ] 4.3 Review the final `dist/` manifest and GitHub workflow permissions to confirm snapshot testing cannot publish and tag publication still includes only the documented artifacts.

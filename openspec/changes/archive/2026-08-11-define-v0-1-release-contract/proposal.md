## Why

sbxflow needs a small, explicit release contract before publishing v0.1.0 so users can understand its license, versioning policy, and Docker Sandboxes compatibility while maintainers have a repeatable release gate.

## What Changes

- License sbxflow under the MIT License.
- Simplify the README around the released CLI's essential usage and safety-relevant behavior.
- State that sbxflow follows Semantic Versioning and that pre-1.0 minor releases may make incompatible CLI or configuration changes.
- Treat the supported `sbx` version range, rather than operating-system or architecture claims, as the external compatibility boundary.
- Add a tag-driven release workflow that validates the repository, builds versioned artifacts, verifies them, generates checksums, and publishes a GitHub release.
- Keep release approval mostly automated, with release-note accuracy and a supported-`sbx` smoke test as explicit human evidence.

## Capabilities

### New Capabilities

- `release-distribution`: Defines how tagged sbxflow versions are validated, built, identified, checksummed, and published.

### Modified Capabilities

- `cli-interface`: Clarifies the pre-1.0 Semantic Versioning contract for the public CLI and configuration format.
- `environment-diagnostics`: Makes the bounded `sbx` version range the documented external compatibility contract for a release.

## Impact

This affects the repository license, README, release documentation, GitHub Actions workflows, release build metadata, and OpenSpec contracts. It adds no runtime dependency and makes no platform-specific support claim.

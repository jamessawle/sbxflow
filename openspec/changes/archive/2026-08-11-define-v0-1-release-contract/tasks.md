## 1. License and User Contract

- [x] 1.1 Add the standard MIT license with the repository owner's copyright.
- [x] 1.2 Simplify the README while retaining essential configuration, command usage, destructive-operation warnings, and the existing-sandbox reconciliation behavior.
- [x] 1.3 Document the supported `sbx >= 0.35.0, < 0.38.0` range and the concise pre-1.0 Semantic Versioning policy without adding an operating-system, architecture, or missing-feature catalogue.

## 2. Release Procedure

- [x] 2.1 Add a concise pre-tag v0.1.0 checklist to the contributor documentation covering validation, release-note review, breaking-change disclosure, and a supported-`sbx` smoke test.
- [x] 2.2 Add a tag-triggered GitHub Actions workflow that rejects non-canonical version tags, runs repository validation, and builds the configured release artifact matrix with tag and commit metadata.
- [x] 2.3 Exercise every built artifact through its version and help interfaces, gather the artifacts, generate a checksum manifest, and publish only after every required job succeeds.

## 3. Verification

- [x] 3.1 Add or update automated checks for release build identity and tag-derived version handling where repository tests can exercise them directly.
- [x] 3.2 Run Markdown formatting and `mise run validate`, resolving all documentation, workflow, OpenSpec, Go, and architecture failures.

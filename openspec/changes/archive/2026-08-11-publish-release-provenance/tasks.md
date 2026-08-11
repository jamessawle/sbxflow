## 1. Release provenance

- [x] 1.1 Add the full-commit-pinned GitHub build-provenance action after
      GoReleaser and restrict its subjects to the published tarballs, ZIP archive,
      and `checksums.txt`.
- [x] 1.2 Add only `id-token: write` and `attestations: write` to the release
      job while preserving the workflow and validation job's read-only defaults.

## 2. Verification documentation

- [x] 2.1 Add GitHub attestation verification for the archive and checksum
      manifest to the README direct-download instructions.
- [x] 2.2 Extend the release runbook with provenance verification, expected
      attested subjects, and recovery guidance for partial publication or
      attestation failure.

## 3. Validation

- [x] 3.1 Format the Markdown and run actionlint to validate the release
      workflow syntax and pinned-action policy.
- [x] 3.2 Run strict OpenSpec validation and the full repository validation
      suite.
- [x] 3.3 Review the final permission scope and artifact subject patterns, and
      record next-release checks for attestation verification and the subsequent
      OpenSSF Scorecard result.

## 4. Scorecard compatibility

- [x] 4.1 Amend the proposal, specification, and design to require the generated
      in-toto bundle as a GitHub release asset.
- [x] 4.2 Upload the attestation action's bundle output with an `.intoto.jsonl`
      filename and align release verification documentation.
- [x] 4.3 Re-run formatting, strict OpenSpec validation, actionlint, and the full
      repository validation suite.

## Context

The tag workflow runs GoReleaser in one job with `contents: write`; GoReleaser
leaves the published archives and `checksums.txt` in `dist/`. Consumers can
verify archive bytes against that manifest, but both files currently share the
same unauthenticated publication channel. See `proposal.md` for motivation and
`specs/release-distribution/spec.md` for the new behavior contract.

The workflow must keep every third-party action pinned to a full commit and
must not broaden the repository-wide read-only permission default. Existing
releases are immutable and out of scope.

## Goals / Non-Goals

**Goals:**

- Bind every GoReleaser archive and checksum manifest to the repository,
  release workflow, tagged commit, and GitHub Actions identity.
- Let consumers verify downloaded files with GitHub's supported CLI rather
  than project-specific verification code or long-lived signing keys.
- Publish the generated in-toto bundle as a release asset that supply-chain
  scanners can discover without querying GitHub's attestation API.
- Keep provenance authority confined to the release job.

**Non-Goals:**

- Key-based signatures, detached signature assets, SBOM generation, or
  reproducible-build claims.
- Attesting unpacked executables or GoReleaser's internal metadata.
- Backfilling attestations for v0.1.0 or any other existing release.

## Decisions

### Use GitHub artifact attestations

Run the SHA-pinned `actions/attest-build-provenance` action after GoReleaser and
attest the exact archive patterns plus `dist/checksums.txt`. The action uses
GitHub's OIDC identity to create signed Sigstore provenance and records it as a
GitHub artifact attestation. This avoids managing a private signing key and is
directly verifiable with `gh attestation verify <file> --repo
jamessawle/sbxflow`.

Keyless Cosign signing was considered, but it would add another CLI and a
separate verification interface while GitHub's native mechanism already binds
the artifact to the repository workflow. GoReleaser's signing support was also
considered, but managing a long-lived signing key is unnecessary for the
issue's provenance requirement.

### Attest only consumer-facing release files

The subject paths will explicitly cover `dist/*.tar.gz`, `dist/*.zip`, and
`dist/checksums.txt`. This matches the files consumers download and prevents an
over-broad `dist/**` glob from attesting unpacked binaries, metadata, or other
GoReleaser intermediates. GitHub records a distinct subject digest for every
matched file.

Give the attestation step an ID, copy its `bundle-path` output to the stable
`sbxflow-provenance.intoto.jsonl` name, and upload that file to the GitHub
release. GitHub still associates the attestations with the repository for CLI
verification; the additional asset lets the OpenSSF Scorecard
Signed-Releases check discover the provenance by its required filename suffix.

### Scope provenance permissions to the release job

Add `id-token: write` and `attestations: write` beside the release job's
existing `contents: write`. Keep the workflow default and validation job at
`contents: read`. The OIDC token is short-lived and issued only while the
attestation step runs; no new secret is introduced.

### Verify provenance independently of checksums

The README direct-download instructions and release runbook will run `gh
attestation verify` for both the selected archive and `checksums.txt`, then run
the existing checksum check. Provenance authenticates the origin of each file;
the checksum command continues to confirm the archive entry and provide a
familiar offline integrity check after download.

## Risks / Trade-offs

- [GoReleaser publishes assets before an attestation step fails] → Mark the
  workflow failed, treat the release as incomplete, preserve the immutable tag,
  and publish a corrected patch release rather than presenting unattested
  assets as successful.
- [The attestation succeeds but its bundle upload fails] → Let the upload step
  fail the workflow and treat the partially published release as incomplete.
- [Artifact globs drift from GoReleaser output names] → Use extension-based
  patterns for the supported archive formats and explicitly name the checksum
  manifest; review the resulting subject list during release verification.
- [Verification depends on GitHub's attestation service and CLI] → Retain
  SHA-256 checksum verification as an additional integrity mechanism and
  document that provenance verification requires network access and a current
  GitHub CLI.
- [GitHub-hosted provenance establishes workflow origin, not reproducibility]
  → Describe it as build provenance and avoid claims that independent rebuilds
  are byte-identical.

## Migration Plan

1. Add the pinned attestation action and job-scoped permissions after the
   GoReleaser step, then attach its bundle output to the release with an
   `.intoto.jsonl` filename.
2. Update consumer and maintainer verification documentation, including
   partial-publication recovery.
3. Validate workflow syntax, action pinning, OpenSpec artifacts, and repository
   checks without publishing a release.
4. On the next version tag, verify every expected subject and confirm a later
   OpenSSF Scorecard run recognizes the release provenance.

Rollback removes the attestation and bundle-upload steps and the two provenance
permissions on a new commit; already published attestations and release tags
remain immutable evidence.

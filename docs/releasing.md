# Release runbook

This runbook covers preparation, publication, verification, and recovery for an
sbxflow release. Releases are immutable: never move or reuse a published tag.

Repository visibility and release publication are separate operations. The
repository was made public before v0.1.0 so its source, policies, and automation
could be reviewed publicly. Making a repository public does not create a tag,
GitHub release, executable archive, checksum manifest, or Homebrew update.

## Prerequisites

- The release commit is on `main` and all required checks pass.
- `HOMEBREW_TAP_GITHUB_TOKEN` is configured as an Actions secret with permission
  to create a branch and pull request in `jamessawle/homebrew-tap`.
- The maintainer can push tags to `jamessawle/sbxflow` and review the tap pull
  request produced by GoReleaser.
- The planned version is a canonical `v`-prefixed Semantic Version, such as
  `v0.1.1`.

## Prepare the release

1. Update the documented sbxflow version and supported Docker Sandboxes range
   when either changes.
2. Review every change since the preceding tag. Confirm that GitHub's generated
   release notes will identify compatibility changes and breaking changes.
3. Run the full repository validation from the release commit:

   ```text
   mise run validate
   ```

4. Build and smoke-test the release commit on a supported host with an `sbx`
   version inside the range documented in `README.md`:

   ```sh
   mise exec -- go build -o /tmp/sbxflow-release-candidate ./cmd/sbxflow
   /tmp/sbxflow-release-candidate --version
   /tmp/sbxflow-release-candidate --help
   /tmp/sbxflow-release-candidate doctor
   ```

   A local build reports a development identity. This test exercises behavior;
   the published-artifact check below verifies linker-injected release identity.

5. Confirm the tap secret is available and that no earlier sbxflow tap pull
   request is awaiting resolution.

## Tag and publish

Create the tag only after the preparation checks pass. Replace the example
version with the intended release:

```sh
version=v0.1.1
git switch main
git pull --ff-only
git tag "${version}"
git push origin "${version}"
```

Pushing the tag starts `.github/workflows/release.yml`. The workflow validates
the tag and repository, then GoReleaser builds the supported archives, generates
`checksums.txt`, creates the GitHub release, and opens a pull request updating
the Homebrew cask.

Monitor it with:

```sh
run_id=$(gh run list --workflow release.yml --branch "${version}" \
  --json databaseId --jq '.[0].databaseId')
gh run watch "${run_id}" --exit-status
```

Review and merge the generated pull request in `jamessawle/homebrew-tap` after
confirming that its version, URLs, and SHA-256 values match the GitHub release.
The release is complete only after the workflow has published the GitHub assets
and the tap update is merged and validated.

## Verify publication

Set the version without the leading `v`, then download the checksum manifest and
the archive native to the verification host. This Linux amd64 example can be
adapted using the artifact names on the release page:

```sh
version=0.1.1
tag="v${version}"
archive="sbxflow_${version}_linux_amd64.tar.gz"
base_url="https://github.com/jamessawle/sbxflow/releases/download/${tag}"
curl -LO "${base_url}/${archive}"
curl -LO "${base_url}/checksums.txt"
gh attestation verify "${archive}" --repo jamessawle/sbxflow
gh attestation verify checksums.txt --repo jamessawle/sbxflow
grep " ${archive}$" checksums.txt | sha256sum --check
tar -xzf "${archive}"
./sbxflow --version
./sbxflow --help
```

The two attestation checks require network access and a current GitHub CLI.
They must identify `jamessawle/sbxflow` and its release workflow as the source
of both files. Provenance authenticates each file independently; the checksum
check still confirms that the archive matches the published manifest.

Confirm that `--version` reports the tag and a short form of the tagged commit.
Also confirm that the GitHub release contains all expected files and that each
one appears as a subject in the workflow's build-provenance summary:

```text
checksums.txt
sbxflow_<version>_darwin_amd64.tar.gz
sbxflow_<version>_darwin_arm64.tar.gz
sbxflow_<version>_linux_amd64.tar.gz
sbxflow_<version>_windows_amd64.zip
sbxflow-provenance.intoto.jsonl
```

Download and run `gh attestation verify <file> --repo jamessawle/sbxflow` for
each archive and checksum subject, not only the host-native archive used for
the smoke test. Confirm that the provenance bundle is attached to the release
with the exact `.intoto.jsonl` suffix required by OpenSSF Scorecard. After the
next scheduled Scorecard run, confirm that its Signed-Releases check recognizes
the release provenance.

After merging the tap update, test a clean Homebrew installation on macOS:

```sh
brew uninstall --cask sbxflow 2>/dev/null || true
brew install --cask jamessawle/tap/sbxflow
sbxflow --version
sbxflow --help
brew info --cask sbxflow
```

Confirm that the installed version is the release just published. Record the
release URL, workflow run, tap pull request, validation result, supported-`sbx`
smoke-test result, attestation results for every subject, subsequent OpenSSF
Scorecard Signed-Releases result, and Homebrew test in the release issue or
pull request.

## Failure recovery

Before any GitHub release exists, fix the problem on a new commit and create a
new patch-version tag. Do not force-update the failed tag: a pushed tag may
already have been fetched or partially processed.

Publication can partially succeed because GitHub assets are created before the
provenance step or Homebrew update completes. A missing or unverifiable
attestation makes the release incomplete even when its asset was uploaded.
Inspect the GitHub release, workflow run, attestations, and tap repository
separately:

```sh
tag=v0.1.1
gh release view "${tag}"
gh run list --workflow release.yml --limit 5
gh attestation verify checksums.txt --repo jamessawle/sbxflow
gh pr list --repo jamessawle/homebrew-tap --search sbxflow
```

- If provenance generation or bundle upload fails, or any expected subject is
  missing or cannot be verified, do not present the release as complete and do
  not merge its tap update. Preserve the immutable tag and assets, fix the
  problem on `main`, and publish a corrected patch release whose complete
  subject set verifies.
- If GitHub publication succeeded but the tap pull request failed, preserve the
  tag and release only after every expected attestation verifies. Correct the
  tap update in a pull request using the published artifact URLs and checksums,
  then perform the Homebrew validation above.
- If an incorrect release is already public, leave its tag immutable, mark the
  release as superseded where appropriate, fix the problem on `main`, and
  publish a new patch version.
- If the tap points to missing or incorrect artifacts, do not merge it. Correct
  the release through a new version and close the invalid tap pull request.
- If credentials caused the failure, repair or rotate the secret before the new
  release attempt. Never put a token in repository files, command output, or an
  issue.

Deleting a public release or moving its tag is not a rollback strategy: clients
and caches may already hold the original artifacts. Recovery produces a valid
tap entry for unchanged published artifacts or a new version that supersedes
them.

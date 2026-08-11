## Why

sbxflow publishes a Homebrew cask. It is the only cask in
`jamessawle/homebrew-tap` and the only entry there that has never released
cleanly: `v0.1.0`, `v0.1.1`, and `v0.1.2` each required a follow-up commit to
repair a cask that GoReleaser had already pushed and the tap's
`brew test-bot --only-tap-syntax` job had already rejected.

Three separate checks reject the generated cask, and none of them can be
satisfied by GoReleaser configuration:

- `Cask/StanzaOrder`. GoReleaser sorts macOS packages by comparing architecture
  strings, and `"amd64" < "arm64"`, so `on_intel` always precedes `on_arm`.
  Fixed upstream in `goreleaser/goreleaser#6752`, which is merged but unreleased.
- `Layout/EmptyLinesAroundBlockBody`. The cask template leaves a blank line
  before the closing `end`. Same unreleased upstream fix.
- `brew readall --os=all --arch=all` and `brew audit`. The cask is macOS-only
  but cannot declare it, because GoReleaser has no `depends_on macos:` support.

The repository worked around this with `scripts/render-homebrew-cask.sh`, which
regenerates the cask from scratch and rewrites the tap branch after publication.
That repairs the tap instead of preventing the invalid push, duplicates package
metadata that already lives in `.goreleaser.yaml`, and produces a cask that
passes `readall` only because dropping the `on_macos` wrapper makes Linux
resolve a macOS archive.

The four sibling projects in the same tap publish Homebrew formulas through
GoReleaser's formula publisher and land in a single commit. A formula avoids all
three failures by construction: cask stanza cops do not apply to formulas, the
formula template emits no trailing blank line, and a formula covering macOS and
Linux resolves a checksum on every platform Homebrew evaluates.

## What Changes

- Publish the Homebrew update as a formula in `Formula/` rather than a cask in
  `Casks/`, matching the sibling projects already in the tap.
- Build `linux/arm64` alongside the existing platforms so the formula resolves an
  archive and checksum for every platform `brew readall --os=all --arch=all`
  evaluates.
- Remove the post-publication cask rewrite from the release workflow and delete
  `scripts/render-homebrew-cask.sh`, so the first commit pushed to the tap is the
  one that ships.
- Replace the release snapshot check, which validated a file it generated itself,
  with one that validates the package GoReleaser actually generates, plus an
  opt-in check that runs the tap's own syntax checks against it on a current
  Homebrew.
- Update installation, contribution, and release documentation from cask
  commands to formula commands.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `release-distribution`: publish each release through a Homebrew formula whose
  first pushed commit passes the tap's syntax checks unmodified, and require the
  pre-tag release check to validate the package GoReleaser generates.

## Impact

The change affects the GoReleaser configuration, the tag-triggered release
workflow, the release snapshot check, the repository's release scripts, and the
installation and release documentation. Published archives gain `linux/arm64`.
It does not change the executable, its interface, the GitHub release contents
for existing platforms, checksum generation, or build provenance.

Merging this change also requires deleting `Casks/sbxflow.rb` from
`jamessawle/homebrew-tap`, and users who installed the cask must reinstall from
the formula.

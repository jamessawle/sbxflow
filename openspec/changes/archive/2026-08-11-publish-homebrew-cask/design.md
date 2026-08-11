# Design

## Evidence

Verified against Homebrew 6.0.17, the version the tap's test-bot job uses, in a
throwaway container running the tap's three syntax checks over the end-state tap
— the generated cask alongside the tap's existing formulas:

```text
6 files inspected, no offenses detected
brew readall --aliases --os=all --arch=all   (clean)
brew audit --except=installed                (clean)
```

No rules were skipped and no generated file was patched.

## Why the cask, having just moved to a formula

The formula migration was made on the belief that formulas published cleanly,
because the tap's other four entries had no recent corrective commits. That was
wrong. `osch` hit the identical `version is redundant with version scanned from
URL` failure and it was repaired by hand; the tap's `osch` formula carries no
`version` stanza because one was deleted manually. `osch` last released ten weeks
before Homebrew began enforcing the rule, so it has not yet regenerated the line.
sbxflow was simply the first project to release after the rule landed.

GoReleaser also describes its own generated formulas as _"hackyish"_ and directs
projects to casks. The formula path was both deprecated and broken.

## Why Linux archives are in a macOS-only package

Homebrew cannot install a cask on Linux, so the Linux stanzas are never used to
install anything. They exist because `brew readall --os=all --arch=all` rejects a
cask that resolves no checksum on a platform it evaluates, and the alternative —
declaring the cask macOS-only with `depends_on macos:` — is not supported by
GoReleaser, which records it as a `TODO` on its cask dependency type.

The stanzas are accurate: those URLs really do serve those Linux binaries. This
is preferable to the previous workaround, which passed the same check by removing
the `on_macos` wrapper so that Linux resolved a macOS archive.

Prefer `depends_on macos:` once GoReleaser supports it.

## Why the nightly, and why no local check

The fix that makes generated casks pass `brew style` is merged upstream but
unreleased. Pinning a specific nightly is not durable: only about a week of dated
nightly releases are retained, so a pinned tag would disappear and break the
release workflow. The action's `version: nightly` resolves the latest immutable
`vX.Y.Z-<sha>-nightly` release through the Releases API and verifies its
checksums, which is the closest available approximation.

Mise cannot install that build. Its registry lists only stable versions, and the
repository's rolling `nightly` git tag resolves to `2.16.0-nightly`, which
predates the fix. A locally pinned GoReleaser would therefore generate a
different cask from the one the workflow publishes, and a check against it would
report on an artifact that is never shipped. The previous local check already
failed this way once: it passed a formula that the tap rejected, because snapshot
versions never match the release URL and so the audit rule could not fire.

Rather than keep a check that reports on the wrong artifact, the tap's own
test-bot job is the gate until the workflow can pin a released version again.

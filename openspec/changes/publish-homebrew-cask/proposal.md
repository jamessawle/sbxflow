## Why

`v0.1.3` published its GitHub release successfully but its tap pull request was
rejected: `brew audit` reported that the generated formula's `version` stanza is
redundant with the version scanned from its URL. GoReleaser's formula template
emits that stanza unconditionally and offers no option to suppress it, so no
configuration avoids it. The same failure has already been repaired by hand in
the tap for `osch`.

The formula publisher was the wrong destination anyway. GoReleaser deprecated it
in v2.10 and hard-deprecated it in v2.16, describing the formulas it generates as
_"hackyish"_ and directing projects to casks instead. Moving sbxflow to a formula
moved it onto a deprecated path that is itself broken.

The cask is the supported approach, and its remaining blockers are now solved:

- `Cask/StanzaOrder` and `Layout/EmptyLinesAroundBlockBody` are fixed by
  `goreleaser/goreleaser#6752`, merged upstream and shipping in 2.18.0.
- `brew readall --os=all --arch=all` rejected the cask because it resolved no
  checksum on Linux. Including the Linux archives, which the build already
  produces, gives every evaluated platform a URL and checksum. GoReleaser's cask
  template supports this directly.
- The cask never removed the macOS quarantine attribute, so the installed binary
  risked being refused by Gatekeeper. GoReleaser documents a postflight hook for
  binaries that are neither signed nor notarized.

## What Changes

- Publish the Homebrew update as a cask again, with the Linux archives included
  so every platform Homebrew evaluates resolves a checksum.
- Add the documented postflight hook that removes the quarantine attribute.
- Track the GoReleaser nightly in the release workflow until 2.18.0 ships, since
  the fix that makes generated casks pass `brew style` is merged but unreleased.
- Remove the local packaging checks and the pinned local GoReleaser. Mise cannot
  install the nightly the workflow uses, so a local check would validate a
  different cask from the one published.
- Restore cask installation commands in the installation and release
  documentation.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `release-distribution`: require the first commit pushed to the tap to pass the
  tap's checks, require the cask to resolve an artifact on every platform
  Homebrew evaluates, and require the published cask to be runnable on macOS
  without the user clearing the quarantine attribute.

## Impact

The change affects the GoReleaser configuration, the release workflow's
GoReleaser version, the Mise tool set and tasks, and the installation and release
documentation. Published archives are unchanged. Homebrew installation returns to
`brew install --cask`; Linux users install from the release archives.

`v0.1.3` has no tap entry. Its pull request is closed unmerged and the tap moves
from `0.1.2` to `0.1.4`.

# Design

## Evidence

The failure and the fix were both reproduced against Homebrew 6.0.17, the
version the tap's `brew test-bot` job uses, in a throwaway container running the
tap's three syntax checks over the whole tap.

The cask GoReleaser 2.17.1 generates today reproduces every failure seen on the
`v0.1.2` tap run:

```text
Casks/sbxflow.rb:6:5:  Cask/StanzaOrder: on_intel stanza out of order
Casks/sbxflow.rb:10:5: Cask/StanzaOrder: on_arm stanza out of order
Casks/sbxflow.rb:27:1: Layout/EmptyLinesAroundBlockBody
Error: Invalid cask (Linux on Intel x86_64) ... Missing Linux stanzas
Error: Invalid cask (Linux on ARM64)        ... Missing Linux stanzas
audit: a sha256 stanza is required, a url stanza is required
```

The formula this change generates, placed in the tap alongside the four existing
formulas with `Casks/sbxflow.rb` removed, passes all three checks:

```text
6 files inspected, no offenses detected
brew readall --aliases --os=all --arch=all   (clean)
brew audit --except=installed                (clean)
```

## Why a formula rather than a corrected cask

A cask cannot be made valid by configuration today. Two of the three failures
are in GoReleaser's cask template and are fixed only in an unreleased upstream
commit; the third needs a `depends_on macos:` stanza that GoReleaser does not
support at all, recorded upstream as a `TODO` on its cask dependency type.

A formula sidesteps all three. `Cask/StanzaOrder` is a cask cop and cannot fire
on a formula, whose template emits `if Hardware::CPU.intel?` conditionals rather
than cask stanzas. The formula template emits no trailing blank line. A formula
covering macOS and Linux resolves a URL and checksum for every platform, so no
`depends_on` declaration is needed.

A formula also removes a defect the cask carried: casks quarantine their
downloads, and sbxflow is neither signed nor notarized, so the installed binary
risked being refused by Gatekeeper. Formula downloads are not quarantined.

## Rejected alternatives

**Keep the cask and wait for the upstream release.** GoReleaser 2.18.0 fixes the
two style offences but not the `readall` failure, so the tap would still reject
the first pushed commit. This would need the upstream `depends_on macos:`
feature as well, which is not yet implemented or filed.

**Keep the cask, stop GoReleaser publishing it, and patch its output before
opening the tap pull request.** This fixes the ordering problem but keeps a
bespoke patch step for output that upstream intends to generate correctly, and
still leaves the Gatekeeper quarantine problem.

**Keep regenerating the cask after publication.** Rejected as the status quo
under review: it repairs the tap after an invalid commit has already triggered a
failing run, and its output only passes `readall` by offering Linux a macOS
archive.

## Deprecation

GoReleaser has soft-deprecated its formula publisher in favour of casks since
v2.10 and hard-deprecated it since v2.16, and prints a deprecation notice on
every run. No removal version is announced, and deprecated options are removed
at major versions, so the publisher is expected to work until GoReleaser v3.

This is accepted rather than overlooked. The four sibling projects in the tap
already use the formula publisher, so the migration away from it is a decision
for the whole tap at v3, driven by upstream's timeline. Until then this change
makes sbxflow consistent with them instead of being the single entry that cannot
release without manual repair. Returning to a cask becomes viable once
GoReleaser both ships the merged style fix and supports `depends_on macos:`.

## Platform coverage

`brew readall --os=all --arch=all` evaluates a formula for macOS and Linux on
both Intel and ARM. The existing build matrix omits `linux/arm64`, so the
formula would leave that platform without a URL or checksum. Adding it is free
for a statically linked Go binary and makes the generated formula's coverage
match the sibling projects'.

Windows is unaffected: GoReleaser's formula publisher filters archives to macOS
and Linux, so the published Windows ZIP stays out of the formula without needing
a separate build definition to exclude it.

## Validating before the tag

The previous check wrote its own cask over the path GoReleaser had just written,
then asserted properties the writing script hardcoded, so it could not fail and
destroyed the only artifact worth inspecting. Its replacement inspects the file
GoReleaser generates.

Structural assertions run in the ordinary validation suite. The full tap syntax
check needs a current Homebrew and a network, so it is a separate opt-in task
rather than part of every validation run, and the release runbook requires it
before tagging.

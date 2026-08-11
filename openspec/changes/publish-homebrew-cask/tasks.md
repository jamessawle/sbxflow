## 1. Cask publication

- [x] 1.1 Replace the GoReleaser formula publisher with the cask publisher
      targeting `Casks/` in `jamessawle/homebrew-tap`.
- [x] 1.2 Include the Linux archives so every platform Homebrew evaluates
      resolves an artifact and checksum.
- [x] 1.3 Add the documented postflight hook that removes the macOS quarantine
      attribute from the installed executable.

## 2. GoReleaser version

- [x] 2.1 Track the GoReleaser nightly in the release workflow until 2.18.0
      ships, recording why in the workflow itself.
- [x] 2.2 Remove the pinned local GoReleaser and the packaging tasks that used
      it, since Mise cannot install the build the workflow uses.

## 3. Documentation

- [x] 3.1 Restore cask installation, upgrade, and uninstall commands in the
      README and note where Linux and Windows users install from.
- [x] 3.2 Update the release runbook for cask publication and explain why there
      is no local packaging check.
- [x] 3.3 Remove the deleted tasks from the contribution guide.

## 4. Validation

- [x] 4.1 Confirm the generated cask passes the tap's style, readall, and audit
      checks in the end-state tap, on the Homebrew version the tap's CI uses.
- [ ] 4.2 Run `mise run fmt`, strict OpenSpec validation, and the full
      repository validation suite.

## 5. Release

- [ ] 5.1 Close the `v0.1.3` tap pull request unmerged.
- [ ] 5.2 Publish `v0.1.4` and confirm the tap receives one commit and one
      passing test-bot run.

## 6. Follow-up

- [ ] 6.1 Pin the workflow back to GoReleaser 2.18.0 when it is released, and
      restore a local packaging check once Mise can install that version.

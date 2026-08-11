## 1. Homebrew formula publication

- [x] 1.1 Replace the GoReleaser Homebrew cask publisher with the formula
      publisher targeting `Formula/` in `jamessawle/homebrew-tap`, carrying the
      existing name, description, homepage, and licence metadata.
- [x] 1.2 Add `linux/arm64` to the build matrix and collapse the cask-driven
      split into a single build definition, keeping the Windows ZIP override.
- [x] 1.3 Add a formula test that runs the published executable's `--version`.

## 2. Remove the post-publication rewrite

- [x] 2.1 Delete the `Normalize generated Homebrew cask` step from the release
      workflow so the first commit pushed to the tap is the one that ships.
- [x] 2.2 Delete `scripts/render-homebrew-cask.sh`.

## 3. Validate what actually ships

- [x] 3.1 Rewrite the release snapshot check to assert the published archive set
      and the structure of the formula GoReleaser generates, without overwriting
      it.
- [x] 3.2 Add a script and opt-in task that run the tap's own `brew style`,
      `readall`, and `audit` checks against the generated formula on a current
      Homebrew.

## 4. Documentation

- [x] 4.1 Update the README installation, upgrade, and uninstall instructions
      from cask commands to formula commands, and note Linux support.
- [x] 4.2 Update the release runbook for formula publication, the expanded
      artifact set, the pre-tag tap check, and clean-install verification.
- [x] 4.3 Update the contribution guide's description of the release check.

## 5. Validation

- [x] 5.1 Run the tap syntax check against the generated formula in the
      end-state tap and confirm it passes on the Homebrew version the tap's CI
      uses.
- [x] 5.2 Run `mise run fmt`, strict OpenSpec validation, and the full
      repository validation suite.

## 6. Follow-up outside this repository

- [ ] 6.1 Delete `Casks/sbxflow.rb` from `jamessawle/homebrew-tap` once the first
      formula release is merged.

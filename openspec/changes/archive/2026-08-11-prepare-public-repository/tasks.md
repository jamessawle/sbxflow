## 1. Security and conduct policy

- [x] 1.1 Add `SECURITY.md` with supported-version guidance, GitHub private vulnerability-reporting instructions, public-reporting cautions, and expectations consistent with the pre-1.0 release contract.
- [x] 1.2 Select a monitored private conduct-reporting route and add the Contributor Covenant 2.1 as `CODE_OF_CONDUCT.md` with project-specific enforcement details.
- [x] 1.3 Link security and conduct expectations from `CONTRIBUTING.md` without duplicating their full policy text.

## 2. Contribution intake

- [x] 2.1 Add YAML bug-report and feature-request issue forms that collect actionable context and direct vulnerability reports to `SECURITY.md`.
- [x] 2.2 Add issue-template configuration that disables blank issues and exposes the private security-reporting contact link.
- [x] 2.3 Review and minimally update the pull request template so its checks agree with contribution, conduct, validation, and release requirements.

## 3. Publication audit

- [x] 3.1 Run and document current-tree and reachable-history scans for credential patterns, private data, internal-only links, suspicious filenames, and unsuitable fixtures.
- [x] 3.2 Manually review public-facing links, repository declarations, examples, commit attribution, license wording, release documentation, and community-profile discovery.
- [x] 3.3 Record the owner's decision to accept `jamessawle@hotmail.com` as public commit attribution or separately authorize and complete a coordinated history rewrite.
- [x] 3.4 Add a publication-readiness checklist recording audit evidence, findings, repository metadata, GitHub setting state, and every step that must accompany the visibility change.

## 4. GitHub security and metadata

- [x] 4.1 Enable dependency vulnerability alerts and Dependabot security updates, retaining the existing weekly version-update configuration, and verify the resulting state.
- [x] 4.2 Add accurate repository topics and verify the existing description, license, issues, merge strategy, branch deletion, and `main` protection settings.
- [x] 4.3 Record CodeQL default setup for Go, secret scanning with push protection, and private vulnerability reporting as mandatory visibility-transition steps; enable and verify any of them that GitHub permits while private.
- [x] 4.4 Verify the resulting GitHub community profile and all security-reporting links without changing repository visibility.

## 5. Validation

- [x] 5.1 Run `mise run fmt` and review the formatted policy, forms, guidance, audit, and OpenSpec artifacts.
- [x] 5.2 Run `mise run validate` and resolve all repository checks.
- [x] 5.3 Recheck the working tree and GitHub settings, confirming that repository visibility remains private and documenting any owner-only publication gates.

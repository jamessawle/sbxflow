# Publication readiness

This checklist records the evidence and owner actions required before changing
`jamessawle/sbxflow` from private to public. Completing repository preparation
does not authorize or perform the visibility change.

## Content and history audit

Reviewed on 2026-08-11 against the current tree and all commits reachable from
local refs:

- [x] `git grep` searches for credential assignments, private-key headers,
      common provider token formats, credential-bearing URLs, private network
      addresses, and local-only hosts returned no matches outside dependency
      checksums.
- [x] Every reachable commit was searched for private-key headers and common
      provider token formats; no matching commit was found.
- [x] Current and historical paths were checked for environment files, private
      keys, certificates, keystores, credential files, and secret files; no
      suspicious path was found.
- [x] Tracked examples and test fixtures were manually reviewed. Reserved
      `example.com` names are intentional test data; no unsuitable fixture or
      private data was found.
- [x] Public-facing URL hosts were reviewed. They refer to Docker, GitHub,
      JSON Schema, Mise, OpenSpec, Semantic Versioning, Conventional Commits, or
      reserved example domains. No internal-only link was found.
- [x] `sbxflow.yaml`, examples, README, contribution and release guidance,
      schema references, and license wording were reviewed. The declaration's
      `jamessawle/sbx-kits` source is public, and the MIT license is detected by
      GitHub.
- [x] Reachable commit identities were reviewed. The owner accepts
      `jamessawle@hotmail.com` as public commit attribution; no history rewrite is
      required. `git-ai@local` and GitHub's no-reply identity contain no personal
      data requiring remediation.

These negative pattern-scan results reduce risk but do not prove that no secret
or private information exists. The manual review is part of the publication
decision, and a final diff and history review remains required immediately
before changing visibility.

## Repository metadata and current settings

Verified through the GitHub API on 2026-08-11:

- [x] Visibility remains private.
- [x] Description accurately summarizes the repository; no homepage is set.
- [x] Issues are enabled, the wiki is disabled, and the MIT license is detected.
- [x] Squash merging is the only merge strategy and merged branches are deleted.
- [x] `main` requires an up-to-date `Validate` check, linear history, and
      administrator enforcement; force pushes and deletion are disabled.
- [x] Repository topics include `cli`, `docker`, `docker-sandboxes`, `go`, and
      `sandbox`.
- [x] Dependency vulnerability alerts and Dependabot security updates are
      enabled; the weekly grouped version-update configuration remains in
      `.github/dependabot.yml`.

GitHub accepted and returned the topics, vulnerability-alert endpoint returned
`204 No Content`, and the automated-security-fixes endpoint returned `200 OK`.
The private repository currently returns `403` for CodeQL default setup, `422`
for secret scanning, and `404` for private vulnerability reporting. Those
features therefore remain mandatory visibility-transition gates below.

The repository-owned community files and issue forms use GitHub's standard
discovery paths, their Markdown and YAML parse successfully during formatting,
and the full repository validation suite passes. GitHub's live community
profile will not discover the new files until they are committed and pushed;
live discovery and private-reporting link behavior remain explicit transition
checks below.

## Implementation validation

Completed on 2026-08-11:

- [x] `mise run fmt`
- [x] `git diff --check`
- [x] `mise run validate`
- [x] Final GitHub API review confirmed visibility remains private, topics and
      dependency security settings remain enabled, and the documented repository
      and branch settings are unchanged.

## Visibility-transition gates

The owner must complete and verify these steps immediately before or after
making the repository public, according to GitHub feature availability:

- [ ] Re-run current-tree and reachable-history scans and manually review the
      final diff, filenames, links, fixtures, and commit attribution.
- [ ] Confirm repository metadata, topics, branch protection, merge settings,
      and dependency security settings still match this document.
- [ ] Enable GitHub CodeQL default setup for Go and verify its first analysis.
- [ ] Enable secret scanning and push protection and verify both report enabled.
- [ ] Enable private vulnerability reporting and verify that the link in
      `SECURITY.md` and the issue-template contact link open a private report.
- [ ] Verify GitHub's community profile discovers the README, contribution
      guide, code of conduct, security policy, license, issue forms, and pull
      request template.
- [ ] Make the separate owner-approved visibility change; this document and its
      implementation do not grant that approval.
- [ ] After the visibility change, inspect the public repository while signed
      out and confirm releases, source archives, actions, issues, security links,
      and community files expose only intended content.

CodeQL, secret scanning with push protection, and private vulnerability
reporting are mandatory publication gates. If GitHub does not permit a feature
while the repository is private, leave visibility private, retain the unchecked
gate, and enable it as part of the coordinated transition.

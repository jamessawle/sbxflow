## Context

See `proposal.md` for motivation. The repository is currently private, has MIT licensing and a documented pre-1.0 release contract, and already carries contribution and pull request guidance. Weekly Dependabot version updates and protected `main` validation are configured, while vulnerability alerts, automated security fixes, CodeQL, secret scanning, private vulnerability reporting, issue templates, and a code of conduct are absent.

The initial audit found no credential-shaped tracked content or history. Reachable commits expose the author's `jamessawle@hotmail.com` address, and the repository declaration references `jamessawle/sbx-kits`, which is public. Some GitHub security features are unavailable without GitHub Advanced Security while the repository is private but become available at no cost for a public repository.

## Goals / Non-Goals

**Goals:**

- Make repository-owned policy concise, usable, and discoverable through GitHub community-health conventions.
- Separate settings that can safely be enabled now from settings that must accompany the visibility transition.
- Produce an auditable publication decision rather than treating a pattern scan as proof that history is safe.
- Keep security automation compatible with the existing pinned-action and least-privilege workflow style.

**Non-Goals:**

- Changing repository visibility.
- Rewriting Git history or force-pushing without a separate, explicit owner decision.
- Changing sbxflow runtime behavior, architecture, licensing, or the v0.1 release contract.
- Promising response or remediation service levels the sole maintainer cannot reliably meet.

## Decisions

### Use standard community-health locations

Add `SECURITY.md` and `CODE_OF_CONDUCT.md` at the repository root, and YAML issue forms plus their configuration under `.github/ISSUE_TEMPLATE/`. Update `CONTRIBUTING.md` and the pull request template only where they need to link or reinforce these policies. Standard locations make the files visible in GitHub's community profile and avoid a bespoke documentation index.

The security policy will support the latest released minor line and explain that older pre-1.0 lines may not receive fixes. Vulnerabilities will be reported through GitHub private vulnerability reporting, with public issues explicitly discouraged. It will avoid fixed response-time promises.

Use the Contributor Covenant 2.1 text for conduct expectations, with a project-specific private enforcement route selected before implementation. A custom short code is easier to maintain but gives contributors less familiar and less complete guidance.

### Use structured issue intake and retain the focused pull request template

Provide forms for bug reports and feature requests, route security reports to `SECURITY.md`, and disable blank issues so required diagnostic and scope information is consistently collected. The existing pull request template remains small; it gains policy-facing checks only where contributors can act on them.

### Stage GitHub security settings around visibility

Enable the dependency graph, vulnerability alerts, and Dependabot security updates while private. Keep the existing weekly grouped version-update configuration, since version maintenance and vulnerable-dependency remediation serve different purposes.

Enable CodeQL default setup for Go, secret scanning with push protection, and private vulnerability reporting immediately before or after making the repository public, depending on GitHub feature availability. Default CodeQL setup is preferred over a committed workflow because this project uses a conventional Go build and does not need custom extraction; it also avoids a workflow that fails while private CodeQL is unavailable. If default setup cannot express a future build requirement, replace it with a SHA-pinned, least-privilege workflow then.

### Record evidence and owner decisions in a publication checklist

Add a repository document that records the audit scope, commands or methods used, findings, GitHub metadata/settings, and remaining visibility-transition steps. It will distinguish automated negative findings from manual review and explicitly record the disposition of the author email in reachable commits.

Do not rewrite history as an incidental implementation step. The owner must either accept the address as public attribution or separately authorize a coordinated rewrite and force-push. This is the only safe treatment because rewriting changes every affected commit identifier and disrupts existing pull request and release references.

### Keep repository metadata minimal and accurate

Retain the current description, MIT license detection, issues, squash-only merging, automatic branch deletion, protected `main`, and disabled wiki. Add accurate GitHub topics during the settings pass. Do not add a homepage until a maintained project site exists, and do not enable Discussions without a demonstrated community need.

## Risks / Trade-offs

- [Security features unavailable while private] → Record them as visibility-transition gates and enable the subset GitHub permits now.
- [A private-reporting link is unavailable before public reporting is enabled] → Phrase `SECURITY.md` clearly and verify the link during the final settings pass before visibility changes.
- [Disabling blank issues may make unusual reports harder] → Give each form an open-ended context field and point questions to an appropriate existing form.
- [Automated scans miss contextual private information] → Pair pattern-based current-tree and reachable-history scans with manual review of filenames, links, fixtures, commit attribution, and repository metadata.
- [Contributor Covenant enforcement needs a private contact] → Resolve and verify a monitored route before adding the document; do not publish an unmonitored placeholder.
- [Accepting the historical email exposes personal data] → Make the disposition an explicit owner gate; a rewrite remains outside this change unless separately authorized.

## Migration Plan

1. Add and validate repository-owned policy, issue forms, guidance, and the publication audit.
2. Enable dependency vulnerability alerts and automated security fixes while the repository is private.
3. Resolve the conduct-reporting contact and historical author-email disposition.
4. Immediately around the visibility change, enable CodeQL default setup, secret scanning and push protection, and private vulnerability reporting; verify community-profile and security links.
5. If settings cause unexpected automation failures, disable the affected feature without removing policy files, document the failure, and restore it after correction. Repository visibility remains a separate owner action.

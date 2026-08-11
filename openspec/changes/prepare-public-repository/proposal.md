## Why

The repository needs explicit community, security, and maintenance expectations before it becomes public. Its tracked content, history, and GitHub security settings also need a recorded publication audit so public visibility does not expose unsuitable material or omit basic safeguards.

## What Changes

- Add a security policy with supported-version guidance and a private vulnerability-reporting route.
- Add explicit contributor conduct expectations and structured issue intake while keeping pull request guidance aligned with the existing workflow.
- Enable repository security maintenance through Dependabot vulnerability alerts and security updates, CodeQL analysis for Go, secret scanning, and private vulnerability reporting where GitHub visibility permits them.
- Audit tracked content and reachable Git history for secrets, private data, unsuitable fixtures, and internal-only links, explicitly resolving whether the existing author email may become public.
- Review and record the repository metadata and GitHub settings required before changing visibility, without changing visibility as part of this change.
- Keep all public policy consistent with the MIT license and the documented pre-1.0 release contract.

## Capabilities

### New Capabilities

None. This change governs repository policy, community health, and development security rather than sbxflow's product behavior.

### Modified Capabilities

None.

## Impact

The change affects root policy documents, `.github` community and workflow configuration, repository validation inputs, and owner-managed GitHub settings. It does not change the CLI, configuration schema, package architecture, runtime dependencies, or repository visibility.

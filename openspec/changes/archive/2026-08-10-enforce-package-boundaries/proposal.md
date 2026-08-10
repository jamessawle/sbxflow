## Why

The current architecture policy mirrors today's package graph rather than
expressing reusable architectural roles. That makes the check difficult for a
new contributor to reason about and can report success without proving that
important boundaries, such as leaf packages, actually reject violations.

## What Changes

- Organize production packages around explicit architectural types:
  entrypoint, inbound adapter, application workflow, domain service/model,
  port, and outbound adapter.
- Move CLI, repository declaration, and Docker Sandbox integrations beneath
  dedicated inbound and outbound adapter roots.
- Treat configuration as the domain concept and validation as the application
  use case, while separating configuration rules from declaration I/O and
  removing dependencies between peer application workflows.
- Make `cmd/sbxflow` the startup and wiring boundary while keeping the public
  CLI behavior unchanged.
- Add capability-scoped declaration, sandbox, and BuildInfo ports so consumers
  share narrow boundaries without adapters importing their callers or
  introducing behavior-free pass-through layers.
- Enforce relationships between architectural types instead of enumerating the
  current package graph. Only the entrypoint may import concrete outbound
  adapters; inbound-to-BuildInfo and application-to-Sandbox are the two narrow
  port shortcuts.
- Replace the Arch-Go prototype with a pinned, declarative component linter
  that fails for unclassified packages, invalid type relationships, and
  prohibited external dependencies without repository-owned checker code.
- Keep the architecture check mandatory in local hooks and pull-request
  validation, with actionable diagnostics for contributors.

## Capabilities

### New Capabilities

None. This change affects repository structure and development tooling rather
than product behavior.

### Modified Capabilities

None. Existing CLI, configuration, diagnostics, and lifecycle requirements are
unchanged.

## Impact

The change reorganizes internal Go packages, dependency wiring, architecture
documentation, and development tooling. It replaces the pinned Arch-Go tool
and its package-specific policy with go-arch-lint after the type-based check is
in place. The executable's commands, configuration format, published JSON
Schema, and observable lifecycle behavior remain unchanged.

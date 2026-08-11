## Why

Repositories that need access to a small number of additional network resources
currently have to select or maintain a complete custom kit. Docker Sandboxes
already supports sandbox-scoped local allow rules, so sbxflow can expose that
capability directly in the repository declaration while preserving stricter
organisation policy.

## What Changes

- Add an optional `sandbox.network.allowedHosts` list to version 1 declarations.
- Strictly validate the list as unique hosts, domains, wildcard subdomains, IP
  literals, or `**`, each with an optional port, rejecting URL forms that Docker
  accepts but can never match.
- Split creating a missing sandbox from entering it, and apply declared hosts as a
  sandbox-scoped local allow rule in between, because Docker scopes a rule only to
  a sandbox that already exists.
- Share removal logic between destroy and recreation so both clean up the
  currently declared sandbox-scoped resources after removal.
- Preserve the existing no-reconciliation behavior for an already-created
  sandbox; users recreate the sandbox to apply declaration changes.
- Document the field, its interaction with organisation governance, and a
  complete example.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `configuration-validation`: Accept and validate declared sandbox network allow
  resources.
- `sandbox-lifecycle`: Apply declared network resources when creating a sandbox
  without weakening organisation-managed policy or reconciling existing
  sandboxes.

## Impact

The public YAML schema, declaration port model, lifecycle plan and sandbox port,
Docker Sandbox outbound adapter, README, and examples change. The sandbox port
gains a `Creator` capability and `Runner` becomes attachment-only, so `up` now
invokes `sbx create` and `sbx run --name` in place of a single `sbx run`. The
implementation uses the existing supported `sbx create`, `sbx run`, and
`sbx policy allow network --sandbox` CLI and adds no dependency or architecture
relationship.

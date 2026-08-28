## Why

Docker Sandboxes 0.39 introduced declarative sandbox environments that overlap
with sbxflow's lifecycle adapter and provide the upstream foundation for future
workspace, resource, credential, port, and MCP configuration. Adopting that
foundation now lets sbxflow concentrate on its differentiated repository
validation, least-privilege kit trust, standalone network declarations, and
safe lifecycle behavior instead of continuing to assemble low-level create and
run flags.

## What Changes

- Keep `sbxflow.yaml` as the repository-owned public declaration and render a
  private temporary `.sbxenv.yaml` outside every mounted workspace for Docker
  Sandboxes lifecycle commands.
- Use `sbx env create`, `sbx env run`, `sbx env exec`, and `sbx env rm` for
  provisioning, interactive entry, initialization, rollback, recreation, and
  destruction; retain machine-readable inspection and `sbx stop` for the
  existing state-aware and state-preserving behavior.
- Preserve `sandbox.network.allowedHosts` as a standalone sandbox-scoped rule
  applied after environment creation and before agent entry, without requiring
  a generated kit.
- Preserve derived kit-source trust as process-local environment overrides for
  every environment command that can resolve the rendered kit references,
  without changing persistent Docker settings.
- Preserve initialization ordering, literal argument vectors, workspace
  selection, stream behavior, and rollback guarantees through `sbx env exec`
  and environment-owned removal.
- Preserve identity-only, idempotent teardown: inspect before removal so an
  absent sandbox remains silent, and render a minimal environment file for
  teardown paths that intentionally do not validate unrelated configuration.
- **BREAKING**: require Docker Sandboxes 0.39.x, the first release line that
  provides `sbx env`, and stop accepting the earlier 0.35-0.37 range. Keep the
  upper bound below 0.40 while the environment interface is experimental.
- Update the repository's own `jamessawle/sbx-kits` pin from the incompatible
  `v2026.08.02` schema to the SBX 0.39-compatible `v2026.08.03` release.
- Update documentation and automated coverage for rendered environment files,
  process-scoped trust, mixed kit/standalone network rules, initialization
  rollback, resilient teardown, and the new compatibility bounds.
- Make the existing sandbox lifecycle requirements backend-neutral where they
  currently prescribe individual `sbx` commands, without changing sbxflow's
  user-visible lifecycle contract.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `sandbox-lifecycle`: Replace command-specific lifecycle wording with
  backend-neutral requirements while preserving sbxflow's validation, trust,
  network, initialization, interactive, recreation, and teardown contracts.

## Impact

- Affects lifecycle planning and the sandbox port request types used by
  `internal/application/lifecycle`.
- Reworks subprocess construction and temporary-file handling in
  `internal/adapters/outbound/sbx` without adding an architectural type or
  broadening the repository dependency matrix.
- Updates doctor compatibility constants and their boundary tests.
- Updates lifecycle unit and executable tests, README lifecycle documentation,
  examples where required, and the repository's own `sbxflow.yaml` kit pin.
- Requires Docker Sandboxes 0.39.x and depends on its experimental environment
  file schema and command behavior.

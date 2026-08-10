## Why

The package dependency rules documented in `docs/structure.md` currently rely
on contributor awareness and review. As the codebase grows, an automated check
will prevent accidental dependency inversions from weakening those boundaries.

## What Changes

- Add a pinned Arch-Go development check that compares production Go imports
  with a declarative package dependency policy.
- Reject unapproved dependency edges with actionable failure messages while
  allowing tests to use broader dependencies for fixtures and test doubles.
- Align `docs/structure.md` with the complete intentional graph, including the
  existing `cmd/sbxflow` to `buildinfo` and `configuration` to `schema` edges.
- Expose the architecture check as a dedicated Mise task and require it from
  aggregate validation without coupling it to the ordinary Go test suite.
- Add repository guidance that architecture policies and their enforcement must
  not be weakened merely to make an unrelated change pass.

## Capabilities

### New Capabilities

None. This change adds repository tooling and does not introduce product
behavior.

### Modified Capabilities

None. Existing CLI, validation, diagnostics, and lifecycle requirements remain
unchanged.

## Impact

The change affects repository architecture policy, pinned development tools,
Mise tasks, agent and contributor guidance, internal architecture
documentation, and validation feedback. It does not alter the `sbxflow`
executable, its public schema, or its runtime dependencies.

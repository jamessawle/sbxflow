## Why

Sandbox creation is not always sufficient to make a repository ready for its
agent: asynchronous kit setup, sandbox-local dependency installation, and
repository bootstrap work may need to finish first. Without a declarative
initialization boundary, consumers must duplicate sbxflow's host-side lifecycle or
allow the agent to start against an incompletely prepared workspace.

## What Changes

- Extend version 1 declarations with an optional ordered list of initialization
  commands, represented as explicit argument vectors rather than implicit shell
  strings.
- Validate command entries before Docker state is inspected or changed, without
  executing them during validation.
- Run every declared initialization command inside the exact sandbox, from the
  declared workspace, only when `up` creates a missing or replacement sandbox.
- Stream command output through the CLI, stop at the first failure, identify the
  failed command, and remove the incomplete sandbox plus its declared scoped
  resources before returning failure.
- Keep ordinary `up` behavior unchanged for existing running or stopped
  sandboxes; changed initialization commands require `up --recreate`.
- Keep declarations without initialization commands and the `validate`, `down`, and
  `destroy` lifecycles behaviorally unchanged.
- Document the declaration contract and demonstrate it in the README, public
  schema, and repository examples, including the `language-node-npm` readiness
  use case from issue #94.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `configuration-validation`: Define the strict version 1 declaration shape and
  semantic validation for ordered initialization command argument vectors.
- `sandbox-lifecycle`: Require `up` to execute declared commands in order inside
  a newly created sandbox and workspace before agent entry, with rollback on
  initialization failure.

## Impact

The change affects the public `sbxflow.yaml` contract and JSON Schema, declaration
decoding and resolved configuration, lifecycle orchestration, the Sandbox port,
Docker subprocess construction, CLI stream wiring and errors, documentation and
examples, and unit plus executable-level tests. It adds no new architectural
package type or dependency direction and preserves compatibility for declarations
that omit the optional lifecycle configuration.

## Context

See `proposal.md` for motivation and the delta specs for observable behavior.
The validated lifecycle plan currently carries sandbox identity, agent,
workspace, selected kits, derived trust, and network resources. The lifecycle
application coordinates provisioning, network policy, and interactive attachment
through narrow interfaces in the Sandbox port; the outbound `sbx` adapter owns
all Docker Sandboxes subprocess construction.

The new declaration data crosses the public schema, declaration adapter, domain
configuration model, lifecycle plan, Sandbox port, outbound adapter, and CLI
streams. It must stay within the dependency directions in `docs/structure.md`.

## Goals / Non-Goals

**Goals:**

- Preserve command and argument order without introducing implicit shell
  evaluation.
- Add one initialization phase between provisioning/policy application and agent
  attachment for newly created and recreated sandboxes.
- Keep command execution and stream behavior behind a narrow Sandbox capability.
- Preserve the invariant that every sandbox left by a successful `up` completed
  its declared initialization.

**Non-Goals:**

- Add declarative timeouts, retries, conditional execution, environment
  overrides, or
  per-command working directories to the declaration.
- Detect whether a command is idempotent or has already completed.
- Run host commands or provide an implicit shell.
- Reconcile hooks, workspace, kits, or network configuration on an existing
  sandbox.
- Define a per-entry hook for ephemeral token or environment loading.

## Decisions

### Model initialization hooks as ordered command objects

Add optional `sandbox.hooks.initialize`, an array of objects whose required
`command` field is a non-empty array of non-empty strings. The object form leaves
room for future command properties, while the `hooks` namespace leaves room for
separately designed lifecycle hooks without defining them now. Version 1 schema
strictness rejects unknown properties at the hooks and entry levels.

A list of shell strings was rejected because it makes quoting platform-dependent
and creates an implicit host- or guest-shell contract. A bare array of arrays was
rejected because it is less self-describing and harder to extend compatibly.

The declaration adapter decodes ordered vectors, the domain configuration model
retains them through validation, and the lifecycle plan owns a defensive copy so
later orchestration cannot accidentally mutate validated input. Empty vectors and
empty arguments are structural errors; no additional semantic restriction is
placed on executable names or arguments.

### Add a narrow sandbox command-execution capability

Extend the existing `ports/sandbox` package with a command execution request
containing sandbox name, workspace, and an argument vector, plus an interface for
executing it with attached streams. The outbound adapter maps that request to
Docker Sandboxes' exec operation and passes the vector directly, without a host
shell. The adapter remains responsible for the exact Docker CLI syntax and for
preserving the attached process exit result.

Reusing the interactive agent attachment request was rejected because attachment
selects an agent while setup executes an arbitrary process with a working
directory. Exposing subprocess primitives to the application was rejected because
only the outbound adapter may construct Docker commands.

### Run initialization only on sandbox creation

Missing and recreated sandboxes complete creation and network-policy application,
then run `hooks.initialize` before agent attachment. Existing running and stopped
sandboxes retain their current direct-attachment behavior, so changing hooks does
not reconcile an existing sandbox and requires `up --recreate`. The phase iterates
in declaration order and passes the CLI's standard output and standard error
through to each sandbox command.

Running hooks on every `up` was rejected because initialization belongs to the
same one-time provisioning boundary as kits and network creation inputs. A future
`beforeEnter` hook for ephemeral concerns has different reconciliation and
failure semantics and needs its own design. Running initialization after
attachment was rejected because it does not establish a readiness boundary.

### Roll back incomplete sandbox creation

The first failed exec returns a contextual error that includes the one-based
command position and a safely rendered argument vector, stops iteration, and
prevents agent attachment. It then reuses lifecycle removal orchestration to
remove the sandbox created by this invocation and its currently declared scoped
network resources. If cleanup fails, the returned error reports both failures and
preserves the initialization error for unwrapping as the primary cause.

Retaining the failed sandbox was rejected because reliably blocking later entry
would require durable initialization state tied to that exact sandbox instance,
workspace, and declaration, and direct `sbx run` could bypass it. Rollback keeps
the invariant that a sandbox successfully left by `up` completed initialization.
It cannot roll back host-workspace changes, so commands must tolerate retry after
partial execution. Continuing later commands was rejected because ordering
expresses dependencies and could hide the primary failure.

### Make execution attached, non-interactive, and cancellable

Initialization attaches stdout and stderr to the matching CLI streams but uses
an empty stdin rather than the command input stream. This prevents hooks from
consuming recreation confirmation input or unexpectedly prompting before agent
attachment. Execution uses the `up` context, so interruption cancels the active
Docker exec and triggers the same rollback path as another command failure.

No fixed wall-clock timeout is added in this change: repository setup duration is
project-dependent, while the motivating readiness command already expresses a
bounded wait. Users can declare an explicit timeout utility or bounded shell loop.
A future declarative timeout can be added without changing argument-vector
semantics.

### Keep setup out of validation and teardown projections

Complete validation checks the hook shape and carries valid commands into the
ordinary `up` plan, but never invokes the execution capability. The narrow
teardown loader continues resolving only data required by `down` and `destroy`;
it need not retain initialization commands. This prevents malformed unrelated setup
content from broadening teardown's established tolerance while guaranteeing that
`up` rejects it before Docker inspection or mutation.

## Risks / Trade-offs

- **[A command hangs until cancellation]** → Document the absence of an implicit
  timeout and show bounded polling in the example; cancellation still invokes
  rollback.
- **[A partially completed command changes the host workspace]** → Document that
  rollback covers sandbox-local state only and require initialization tasks to
  tolerate retry.
- **[Rendering a failed vector is confusing or leaks quoting semantics]** →
  identify its stable declaration index and render arguments for diagnostics only;
  never reuse rendered text for execution.
- **[Cleanup fails after initialization fails]** → report both errors, preserve
  initialization as primary, and identify any remaining sandbox or resource for
  manual diagnosis.
- **[Setup requires interactive input]** → keep hooks non-interactive in this
  contract and reserve standard input for confirmation and agent attachment.

## Migration Plan

The field is optional, so existing version 1 declarations require no changes.
Adopters add ordered `hooks.initialize` entries and recreate an existing sandbox
when they need those commands applied. Rollback consists of removing
`sandbox.hooks`; no state migration is required.

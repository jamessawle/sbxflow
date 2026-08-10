## Context

See `proposal.md` for the motivation. The existing `down` implementation
already provides the relevant teardown boundaries: a target-only resolver for
the nearest declaration and exact sandbox name, a bounded `sbx ls --quiet`
lookup, and an attached subprocess runner that preserves Docker output and
process results. Unlike `down`, removal is interactive by default and therefore
must also receive the command's standard input.

Docker's removal interface combines two behaviors under `--force`: skipping the
confirmation prompt and permitting deletion while a sandbox has an active
session. sbxflow should expose that behavior transparently without creating a
second confirmation protocol or broadening repository-scoped targeting.

## Goals / Non-Goals

**Goals:**

- Reuse the existing safe lifecycle target and exact-name lookup boundaries.
- Keep absent-target destruction idempotent and permanent removal explicit.
- Preserve Docker's confirmation, active-session protection, output, and
  process result.
- Keep lifecycle orchestration injectable and testable without a real sandbox.

**Non-Goals:**

- Accept arbitrary sandbox names or expose Docker's remove-all operation.
- Reimplement Docker's confirmation prompt or infer confirmation from terminal
  capabilities.
- Stop a sandbox before removal, retain snapshots, or provide undo.
- Reconcile configuration, inspect sandbox contents, or run complete validation.
- Change the shared lifecycle-target loading contract introduced for `down`.

## Decisions

### Reuse lifecycle-target resolution and exact-name lookup

The destroy runner will depend on the existing `TargetResolver`, locate `sbx`,
and call the shared exact-name lookup before attempting removal. An absent exact
name is already in the requested state and returns success; similar names are
ignored.

Directly invoking removal and interpreting a not-found diagnostic was rejected
because Docker's human-readable errors are not a stable idempotency contract.
Running complete validation was rejected because agent, kit, local-path, and
trust configuration do not affect which existing sandbox is removed.

### Delegate confirmation and force semantics to Docker

Default destruction will invoke `sbx rm <name>`. Forced destruction will invoke
`sbx rm --force <name>`. The Cobra command will expose only `--force` and `-f`
and pass the selected mode to the lifecycle runner; it will never expose
`--all` or positional sandbox operands.

An sbxflow-owned confirmation was rejected because invoking unforced removal
afterward would prompt twice, while invoking forced removal after an sbxflow
prompt would also bypass Docker's active-session protection. Giving sbxflow's
flag a narrower meaning than Docker's flag was rejected because it would
require another unsupported removal protocol.

### Attach all three command streams to removal

The removal subprocess will use the existing no-shell attached runner with
standard input, output, and error connected. This allows Docker to render and
read its default prompt, keeps cancellation attached to the command context,
and avoids an artificial lifecycle timeout. An attached failure will use the
existing rendered-error marker so Cobra does not duplicate Docker diagnostics.

Using the captured lookup runner for removal was rejected because it cannot
support an interactive prompt and would buffer output behind the short lookup
timeout. The current `down` CLI intentionally omits stdin; `destroy` must pass
`cmd.InOrStdin()` explicitly.

### Remove directly without stopping first

The destroy runner will not compose or invoke the down runner. Docker's removal
operation already stops running sandboxes, and a separate stop would add a
second state-changing command, new partial-failure states, and unnecessary
latency before confirmation.

### Register a separately injected destroy runner

The CLI package will define a narrow destroy-runner interface and add it to root
construction beside the current lifecycle dependencies. This follows the
existing command test pattern and keeps help, argument parsing, force selection,
and stream forwarding testable without Docker.

## Risks / Trade-offs

- **Listing and removal are not atomic** → Preserve a subsequent Docker failure
  rather than retrying or converting it to success.
- **Force couples prompt bypass with active-session deletion** → Describe both
  effects in help and forward the option only after explicit user selection.
- **Default removal depends on usable input** → Attach the caller's input and
  document `--force` as the intentional non-interactive interface; preserve
  Docker's failure on EOF rather than guessing consent.
- **Docker owns some host-side managed worktrees** → Document that the user's
  repository workspace is preserved while Docker-managed resources follow
  Docker's removal behavior.
- **Root constructor gains another dependency** → Update all CLI test helpers
  together and retain narrow fake runners to keep failures localized.

## Migration Plan

Add the destroy runner and command without changing existing lifecycle behavior,
then update root wiring, tests, specifications, and README documentation.
Rollback consists of unregistering and removing the new command; the
configuration format and existing sandboxes require no migration.

## Context

See `proposal.md` for motivation and the delta specs for the behavioral
contract. The existing `up` workflow performs complete validation, builds a
creation plan, looks up the exact declared sandbox name, and asks the Sandbox
runner either to create or enter it. The same Sandbox port and outbound adapter
already expose forced removal for `destroy --force`.

The design must preserve the package relationships in `docs/structure.md`: the
CLI adapter translates the flag, the lifecycle application package owns branch
ordering, the Sandbox port carries operation inputs, and the `sbx` adapter
constructs subprocess invocations.

## Goals / Non-Goals

**Goals:**

- Represent recreation explicitly at the CLI-to-application boundary.
- Reuse exact-name lookup, forced removal, creation, trust, and attached-stream
  behavior already available through the Sandbox boundary.
- Keep complete validation ahead of every destructive or interactive operation.
- Make every failure boundary deterministic and testable without a real
  sandbox.

**Non-Goals:**

- Reconcile or compare an existing sandbox with the declaration before deciding
  whether to replace it.
- Add confirmation, a short flag, sandbox-name selection, automatic retries, or
  race recovery.
- Make removal and replacement creation transactional.
- Change `destroy`, the configuration schema, or the architecture relationship
  matrix.

## Decisions

### Pass recreation as an application request option

The `up` command will bind a long-form `--recreate` boolean and pass it in a
lifecycle-owned options value to `UpRunner.Run`. The application request type
keeps Cobra concepts out of the workflow and is clearer to extend than adding
an unlabeled boolean parameter.

The CLI help will explicitly say that an existing sandbox and its persisted
state are force-removed. No `-r` shorthand will be registered. Leaving the
option false preserves the current `up` call path.

An alternative was to implement `up --recreate` in the CLI by invoking destroy
and up workflows sequentially. That would put lifecycle coordination in the
inbound adapter, use destroy's partial target resolution rather than up's
complete validation, and require one application workflow to depend on another
or duplicate error rendering.

### Compose lookup, forced removal, and run inside `UpRunner`

`UpRunner` will extend its existing Sandbox dependency with the `Remover` port.
After validation and plan creation, it will perform the existing exact-name
lookup. Its branches will be:

1. Absent name: issue the existing create-and-enter request, regardless of the
   option.
2. Existing name without recreation: issue the existing enter request.
3. Existing name with recreation: call `RemoveSandbox` with `Force: true` and
   the command streams, then issue a create-and-enter request with `Exists:
false` only after removal succeeds.

The already validated plan remains the source for the replacement's name,
agent, workspace, kits, and process-local trust. The workflow will not use the
lighter lifecycle target resolver because recreation immediately consumes the
full declaration to create the replacement.

An alternative was to add a combined recreate operation to the Sandbox port.
Keeping the orchestration in the application layer avoids baking a multi-step
sbxflow use case into the external-process adapter and permits focused tests at
each failure boundary.

### Preserve attached Docker diagnostics and fail closed

Forced removal will use the same attached streams and `AttachedProcessError`
handling as destruction and agent execution, so Docker diagnostics are visible
once and the CLI does not append a duplicate error. A failed lookup stops before
removal or execution; a failed removal stops before creation; a failed creation
leaves the successfully removed sandbox absent and returns Docker's status.

No retry will reinterpret a removal failure as an absent sandbox. This makes
the destructive sequence predictable and avoids hiding concurrent lifecycle
changes.

### Reuse existing ports and adapter behavior

No new production package or Sandbox capability is needed. The entrypoint will
narrow the lifecycle client to a composite of `Lookup`, `Runner`, and `Remover`
for `UpRunner`. The existing outbound adapter already maps a forced
`RemoveRequest` to `sbx rm --force <declared-name>` and attaches all three
streams.

This retains the current dependency direction and keeps subprocess construction
in the sole outbound `sbx` adapter.

## Risks / Trade-offs

- **Forced recreation terminates active sessions and permanently loses sandbox
  state** -> Require an explicit destructive flag and state the effect in
  command help and README documentation.
- **Removal can succeed while replacement creation fails** -> Preserve Docker's
  diagnostics and non-zero result; document the sequence as non-transactional
  so a later ordinary `up` can create the missing sandbox.
- **The sandbox can change between lookup and removal** -> Fail the removal and
  stop; do not retry or broaden the exact-name target.
- **Adding removal to `UpRunner` expands its dependency surface** -> Depend only
  on the existing narrow `Remover` capability and cover call ordering with
  application fakes.

## Migration Plan

1. Extend the internal up request and dependency wiring, then add lifecycle and
   CLI tests before exposing the option in documentation.
2. Update README guidance that currently says `up` accepts no command-specific
   flags and directs users to remove a sandbox separately.
3. Run formatting, architecture checks, unit and executable tests, builds, and
   strict OpenSpec validation through the repository's Mise tasks.

Rollback removes the flag and request option while leaving the reused Sandbox
removal capability intact for `destroy`; no user data or configuration migration
is required.

## Context

See `proposal.md` for motivation. The current `up` workflow validates the full
declaration, uses `sbx ls --quiet` for exact-name existence, and immediately
issues attached `sbx rm --force` when recreation is requested. Quiet listing
does not expose lifecycle state. The architecture requires lifecycle decisions
to remain in `internal/application/lifecycle`, subprocess construction in the
outbound `sbx` adapter, and terminal interaction in the inbound CLI adapter.

The supported Docker Sandboxes versions provide machine-readable sandbox
listing data from which the adapter can extract exact-name lifecycle state.

## Goals / Non-Goals

**Goals:**

- Give the lifecycle workflow a narrow absent, stopped, or running result for
  the exact declared sandbox.
- Keep state inspection captured and bounded while confirmation and removal
  remain attached to command streams.
- Make every unknown state or unavailable confirmation fail closed.
- Preserve existing stopped-sandbox convenience and post-removal creation.

**Non-Goals:**

- Add a confirmation-bypass flag, change `destroy`, or generalize sandbox status
  into a user-facing command.
- Detect how many sessions are attached or identify their terminals.
- Make inspection and removal atomic or make replacement transactional.
- Change package classifications or broaden the architecture matrix.

## Decisions

### Replace existence lookup for `up` with narrow lifecycle-state inspection

Extend the Sandbox port with an `InspectSandbox`-style capability returning a
small port-owned state value: absent, stopped, or running. The outbound adapter
will invoke Docker's machine-readable sandbox listing, select only an exact name,
and reject missing, duplicate, or unrecognized state data. The lifecycle runner
will use this single inspection for all `up` branches.

Keeping the normalized result in the port prevents Docker's larger JSON schema
from leaking into the application. Reusing `SandboxExists` and adding a second
query only during recreation was rejected because it adds latency and creates
two potentially inconsistent observations. Parsing human-readable listing was
rejected as unstable.

### Inject a narrow confirmer into the lifecycle request

The lifecycle runner will own the decision to request confirmation only for the
running-plus-recreate branch. It will call a small confirmation interface whose
CLI implementation writes the warning and default-negative prompt to the error
stream and reads one response from the input stream. Only a documented,
case-insensitive affirmative response will return approval; EOF and read errors
will return a cancellation or confirmation error.

Putting prompt parsing directly in the application was rejected because it
would mix presentation policy with orchestration. Prompting entirely in Cobra
before calling the runner was rejected because the CLI does not know the
validated sandbox name or inspected state and would duplicate lifecycle work.
Using Docker's unforced removal prompt was rejected because Docker may protect
an active session by refusing removal even after the user approves recreation;
the approved operation still needs `--force`.

### Treat every non-approval as a safe, non-zero stop

Declining, submitting an empty or unrecognized response, reaching EOF, or
encountering a read error will stop before removal. The workflow will return a
distinct non-attached error so the CLI reports why no action occurred. No
`--yes`, `--force`, environment variable, or implicit non-interactive approval
will be introduced in this change.

A successful no-op exit was considered for a declined prompt, but a non-zero
result better signals to scripts and callers that the requested recreation did
not happen.

### Preserve Docker authority across the inspection-removal race

After confirmation, the workflow will issue the same exact-name forced removal
and will not re-inspect or retry. If state changes between inspection and
removal, Docker's result is preserved. Re-inspection cannot eliminate the race
and could create confusing repeated prompts.

## Risks / Trade-offs

- **Docker listing JSON changes within the supported version range** → Add
  fixtures for supported shapes, reject ambiguous data, and fail before any
  mutation.
- **A sandbox becomes running after it was observed stopped** → Docker remains
  authoritative; document the unavoidable race and preserve removal failures.
- **A user approves without knowing which other session is active** → Name the
  sandbox and consequence clearly; session enumeration is out of scope.
- **Scripts that relied on unconditional recreation of running sandboxes now
  fail without input** → Treat this as the intended safety change and document
  it; do not add an implicit bypass.

## Migration Plan

1. Add state parsing and the narrow Sandbox port contract with adapter fixtures.
2. Add lifecycle branching and the CLI confirmer, then update wiring and tests.
3. Align help, README, and OpenSpec behavior documentation and run the full Mise
   validation workflow.

Rollback can restore existence-only lookup and unconditional recreation. No
repository configuration or persisted data migration is required.

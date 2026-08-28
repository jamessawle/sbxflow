## Context

See [proposal.md](proposal.md) for the motivation. The lifecycle application
currently sends separate create, run, execute, and remove request shapes through
`internal/ports/sandbox`; the outbound SBX adapter translates those requests
into individual `sbx create`, `sbx run`, `sbx exec`, and `sbx rm` invocations.
The application also applies standalone network policy between creation and
agent entry and performs explicit per-resource cleanup after removal.

SBX 0.39 environments overlap with that adapter but do not replace two sbxflow
controls:

- standalone allowed hosts still require `sbx policy allow network --sandbox`;
- non-default and local kit sources still require process-scoped trust because
  the environment schema has no allowed-source field.

The spike established that environment creation, entry, execution, and removal
work with full repository kits; standalone and kit-derived network rules
coexist; environment removal cleans up both rule sources; process-scoped trust
does not change persistent settings; and removal identifies an environment by
name even when other document fields differ. It also found that absent removal
exits successfully while printing an error-looking diagnostic, and that the
repository's current `jamessawle/sbx-kits@v2026.08.02` pin is incompatible with
SBX 0.39 while `v2026.08.03` works.

This remains an outbound integration change. New production packages must not
be introduced, and the dependency directions in `docs/structure.md` remain
unchanged.

## Goals / Non-Goals

**Goals:**

- Keep application lifecycle policy independent from the SBX command syntax.
- Render environment documents privately and remove them on every completion
  path.
- Preserve the existing ordering, stream, trust, network, initialization,
  recreation, rollback, and identity-only teardown behavior described by the
  lifecycle spec.
- Contain the experimental SBX 0.39 interface behind the outbound adapter so it
  can be revised without changing `sbxflow.yaml`.

**Non-Goals:**

- Exposing `.sbxenv.yaml` as a second public declaration format.
- Replacing standalone `sandbox.network.allowedHosts` with generated kits.
- Adding environment credentials, ports, resources, MCP configuration, or kit
  authoring features.
- Changing `down`, inspection, doctor diagnostics other than compatibility
  bounds, or organisation-managed policy behavior.
- Supporting both the old individual lifecycle commands and the environment
  backend simultaneously.

## Decisions

### Keep `sbxflow.yaml` public and render private environment documents

The outbound SBX adapter will translate a lifecycle request into a
schema-version-1 environment document containing the exact sandbox name,
declared agent, absolute workspace, and linked kit references in declaration
order. It will create a private temporary directory outside the declared
workspace, write `.sbxenv.yaml` with owner-only permissions, pass its explicit
path to SBX, and remove the whole temporary directory with deferred cleanup.
Rendering and cleanup belong in `internal/adapters/outbound/sbx`, not the
application layer.

The renderer will reject a temporary location that resolves within the mounted
workspace. Removal will use a separate identity document containing the exact
name, a valid inert agent, and a private empty workspace adjacent to (not
containing) the document. It will not include kits or require complete
configuration validation. This preserves `destroy` when unrelated declaration
fields are invalid and relies only on the name-based removal behavior proven by
the spike.

Alternatives considered:

- Treating `.sbxenv.yaml` as repository configuration would expose an
  overlapping source of truth and still could not express allowed sources or
  standalone networks.
- Placing generated material in the workspace would expose implementation
  state to the sandbox and risk accidental persistence.
- Reusing the full declaration for removal would regress identity-only
  teardown by resolving fields that removal does not need.

### Share one environment definition across lifecycle port requests

The sandbox port will define a backend-neutral environment value containing
name, agent, workspace, ordered kits, and derived trust. Create, run, and
command requests will carry that value; command requests will additionally
carry the literal argument vector. This gives every environment command the
same definition and trust without exposing YAML or filesystem details across
the port. Removal remains name, force, and streams because its identity
document is an adapter concern.

The application will build this value from its existing validated `Plan` and
continue to orchestrate creation, standalone network application,
initialization, and entry separately. Port interfaces remain capability-sized;
no new package or architecture type is required.

An alternative was to let the adapter cache the last rendered path between
calls. That would introduce hidden mutable lifecycle state, complicate cleanup,
and fail across separate command invocations. Each adapter call will instead
render its own short-lived document.

### Map lifecycle capabilities to SBX environment commands

The adapter mapping will be:

| Port capability        | SBX operation                           |
| ---------------------- | --------------------------------------- |
| Create                 | `sbx env create <file>`                 |
| Interactive entry      | `sbx env run <file>`                    |
| Initialization command | `sbx env exec <file> -- <arguments...>` |
| Removal and rollback   | `sbx env rm [--force] <file>`           |

Inspection remains `sbx ls --json`, stopping remains `sbx stop`, and standalone
network policy remains the existing `sbx policy` operation. Keeping creation
separate from entry leaves the existing network and initialization ordering
intact. `env exec` will use the environment's declared workspace default, keep
stdin detached, forward stdout and stderr, and append the initialization vector
after `--` without shell reconstruction.

The old raw lifecycle commands will not remain as a fallback. A dual backend
would multiply behavioral tests and obscure which upstream resource model owns
cleanup.

### Apply kit trust to every operation that can resolve kits

The adapter will construct `DOCKER_SANDBOXES_KIT_ALLOWED_SOURCES` and
`DOCKER_SANDBOXES_KIT_ALLOW_LOCAL` from the validated plan and attach them only
to create, run, and exec subprocesses. This is intentionally broader than only
creation because any environment command may parse the full document. The
identity-only removal document has no kits and needs no trust override.

Persistent SBX settings will not be changed. Relying on user-global settings
was rejected because it would weaken repository-specific least privilege and
would make behavior depend on host state.

### Let environment removal own sandbox-scoped resource cleanup

The shared removal path will invoke environment removal and stop issuing
per-resource network removal calls afterwards. The spike showed that SBX
removes both kit-derived and separately applied sandbox-scoped policies with the
environment. This also avoids errors caused by trying to remove resources that
SBX already discarded.

Lifecycle state inspection stays in front of `destroy` and recreation. Besides
preserving exact-name and running-session safeguards, this keeps absent removal
silent despite SBX 0.39 emitting `sandbox not found` while returning success.
Rollback will continue to use a cancellation-independent context and report
both the primary failure and a failed removal.

Keeping raw `sbx rm` or manual policy cleanup was rejected because it splits
ownership between the old and new resource models and duplicates upstream
cleanup.

### Pin the experimental compatibility window

Doctor will accept SBX versions from 0.39.0 inclusive to 0.40.0 exclusive.
Unlike a general minimum-version policy, the upper bound is deliberate because
the environment interface is experimental and has already changed across
pre-release lines. The repository declaration will move all
`jamessawle/sbx-kits` references from `v2026.08.02` to `v2026.08.03` in the same
change so the checked-in example remains runnable.

These are compatibility constants, tests, and documentation updates under the
existing diagnostics requirement; they do not add a new user-visible
capability.

## Risks / Trade-offs

- [SBX changes the experimental environment schema or command behavior] → Keep
  parsing and invocation construction isolated in the outbound adapter, pin the
  0.39.x range, and cover exact command and document boundaries with executable
  tests.
- [Temporary environment material leaks configuration] → Use owner-only files
  in an owner-only directory outside the workspace and test cleanup after
  success, failure, and cancellation.
- [A host redirects its temporary directory into the workspace] → Compare
  canonical locations before invocation, clean up, and fail rather than mount
  the generated document.
- [Environment entry reconciles fields differently in a later SBX release] →
  Retain sbxflow's no-explicit-reconciliation contract and re-spike before
  widening the compatibility window.
- [Environment removal reports success without complete resource cleanup] →
  Treat SBX as the resource owner for 0.39.x and cover mixed policy cleanup in
  executable integration tests.
- [Dropping the old backend temporarily narrows compatible SBX versions] →
  Make the 0.39.x requirement prominent in doctor and README output and ship the
  kit pin update atomically.

## Migration Plan

1. Introduce the shared backend-neutral environment value in the sandbox port
   and update lifecycle planning and fakes to pass it through create, run, and
   initialization requests.
2. Add private environment rendering and cleanup in the SBX adapter, then map
   create, run, exec, and remove to their environment commands.
3. Remove explicit post-removal network cleanup while retaining standalone
   policy application, exact-name inspection, rollback, and force safeguards.
4. Update unit and executable tests for document contents, permissions,
   cleanup, argument fidelity, trust, mixed policies, and identity-only
   teardown.
5. Change doctor bounds to 0.39.x, update the repository kit pin to
   `v2026.08.03`, and align README and examples.
6. Run repository validation and repeat the disposable full-kit smoke test on
   SBX 0.39 before release.

Rollback is a source revert of the adapter, port request changes, compatibility
bounds, documentation, and kit pin. No user configuration or persisted sbxflow
state needs migration because `sbxflow.yaml` is unchanged and SBX owns the
sandbox resources.

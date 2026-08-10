## Context

See `proposal.md` for the motivation. The current CLI has a repository
validation pipeline that returns the declaration path, linked configuration,
derived trust settings, and validated local-kit targets. Its shared subprocess
runner is intentionally bounded and captures output, which suits diagnostics
and local-kit validation but cannot host an interactive agent session.

Docker Sandboxes already gives `sbx run` the lifecycle semantics required here:
it creates a missing named sandbox when given creation inputs, attaches to a
running named sandbox, and restarts a stopped named sandbox before attaching.
Kit arguments are creation-only, so sbxflow must determine existence before it
can construct the correct `run` invocation.

## Goals / Non-Goals

**Goals:**

- Reuse the complete validation and linking boundary before lifecycle work.
- Construct one ordered set of Docker kit references for initial creation.
- Delegate creation, restart, agent verification, and attachment to `sbx run`.
- Preserve terminal interactivity and apply least-privilege trust only to the
  child process.
- Keep command orchestration injectable and testable without a real sandbox.

**Non-Goals:**

- Inspect or compare an existing sandbox's workspace, kits, or creation spec.
- Add fingerprints, ownership records, state files, or content hashing.
- Add, remove, reorder, or refresh kits on an existing sandbox.
- Recreate a sandbox or recover automatically from a concurrent creation race.
- Forward agent-specific arguments or add `up` flags in the initial interface.
- Implement `down` or `destroy`.

## Decisions

### Validate before inspecting sandbox existence

`up` will run the existing validation pipeline from the current working
directory and stop on any reported error. On success, the declaration directory
becomes the workspace used for creation. This keeps lifecycle commands from
acting on malformed configuration and preserves the nearest-declaration behavior
already established by `validate`.

Allowing an existing sandbox to bypass validation was considered because its
kits will not be reconciled, but it would make `up` behave differently depending
on external state and permit an invalid repository contract to launch an agent.

### Use quiet name listing only to choose the Docker run shape

After validation, sbxflow will invoke `sbx ls --quiet` through a captured,
bounded subprocess and compare each complete output line with the declared name.
The result selects one of two invocations:

```text
missing:
  sbx run --name <name> --kit <ref>... <agent> <repository-root>

existing:
  sbx run <agent> --name <name>
```

The existing path deliberately supplies the declared agent so Docker verifies
it, but omits workspace and kits because those are creation inputs. It does not
inspect status: the same `run` operation attaches when running and restarts when
stopped.

Using `sbx inspect --json` was considered but would couple this first pass to a
larger experimental schema while encouraging configuration comparison that is
out of scope. Blindly passing kit flags on every call was rejected because
Docker accepts them only while creating a sandbox.

### Build execution kit references from validated results

The lifecycle planner will walk the linked selections in declaration order.
Git and OCI selections use their existing `RemoteReference`; local selections
are matched by selection index to the canonical absolute paths already returned
by local-kit validation. No remote materialization or second local path
resolution occurs before `sbx run`.

Keeping the selection index as the join key preserves duplicates that select
different configuration objects and avoids relying on source or kit names being
globally unique.

### Add a separate interactive subprocess boundary

The captured `command.Runner` will remain the interface for bounded inspection
and validation. Lifecycle attachment will use a narrow interactive runner that
accepts arguments, environment overrides, and the CLI streams; connects those
streams directly to an `os/exec` child; and has no command timeout. Separating
the modes prevents interactive concerns from weakening the timeout and output
contracts used by doctor and validation.

The interactive runner will inherit the host environment and replace only
`DOCKER_SANDBOXES_KIT_ALLOWED_SOURCES` and
`DOCKER_SANDBOXES_KIT_ALLOW_LOCAL` with values derived from the linked
configuration. It will invoke `sbx` directly without a shell. The returned
process result becomes the `up` result, while Docker's own diagnostic output is
already present on the attached error stream and must not be duplicated by an
additional rendered subprocess error.

### Treat existing configuration as Docker-owned state

For an existing name, sbxflow will not claim that the sandbox is current or up
to date. Changes to workspace or kits will not be detected or applied. This is
an intentional first-pass boundary: the command's promise is to enter the named
sandbox, not to converge it on the declaration.

Adding partial comparisons was considered but would create an ambiguous middle
state where some changes are enforced while others are silently ignored. A
future reconciliation change can define fingerprints, ownership, active-session
safety, and destructive recovery together.

## Risks / Trade-offs

- **A changed declaration does not update an existing sandbox** -> State the
  limitation in command help and README documentation; users can remove the
  sandbox explicitly before running `up` to create it again.
- **Name listing and creation have a race** -> Let Docker reject the conflicting
  create attempt and preserve its diagnostic; a later invocation will take the
  existing path.
- **Quiet listing output could change across pre-1.0 Docker releases** -> Keep
  the parser to exact non-empty lines and cover the supported version contract
  with controlled process tests.
- **Interactive execution can regress signal or stream behavior** -> Test the
  execution boundary with a fake `sbx` process that reads stdin, writes both
  output streams, and returns controlled exit statuses.
- **Docker output may be followed by a redundant wrapper error** -> Mark
  downstream process failures as already rendered while preserving a non-zero
  command result.

## Migration Plan

No sandbox data migration is required. Register `up`, update README lifecycle
documentation, and ship the command behind the existing versioned executable.
Rollback removes the command registration and documentation while leaving any
sandbox created through Docker untouched.

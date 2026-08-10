## Context

See `proposal.md` for the motivation. The lifecycle runner currently completes
validation before it builds a plan or invokes Docker, while the CLI receives
the validation report only after the interactive session returns. The lifecycle
runner already owns the terminal streams passed to the attached Docker process.

## Goals / Non-Goals

**Goals:**

- Make successful startup validation visible at the moment it completes.
- Preserve the ordering between validation, the status message, and Docker
  inspection.
- Keep agent standard output free of sbxflow status text.

**Non-Goals:**

- Add verbose, quiet, or output-format flags.
- Add progress messages for sandbox lookup, creation, or attachment.
- Change the standalone validation report or any validation semantics.

## Decisions

### Render the status inside lifecycle orchestration

The lifecycle runner will write the status immediately after `Report.Valid()`
succeeds and before plan construction or `sbx` lookup. This is the only layer
that has both the completed report and the correct timing relative to the
interactive session.

Rendering in the Cobra command after `Run` returns was considered, but that
would display the message only after the agent session ends. Running validation
separately in the Cobra layer was rejected because it would duplicate work and
weaken the lifecycle runner's validation boundary.

### Write a plain path-bearing line to standard error

The status will be `Configuration valid: <declaration path>` with a trailing
newline. The declaration path confirms which nearest configuration was used,
and plain text remains understandable without terminal styling. Standard error
is appropriate for lifecycle status so the attached agent's standard output is
not polluted.

The shorter `Configuration file valid` message was considered, but it does not
identify the discovered file. Standard output was considered, but it would mix
sbxflow lifecycle status with the agent's primary output stream.

## Risks / Trade-offs

- **The extra line changes successful `up` output** -> Document it and cover
  stream and ordering behavior in CLI, lifecycle, and executable-level tests.
- **A later Docker failure follows a success message** -> Keep the wording
  scoped specifically to configuration validation so it cannot imply that the
  whole `up` operation succeeded.

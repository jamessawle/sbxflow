## Context

See `proposal.md` for motivation. The inbound CLI adapter reads confirmation one
byte at a time so it can stop at a newline without consuming later input. Its
current reader returns every read error immediately, discarding bytes already
collected when a reader returns them before a subsequent `io.EOF`. Confirmation
normalization and the destructive-operation decision already happen separately
after line reading.

## Goals / Non-Goals

**Goals:**

- Preserve buffered confirmation text when EOF is its terminator.
- Keep all unavailable-input and non-affirmative paths fail-closed.
- Exercise the behavior both at the reader/CLI boundary and through the built
  executable's running-sandbox recreation path.

**Non-Goals:**

- Change accepted affirmative words, whitespace normalization, prompt text, or
  lifecycle ownership of the recreation decision.
- Change Docker Sandbox subprocess behavior or confirmation for `destroy`.
- Introduce a general-purpose line-reading package or alter architecture
  boundaries.

## Decisions

### Handle EOF according to whether response bytes were buffered

The existing byte-oriented reader will return its buffered response when it
encounters `io.EOF` after at least one byte. Immediate EOF will remain an error,
and every non-EOF error will still be returned even when bytes precede it. This
keeps EOF from implying consent and lets the existing trim, case-fold, and
explicit `y`/`yes` check decide completed responses.

Using a scanner or buffered reader was considered, but would be broader than
the fault and could change buffering or input-consumption behavior. Treating all
EOF as an empty line was rejected because it would conflate unavailable input
with a completed negative response and weaken error reporting.

### Cover both focused parsing and executable lifecycle behavior

Table-driven CLI tests will distinguish newline termination, EOF after
affirmative/negative/malformed bytes, immediate EOF, and a non-EOF failure after
buffered bytes. Executable-level cases will run the built command against the
fake `sbx` harness to prove that EOF-terminated affirmative input reaches
removal and replacement, while negative and unavailable input leave the sandbox
untouched.

Testing only `readConfirmationLine` was considered, but would not prove that
the response still passes through normalization and gates the destructive
lifecycle path correctly.

## Risks / Trade-offs

- [A custom reader can return bytes and an error in the same call] → Process the
  byte first, but authorize EOF termination only after confirming at least one
  byte was accumulated; preserve all non-EOF errors.
- [Executable tests can accidentally assert implementation details] → Assert
  observable exit status, diagnostics, and fake Docker Sandbox calls that map
  directly to the changed contract.

## Migration Plan

No data or configuration migration is needed. Deploy as a compatible bug fix.
Rollback restores the previous reader and tests, with no persisted-state
conversion required.

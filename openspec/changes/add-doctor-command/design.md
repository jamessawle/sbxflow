## Context

The executable currently contains a thin process entry point and a constructed
Cobra root command, with no domain commands or subprocess abstraction. See
`proposal.md` for the motivation and the environment-diagnostics and
cli-interface delta specs for the observable contract.

Docker Sandboxes already exposes machine-readable diagnostic, policy, and
settings commands. The design must use those interfaces without copying
Docker's individual health checks into sbxflow or coupling `doctor` to the
repository declaration that a future `validate` command will own.

## Goals / Non-Goals

**Goals:**

- Establish a small strategy-based diagnostic runner that can grow through
  independent checks.
- Keep command execution and Docker JSON parsing isolated from Cobra
  presentation.
- Produce deterministic, plain-text results and reliable exit status from
  injected, testable dependencies.
- Limit integration with Docker diagnostics to their versioned summary
  envelope.

**Non-Goals:**

- Load or validate `sbxflow.yaml`, resolve local kit paths, or assess whether
  global settings permit one declaration.
- Detect host platform, architecture, virtualization, or Docker Engine/Desktop
  state independently of Docker's diagnostics.
- Reproduce Docker diagnostic details, inspect agent credentials, initialize
  policy, alter settings, start services, or provide an automatic fix mode.
- Add structured sbxflow output or a general-purpose logging or presentation
  framework.

## Decisions

### Run four ordered diagnostic strategies

An internal doctor package will define a strategy contract that accepts context
and a shared environment and returns one result. Results have a stable
identifier, pass/warn/fail/skip status, concise summary, and optional guidance.
The initial ordered strategies are:

1. Compatibility: locate `sbx`, obtain its version, and compare it with
   sbxflow's centralized supported range.
2. Docker diagnostics: run `sbx diagnose --output json` and summarize its
   totals.
3. Network policy: inspect global policy initialization and management source.
4. Kit sources: inspect the effective `kit.allowedSources` value and source.

Later strategies can be registered without changing the runner or Cobra
command. A single monolithic doctor function was considered, but it would mix
prerequisites, Docker schemas, grading, and presentation. Modeling every
individual Docker diagnostic as a strategy was also rejected because it would
duplicate Docker's ownership and make sbxflow fragile.

### Use explicit prerequisites and deterministic sequential execution

The runner will execute strategies in registration order. Strategies declare
the facts they require, and the runner marks them skipped when a prerequisite
is unavailable. Compatibility supplies the executable and accepted-version
facts required by the other three strategies. Docker diagnostics, policy, and
settings then run independently because aggregate diagnostic totals do not
reveal whether any particular Docker check, including daemon health, passed.

Sequential execution keeps output stable and avoids concurrent `sbx` processes
contending for shared locks. A generic dependency graph and parallel scheduler
were considered unnecessary for four checks.

### Inject a subprocess runner and keep stdout and stderr separate

An internal command-runner interface will execute commands directly without a
shell, accept context cancellation, and return stdout, stderr, and exit status
separately. Production wiring uses `os/exec`; tests use deterministic fakes.
Each invocation receives a bounded timeout so an unhealthy daemon cannot hang
`doctor` indefinitely.

Separating streams is essential because `sbx` may emit warnings on stderr while
returning valid JSON on stdout. Combined output would corrupt otherwise usable
machine-readable results.

### Treat compatibility as sbxflow's concern

Docker diagnostics owns update availability and CLI/daemon version matching;
sbxflow separately owns whether the installed pre-1.0 CLI exposes the commands
and schemas this release supports. The supported lower and upper bounds will be
centralized and compared using the semantic components parsed from `sbx
version`. An unparseable or out-of-range version fails compatibility and skips
strategies whose schemas cannot be trusted.

The initial bounds will be established from the versions verified against all
four required commands during implementation and documented alongside the
constants. Depending only on Docker's update check was rejected because a
current binary can still be outside sbxflow's tested contract.

### Decode only the Docker diagnostic summary envelope

The diagnostic strategy will decode only the top-level schema version and the
summary counts for pass, warning, failure, and skipped checks. Unknown JSON
fields, including the individual `checks` collection, are ignored. A valid
summary is parsed even if the subprocess exits non-zero, because Docker may use
that status to represent diagnostic failures.

Any failed total makes this required strategy fail; warnings without failures
make it warn. An unsupported envelope version, malformed summary, or execution
failure produces one sbxflow failure with guidance to run `sbx diagnose`.
Passing through names, details, or remediation from Docker was rejected because
those fields can evolve independently and Docker already presents them.

### Inspect network policy without grading valid presets

The policy strategy will use Docker's machine-readable policy listing to decide
whether a global local policy or organisation-managed policy is active. An
active managed policy is informationally identified and accepted rather than
prompting for a conflicting local setup. An absent policy produces an advisory
warning; Open, Balanced, and Locked Down are not graded because their fitness
depends on the user's environment.

The policy strategy runs independently after compatibility. If Docker cannot
return machine-readable policy state, the result is an advisory warning that
policy could not be inspected. It does not infer daemon health from aggregate
diagnostic totals or turn a configuration-inspection failure into another
required health failure.

### Grade only the universally unsafe kit-source wildcard

The kit-source strategy will run `sbx settings get --json kit.allowedSources`
and decode only the evaluated value and its reported source. Any array member
exactly equal to `"*"` makes the result an advisory warning even when narrower
entries are also present; otherwise the result passes. The strategy does not
judge whether particular publisher prefixes are sufficient.

Guidance varies by source: local overrides can be unset or narrowed,
environment-provided values point to the corresponding environment variable,
and remotely managed values direct the user to their administrator. Query or
schema failures produce a warning because configuration posture could not be
verified, but do not make an otherwise healthy installation unusable.

### Render all results and fail only required health checks

The Cobra command will render ordered plain-text results to stdout. It will
continue independent strategies, include skipped checks, and return a non-zero
status when any result is a required failure. Advisory configuration warnings
retain a successful status. Command construction receives the doctor runner as
an injected dependency, while `Execute` wires the production runner.

This preserves the root command's existing stream and context conventions. A
`--fix` mode and a public JSON output flag were considered but deferred until
there is a demonstrated consumer and a stable sbxflow result schema.

### Remove the sole-subcommand help workaround

The root currently wraps Cobra's help renderer so the designated `help`
command remains visible when it is the only subcommand. Registering `doctor`
removes the condition that required this workaround. The custom help command
and its existing behavior will remain, but the `HelpFunc` wrapper that
temporarily removes and restores that command will be deleted so Cobra renders
the normal multi-command tree.

Leaving the wrapper in place was considered, but retaining stateful command
tree mutation after its motivating condition disappears would add unnecessary
complexity and risk as more commands are registered. Existing root-help and
contextual-help tests will continue to protect the public behavior.

## Risks / Trade-offs

- **Docker changes a machine-readable schema** -> Gate parsing by sbx version
  and envelope version, decode only required fields, and provide one actionable
  compatibility or summary failure instead of guessing.
- **A non-zero diagnostic process still contains valid totals** -> Parse stdout
  before interpreting process status and grade the known summary when valid.
- **Policy output differs under organisation governance** -> Test local,
  managed, absent, and inactive-policy fixtures; never recommend overriding an
  active managed source.
- **A wildcard is intentional for local experimentation** -> Keep it advisory
  and non-fatal while explaining the supply-chain trade-off.
- **Subprocess checks are slow or hang** -> Apply context-aware timeouts and run
  sequentially with concise failure guidance.
- **The supported version window needs frequent maintenance** -> Keep the range
  in one location, cover both boundaries with tests, and update it deliberately
  when Docker's CLI contract is verified.

## Migration Plan

This adds a new command and no persistent sbxflow state. Register the command,
remove the obsolete root help-rendering workaround, ship the strategy runner
and Docker adapter, update README documentation, and validate against supported
`sbx` fixtures. Rollback consists of unregistering the command, restoring the
sole-subcommand help workaround if `help` again becomes the only subcommand,
and removing the internal doctor package and documentation; Docker Sandboxes
state is never modified.

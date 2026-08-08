## Context

The repository currently contains the intended configuration and lifecycle
interface but no application code or application-language toolchain. See
`proposal.md` for the motivation and `specs/cli-interface/spec.md` for the
initial observable CLI contract.

The foundation must remain small while providing a command structure that can
accept `up`, `down`, `destroy`, `validate`, and `doctor` later. It also needs to
fit the repository's existing Mise-managed development and validation workflow.

## Goals / Non-Goals

**Goals:**

- Establish a testable Go executable with a thin process entry point.
- Centralize command construction so later commands share help, output, error,
  and context conventions.
- Use compact, conventional presentation that remains predictable in terminals
  and redirected output.
- Make development and release version output deterministic.

**Non-Goals:**

- Implement or stub any lifecycle, validation, or diagnostic commands.
- Add configuration loading, logging, telemetry, shell completion, man pages, or
  a release pipeline.
- Establish global application configuration through Viper.

## Decisions

### Use Go for the application and pin it with Mise

The repository will add a Go module and pin a supported Go toolchain in
`mise.toml`. Go provides a single native executable, straightforward process
execution, and strong testing support for a CLI that will orchestrate Docker
Sandbox commands.

Alternatives considered were TypeScript, which would require an additional
runtime or packaging decision, and Rust, whose additional implementation
complexity is not justified for this command-oriented application.

### Use Cobra directly for the command model and presentation

Cobra will own the root command, flags, argument validation, help and version
semantics, error presentation, and future command tree. Its compact default
presentation is predictable in interactive terminals and redirected output
without requiring terminal styling or an additional presentation dependency.

Cobra's default completion command will be disabled, and no man-page command
will be added. Both can be introduced later as intentional capabilities. A
presentation wrapper such as Fang can likewise be added around the independently
constructed Cobra command tree later without restructuring commands.

### Construct the command tree rather than storing global commands

An internal CLI package will expose a root-command constructor. It will accept
or otherwise allow injection of output streams and build metadata, while the
`cmd/sbxflow` package remains responsible only for process context, execution,
and translating errors into exit status.

This avoids package-global command state, permits isolated tests, and gives
future commands a consistent composition point. Direct construction in
`main()` was considered but would couple command behavior to process globals
and make tests more expensive.

### Show help when no action is selected

The root command will be a help-only entry point and will reject positional
arguments. It will support the `help` command from the first release even though
no domain subcommands exist yet; later Cobra commands can participate in the
same contextual help system.

No placeholder lifecycle commands will be registered. Advertising commands
that only report "not implemented" would conflict with the README's early-stage
warning and give users a misleading executable interface.

### Use injected build metadata with an explicit development fallback

Build metadata will have a deterministic development default and allow release
builds to inject version and commit values through Go linker flags. The values
will be passed to Cobra so both version flags share one source of truth.

Relying only on automatic Go build information was considered, but local and
direct binary builds do not always contain a useful module version. An explicit
development fallback makes local output and tests predictable without requiring
a release workflow in this change.

### Validate the CLI at command and process boundaries

Command-level tests will execute fresh root commands with controlled arguments
and buffered streams. A small built-binary test will cover actual exit status
and stdout/stderr separation where in-process Cobra execution is insufficient.
Assertions will focus on stable content and behavior rather than snapshots of
ANSI styling.

Mise tasks will format, vet, test, and build the Go module, and the repository's
aggregate validation task will include the relevant non-mutating checks.

## Risks / Trade-offs

- **Plain output may feel less distinctive** → Prefer conventional, compact
  output for the foundation and keep the command constructor independent so a
  presentation layer can be added later if it provides clear value.
- **Cobra reserves `-v` for version once version metadata is configured** →
  Treat `-v` as part of the public contract; any future verbosity setting must
  use another shorthand or only `--verbose`.
- **Help output will evolve as real commands arrive** → Test required sections
  and command visibility instead of freezing the entire rendered document.
- **Linker metadata can drift from release tooling** → Keep one build-metadata
  package and one documented injection path for future release automation.

## Migration Plan

This is a new executable with no installed predecessor or persistent state to
migrate. Implementation can be rolled back by removing the Go module,
executable packages, Go-related Mise configuration, and associated
documentation changes. A presentation wrapper can be introduced later around
the root-command constructor without changing the public command contract.

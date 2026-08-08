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
- Use attractive terminal-aware presentation without making color or styling
  necessary to understand output.
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

### Use Cobra for the command model and Fang for presentation

Cobra will own the root command, flags, argument validation, help semantics,
and future command tree. Fang will execute that command while supplying styled
help and errors and the `-v`/`--version` integration.

Fang will be configured without completions and without its hidden man-page
command. Those are useful future capabilities, but enabling them implicitly
would expand the initial interface and dependency behavior beyond the spec.

Using Cobra alone was considered and would be functionally sufficient, but
would require either accepting plain presentation or building presentation
customization. Other command frameworks were rejected because Cobra and Fang
together match both the mature command model and the desired polished output.

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
will be passed to Fang for rendering so both version flags share one source of
truth.

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

- **Fang describes itself as experimental** → Pin an exact compatible version,
  keep Cobra commands independent of Fang-specific presentation, and cover the
  public output contract with tests so Fang can be upgraded or replaced.
- **Styled output can make tests and redirected output brittle** → Assert
  semantic text and stream behavior, use controlled non-interactive output in
  tests, and avoid ANSI snapshots.
- **Fang reserves `-v` for version** → Treat `-v` as part of the public contract;
  any future verbosity setting must use another shorthand or only `--verbose`.
- **Help output will evolve as real commands arrive** → Test required sections
  and command visibility instead of freezing the entire rendered document.
- **Linker metadata can drift from release tooling** → Keep one build-metadata
  package and one documented injection path for future release automation.

## Migration Plan

This is a new executable with no installed predecessor or persistent state to
migrate. Implementation can be rolled back by removing the Go module,
executable packages, Go-related Mise configuration, and associated
documentation changes.

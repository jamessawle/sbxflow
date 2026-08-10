## Context

The production code currently follows the dependency direction described in
`docs/structure.md`, but no automated check prevents a future import from
crossing those boundaries. The repository uses Mise for pinned development
tools and aggregate validation, has a small and explicit package graph, and
keeps tooling dependencies out of the runtime module where possible. See
`proposal.md` for the motivation and scope.

Two documented constraints have different levels of mechanical enforceability:
direct import relationships can be checked reliably, while responsibility
statements such as keeping policy out of `cli` still require design review.

## Goals / Non-Goals

**Goals:**

- Fail validation when production Go code introduces an unapproved first-party
  package dependency.
- Detect Cobra imports from application packages.
- Fail when a production package is not covered by the architecture policy.
- Keep the check visibly separate from ordinary Go tests while making it
  mandatory in aggregate validation.
- Use a pinned, maintained enforcement engine and a small declarative policy
  rather than owning package loading and dependency analysis code.

**Non-Goals:**

- Classify business policy or presentation logic by inspecting function bodies.
- Restrict test-only dependencies used for fixtures, fakes, and integration
  coverage.
- Replace Go's compiler, `go vet`, or human architectural review.
- Add an application runtime dependency or a custom architecture test engine.
- Add external approval controls such as CODEOWNERS or branch protection before
  the project identifies the appropriate owners and policy.

## Decisions

### Use pinned Arch-Go as the enforcement engine

Use the Arch-Go v2 command-line tool, initially pinned to v2.1.2 through Mise,
and keep it out of the application's `go.mod`. Arch-Go owns module package
discovery, dependency analysis, package matching, coverage calculation, and
failure reporting. The repository owns only the architecture policy in
`arch-go.yml`.

Arch-Go is preferred over a custom `go/parser` or `go/packages` implementation
because those approaches would make the repository responsible for build
constraints, package discovery, test-package handling, matching semantics, and
diagnostics. `go-arch-lint` was also considered; its component model maps well
to the documented layers, but Arch-Go's explicit package coverage threshold is
a stronger fit for the requirement that new packages fail until classified.
Depguard would require adopting a broader lint runner for a narrower import
rule and would still need path-specific policy configuration.

Changing the pinned Arch-Go major or minor version is an architecture-tooling
change and must be reviewed together with any configuration compatibility
updates.

### Store the complete dependency policy in `arch-go.yml`

Set both Arch-Go thresholds to 100 percent:

- `compliance: 100` requires every evaluated dependency rule to pass.
- `coverage: 100` requires every production module package to be evaluated by
  at least one rule, so a newly introduced package cannot silently avoid the
  policy.

Define exact package rules for the current graph:

| Source package                                 | Allowed first-party dependencies                                            |
| ---------------------------------------------- | --------------------------------------------------------------------------- |
| `cmd/sbxflow`                                  | `internal/cli`, `internal/buildinfo`                                        |
| `internal/cli`                                 | the three current application packages and `internal/buildinfo`             |
| `internal/application/lifecycle`               | `internal/application/validation`, `internal/configuration`, `internal/sbx` |
| `internal/application/validation`              | `internal/configuration`, `internal/sbx`                                    |
| `internal/application/doctor`                  | `internal/sbx`                                                              |
| `internal/configuration`                       | `schema`                                                                    |
| `internal/sbx`, `internal/buildinfo`, `schema` | none                                                                        |

The rules use the full module package paths and direct dependency allowlists.
Application package rules also explicitly prohibit the external
`github.com/spf13/cobra` dependency. The semantic constraint that application
code must not construct raw `sbx` arguments remains a review concern;
identifier or string matching would be brittle.

Arch-Go should evaluate production module packages without applying dependency
rules to test-only imports. The current repository provides a useful acceptance
case: CLI tests import `internal/configuration`, while production `internal/cli`
does not and the policy forbids that production edge. The dedicated check must
pass this existing test-only relationship before the task is considered
complete.

### Run architecture validation as a dedicated mandatory task

Add `mise run test:architecture` as the direct entry point for Arch-Go. Keep the
ordinary `go test ./...` command focused on Go tests, then invoke both commands
from `mise run validate`:

```text
mise run validate
├── go test ./...
├── mise run test:architecture
├── go vet ./...
└── go build ./...
```

This gives architecture failures a clear command and output boundary without
making the policy optional in the normal contributor or CI workflow. A
separate CI job is not required initially because the existing required
`Validate` job invokes the aggregate task; it can be split later if independent
branch protection becomes useful.

### Keep documentation authoritative for architectural intent

Update `docs/structure.md` to show the existing `cmd/sbxflow` to `buildinfo`
and `configuration` to `schema` edges, distinguish Arch-Go-enforced import rules
from review-only responsibility rules, and document the dedicated command. The
declarative policy mirrors the graph, while the document continues to explain
why the boundaries exist.

### Protect architecture policy from incidental fixes

Update `AGENTS.md` to prohibit changing `arch-go.yml`, the pinned enforcement
tool, or the documented graph merely to make an unrelated change pass. When a
requested implementation conflicts with a boundary, an agent must stop and
request an explicit architecture decision. `CONTRIBUTING.md` will identify the
dedicated command and explain that changes to the governed graph require
deliberate architectural review.

The dedicated task and declarative configuration make policy changes visible,
but cannot by themselves distinguish an intentional decision from a weakened
rule. CODEOWNERS or a separately required CI job can add external approval
controls later.

## Risks / Trade-offs

- **[A third-party tool becomes part of repository validation]** → Pin Arch-Go
  through Mise, keep it out of the runtime module, and review upgrades as
  architecture-tooling changes.
- **[The graph is represented in documentation and configuration]** → Keep both
  representations small and require them to change in the same architecture
  decision.
- **[A legitimate new package fails the 100 percent coverage threshold]** →
  Treat this as the intended review point and classify the package explicitly.
- **[An agent could edit the policy to silence a failure]** → Add an explicit
  repository instruction to stop and request an architecture decision instead;
  add external ownership controls later if needed.
- **[The dedicated task could be omitted from an ad hoc test run]** → Keep it
  mandatory in `mise run validate` and document its direct command.
- **[Tool package loading may depend on the active build context]** → There are
  currently no platform-specific production files; expand the architecture
  task to a build-context matrix if such files are introduced.
- **[Semantic responsibility violations remain possible]** → Continue using
  code review for policy placement and raw adapter argument construction.

## Migration Plan

1. Pin Arch-Go v2.1.2 in Mise and add `arch-go.yml` with 100 percent compliance
   and coverage thresholds.
2. Encode the complete current internal dependency graph and Cobra restriction,
   then confirm the current production graph passes despite broader test-only
   imports.
3. Add `mise run test:architecture`, require it from aggregate validation, and
   update `docs/structure.md`, `CONTRIBUTING.md`, and `AGENTS.md`.
4. Run formatting and the full Mise validation task.

The current production graph already conforms, so no package refactor or
runtime migration is expected. Rollback consists of removing the pinned tool,
policy file, Mise task, and associated documentation.

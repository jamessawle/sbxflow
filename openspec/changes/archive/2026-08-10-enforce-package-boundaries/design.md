## Context

The implemented Arch-Go prototype checks exact outbound dependencies for the
current packages. Exploration exposed two problems with that approach:

- The policy describes package instances and their exceptions instead of a
  small set of architectural types and relationships.
- Arch-Go v2.1.2 treats an empty allowed-dependency list as no restriction and
  its configuration validator rejects rules that combine allowed and
  prohibited dependencies, even though its default check executes the current
  file.

The current code also mixes roles. `cli.Execute` performs dependency wiring,
application validation combines domain rules with filesystem and Docker I/O,
configuration combines its model and semantic rules with declaration adapters,
and lifecycle depends on the peer validation workflow. See `proposal.md` for
the motivation to make these boundaries reliable for contributors.

## Goals / Non-Goals

**Goals:**

- Give every production package exactly one architectural type derived from
  its import path.
- Enforce a small relationship matrix between types, including relationships
  involving packages added later.
- Make new unclassified packages and disallowed dependencies fail with source,
  target, and applicable-rule diagnostics.
- Keep configuration rules in the domain and make validation an application
  use case without changing product behavior.
- Introduce explicit, capability-scoped ports so adapters do not depend on
  domain or application packages, including a self-contained BuildInfo port.
- Keep architecture validation separate and mandatory in aggregate validation.

**Non-Goals:**

- Change any public command, configuration field, lifecycle behavior, or
  diagnostic result.
- Apply production dependency restrictions to test-only imports.
- Introduce marker files or require particular Go source filenames.
- Infer semantic concerns such as presentation or business policy from
  function bodies.

## Decisions

### Classify packages by architectural type

Use package paths, which are Go's native compilation and dependency units, as
the source of architectural classification. Directory containers such as
`adapters/inbound` and `adapters/outbound` make types visible to contributors;
they do not contain Go files themselves.

The target structure is:

```text
cmd/sbxflow/                              entrypoint and wiring

internal/
├── adapters/
│   ├── inbound/
│   │   └── cli/                          Cobra and rendering
│   └── outbound/
│       ├── declaration/                  discovery, decoding, filesystem
│       └── sbx/                          Docker Sandbox subprocesses
├── application/
│   ├── doctor/                           workflow
│   ├── lifecycle/                        workflows
│   └── validation/                       validation use case
├── domain/
│   └── configuration/                    linking, trust, rules
├── ports/
│   ├── buildInfo/                        linker-provided build identity
│   ├── declaration/                      declaration boundary and DTOs
│   └── sandbox/                          Docker Sandbox capabilities

schema/                                   published resource
```

The entrypoint is an existing package, not a new `composition` package. It is
the sole startup and wiring boundary and may import concrete implementations
that normal architectural types cannot. Wiring currently performed by
`cli.Execute` moves to `cmd/sbxflow`; CLI construction remains independently
testable through injected runner interfaces.

Alternatives considered:

- Required marker files were rejected because Go packages are defined by
  directories and imports; marker filenames add ceremony without defining
  dependency semantics.
- Keeping the current directories and mapping every package in policy was
  rejected because it recreates the instance-specific graph.

### Enforce relationships between types

Every production file must be assigned to an architectural component.
Recursive patterns let subpackages inherit their enclosing type while a new
top-level architectural concept must be added deliberately. Where component
globs overlap, go-arch-lint assigns the most-specific match.

The allowed direct dependency matrix is:

| Source type      | Allowed first-party targets    |
| ---------------- | ------------------------------ |
| Entry point      | Any production package         |
| Inbound adapter  | Application and BuildInfo port |
| Application      | Domain and Sandbox port        |
| Domain           | General ports                  |
| Port             | None                           |
| Outbound adapter | General ports and resources    |
| Resource         | None                           |

Additional invariants are:

- Inbound adapters never import domain, general ports, or outbound adapters.
- Application packages never import peer application packages.
- Application packages access no general port; Sandbox is their one explicit
  port capability.
- Domain packages never import peer domain packages, adapters, application, or
  presentation.
- Ports have no first-party dependencies and contain no business rules;
  BuildInfo's linker values are the documented implementation exception.
- Outbound adapters depend on ports rather than domain or application callers.
- Outbound adapters never import one another.
- Cobra may be imported only by inbound adapters.
- The entrypoint's broader imports are for startup and wiring; semantic use of
  those implementations remains outside automated inspection.

Direct imports are evaluated. Transitive flow such as entrypoint to CLI to
lifecycle is expected and does not create an additional direct relationship.

### Make configuration the domain and validation the use case

Split validation and configuration responsibilities around capability-scoped
ports:

- The declaration port owns the decoded declaration boundary model and narrow
  loading, local-path resolution, and lifecycle-target interfaces.
- The sandbox port owns narrow kit-validation, inspection, and lifecycle
  interfaces plus their boundary request and result types.
- The BuildInfo port directly owns linker-provided identity because a separate
  adapter would only forward immutable process values.
- Domain configuration owns source linking, trust derivation, semantic rules,
  safe path resolution, and the resulting configuration validity model.
- The declaration adapter owns nearest-file discovery, file reads, YAML and
  JSON Schema decoding, and local filesystem resolution.
- Application validation coordinates resolved configuration with the Sandbox
  port's local-kit validation capability.
- Lifecycle owns a validator interface that application validation satisfies,
  avoiding a dependency between peer application packages.
- Application lifecycle and doctor consume sandbox ports rather than the
  concrete `sbx` adapter.

The entrypoint constructs the declaration and `sbx` adapters, supplies them to
configuration and validation, constructs application workflows, and passes
those workflows to the CLI command tree. The CLI consumes the BuildInfo port
directly for version rendering. Wiring narrows concrete adapters to their port
interfaces before injection so deep dependency analysis observes the same
boundary as direct-import analysis.

Generic shared contracts were rejected because they would create an
unbounded dumping ground. Port packages remain narrow and capability-specific,
and each contains only interfaces and boundary data required to cross that
port.

### Replace the package-specific Arch-Go prototype

Use pinned go-arch-lint and its declarative component configuration instead of
maintaining either a package-instance allowlist or a repository-owned checker.
Recursive component globs classify future packages by architectural type, and
unmatched production files fail the check. Test files are excluded explicitly.

The policy classifies `internal/ports/**` separately from domain, application,
and adapter packages, with more-specific BuildInfo and Sandbox components for
the two declared shortcuts. Vendor rules allow Cobra only from the inbound
component and allow decoding libraries only from outbound adapters.

Keep the check behind `mise run test:architecture` and invoke it from
`mise run validate`. Review the tool's component mapping when changing package
roots or overlapping component patterns. Arch-Go, its Mise pin, and
`arch-go.yml` are removed only after go-arch-lint passes the reorganized
repository, avoiding a validation gap during migration.

### Keep intent and enforcement reviewable

Update `docs/structure.md` to document types, responsibilities, ports, and the
dependency matrix rather than a snapshot of imports.
`CONTRIBUTING.md` and `AGENTS.md` continue to require deliberate review of
architecture policy changes. The existing required `Validate` CI job and
pre-commit hook remain the enforcement entry points.

External ownership controls such as CODEOWNERS remain deferred until the
repository has named maintainers.

## Risks / Trade-offs

- **[The refactor touches most internal imports without changing behavior]** →
  Move one architectural type at a time, retain tests during each move, and run
  the full suite after every stage.
- **[Architecture enforcement could become bespoke and brittle]** → Use a
  pinned component linter with a small declarative policy and no
  repository-owned checker implementation.
- **[The entrypoint can import broadly]** → Limit it to startup, wiring, stream
  setup, build metadata, and exit handling through documentation and review.
- **[Ports could become a generic shared-types layer]** → Split them by external
  capability and prohibit first-party dependencies, rules, and implementations
  within port packages, except for BuildInfo's explicitly documented linker
  values.
- **[Configuration responsibilities may be difficult to split mechanically]**
  → Move pure models and rules first, then adapt discovery and parsing around
  their stable inputs and outputs.
- **[Type directories add path depth]** → Accept the additional nesting because
  it makes classification automatic and contributor placement explicit.

## Migration Plan

1. Introduce declaration, sandbox, and BuildInfo ports plus the domain
   configuration and application validation packages with tests while the
   current packages compile.
2. Introduce declaration, `sbx`, and CLI adapter paths, then migrate callers and
   tests incrementally onto the ports.
3. Move dependency construction from CLI to the existing `cmd/sbxflow`
   entrypoint and remove obsolete package paths.
4. Add the type-based go-arch-lint policy, review its component mapping, and
   confirm the refactored repository passes.
5. Remove Arch-Go, revise architecture documentation, format, and run aggregate
   validation.

Rollback before the final removal step consists of retaining the current
package paths and Arch-Go task. After migration, normal version control can
revert the package moves and checker together; no data or runtime migration is
involved.

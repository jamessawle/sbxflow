## Context

The executable currently contains a Cobra command tree and a repository-
independent diagnostic runner, but it has no configuration model or repository
discovery. The documented version 1 declaration uses named Git, OCI, and local
sources plus an ordered `use` list. JSON Schema can describe both sides of that
document but cannot follow a selection's dynamic source name to enforce rules
based on the referenced source type.

Kits selected through a source declared as local are a deliberate exception to
offline reference validation. Docker Sandboxes already provides
`sbx kit validate` for filesystem artifacts, which it may interpret as
directories or ZIP files. Packaging does not establish provenance: the local
source type and a safely resolved host path are the only reasons sbxflow invokes
the local validator. The implementation must call that interface without
allowing invalid configuration, URI-like references, or path traversal to choose
an arbitrary reference.

## Goals / Non-Goals

**Goals:**

- Make the published version 1 schema the structural contract used by the CLI
  and available to editor and CI tooling.
- Reject agent names that are not supported by sbxflow's bounded Docker
  Sandboxes compatibility range.
- Produce a typed, linked configuration that later lifecycle commands can reuse.
- Keep the custom semantic layer limited to relationships JSON Schema cannot
  express portably.
- Validate every selected local kit with Docker's own validator while keeping
  validation inspection-only.
- Produce deterministic, least-privilege trust settings from the ordered
  selections.

**Non-Goals:**

- Fetch, inspect, or authenticate to Git or OCI sources.
- Dynamically inspect the installed `sbx` executable to discover agent names.
- Create an execution fingerprint or reconcile sandbox state.
- Add configuration migration, automatic correction, JSON output, or a custom
  declaration path flag.
- Validate unused local kit artifacts or unused declared sources beyond their
  schema-defined fields.

## Decisions

### Use a gated validation pipeline

Validation will proceed through explicit phases:

```text
nearest sbxflow.yaml
        |
        v
YAML parse + JSON Schema
        |
        v
source linking + local path containment
        |
        +------> selected local kits --> sbx kit validate
        |
        v
ordered trust derivation --> report
```

Structural errors stop before typed decoding and semantic linking. Semantic or
path-containment errors stop before any Docker subprocess is invoked. Once those
phases pass, local kit validations run sequentially in selection order; an
invalid artifact is recorded without preventing later independent artifacts
from being checked. Trust is derived from the already linked selections and is
reported with the overall result.

This separation prevents cascades of misleading errors and makes the linked
configuration a reusable boundary for `up`, `down`, and `destroy`. Running all
checks in one permissive pass was considered, but invalid references and paths
would make subprocess selection unsafe and error attribution unclear.

### Discover the nearest declaration by walking ancestors

The command will start at its working directory and walk parent directories
until it finds `sbxflow.yaml`, choosing the nearest match. Discovery will not
depend on Git, so the CLI remains useful for repositories managed by other
systems and for directories not yet committed.

Requiring invocation from the declaration directory was considered but would be
awkward in scripts and nested terminals. Stopping at a Git root was considered
but would make Git an undeclared runtime dependency and behave poorly for nested
or non-Git workspaces.

### Publish one JSON Schema and embed that same file at runtime

A checked-in Draft 2020-12 JSON Schema will define configuration version 1. It
will use strict object shapes, discriminated source unions, and `uniqueItems`
for exact duplicate selections. The Go executable will embed the checked-in
schema rather than maintain a second programmatic copy.

YAML will first be parsed with duplicate-key and multi-document rejection, then
converted to the JSON data model for schema validation. Only schema-valid data
will be decoded into Go types. Schema validation errors will retain YAML-facing
field paths where possible.

Hand-written structural validation was considered but duplicates standard
schema behavior and would not provide an artifact editors can consume. Generating
the schema from Go types was considered, but source-type unions and strict
conditional fields are clearer and more stable when the public schema is the
source of truth.

### Pin the supported agent set in the versioned schema

The schema will enumerate the agent identifiers supported across sbxflow's
verified Docker Sandboxes version range: `claude`, `codex`, `copilot`, `cursor`,
`docker-agent`, `droid`, `gemini`, `kiro`, `opencode`, and `shell`. Validation
will remain offline and deterministic rather than invoking `sbx` to discover
commands dynamically. `shell` remains the escape hatch for manual or custom
agent installation.

Accepting arbitrary strings was considered, but it defers simple declaration
typos until a mutating lifecycle command. Discovering agents from the installed
binary was rejected because validation must work without `sbx` when no local kit
is selected and because a local binary may not be in sbxflow's supported range.
When Docker adds or removes an agent, sbxflow will update the enum as part of
verifying a new compatibility range.

### Keep a self-validating project declaration

The repository root will contain `sbxflow.yaml` selecting `codex`, the community
`mise` kit, and `kits/mise/network-go` from the pinned personal sbx-kits source.
Both selections are remote, so validating sbxflow itself exercises discovery,
schema validation, linking, and trust derivation without requiring `sbx` or
network materialization.

### Keep cross-reference validation as a small semantic linking pass

The linker will preserve the ordered `use` slice and attach each selection to
its named source. It will report an unknown source, require `version` for OCI
selections, and reject `version` for Git and local selections. Those are the
only field-shape rules kept outside JSON Schema because the selection's source
type is available only after the dynamic name lookup.

Adding a repeated `type` discriminator to every selection was considered but
would create two values that could disagree. Embedding the full source in every
selection was considered but would discard named-source reuse. Non-standard
schema data-reference extensions were rejected to keep the published schema
portable.

### Require host filesystem references and contain them before invoking Docker

A local source `root` must be a host filesystem path rather than an HTTP, Git,
OCI, or other URI-like reference. A relative root will be resolved from the
directory containing `sbxflow.yaml`; an absolute root remains absolute. The root
must exist and resolve to a directory. Each selected `kit` must likewise be a
relative filesystem path rather than a URI. It is joined to the root, cleaned,
and resolved through symbolic links. The selected target must exist and remain
at or below the canonical root.

Each accepted local target will be passed as an absolute path to
`sbx kit validate`. Docker, rather than sbxflow, will determine whether that
filesystem artifact is a valid kit directory or ZIP. The subprocess will be
invoked without a shell, with a bounded timeout and captured stdout and stderr.
A failure will be attributed to the source and kit selection and will retain
Docker's useful diagnostic text. If `sbx` is unavailable, local-kit validation
fails clearly; configurations with no kits selected from local sources do not
look up or invoke `sbx`.

"Local" describes how sbxflow addresses the artifact, not the underlying
storage technology. A network-mounted volume that presents a normal host path is
therefore treated as local; reliably distinguishing its physical storage from a
local disk is neither portable nor relevant to Docker's local-kit trust setting.

Lexical path cleaning alone was considered but would permit a symlink beneath a
source root to escape it. Inferring locality from a `.zip` extension was rejected
because file format says nothing about provenance. Reimplementing Docker's kit
format was rejected because the experimental format can evolve within sbxflow's
supported Docker Sandbox version range.

### Treat remote normalization as planning, not materialization

The linker will normalize enough of each selected remote reference to derive its
execution reference and trust prefix. It will not pass Git or OCI references to
`sbx kit validate`, because that command can resolve remote content.

Remote allowlist output begins with `docker.io/`. A Git selection contributes
the repository host and path without transport syntax or a trailing `.git`; an
OCI selection contributes its registry, namespace, and selected artifact name.
Prefixes are emitted once in first-selection order. The local-kit value is true
only when a local source appears in `use`. Unused sources never widen trust.

This creates the same deterministic input future lifecycle commands need while
keeping `validate` independent of network availability and credentials.

### Share subprocess execution without coupling validation to doctor

The generic command output and execution abstractions currently housed under
`internal/doctor` will move to a neutral internal package. Doctor and local-kit
validation will both depend on that package and inject fakes in tests. The move
does not change doctor behavior.

Importing `internal/doctor` from configuration validation was considered but
would couple unrelated domains. Duplicating the subprocess wrapper was rejected
because future lifecycle commands will also need the same timeout, exit-code,
and stream handling.

### Use structured human-readable output in the initial command

The command will render one cohesive, YAML-like report with three sections:
the discovered declaration, derived kit trust state, and a terminal validation
state plus findings. Allowed sources will be listed in first-selection order.
Successful validation will report state `pass` and `Findings: []`; failed
validation will report state `fail` with ordered actionable findings. If a
gated phase stops before trust derivation, the report will state that derived
state is unavailable instead of rendering misleading empty defaults.

The complete report will use standard output on success and standard error on
failure so one invocation never splits related state across streams. Cobra will
still return a non-zero error through the existing execution boundary. Local-kit
diagnostic text will be attached to its corresponding failure finding rather
than repeated as a separate result block.

The first release will not add `--json`. A machine-readable mode can be designed
once the result model has a real automation consumer rather than freezing an
unproven output schema now. The YAML-like presentation is human-readable and is
not a stable machine interface.

## Risks / Trade-offs

- **The public schema and typed Go model can drift** -> Tests will load every
  repository example through the embedded schema and assert representative
  invalid documents at both schema and decode boundaries.
- **Docker's agent set can change** -> Treat the enum as part of the versioned
  configuration contract and update it when verifying a new supported `sbx`
  range.
- **Docker's experimental kit validator can change** -> Keep the integration to
  the documented `sbx kit validate <local-reference>` command, inject the
  runner, and cover output and exit handling without parsing unstable prose.
- **Local validation depends on filesystem and installed tooling** -> Unit tests
  will use temporary paths and fake subprocesses; process tests will supply a
  fake `sbx` executable rather than depend on the developer's installation.
- **Ancestor discovery can select an unintended outer declaration** -> Always
  report the selected absolute path before validation, and choose the nearest
  file deterministically.
- **Trust normalization mistakes could grant too much access later** -> Test
  exact Git and OCI prefix boundaries, deduplication, first-selection order, and
  exclusion of unused sources before lifecycle commands consume the output.

## Migration Plan

This adds a command and configuration contract without persistent application
state. Introduce the schema and validation packages first, register the command
after their behavior is covered, then update the documented interface. Rollback
removes the new command and packages and restores the shared subprocess types to
the doctor package; no repository or Docker Sandbox state requires migration.

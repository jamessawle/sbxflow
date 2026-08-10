## Purpose

Defines how sbxflow discovers and validates a repository declaration, validates
selected local kit artifacts, and reports the least-privilege Docker kit-source
settings required by the declaration.

## ADDED Requirements

### Requirement: Validate discovers the nearest declaration

The `validate` command SHALL search the current directory and then its ancestors
for the nearest file named `sbxflow.yaml` and SHALL validate that declaration.

#### Scenario: Run from the declaration directory

- **WHEN** a user runs `sbxflow validate` in a directory containing
  `sbxflow.yaml`
- **THEN** the command validates that file

#### Scenario: Run from a nested directory

- **WHEN** a user runs `sbxflow validate` below a directory containing
  `sbxflow.yaml`
- **THEN** the command validates the nearest ancestor declaration

#### Scenario: Declaration is absent

- **WHEN** no `sbxflow.yaml` exists in the current directory or any ancestor
- **THEN** the command reports that no repository declaration was found
- **AND** exits with a non-zero status

### Requirement: Declaration structure is strictly validated

The `validate` command SHALL parse exactly one YAML document and validate it
against the published schema for configuration version 1. The schema SHALL
enforce required fields, field types, supported Docker Sandbox agent names,
source-specific shapes, exact duplicate selection rejection, and rejection of
unknown fields.

#### Scenario: Structurally valid declaration

- **WHEN** a declaration conforms to the version 1 schema
- **THEN** structural validation succeeds

#### Scenario: YAML is malformed or ambiguous

- **WHEN** a declaration contains malformed YAML, multiple YAML documents, or a
  duplicate mapping key
- **THEN** the command reports a structural validation error
- **AND** exits with a non-zero status

#### Scenario: Configuration version is unsupported

- **WHEN** a declaration specifies a version other than 1
- **THEN** the command reports the unsupported version
- **AND** exits with a non-zero status

#### Scenario: Agent is supported

- **WHEN** a declaration selects one of the Docker Sandbox agents supported by
  sbxflow's compatible `sbx` range
- **THEN** structural validation accepts the agent

#### Scenario: Agent is unsupported

- **WHEN** a declaration selects an agent name outside the supported set
- **THEN** the command reports the invalid `sandbox.agent` value
- **AND** exits with a non-zero status

#### Scenario: Source shape is invalid

- **WHEN** a source omits a field required by its type, includes a field from a
  different source type, or contains an unknown field
- **THEN** the command reports the affected source and field
- **AND** exits with a non-zero status

#### Scenario: Exact selection is repeated

- **WHEN** the ordered `use` list contains two structurally identical selection
  objects
- **THEN** the command reports the duplicate selection
- **AND** exits with a non-zero status

### Requirement: Kit selections reference compatible sources

The `validate` command SHALL link every selection to a source declared under
`sandbox.kits.sources` and SHALL apply selection rules that depend on the
referenced source type while preserving declaration order.

#### Scenario: Selection references an unknown source

- **WHEN** a `use` entry names a source that is not declared
- **THEN** the command identifies the selection and unknown source
- **AND** exits with a non-zero status

#### Scenario: OCI selection omits its version

- **WHEN** a selection references an OCI source without declaring `version`
- **THEN** the command reports that the OCI selection requires a version
- **AND** exits with a non-zero status

#### Scenario: Non-OCI selection declares a version

- **WHEN** a selection referencing a Git or local source declares `version`
- **THEN** the command reports that the field is not valid for that source type
- **AND** exits with a non-zero status

#### Scenario: Selections are valid

- **WHEN** every selection references a declared source and contains the fields
  required for that source type
- **THEN** the command retains the selections in their declared order for local
  validation and trust derivation

### Requirement: Remote selections are validated without materialization

The `validate` command SHALL validate the declared shape and semantics of Git and
OCI selections without cloning repositories, pulling artifacts, checking remote
existence, or testing registry credentials.

#### Scenario: Valid Git selection is unreachable

- **WHEN** a Git selection is structurally and semantically valid but its remote
  repository cannot be reached
- **THEN** repository reachability does not affect validation

#### Scenario: Valid OCI selection is unpublished

- **WHEN** an OCI selection is structurally and semantically valid but its
  artifact is not present in the registry
- **THEN** artifact availability does not affect validation

### Requirement: Kits selected from local sources are validated by Docker Sandboxes

For each selection referencing a source declared as local, the `validate`
command SHALL require the source root to be a host filesystem path and the
selected kit to be a relative filesystem path rather than a URI-like reference.
It SHALL resolve the selected path relative to the declaration, require the
result to remain beneath the source root, and invoke `sbx kit validate` with the
resulting absolute host path. These selections SHALL be validated in declaration
order. Docker SHALL determine whether each filesystem artifact is a valid kit
directory or ZIP; its packaging SHALL NOT determine whether sbxflow considers it
local.

#### Scenario: Local source root is not a filesystem reference

- **WHEN** a source declared as local uses an HTTP, Git, OCI, or other URI-like
  value as its root
- **THEN** the command reports that the local source requires a host filesystem
  path
- **AND** does not invoke Docker's kit validator for that source
- **AND** exits with a non-zero status

#### Scenario: Local source root is unavailable

- **WHEN** a local source root does not exist or does not resolve to a directory
- **THEN** the command reports that the local source root is unavailable
- **AND** does not invoke Docker's kit validator for that source
- **AND** exits with a non-zero status

#### Scenario: Local kit is not a relative filesystem reference

- **WHEN** a selection from a local source uses an absolute path or an HTTP, Git,
  OCI, or other URI-like value for `kit`
- **THEN** the command reports that the selection requires a relative host
  filesystem path
- **AND** does not invoke Docker's kit validator for that selection
- **AND** exits with a non-zero status

#### Scenario: Selected local directory is valid

- **WHEN** a local selection resolves beneath its source root and Docker
  Sandboxes accepts the directory as a valid kit
- **THEN** the command reports that local selection as valid

#### Scenario: Kit selected from a local source is a valid ZIP

- **WHEN** a local selection resolves beneath its source root and Docker
  Sandboxes accepts the ZIP file as a valid kit
- **THEN** the command reports that local selection as valid

#### Scenario: Remote selection may use ZIP packaging

- **WHEN** a Git or OCI selection refers to an artifact that may be packaged as
  a ZIP file
- **THEN** the packaging does not cause sbxflow to treat the selection as local
- **AND** the command does not invoke Docker's local kit validator for it

#### Scenario: Local selection escapes its source root

- **WHEN** a local selection resolves outside its declared source root,
  including through a symbolic link
- **THEN** the command rejects the selection without invoking Docker's kit
  validator for it
- **AND** exits with a non-zero status

#### Scenario: Selected local kit is invalid

- **WHEN** `sbx kit validate` rejects a selected local directory or ZIP file
- **THEN** the command identifies the selection and reports Docker's validation
  failure
- **AND** continues validating other independently selected local kits
- **AND** exits with a non-zero status

#### Scenario: Docker kit validator is unavailable

- **WHEN** the declaration selects at least one local kit and `sbx kit validate`
  cannot be executed
- **THEN** the command reports that local kit validation could not be completed
- **AND** exits with a non-zero status

#### Scenario: Declaration has no kits selected from local sources

- **WHEN** all selections reference Git or OCI sources
- **THEN** the command does not require or invoke `sbx kit validate`

### Requirement: Effective kit-source trust is derived from selections

The `validate` command SHALL report the effective `kit.allowedSources` and
`kit.allowLocalKits` values that sbxflow would apply only to its future Docker
Sandbox commands. Derivation SHALL begin with Docker Hub, add each selected Git
or OCI artifact's narrowest required remote prefix in first-selection order,
deduplicate prefixes, and enable local kits only when at least one local source
is selected.

#### Scenario: Git and OCI sources are selected

- **WHEN** the ordered selections use a Git repository and an OCI artifact
- **THEN** the reported remote allowlist contains `docker.io/` followed by the
  narrowest prefixes required for those selected artifacts
- **AND** preserves their first-selection order

#### Scenario: The same remote prefix is selected more than once

- **WHEN** multiple selections derive the same remote prefix
- **THEN** the reported allowlist contains that prefix only once

#### Scenario: Local source is selected

- **WHEN** at least one selection references a local source
- **THEN** the reported `kit.allowLocalKits` value is `true`

#### Scenario: Declared source is unused

- **WHEN** a source is declared but no `use` entry selects it
- **THEN** that source does not affect the reported trust settings

### Requirement: Validation does not change sandbox state

The `validate` command SHALL NOT create, start, stop, recreate, or remove a
sandbox and SHALL NOT modify global Docker Sandbox settings.

#### Scenario: Validation succeeds

- **WHEN** a declaration and its selected local kits are valid
- **THEN** the command reports the validated declaration and derived trust
  settings
- **AND** exits successfully without changing sandbox or global settings state

#### Scenario: Validation fails

- **WHEN** structural, semantic, or local kit validation fails
- **THEN** the command reports actionable validation errors
- **AND** exits with a non-zero status without changing sandbox or global
  settings state

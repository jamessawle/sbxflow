# configuration-validation Specification

## Purpose

Defines how sbxflow discovers and validates a repository declaration, validates
selected local kit artifacts, and reports the least-privilege Docker kit-source
settings required by the declaration.

## Requirements

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

### Requirement: Sandbox network allow resources are strictly declared

The `validate` command SHALL accept an optional
`sandbox.network.allowedHosts` array of unique, non-empty strings and SHALL
reject unknown network fields. Each entry SHALL be a host, domain, wildcard
subdomain, bracketed IPv6 literal, or `**`, each with an optional `:port` suffix
from 1 to 65535, and SHALL be retained exactly as declared. A bracketed literal
SHALL be a well-formed IPv6 address. Docker Sandboxes matches network requests
by host and port, so a resource carrying a scheme or path, a port outside the
matchable range, or a malformed literal is accepted by its policy CLI but can
never match a request; validation SHALL reject such an entry rather than let it
appear applied.

#### Scenario: Network allow resources are declared

- **WHEN** a version 1 declaration contains unique, non-empty entries under
  `sandbox.network.allowedHosts`
- **THEN** structural validation succeeds
- **AND** the entries remain in declaration order

#### Scenario: Network allow resource is empty

- **WHEN** `sandbox.network.allowedHosts` contains an empty string
- **THEN** structural validation reports the invalid entry
- **AND** exits with a non-zero status

#### Scenario: Network allow resource is repeated

- **WHEN** `sandbox.network.allowedHosts` contains the same string more than
  once
- **THEN** structural validation reports the duplicate entry
- **AND** exits with a non-zero status

#### Scenario: Network allow resource is not a host

- **WHEN** `sandbox.network.allowedHosts` contains a URL, a path, or any other
  value that is not a host, domain, wildcard subdomain, IP literal, or `**` with
  an optional port
- **THEN** structural validation reports the invalid entry
- **AND** exits with a non-zero status

#### Scenario: Network allow resource cannot match a request

- **WHEN** `sandbox.network.allowedHosts` contains a port outside 1 to 65535, or
  a bracketed literal that is not a well-formed IPv6 address
- **THEN** structural validation reports the invalid entry
- **AND** exits with a non-zero status

#### Scenario: Network declaration contains an unknown field

- **WHEN** `sandbox.network` contains a field other than `allowedHosts`
- **THEN** structural validation reports the unknown field
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
- **THEN** the command reports the declaration and derived kit trust settings
  on standard output
- **AND** reports validation state `pass` with an empty findings list
- **AND** exits successfully without changing sandbox or global settings state

#### Scenario: Validation fails

- **WHEN** structural, semantic, or local kit validation fails
- **THEN** the command reports one cohesive report on standard error
- **AND** reports validation state `fail` with actionable findings
- **AND** reports derived state as unavailable when validation stopped before
  trust derivation
- **AND** exits with a non-zero status without changing sandbox or global
  settings state

### Requirement: Teardown lifecycle operations resolve only the declared sandbox identity

Lifecycle operations that only need to target an existing sandbox SHALL discover
the nearest `sbxflow.yaml` and resolve a lifecycle target from configuration
version 1 and its non-empty `sandbox.name`. Target resolution SHALL parse exactly
one unambiguous YAML document, but SHALL NOT require the declaration's agent,
kit sources, kit selections, local paths, linked references, or derived trust to
be valid.

#### Scenario: Target is resolved from a nested directory

- **WHEN** target resolution starts below a directory containing an identifiable
  version 1 `sbxflow.yaml`
- **THEN** it returns that nearest ancestor declaration and its exact
  `sandbox.name`

#### Scenario: Declaration is absent

- **WHEN** no `sbxflow.yaml` exists in the current directory or any ancestor
- **THEN** target resolution reports that no repository declaration was found
- **AND** does not invoke Docker Sandboxes

#### Scenario: Declaration identity is unsafe to interpret

- **WHEN** the declaration is malformed YAML, contains multiple YAML documents,
  contains a duplicate mapping key, uses an unsupported or missing version, or
  has a missing, non-string, or empty `sandbox.name`
- **THEN** target resolution reports the identity error
- **AND** does not invoke Docker Sandboxes

#### Scenario: Non-identity configuration is invalid

- **WHEN** a version 1 declaration has a non-empty `sandbox.name` but its agent,
  kit sources, kit selections, local paths, linking, or trust derivation would
  fail complete validation
- **THEN** target resolution still returns the declaration and sandbox name
- **AND** does not resolve local paths or invoke `sbx kit validate`

### Requirement: Initialization hooks are strictly declared

The `validate` command SHALL accept an optional `sandbox.hooks.initialize` array
whose entries each contain a `command` argument vector of one or more non-empty
strings. It SHALL preserve both command order and argument order exactly, reject
unknown lifecycle or command-entry fields, and validate the declaration without
executing any command.

#### Scenario: Ordered commands are declared

- **WHEN** a version 1 declaration contains multiple valid
  `sandbox.hooks.initialize` entries
- **THEN** structural and semantic validation succeed
- **AND** each command and its arguments remain in declaration order
- **AND** no declared command is executed

#### Scenario: Lifecycle configuration is omitted

- **WHEN** a version 1 declaration omits `sandbox.hooks` or declares no
  initialization commands
- **THEN** validation succeeds subject to the existing declaration requirements

#### Scenario: Command vector is empty

- **WHEN** an `initialize` entry has an absent or empty `command` vector
- **THEN** structural validation identifies the malformed entry
- **AND** exits with a non-zero status without inspecting or mutating Docker state

#### Scenario: Command argument is empty

- **WHEN** an `initialize` command contains an empty string argument
- **THEN** structural validation identifies the malformed argument
- **AND** exits with a non-zero status without inspecting or mutating Docker state

#### Scenario: Hooks declaration contains an unknown field

- **WHEN** `sandbox.hooks` or one of its `initialize` entries contains an unknown
  field
- **THEN** structural validation reports the unknown field
- **AND** exits with a non-zero status without inspecting or mutating Docker state

#### Scenario: Shell syntax is declared as plain text

- **WHEN** a command argument contains shell metacharacters without an explicit
  shell executable and evaluation option in the same command vector
- **THEN** validation retains that argument literally rather than interpreting it
  through a host shell

### Requirement: Sandbox workspace mode is explicitly selectable

The version 1 declaration SHALL accept an optional `sandbox.workspace` object with at least one property and no unknown properties. Its `mode` property SHALL accept exactly `direct` or `clone`. Only when `sandbox.workspace` is omitted SHALL validation retain `direct` as the effective compatibility behavior; an explicit `direct` or `clone` SHALL be retained exactly as the repository's intended workspace mode.

#### Scenario: Direct mode is explicit

- **WHEN** a declaration contains `sandbox.workspace.mode: direct`
- **THEN** structural validation succeeds
- **AND** the effective workspace mode is `direct`

#### Scenario: Clone mode is explicit

- **WHEN** a declaration contains `sandbox.workspace.mode: clone`
- **THEN** structural validation succeeds
- **AND** the effective workspace mode is `clone`

#### Scenario: Workspace is omitted

- **WHEN** a declaration omits `sandbox.workspace`
- **THEN** the declaration remains structurally valid
- **AND** its effective workspace mode is `direct`

#### Scenario: Workspace object is empty

- **WHEN** a declaration contains an empty `sandbox.workspace` object
- **THEN** structural validation reports that at least one workspace property is required
- **AND** exits with a non-zero status

#### Scenario: Workspace mode is unsupported

- **WHEN** a declaration selects a workspace mode other than `direct` or `clone`
- **THEN** structural validation identifies the unsupported `sandbox.workspace.mode` value
- **AND** exits with a non-zero status

#### Scenario: Workspace property is unknown

- **WHEN** a declaration contains a `sandbox.workspace` property other than `mode`
- **THEN** structural validation reports the unknown property
- **AND** exits with a non-zero status

### Requirement: Validation distinguishes advisory warnings from errors

The validation result SHALL retain advisory warnings separately from errors. Warnings SHALL NOT make an otherwise valid declaration invalid or cause `validate` to exit unsuccessfully. An omitted `sandbox.workspace.mode` SHALL produce a warning that identifies the current effective `direct` behavior, asks the user to declare `direct` or `clone` explicitly, and discloses that the implicit mode will change to `clone` in a future pre-1.0 release.

#### Scenario: Workspace mode is omitted

- **WHEN** validation succeeds for a declaration that omits `sandbox.workspace.mode`
- **THEN** the validation result contains the omitted-mode warning
- **AND** remains valid
- **AND** `validate` exits successfully

#### Scenario: Workspace mode is explicit

- **WHEN** validation succeeds for a declaration that explicitly selects `direct` or `clone`
- **THEN** the validation result does not contain an omitted-mode warning

#### Scenario: Warning and error coexist

- **WHEN** a declaration omits the workspace mode and also has an independent validation error
- **THEN** the validation result retains the warning and the error separately
- **AND** remains invalid because of the error

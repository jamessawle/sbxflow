## ADDED Requirements

### Requirement: Sandbox workspace mode is explicitly selectable

The version 1 declaration SHALL accept an optional `sandbox.workspace` object with at least one property and no unknown properties. Its `mode` property SHALL accept exactly `direct` or `clone`. When `sandbox.workspace` or its mode is omitted, validation SHALL retain `direct` as the effective compatibility behavior; an explicit `direct` or `clone` SHALL be retained exactly as the repository's intended workspace mode.

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

## ADDED Requirements

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

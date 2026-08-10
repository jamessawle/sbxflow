## Why

sbxflow needs a safe way to establish that a repository declaration is usable
before lifecycle commands rely on it. Adding validation now defines the
configuration contract, catches invalid kit selections early, and exposes the
least-privilege kit-source settings that future lifecycle execution will use.

## What Changes

- Add a repository-aware `sbxflow validate` command that discovers and validates
  `sbxflow.yaml` without changing sandbox state.
- Publish and apply a JSON Schema for the version 1 document structure, including
  strict source-specific shapes and rejection of unknown fields.
- Add semantic validation for named source references and selection rules that
  depend on the referenced source type.
- Resolve selected local kit paths relative to the declaration and validate each
  selected directory or ZIP through `sbx kit validate`.
- Derive and report the effective remote-source allowlist and local-kit setting
  required by the ordered selections without modifying Docker's global settings.
- Keep Git and OCI validation offline: validate their declared reference shapes
  without fetching repositories, pulling artifacts, or testing credentials.

## Capabilities

### New Capabilities

- `configuration-validation`: Defines declaration discovery, structural and
  semantic validation, selected local-kit validation, trust derivation, output,
  and exit behavior.

### Modified Capabilities

- `cli-interface`: Makes the `validate` command discoverable through root and
  contextual help.

## Impact

- Adds configuration, schema, semantic-validation, kit-selection, and
  trust-derivation code used by `validate` and future lifecycle commands.
- Adds a YAML parser and JSON Schema validator to the Go module.
- Invokes the supported `sbx kit validate` interface for selected local kits but
  performs no remote resolution or state mutation.
- Extends CLI and process-level tests, examples, README documentation, and the
  aggregate repository validation suite.

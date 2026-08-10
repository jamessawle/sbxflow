## Why

sbxflow needs a safe way to establish that a repository declaration is usable
before lifecycle commands rely on it. Adding validation now defines the
configuration contract, catches invalid kit selections early, and exposes the
least-privilege kit-source settings that future lifecycle execution will use.

## What Changes

- Add a repository-aware `sbxflow validate` command that discovers and validates
  `sbxflow.yaml` without changing sandbox state.
- Publish and apply a JSON Schema for the version 1 document structure, including
  strict source-specific shapes, the supported Docker Sandbox agent names, and
  rejection of unknown fields.
- Add semantic validation for named source references and selection rules that
  depend on the referenced source type.
- Require every source declared as local to use a host filesystem root and every
  selection from it to use a relative filesystem path, rejecting URI-like local
  references before resolving them.
- Resolve kits selected from local sources relative to the declaration, enforce
  source-root containment, and validate each resulting filesystem artifact
  through `sbx kit validate`, whether Docker interprets it as a directory or ZIP.
- Derive and report the effective remote-source allowlist and local-kit setting
  required by the ordered selections without modifying Docker's global settings.
- Render a cohesive human-readable report containing the declaration, derived
  kit state, validation state, and actionable findings.
- Keep Git and OCI validation offline: validate their declared reference shapes
  without fetching repositories, pulling artifacts, or testing credentials.
- Add this repository's own declaration using the community Mise kit and the
  sbx-kits Go networking companion so the project continuously exercises the
  configuration contract it publishes.

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
- Invokes the supported `sbx kit validate` interface only for kits selected from
  local sources; directory or ZIP packaging does not itself determine locality,
  and validation performs no remote resolution or state mutation.
- Extends CLI and process-level tests, examples, README documentation, and the
  aggregate repository validation suite.
- Adds a root `sbxflow.yaml` that is both a usable project sandbox declaration
  and a schema-validation fixture.

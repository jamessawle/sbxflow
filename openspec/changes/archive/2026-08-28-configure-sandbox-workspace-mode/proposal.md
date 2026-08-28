## Why

sbxflow always exposes the repository's host workspace directly to a sandbox, so users cannot opt into Docker Sandboxes' safer private-clone workflow. The declaration needs an explicit workspace mode now so repositories can choose their intended behavior and avoid relying on an implicit default that is expected to change before sbxflow 1.0.

## What Changes

- Add an optional `sandbox.workspace.mode` declaration with `direct` and `clone` as the initially supported values.
- Preserve compatibility by treating an omitted workspace mode as `direct`, while warning users to declare either mode explicitly because the implicit mode will change to `clone` in a future pre-1.0 release.
- Render `direct` using the repository workspace and render `clone` using Docker Sandboxes' private-clone environment representation when creating or recreating a sandbox.
- Introduce successful validation warnings and show the omitted-mode warning from both `validate` and `up` without preventing lifecycle work.
- Keep Docker Sandboxes authoritative for whether the repository can be used in clone mode.
- Document mode-specific host visibility, persistence, initialization, destruction, and recreation behavior; recommend explicit `clone` for new sandbox declarations while showing how to preserve direct access explicitly.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `configuration-validation`: Validate and preserve the optional workspace mode, default omitted declarations compatibly, and report advisory warnings without making a declaration invalid.
- `sandbox-lifecycle`: Provision and recreate sandboxes in the declared workspace mode and describe the resulting persistence, initialization, and removal behavior.
- `cli-interface`: Display successful validation warnings through both `validate` and `up` while retaining successful exit and lifecycle behavior.

## Impact

The public `sbxflow.yaml` schema, declaration and configuration models, lifecycle planning, sandbox port, private Docker environment rendering, validation reporting, CLI output, examples, README, OpenSpec requirements, and automated tests are affected. The change stays within the existing declaration, domain, application, port, outbound-adapter, and inbound-adapter responsibilities and does not require a new package or architecture relationship.

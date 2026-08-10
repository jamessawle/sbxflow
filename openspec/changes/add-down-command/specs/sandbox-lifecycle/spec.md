## ADDED Requirements

### Requirement: Down resolves the nearest declared sandbox without complete validation

The `down` command SHALL use lifecycle-target resolution to obtain the exact
sandbox name from the nearest repository declaration before inspecting or
stopping Docker Sandbox state. It SHALL NOT run the complete configuration
validation pipeline.

#### Scenario: Down is invoked from a nested directory

- **WHEN** a user invokes `sbxflow down` below a directory containing an
  identifiable version 1 `sbxflow.yaml`
- **THEN** the command targets the name from the nearest ancestor declaration

#### Scenario: Target resolution fails

- **WHEN** the nearest declaration cannot be found or its lifecycle identity
  cannot be resolved safely
- **THEN** `down` reports the target-resolution failure on standard error
- **AND** does not list or stop a sandbox
- **AND** exits with a non-zero status

#### Scenario: Unrelated configuration is invalid

- **WHEN** the declaration identity is valid but complete validation would fail
  for its agent or kit configuration
- **THEN** `down` continues to inspect and stop the declared sandbox
- **AND** does not invoke local-kit validation

### Requirement: Down idempotently stops the declared sandbox

After resolving the lifecycle target, `down` SHALL determine whether a sandbox
whose name exactly equals the declared `sandbox.name` exists. When it exists,
`down` SHALL ask Docker Sandboxes to stop that exact sandbox without removing
it. When it does not exist or is already stopped, `down` SHALL succeed without
creating or starting a sandbox.

#### Scenario: Declared sandbox is running

- **WHEN** Docker Sandboxes reports the exact declared name as an existing,
  running sandbox
- **THEN** `down` invokes `sbx stop` with that name
- **AND** exits successfully when Docker completes the stop

#### Scenario: Declared sandbox is already stopped

- **WHEN** the declared sandbox exists but is already stopped
- **THEN** `down` succeeds without changing its persisted state

#### Scenario: Declared sandbox is absent

- **WHEN** Docker Sandboxes reports no sandbox whose name exactly equals the
  declared name
- **THEN** `down` exits successfully without invoking `sbx stop`

#### Scenario: Only a similar sandbox name exists

- **WHEN** Docker Sandboxes reports a name that only prefixes, suffixes, or
  otherwise resembles the declared name
- **THEN** `down` treats the declared sandbox as absent
- **AND** does not stop the similarly named sandbox

#### Scenario: Sandbox names cannot be listed

- **WHEN** Docker Sandboxes cannot report existing sandbox names
- **THEN** `down` reports the Docker failure
- **AND** does not attempt to stop a sandbox
- **AND** exits with a non-zero status

#### Scenario: Docker cannot stop the sandbox

- **WHEN** Docker Sandboxes rejects or cannot complete the stop request
- **THEN** its diagnostic output remains visible to the user
- **AND** `down` exits with a non-zero status

### Requirement: Down preserves sandbox state

Stopping through `down` SHALL preserve the sandbox and its persisted tools,
Docker images, agent history, configuration changes, and other sandbox state so
that a later `up` can restart and enter it.

#### Scenario: A stopped sandbox is brought up again

- **WHEN** `down` successfully stops a declared sandbox and the user later runs
  `up` for the same declaration
- **THEN** Docker Sandboxes restarts and enters the existing sandbox
- **AND** its persisted state remains available

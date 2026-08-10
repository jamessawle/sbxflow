## ADDED Requirements

### Requirement: Destroy resolves the nearest declared sandbox without complete validation

The `destroy` command SHALL use lifecycle-target resolution to obtain the exact
sandbox name from the nearest repository declaration before inspecting or
removing Docker Sandbox state. It SHALL NOT run the complete configuration
validation pipeline.

#### Scenario: Destroy is invoked from a nested directory

- **WHEN** a user invokes `sbxflow destroy` below a directory containing an
  identifiable version 1 `sbxflow.yaml`
- **THEN** the command targets the name from the nearest ancestor declaration

#### Scenario: Target resolution fails

- **WHEN** the nearest declaration cannot be found or its lifecycle identity
  cannot be resolved safely
- **THEN** `destroy` reports the target-resolution failure on standard error
- **AND** does not list or remove a sandbox
- **AND** exits with a non-zero status

#### Scenario: Unrelated configuration is invalid

- **WHEN** the declaration identity is valid but complete validation would fail
  for its agent or kit configuration
- **THEN** `destroy` continues to inspect and remove the declared sandbox
- **AND** does not invoke local-kit validation

### Requirement: Destroy idempotently removes the exact declared sandbox

After resolving the lifecycle target, `destroy` SHALL determine whether a
sandbox whose name exactly equals the declared `sandbox.name` exists. When it
exists, `destroy` SHALL ask Docker Sandboxes to remove that exact sandbox and
all associated sandbox resources. When it does not exist, `destroy` SHALL
succeed without creating, starting, stopping, or removing a sandbox.

#### Scenario: Declared sandbox exists

- **WHEN** Docker Sandboxes reports the exact declared name as an existing
  sandbox
- **THEN** `destroy` invokes `sbx rm` with that name
- **AND** exits successfully when Docker completes the removal

#### Scenario: Declared sandbox is absent

- **WHEN** Docker Sandboxes reports no sandbox whose name exactly equals the
  declared name
- **THEN** `destroy` exits successfully without invoking `sbx rm`

#### Scenario: Only a similar sandbox name exists

- **WHEN** Docker Sandboxes reports a name that only prefixes, suffixes, or
  otherwise resembles the declared name
- **THEN** `destroy` treats the declared sandbox as absent
- **AND** does not remove the similarly named sandbox

#### Scenario: Sandbox names cannot be listed

- **WHEN** Docker Sandboxes cannot report existing sandbox names
- **THEN** `destroy` reports the Docker failure
- **AND** does not attempt to remove a sandbox
- **AND** exits with a non-zero status

#### Scenario: Docker cannot remove the sandbox

- **WHEN** Docker Sandboxes rejects or cannot complete the removal request
- **THEN** its diagnostic output remains visible to the user
- **AND** `destroy` exits with a non-zero status

### Requirement: Destroy preserves Docker removal safeguards by default

Without the force option, `destroy` SHALL delegate removal confirmation and
active-session protection to Docker Sandboxes. If the user declines removal or
Docker refuses it, sbxflow SHALL preserve Docker's resulting output and process
status without retrying with force.

#### Scenario: User confirms removal

- **WHEN** Docker prompts during `sbxflow destroy` and the user confirms
- **THEN** Docker proceeds to remove the declared sandbox

#### Scenario: User declines removal

- **WHEN** Docker prompts during `sbxflow destroy` and the user declines
- **THEN** the declared sandbox remains
- **AND** sbxflow preserves Docker's resulting output and process status

#### Scenario: Sandbox has an active session

- **WHEN** Docker refuses an unforced removal because the declared sandbox has
  an active session
- **THEN** the declared sandbox remains
- **AND** `destroy` does not retry the removal with force

### Requirement: Force explicitly bypasses removal safeguards

When the user supplies `--force` or `-f`, `destroy` SHALL ask Docker Sandboxes
to skip confirmation and permit removal of the declared sandbox when it has an
active session. Force SHALL NOT broaden the target beyond the exact declared
sandbox.

#### Scenario: Existing sandbox is forcibly destroyed

- **WHEN** the exact declared sandbox exists and the user invokes
  `sbxflow destroy --force`
- **THEN** `destroy` invokes `sbx rm --force` with that sandbox name
- **AND** Docker does not ask for confirmation

#### Scenario: Forced destruction target is absent

- **WHEN** the declared sandbox is absent and the user invokes
  `sbxflow destroy --force`
- **THEN** `destroy` succeeds without invoking `sbx rm`

### Requirement: Destroy permanently removes sandbox state while preserving the host workspace

Successful destruction SHALL remove the declared sandbox's persisted tools,
Docker images, agent history, configuration changes, and other
sandbox-associated resources. It SHALL NOT delete files from the repository's
host workspace; Docker-managed worktrees remain subject to Docker's removal
behavior.

#### Scenario: A destroyed sandbox is brought up again

- **WHEN** `destroy` removes a declared sandbox and the user later runs `up` for
  the same declaration
- **THEN** `up` creates a new sandbox from the current declaration
- **AND** state from the destroyed sandbox is unavailable

#### Scenario: Host workspace contains repository changes

- **WHEN** the declared sandbox is destroyed
- **THEN** files in the repository's host workspace remain intact

# sandbox-lifecycle Specification

## Purpose

Defines how sbxflow validates, creates, restarts, and interactively enters the
Docker Sandbox named by the nearest repository declaration.

## Requirements

### Requirement: Up validates the nearest declaration before lifecycle execution

The `up` command SHALL discover and run the complete validation pipeline for the
nearest `sbxflow.yaml` before inspecting or entering its declared sandbox.

#### Scenario: Up is invoked from a nested directory

- **WHEN** a user invokes `sbxflow up` below a directory containing a valid
  `sbxflow.yaml`
- **THEN** the command uses the nearest ancestor declaration
- **AND** treats the directory containing that declaration as the repository
  workspace

#### Scenario: Declaration validation fails

- **WHEN** discovery, structural validation, semantic linking, local path
  resolution, or selected local-kit validation fails
- **THEN** the command reports the validation failure on standard error
- **AND** does not list, create, restart, or enter a sandbox
- **AND** exits with a non-zero status

#### Scenario: Existing sandbox and newly invalid declaration

- **WHEN** the declared sandbox already exists but the current declaration is
  invalid
- **THEN** validation still fails before the command attempts to enter the
  existing sandbox

### Requirement: Up determines existence by declared sandbox name

After successful validation, the `up` command SHALL determine whether a sandbox
whose name exactly equals `sandbox.name` already exists. This inspection SHALL
NOT change sandbox state.

#### Scenario: Declared name is absent

- **WHEN** Docker Sandboxes reports no sandbox with the declared name
- **THEN** the command selects the create-and-enter path

#### Scenario: Declared name exists

- **WHEN** Docker Sandboxes reports a sandbox with the declared name
- **THEN** the command selects the enter-existing path regardless of whether the
  sandbox is running or stopped

#### Scenario: Sandbox names cannot be listed

- **WHEN** Docker Sandboxes cannot report the existing sandbox names
- **THEN** the command reports the Docker failure
- **AND** does not attempt to run an agent session
- **AND** exits with a non-zero status

### Requirement: Missing sandbox is created and entered from the declaration

When the declared sandbox name is absent, `up` SHALL ask Docker Sandboxes to run
the declared agent under that name with the repository workspace and every
selected kit in declaration order. Remote kits SHALL use their linked execution
references, and local kits SHALL use their safely resolved absolute host paths.
The resulting agent session SHALL be entered interactively.

#### Scenario: Missing sandbox uses a Git kit

- **WHEN** the declared sandbox is absent and selects a Git kit
- **THEN** the create-and-enter request includes the linked Git execution
  reference
- **AND** includes the declared name, agent, and repository workspace

#### Scenario: Missing sandbox uses an OCI kit

- **WHEN** the declared sandbox is absent and selects an OCI kit
- **THEN** the create-and-enter request includes the linked OCI execution
  reference with its declared version

#### Scenario: Missing sandbox uses a local kit

- **WHEN** the declared sandbox is absent and selects a valid local kit
- **THEN** the create-and-enter request includes the canonical absolute path
  produced by local-kit resolution

#### Scenario: Multiple kits are selected

- **WHEN** a missing sandbox declares multiple valid kit selections
- **THEN** every kit is passed to Docker Sandboxes in declaration order

#### Scenario: Docker cannot create the sandbox

- **WHEN** Docker Sandboxes rejects or cannot complete the create-and-enter
  request
- **THEN** its diagnostic output remains visible to the user
- **AND** `up` exits with a non-zero status

### Requirement: Existing sandbox is entered without reconciliation

When the declared sandbox name already exists, `up` SHALL ask Docker Sandboxes
to run the declared agent in that named sandbox without passing workspace or kit
creation inputs. `up` SHALL NOT inspect or reconcile the existing sandbox's
workspace, kits, or other creation configuration.

#### Scenario: Existing sandbox is running

- **WHEN** the declared sandbox exists and is already running
- **THEN** Docker Sandboxes attaches the user to its agent session

#### Scenario: Existing sandbox is stopped

- **WHEN** the declared sandbox exists and is stopped
- **THEN** Docker Sandboxes restarts it and attaches the user to its agent
  session

#### Scenario: Existing sandbox has a different agent

- **WHEN** Docker Sandboxes determines that the existing named sandbox does not
  use the declared agent
- **THEN** Docker's agent-verification failure remains visible to the user
- **AND** `up` exits with a non-zero status

#### Scenario: Declaration kits changed after creation

- **WHEN** the declared name already exists and the declaration's kit selections
  differ from those used to create it
- **THEN** `up` does not add, remove, reorder, or compare the existing sandbox's
  kits
- **AND** enters the existing sandbox through Docker Sandboxes

### Requirement: Kit trust is scoped to the Docker run process

The `up` command SHALL provide the linked configuration's derived remote-source
allowlist and local-kit permission to its Docker Sandbox agent-run process
without changing persistent global Docker Sandbox settings.

#### Scenario: Missing sandbox uses non-default sources

- **WHEN** creation requires selected Git, OCI, or local kit sources
- **THEN** the Docker run process receives exactly the derived trust values
- **AND** source resolution remains subject to any stricter organisation policy

#### Scenario: Existing sandbox is entered

- **WHEN** the declared sandbox already exists
- **THEN** the same process-local trust values are provided to the Docker run
  process
- **AND** no global setting is modified

### Requirement: Up preserves the interactive agent session

The `up` command SHALL connect the Docker Sandbox agent process directly to the
user's standard input, standard output, and standard error, SHALL allow normal
terminal signals to reach the session, SHALL NOT impose a lifecycle timeout,
and SHALL return success or failure according to the Docker process result.

#### Scenario: User interacts with the agent

- **WHEN** Docker Sandboxes starts or attaches to the agent session
- **THEN** user input reaches the agent and agent output remains visible on the
  corresponding terminal streams

#### Scenario: Session remains open

- **WHEN** the interactive agent session continues running
- **THEN** `sbxflow up` remains attached rather than timing out or returning
  early

#### Scenario: Docker run exits successfully

- **WHEN** the Docker agent-run process exits successfully
- **THEN** `sbxflow up` exits successfully

#### Scenario: Docker run exits unsuccessfully

- **WHEN** the Docker agent-run process exits with a failure
- **THEN** `sbxflow up` exits with a non-zero status

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

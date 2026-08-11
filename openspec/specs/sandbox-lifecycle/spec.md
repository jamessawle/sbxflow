# sandbox-lifecycle Specification

## Purpose

Defines how sbxflow validates, creates, restarts, and interactively enters the
Docker Sandbox named by the nearest repository declaration.

## Requirements

### Requirement: Up validates the nearest declaration before lifecycle execution

The `up` command SHALL discover and run the complete validation pipeline for the
nearest `sbxflow.yaml` before inspecting, removing, creating, restarting, or
entering its declared sandbox.

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
- **AND** does not list, remove, create, restart, or enter a sandbox
- **AND** exits with a non-zero status

#### Scenario: Existing sandbox and newly invalid declaration

- **WHEN** the declared sandbox already exists but the current declaration is
  invalid
- **THEN** validation still fails before the command attempts to enter or
  recreate the existing sandbox

### Requirement: Up determines existence by declared sandbox name

After successful validation, the `up` command SHALL determine whether a sandbox
whose name exactly equals `sandbox.name` already exists. The inspection itself
SHALL NOT change sandbox state. Without recreation, existence SHALL continue to
select between the create-and-enter and enter-existing paths; with recreation,
an existing exact match SHALL select the replace-and-enter path.

#### Scenario: Declared name is absent

- **WHEN** Docker Sandboxes reports no sandbox with the declared name
- **THEN** the command selects the create-and-enter path whether or not
  recreation was requested
- **AND** does not attempt sandbox removal

#### Scenario: Declared name exists

- **WHEN** Docker Sandboxes reports a sandbox with the declared name and
  recreation was not requested
- **THEN** the command selects the enter-existing path regardless of whether the
  sandbox is running or stopped

#### Scenario: Declared name exists with recreation

- **WHEN** Docker Sandboxes reports a sandbox with the declared name and
  recreation was requested
- **THEN** the command selects the replace-and-enter path regardless of whether
  the sandbox is running or stopped

#### Scenario: Sandbox names cannot be listed

- **WHEN** Docker Sandboxes cannot report the existing sandbox names
- **THEN** the command reports the Docker failure
- **AND** does not attempt to remove or run a sandbox
- **AND** exits with a non-zero status

### Requirement: Up can replace an existing sandbox from the current declaration

When recreation is requested, `up` SHALL inspect the lifecycle state of the
exact declared sandbox before removal. An absent sandbox SHALL follow the
ordinary creation path, and a stopped sandbox SHALL be force-removed without
additional confirmation. A running sandbox SHALL be force-removed only after
the user explicitly confirms a warning that recreation can terminate other
attached sessions. The removal SHALL remain limited to the exact declared name
and SHALL permanently discard the removed sandbox's persisted state without
deleting files from the repository's host workspace.

If lifecycle state cannot be determined, confirmation cannot be obtained, or
the user does not confirm, `up` SHALL stop without removing, creating, entering,
or otherwise changing the sandbox.

#### Scenario: Existing sandbox is recreated

- **WHEN** the exact declared sandbox exists, a user invokes
  `sbxflow up --recreate`, and any required running-sandbox confirmation is
  granted
- **THEN** `up` invokes `sbx rm --force` with that sandbox name
- **AND** Docker does not ask for an additional removal confirmation
- **AND** after successful removal, `up` creates and enters a sandbox with the
  declared name, agent, repository workspace, selected kits, and derived trust

#### Scenario: Running sandbox recreation is confirmed

- **WHEN** the exact declared sandbox is running, recreation is requested, and
  the user confirms the warning
- **THEN** `up` invokes `sbx rm --force` with that sandbox name
- **AND** after successful removal, `up` creates and enters a sandbox with the
  declared name, agent, repository workspace, selected kits, and derived trust

#### Scenario: Running sandbox recreation is declined

- **WHEN** the exact declared sandbox is running, recreation is requested, and
  the user declines or provides no affirmative response
- **THEN** `up` cancels recreation without removing, creating, or entering a
  sandbox
- **AND** exits with a non-zero status

#### Scenario: Running sandbox confirmation is unavailable

- **WHEN** the exact declared sandbox is running and confirmation input cannot
  be obtained
- **THEN** `up` reports that recreation was not confirmed
- **AND** does not remove, create, or enter a sandbox
- **AND** exits with a non-zero status

#### Scenario: Existing stopped sandbox is recreated

- **WHEN** the exact declared sandbox exists in a stopped state and recreation
  is requested
- **THEN** `up` force-removes it without an additional confirmation prompt
- **AND** creates and enters its replacement from the current declaration

#### Scenario: Recreation state inspection fails

- **WHEN** Docker Sandboxes cannot report a recognized lifecycle state for the
  exact declared sandbox
- **THEN** `up` reports the inspection failure
- **AND** does not remove, create, or enter a sandbox
- **AND** exits with a non-zero status

#### Scenario: Recreation removal fails

- **WHEN** Docker Sandboxes cannot force-remove an approved or stopped declared
  sandbox
- **THEN** its diagnostic output remains visible to the user
- **AND** `up` does not create or enter a sandbox
- **AND** exits with a non-zero status

#### Scenario: Recreated sandbox starts with new state

- **WHEN** forced removal succeeds and `up` creates the replacement sandbox
- **THEN** state from the removed sandbox is unavailable in the replacement
- **AND** files in the repository's host workspace remain intact

### Requirement: Missing sandbox is created and entered from the declaration

When the declared sandbox name is absent, `up` SHALL provision a sandbox under
that name with the repository workspace and every selected kit in declaration
order, and SHALL then attach to that named sandbox so its agent session is
entered interactively. Provisioning and attachment SHALL be separate requests so
that a declared network rule can be applied between them: the provisioning
request SHALL carry the creation inputs, and the attachment request SHALL name
only the sandbox and its declared agent. Remote kits SHALL use their linked
execution references, and local kits SHALL use their safely resolved absolute
host paths.

#### Scenario: Missing sandbox uses a Git kit

- **WHEN** the declared sandbox is absent and selects a Git kit
- **THEN** the provisioning request includes the linked Git execution
  reference
- **AND** includes the declared name, agent, and repository workspace

#### Scenario: Missing sandbox uses an OCI kit

- **WHEN** the declared sandbox is absent and selects an OCI kit
- **THEN** the provisioning request includes the linked OCI execution
  reference with its declared version

#### Scenario: Missing sandbox uses a local kit

- **WHEN** the declared sandbox is absent and selects a valid local kit
- **THEN** the provisioning request includes the canonical absolute path
  produced by local-kit resolution

#### Scenario: Multiple kits are selected

- **WHEN** a missing sandbox declares multiple valid kit selections
- **THEN** every kit is passed to Docker Sandboxes in declaration order

#### Scenario: Provisioned sandbox is entered

- **WHEN** provisioning succeeds and any declared network rule has been applied
- **THEN** `up` attaches to that sandbox by name
- **AND** the agent session is entered interactively

#### Scenario: Docker cannot provision the sandbox

- **WHEN** Docker Sandboxes rejects or cannot complete the provisioning
  request
- **THEN** its diagnostic output remains visible to the user
- **AND** `up` does not attach to a sandbox
- **AND** exits with a non-zero status

### Requirement: Declared network access is applied to a newly created sandbox

Docker Sandboxes accepts a sandbox-scoped local rule only for a sandbox that
already exists. When creating a missing sandbox, `up` SHALL therefore provision
the sandbox without attaching to its agent, add the declared network resources as
a local allow rule scoped to the exact declared sandbox name, and only then enter
the sandbox, so the rule is in force before any agent traffic. Docker Sandboxes
SHALL remain authoritative for the effect of stricter organisation-managed
policy.

#### Scenario: Missing sandbox declares allowed hosts

- **WHEN** the declared sandbox is absent and `allowedHosts` contains one or
  more resources
- **THEN** `up` provisions the sandbox without attaching to its agent
- **AND** asks Docker Sandboxes to allow those resources only for the declared
  sandbox
- **AND** enters the sandbox afterwards so agent traffic can use the rule

#### Scenario: Missing sandbox has no declared allowed hosts

- **WHEN** the declared sandbox is absent and `allowedHosts` is absent or empty
- **THEN** `up` does not ask Docker Sandboxes to add a network rule
- **AND** provisions and enters the sandbox normally

#### Scenario: Docker rejects a declared network resource

- **WHEN** Docker Sandboxes rejects the sandbox-scoped allow request
- **THEN** its diagnostic output remains visible to the user
- **AND** `up` removes the sandbox it just created rather than entering it
  without the declared network access
- **AND** exits with a non-zero status

#### Scenario: The rejected sandbox cannot be removed again

- **WHEN** the sandbox-scoped allow request is rejected and the just-created
  sandbox cannot be removed
- **THEN** `up` reports both the rejected resource and the failed removal
- **AND** exits with a non-zero status

#### Scenario: Organisation governance is stricter

- **WHEN** organisation-managed policy does not permit a declared network
  resource
- **THEN** the local sandbox-scoped rule does not expand the organisation policy
- **AND** Docker Sandboxes remains authoritative for effective access

### Requirement: Existing sandbox is entered without reconciliation

When the declared sandbox name already exists, `up` SHALL ask Docker Sandboxes
to run the declared agent in that named sandbox without passing workspace, kit,
or network creation inputs. `up` SHALL NOT inspect or reconcile the existing
sandbox's workspace, kits, declared network access, or other creation
configuration unless recreation was requested.

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

#### Scenario: Declaration network access changed after creation

- **WHEN** the declared name already exists and `allowedHosts` differs from the
  values applied when it was created
- **THEN** `up` does not add, remove, or compare its sandbox-scoped network rule
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

### Requirement: Sandbox removal cleans up declared network access

The shared lifecycle removal operation used by `destroy` and `up --recreate`
SHALL remove the exact sandbox and SHALL then remove every network resource
currently declared for that sandbox, each by sandbox name and resource. Removing
a resource that is already absent SHALL succeed, because Docker Sandboxes
ordinarily discards a sandbox-scoped policy along with its sandbox and rejects a
removal request naming any resource it cannot find. Declared resources are owned
by sbxflow; manually modifying an overlapping sandbox-scoped resource is outside
sbxflow's ownership guarantees.

#### Scenario: Destroy removes a sandbox with declared network access

- **WHEN** `destroy` successfully removes a sandbox whose declaration contains
  allowed hosts
- **THEN** it removes each declared resource from the local rule scoped to that
  sandbox

#### Scenario: Declared network access is already absent

- **WHEN** removal reaches a declared resource that Docker Sandboxes no longer
  holds, or a sandbox that has no scoped policy at all
- **THEN** cleanup treats that resource as already removed
- **AND** the lifecycle operation succeeds

#### Scenario: Recreation replaces declared network access

- **WHEN** `up --recreate` removes an existing sandbox
- **THEN** it uses the same removal and network-cleanup behavior as `destroy`
- **AND** applies the current declaration's allowed hosts to the replacement
  between provisioning it and entering it

#### Scenario: Sandbox removal fails

- **WHEN** Docker Sandboxes does not remove the sandbox
- **THEN** the lifecycle operation reports the removal failure
- **AND** does not remove its declared sandbox-scoped network resources

#### Scenario: Network cleanup fails after removal

- **WHEN** the sandbox was removed but Docker Sandboxes cannot remove a declared
  network resource for a reason other than its absence
- **THEN** the lifecycle operation reports that the sandbox was removed but
  network cleanup was incomplete
- **AND** exits with a non-zero status

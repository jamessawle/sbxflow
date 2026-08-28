## RENAMED Requirements

- FROM: `### Requirement: Kit trust is scoped to the Docker run process`
- TO: `### Requirement: Kit trust is process-scoped`

## MODIFIED Requirements

### Requirement: Up can replace an existing sandbox from the current declaration

When recreation is requested, `up` SHALL inspect the lifecycle state of the
exact declared sandbox before removal. An absent sandbox SHALL follow the
ordinary creation path, and a stopped sandbox SHALL be force-removed without
additional confirmation. A running sandbox SHALL be force-removed only after
the user explicitly confirms a warning that recreation can terminate other
attached sessions or explicitly requests forced recreation. The shared removal
operation SHALL remain limited to the exact declared name and SHALL permanently
discard the removed sandbox's persisted state without deleting files from the
repository's host workspace.

If lifecycle state cannot be determined, or unforced confirmation cannot be
obtained or is declined, `up` SHALL stop without removing, creating, entering,
or otherwise changing the sandbox. A force request without recreation SHALL be
rejected before configuration validation, sandbox inspection, or any other
lifecycle work.

#### Scenario: Existing sandbox is recreated

- **WHEN** the exact declared sandbox exists, a user invokes
  `sbxflow up --recreate`, and any required running-sandbox confirmation is
  granted
- **THEN** `up` invokes the shared removal operation with force for that exact
  name
- **AND** Docker does not ask for an additional removal confirmation
- **AND** after successful removal, `up` creates and enters a sandbox with
  the declared name, agent, repository workspace, selected kits, and derived
  trust

#### Scenario: Running sandbox recreation is confirmed

- **WHEN** the exact declared sandbox is running, recreation is requested
  without force, and the user confirms the warning
- **THEN** `up` invokes the shared removal operation with force for the exact
  name
- **AND** after successful removal, `up` creates and enters its replacement
  from the current declaration

#### Scenario: Running sandbox recreation is forced

- **WHEN** the exact declared sandbox is running and forced recreation is
  requested
- **THEN** `up` does not request interactive confirmation
- **AND** invokes the shared removal operation with force for the exact declared
  name
- **AND** after successful removal, creates and enters its replacement from the
  current declaration

#### Scenario: Running sandbox recreation is declined

- **WHEN** the exact declared sandbox is running, unforced recreation is
  requested, and the user declines or provides no affirmative response
- **THEN** `up` cancels recreation without removing, creating, or entering a
  sandbox
- **AND** exits with a non-zero status

#### Scenario: Running sandbox confirmation is unavailable

- **WHEN** the exact declared sandbox is running, unforced recreation is
  requested, and confirmation input cannot be obtained
- **THEN** `up` reports that recreation was not confirmed
- **AND** does not remove, create, or enter a sandbox
- **AND** exits with a non-zero status

#### Scenario: Existing stopped sandbox is recreated

- **WHEN** the exact declared sandbox exists in a stopped state and recreation
  is requested with or without force
- **THEN** `up` force-removes it without an additional confirmation prompt
- **AND** creates and enters its replacement from the current declaration

#### Scenario: Missing sandbox follows creation path

- **WHEN** the exact declared sandbox is absent and recreation is requested
  with or without force
- **THEN** `up` creates and enters the sandbox through the ordinary missing-
  sandbox path

#### Scenario: Force without recreation is rejected

- **WHEN** lifecycle is requested with force enabled and recreation disabled
- **THEN** `up` reports the invalid option combination before configuration
  validation or sandbox inspection
- **AND** does not remove, create, or enter a sandbox

#### Scenario: Recreation state inspection fails

- **WHEN** Docker Sandboxes cannot report a recognized lifecycle state for the
  exact declared sandbox
- **THEN** `up` reports the inspection failure
- **AND** does not remove, create, or enter a sandbox
- **AND** exits with a non-zero status

#### Scenario: Recreation removal fails

- **WHEN** Docker Sandboxes cannot force-remove an approved, forced, or stopped
  declared sandbox
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
order, and SHALL then enter its agent session interactively. Provisioning and
interactive entry SHALL remain separate lifecycle operations so a declared
network rule can be applied between them. Remote kits SHALL use their linked
execution references, and local kits SHALL use their safely resolved absolute
host paths.

#### Scenario: Missing sandbox uses a Git kit

- **WHEN** the declared sandbox is absent and selects a Git kit
- **THEN** the provisioning operation includes the linked Git execution
  reference
- **AND** includes the declared name, agent, and repository workspace

#### Scenario: Missing sandbox uses an OCI kit

- **WHEN** the declared sandbox is absent and selects an OCI kit
- **THEN** the provisioning operation includes the linked OCI execution
  reference
  with its declared version

#### Scenario: Missing sandbox uses a local kit

- **WHEN** the declared sandbox is absent and selects a valid local kit
- **THEN** the provisioning operation includes the canonical absolute path
  produced by local-kit resolution

#### Scenario: Multiple kits are selected

- **WHEN** a missing sandbox declares multiple valid kit selections
- **THEN** every kit is passed to Docker Sandboxes in declaration order

#### Scenario: Provisioned sandbox is entered

- **WHEN** provisioning succeeds and any declared network rule has been applied
- **THEN** `up` enters the declared sandbox
- **AND** the agent session is entered interactively

#### Scenario: Docker cannot provision the sandbox

- **WHEN** Docker Sandboxes rejects or cannot complete provisioning
- **THEN** its diagnostic output remains visible to the user
- **AND** `up` does not attach to a sandbox
- **AND** exits with a non-zero status

### Requirement: Existing sandbox is entered without reconciliation

When the declared sandbox name already exists, `up` SHALL ask Docker Sandboxes
to enter that exact sandbox. `up` SHALL NOT inspect, compare, or explicitly
reconcile the existing sandbox's workspace, kits, declared network access, or
other creation configuration unless recreation was requested. Docker Sandboxes
SHALL remain authoritative for the existing sandbox's configuration.

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

### Requirement: Kit trust is process-scoped

The `up` command SHALL provide the linked configuration's derived remote-source
allowlist and local-kit permission to Docker Sandboxes when resolving selected
kit references, without changing persistent global Docker Sandbox settings.

#### Scenario: Missing sandbox uses non-default sources

- **WHEN** creation requires selected Git, OCI, or local kit sources
- **THEN** Docker Sandboxes receives exactly the derived trust values for kit
  resolution
- **AND** source resolution remains subject to any stricter organisation policy

#### Scenario: Existing sandbox is entered

- **WHEN** the declared sandbox already exists
- **THEN** the same process-local trust values are available to Docker
  Sandboxes for kit resolution during entry
- **AND** no global setting is modified

### Requirement: Destroy idempotently removes the exact declared sandbox

After resolving the lifecycle target, `destroy` SHALL determine whether a
sandbox whose name exactly equals the declared `sandbox.name` exists. When it
exists, `destroy` SHALL ask Docker Sandboxes to remove that exact sandbox and
its sandbox-scoped resources through the shared removal operation. When it does
not exist, `destroy` SHALL succeed without creating, starting, stopping, or
removing a sandbox.

#### Scenario: Declared sandbox exists

- **WHEN** Docker Sandboxes reports the exact declared name as an existing
  sandbox
- **THEN** `destroy` invokes the shared removal operation for that exact name
- **AND** exits successfully when removal completes

#### Scenario: Declared sandbox is absent

- **WHEN** Docker Sandboxes reports no sandbox whose name exactly equals the
  declared name
- **THEN** `destroy` exits successfully without invoking the shared removal
  operation

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

### Requirement: Force explicitly bypasses removal safeguards

When the user supplies `--force` or `-f`, `destroy` SHALL ask Docker Sandboxes
to skip confirmation and permit removal of the declared sandbox when it has an
active session. Force SHALL NOT broaden the target beyond the exact declared
sandbox.

#### Scenario: Existing sandbox is forcibly destroyed

- **WHEN** the exact declared sandbox exists and the user invokes
  `sbxflow destroy --force`
- **THEN** `destroy` invokes the shared removal operation with force for that
  exact name
- **AND** Docker does not ask for confirmation

#### Scenario: Forced destruction target is absent

- **WHEN** the declared sandbox is absent and the user invokes
  `sbxflow destroy --force`
- **THEN** `destroy` succeeds without invoking the shared removal operation

### Requirement: Sandbox removal cleans up declared network access

The shared lifecycle removal operation used by `destroy` and `up --recreate`
SHALL remove the exact sandbox and its declared sandbox-scoped network access.
Successful removal SHALL leave no declared network resource scoped to that
sandbox. Network access that is already absent SHALL be treated as removed.
Declared resources remain owned by sbxflow; manually modifying an overlapping
sandbox-scoped resource is outside sbxflow's ownership guarantees.

#### Scenario: Destroy removes a sandbox with declared network access

- **WHEN** `destroy` successfully removes a sandbox whose declaration contains
  allowed hosts
- **THEN** its declared sandbox-scoped network access is also removed

#### Scenario: Declared network access is already absent

- **WHEN** the declared network resources are already absent when the sandbox
  is removed
- **THEN** the shared removal operation treats that access as already removed
- **AND** the lifecycle operation succeeds

#### Scenario: Recreation replaces declared network access

- **WHEN** `up --recreate` removes an existing sandbox
- **THEN** it uses the same removal and network-cleanup behavior as `destroy`
- **AND** applies the current declaration's allowed hosts to the replacement
  between provisioning it and entering it

#### Scenario: Sandbox removal fails

- **WHEN** Docker Sandboxes does not remove the sandbox
- **THEN** the lifecycle operation reports the removal failure with Docker's
  diagnostic output
- **AND** does not report its declared sandbox-scoped network access as removed
- **AND** exits with a non-zero status

#### Scenario: Network cleanup fails after removal

- **WHEN** Docker Sandboxes reports that it removed the sandbox but could not
  completely remove the network access scoped to it
- **THEN** the lifecycle operation reports that removal was incomplete with
  Docker's diagnostic output
- **AND** exits with a non-zero status

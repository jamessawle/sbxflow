## ADDED Requirements

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

## MODIFIED Requirements

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

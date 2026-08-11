## ADDED Requirements

### Requirement: Declared network access is applied to a missing sandbox

Before creating a missing sandbox, `up` SHALL ask Docker Sandboxes to add the
declared network resources as a local allow rule scoped to the exact declared
sandbox name. Docker Sandboxes SHALL remain authoritative for resource syntax
and for the effect of stricter organisation-managed policy.

#### Scenario: Missing sandbox declares allowed hosts

- **WHEN** the declared sandbox is absent and `allowedHosts` contains one or
  more resources
- **THEN** `up` asks Docker Sandboxes to allow those resources only for the
  declared sandbox
- **AND** applies the rule before starting creation so kit and agent traffic can
  use it

#### Scenario: Missing sandbox has no declared allowed hosts

- **WHEN** the declared sandbox is absent and `allowedHosts` is absent or empty
- **THEN** `up` does not ask Docker Sandboxes to add a network rule
- **AND** continues creating the sandbox normally

#### Scenario: Docker rejects a declared network resource

- **WHEN** Docker Sandboxes rejects the sandbox-scoped allow request
- **THEN** its diagnostic output remains visible to the user
- **AND** `up` does not create or enter the sandbox
- **AND** exits with a non-zero status

#### Scenario: Organisation governance is stricter

- **WHEN** organisation-managed policy does not permit a declared network
  resource
- **THEN** the local sandbox-scoped rule does not expand the organisation policy
- **AND** Docker Sandboxes remains authoritative for effective access

### Requirement: Sandbox removal cleans up declared network access

The shared lifecycle removal operation used by `destroy` and `up --recreate`
SHALL remove the exact sandbox and SHALL remove every network resource currently
declared for that sandbox by sandbox name and resource. Declared resources are
owned by sbxflow; manually modifying an overlapping sandbox-scoped resource is
outside sbxflow's ownership guarantees.

#### Scenario: Destroy removes a sandbox with declared network access

- **WHEN** `destroy` successfully removes a sandbox whose declaration contains
  allowed hosts
- **THEN** it removes each declared resource from the local rule scoped to that
  sandbox

#### Scenario: Recreation replaces declared network access

- **WHEN** `up --recreate` removes an existing sandbox
- **THEN** it uses the same removal and network-cleanup behavior as `destroy`
- **AND** applies the current declaration's allowed hosts before creating the
  replacement

#### Scenario: Sandbox removal fails

- **WHEN** Docker Sandboxes does not remove the sandbox
- **THEN** the lifecycle operation reports the removal failure
- **AND** does not remove its declared sandbox-scoped network resources

#### Scenario: Network cleanup fails after removal

- **WHEN** the sandbox was removed but Docker Sandboxes cannot remove a declared
  network resource
- **THEN** the lifecycle operation reports that the sandbox was removed but
  network cleanup was incomplete
- **AND** exits with a non-zero status

## MODIFIED Requirements

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

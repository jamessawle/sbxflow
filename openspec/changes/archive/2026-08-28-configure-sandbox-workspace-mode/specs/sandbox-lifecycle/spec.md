## MODIFIED Requirements

### Requirement: Up can replace an existing sandbox from the current declaration

When recreation is requested, `up` SHALL inspect the lifecycle state of the exact declared sandbox before removal. An absent sandbox SHALL follow the ordinary creation path, and a stopped sandbox SHALL be force-removed without additional confirmation. A running sandbox SHALL be force-removed only after the user explicitly confirms a warning that recreation can terminate other attached sessions or explicitly requests forced recreation. The shared removal operation SHALL remain limited to the exact declared name and SHALL permanently discard the removed sandbox's persisted state, including work stored only inside the sandbox, without deleting files from the repository's host workspace.

If lifecycle state cannot be determined, or unforced confirmation cannot be obtained or is declined, `up` SHALL stop without removing, creating, entering, or otherwise changing the sandbox. A force request without recreation SHALL be rejected before configuration validation, sandbox inspection, or any other lifecycle work.

#### Scenario: Existing sandbox is recreated

- **WHEN** the exact declared sandbox exists, a user invokes `sbxflow up --recreate`, and any required running-sandbox confirmation is granted
- **THEN** `up` invokes the shared removal operation with force for that exact name
- **AND** Docker does not ask for an additional removal confirmation
- **AND** after successful removal, `up` creates and enters a sandbox with the declared name, agent, effective workspace mode, repository path, selected kits, and derived trust

#### Scenario: Running sandbox recreation is confirmed

- **WHEN** the exact declared sandbox is running, recreation is requested without force, and the user confirms the warning
- **THEN** `up` invokes the shared removal operation with force for the exact name
- **AND** after successful removal, `up` creates and enters its replacement from the current declaration

#### Scenario: Running sandbox recreation is forced

- **WHEN** the exact declared sandbox is running and forced recreation is requested
- **THEN** `up` does not request interactive confirmation
- **AND** invokes the shared removal operation with force for the exact declared name
- **AND** after successful removal, creates and enters its replacement from the current declaration

#### Scenario: Running sandbox recreation is declined

- **WHEN** the exact declared sandbox is running, unforced recreation is requested, and the user declines or provides no affirmative response
- **THEN** `up` cancels recreation without removing, creating, or entering a sandbox
- **AND** exits with a non-zero status

#### Scenario: Running sandbox confirmation is unavailable

- **WHEN** the exact declared sandbox is running, unforced recreation is requested, and confirmation input cannot be obtained
- **THEN** `up` reports that recreation was not confirmed
- **AND** does not remove, create, or enter a sandbox
- **AND** exits with a non-zero status

#### Scenario: Existing stopped sandbox is recreated

- **WHEN** the exact declared sandbox exists in a stopped state and recreation is requested with or without force
- **THEN** `up` force-removes it without an additional confirmation prompt
- **AND** creates and enters its replacement from the current declaration

#### Scenario: Missing sandbox follows creation path

- **WHEN** the exact declared sandbox is absent and recreation is requested with or without force
- **THEN** `up` creates and enters the sandbox through the ordinary missing-sandbox path

#### Scenario: Force without recreation is rejected

- **WHEN** lifecycle is requested with force enabled and recreation disabled
- **THEN** `up` reports the invalid option combination before configuration validation or sandbox inspection
- **AND** does not remove, create, or enter a sandbox

#### Scenario: Recreation state inspection fails

- **WHEN** Docker Sandboxes cannot report a recognized lifecycle state for the exact declared sandbox
- **THEN** `up` reports the inspection failure
- **AND** does not remove, create, or enter a sandbox
- **AND** exits with a non-zero status

#### Scenario: Recreation removal fails

- **WHEN** Docker Sandboxes cannot force-remove an approved, forced, or stopped declared sandbox
- **THEN** its diagnostic output remains visible to the user
- **AND** `up` does not create or enter a sandbox
- **AND** exits with a non-zero status

#### Scenario: Recreated sandbox starts with new state

- **WHEN** forced removal succeeds and `up` creates a direct-mode replacement sandbox
- **THEN** state from the removed sandbox is unavailable in the replacement
- **AND** files in the repository's host workspace remain intact

#### Scenario: Recreated clone-mode sandbox discards private work

- **WHEN** forced removal succeeds for a clone-mode sandbox whose private clone contains work not stored elsewhere
- **THEN** that private work is unavailable in the replacement
- **AND** files in the repository's host workspace remain intact

### Requirement: Missing sandbox is created and entered from the declaration

When the declared sandbox name is absent, `up` SHALL provision a sandbox under that name with the declaration directory as its repository path, the effective workspace mode, and every selected kit in declaration order, and SHALL then enter its agent session interactively. `direct` mode SHALL expose the host repository workspace directly. `clone` mode SHALL request Docker Sandboxes' private clone of that repository while keeping the host repository files outside the agent's writable workspace. Provisioning and interactive entry SHALL remain separate lifecycle operations so a declared network rule can be applied between them. Remote kits SHALL use their linked execution references, and local kits SHALL use their safely resolved absolute host paths. Docker Sandboxes SHALL remain authoritative for whether the repository can be provisioned in the selected mode.

#### Scenario: Missing sandbox uses direct mode

- **WHEN** the declared sandbox is absent and its effective workspace mode is `direct`
- **THEN** the provisioning operation uses the declaration directory as a directly mounted workspace

#### Scenario: Missing sandbox uses clone mode

- **WHEN** the declared sandbox is absent and its effective workspace mode is `clone`
- **THEN** the provisioning operation requests a private clone of the repository at the declaration directory
- **AND** does not expose that host repository as the agent's writable workspace

#### Scenario: Missing sandbox uses a Git kit

- **WHEN** the declared sandbox is absent and selects a Git kit
- **THEN** the provisioning operation includes the linked Git execution reference
- **AND** includes the declared name, agent, repository path, and effective workspace mode

#### Scenario: Missing sandbox uses an OCI kit

- **WHEN** the declared sandbox is absent and selects an OCI kit
- **THEN** the provisioning operation includes the linked OCI execution reference with its declared version

#### Scenario: Missing sandbox uses a local kit

- **WHEN** the declared sandbox is absent and selects a valid local kit
- **THEN** the provisioning operation includes the canonical absolute path produced by local-kit resolution

#### Scenario: Multiple kits are selected

- **WHEN** a missing sandbox declares multiple valid kit selections
- **THEN** every kit is passed to Docker Sandboxes in declaration order

#### Scenario: Provisioned sandbox is entered

- **WHEN** provisioning succeeds and any declared network rule has been applied
- **THEN** `up` enters the declared sandbox
- **AND** the agent session is entered interactively

#### Scenario: Docker cannot provision the sandbox

- **WHEN** Docker Sandboxes rejects or cannot complete provisioning, including because the repository cannot be cloned
- **THEN** its diagnostic output remains visible to the user
- **AND** `up` does not attach to a sandbox
- **AND** exits with a non-zero status

### Requirement: Up initializes newly created sandboxes before agent entry

After validation, sandbox provisioning, and application of any declared network resources, `up` SHALL execute every declared `sandbox.hooks.initialize` command in order inside the exact newly created sandbox with its effective workspace as the working directory. In direct mode that workspace SHALL be the host-mounted repository; in clone mode it SHALL be the private clone. The commands SHALL run when `up` creates a missing sandbox, including creation following recreation, and agent entry SHALL begin only after every command succeeds. They SHALL NOT run for an existing running or stopped sandbox, or during `validate`, `down`, or `destroy`.

#### Scenario: Missing sandbox has initialization commands

- **WHEN** `up` creates a missing sandbox with declared initialization commands
- **THEN** it applies any declared network resources to the created sandbox
- **AND** executes the commands inside that sandbox and its effective workspace in declaration order
- **AND** enters the agent only after every command succeeds

#### Scenario: Direct-mode initialization changes host files

- **WHEN** an initialization command changes files in a direct-mode workspace
- **THEN** those changes are immediately visible in the repository's host workspace

#### Scenario: Clone-mode initialization changes private files

- **WHEN** an initialization command changes files in a clone-mode workspace
- **THEN** those changes remain in the private clone until transferred through Git
- **AND** the corresponding host repository files remain unchanged

#### Scenario: Recreated sandbox has initialization commands

- **WHEN** `up --recreate` successfully replaces an existing sandbox with one whose declaration contains initialization commands
- **THEN** it executes every command inside the replacement sandbox before entering the agent

#### Scenario: Stopped sandbox is not reinitialized

- **WHEN** ordinary `up` targets an existing stopped sandbox whose declaration contains initialization commands
- **THEN** it enters the sandbox without executing those commands

#### Scenario: Running sandbox is not reinitialized

- **WHEN** ordinary `up` targets an existing running sandbox whose declaration contains initialization commands
- **THEN** it enters the sandbox without executing those commands

#### Scenario: Initialization declaration changes for an existing sandbox

- **WHEN** initialization commands change after a sandbox has been created
- **THEN** ordinary `up` does not reconcile or execute the changed commands
- **AND** the user must request recreation to apply them to a replacement

#### Scenario: No initialization commands are declared

- **WHEN** a valid declaration has no initialization commands
- **THEN** `up` preserves its existing create, recreate, or enter behavior

#### Scenario: Another lifecycle command is invoked

- **WHEN** a user invokes `validate`, `down`, or `destroy` for a declaration that contains initialization commands
- **THEN** none of the declared commands is executed

### Requirement: Initialization failure rolls back the new sandbox

If an initialization command exits unsuccessfully, cannot be executed, or is cancelled, `up` SHALL stop before executing later commands or entering the agent, return a non-zero result with context identifying the failed command, and remove the sandbox created by that invocation plus its declared sandbox-scoped resources. If cleanup also fails, `up` SHALL report both failures while preserving the initialization failure as the primary cause. Cleanup SHALL NOT revert changes already made to a direct host-mounted workspace; changes confined to a successfully removed private clone SHALL be discarded with that sandbox.

#### Scenario: A command fails

- **WHEN** an initialization command returns a non-zero result
- **THEN** its diagnostic output remains visible
- **AND** `up` identifies the failed command and returns a non-zero result
- **AND** does not execute any later command or enter the agent
- **AND** removes the newly created sandbox and its declared scoped resources

#### Scenario: Cleanup also fails

- **WHEN** initialization fails and cleanup of the new sandbox or a declared scoped resource also fails
- **THEN** `up` reports the initialization and cleanup failures
- **AND** preserves the initialization failure as the primary cause

#### Scenario: A retry follows initialization rollback

- **WHEN** initialization failed, cleanup succeeded, and the user invokes ordinary `up` again
- **THEN** `up` creates a new sandbox and reruns the complete initialization sequence

#### Scenario: A failed command changed the host workspace

- **WHEN** an initialization command changes the direct host-mounted workspace before failing
- **THEN** sandbox rollback does not claim to revert that host filesystem change

#### Scenario: A failed command changed the private clone

- **WHEN** an initialization command changes only the clone-mode private workspace before failing and sandbox cleanup succeeds
- **THEN** the private changes are discarded with the sandbox
- **AND** the host repository remains unchanged

## ADDED Requirements

### Requirement: Existing sandbox workspace mode is not reconciled

Workspace mode SHALL be a creation-time setting. Ordinary `up` SHALL enter an existing exact-name sandbox without comparing its actual workspace mode to the current declaration. A changed mode SHALL take effect only after `up --recreate` successfully replaces the sandbox.

#### Scenario: Workspace mode changes after creation

- **WHEN** the declared workspace mode differs from the mode used to create an existing sandbox and recreation is not requested
- **THEN** `up` enters the existing sandbox without changing or comparing its mode

#### Scenario: Workspace mode changes before recreation

- **WHEN** the declared workspace mode differs from the existing sandbox and recreation is successfully requested
- **THEN** the replacement is provisioned using the currently declared mode

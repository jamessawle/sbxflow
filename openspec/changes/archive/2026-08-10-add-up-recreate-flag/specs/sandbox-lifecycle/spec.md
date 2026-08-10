## ADDED Requirements

### Requirement: Up can replace an existing sandbox from the current declaration

When recreation is requested and the exact declared sandbox exists, `up` SHALL
force-remove that sandbox without confirmation before creating and entering a
replacement from the validated current declaration. The removal SHALL remain
limited to the exact declared name and SHALL permanently discard the removed
sandbox's persisted state without deleting files from the repository's host
workspace.

#### Scenario: Existing sandbox is recreated

- **WHEN** the exact declared sandbox exists and a user invokes
  `sbxflow up --recreate`
- **THEN** `up` invokes `sbx rm --force` with that sandbox name
- **AND** Docker does not ask for removal confirmation
- **AND** after successful removal, `up` creates and enters a sandbox with the
  declared name, agent, repository workspace, selected kits, and derived trust

#### Scenario: Existing stopped sandbox is recreated

- **WHEN** the exact declared sandbox exists in a stopped state and recreation
  is requested
- **THEN** `up` force-removes it rather than restarting and entering it
- **AND** creates and enters its replacement from the current declaration

#### Scenario: Recreation removal fails

- **WHEN** Docker Sandboxes cannot force-remove the existing declared sandbox
- **THEN** its diagnostic output remains visible to the user
- **AND** `up` does not create or enter a sandbox
- **AND** exits with a non-zero status

#### Scenario: Recreated sandbox starts with new state

- **WHEN** forced removal succeeds and `up` creates the replacement sandbox
- **THEN** state from the removed sandbox is unavailable in the replacement
- **AND** files in the repository's host workspace remain intact

## MODIFIED Requirements

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

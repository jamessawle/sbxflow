## ADDED Requirements

### Requirement: Up initializes newly created sandboxes before agent entry

After validation, sandbox provisioning, and application of any declared network
resources, `up` SHALL execute every declared `sandbox.hooks.initialize` command
in order inside the exact newly created sandbox with the declared repository
workspace as its working directory. The commands SHALL run when `up` creates a
missing sandbox, including creation following recreation, and agent entry SHALL
begin only after every command succeeds. They SHALL NOT run for an existing
running or stopped sandbox, or during `validate`, `down`, or `destroy`.

#### Scenario: Missing sandbox has initialization commands

- **WHEN** `up` creates a missing sandbox with declared initialization commands
- **THEN** it applies any declared network resources to the created sandbox
- **AND** executes the commands inside that sandbox and workspace in declaration
  order
- **AND** enters the agent only after every command succeeds

#### Scenario: Recreated sandbox has initialization commands

- **WHEN** `up --recreate` successfully replaces an existing sandbox with one
  whose declaration contains initialization commands
- **THEN** it executes every command inside the replacement sandbox before
  entering the agent

#### Scenario: Stopped sandbox is not reinitialized

- **WHEN** ordinary `up` targets an existing stopped sandbox whose declaration
  contains initialization commands
- **THEN** it enters the sandbox without executing those commands

#### Scenario: Running sandbox is not reinitialized

- **WHEN** ordinary `up` targets an existing running sandbox whose declaration
  contains initialization commands
- **THEN** it enters the sandbox without executing those commands

#### Scenario: Initialization declaration changes for an existing sandbox

- **WHEN** initialization commands change after a sandbox has been created
- **THEN** ordinary `up` does not reconcile or execute the changed commands
- **AND** the user must request recreation to apply them to a replacement

#### Scenario: No initialization commands are declared

- **WHEN** a valid declaration has no initialization commands
- **THEN** `up` preserves its existing create, recreate, or enter behavior

#### Scenario: Another lifecycle command is invoked

- **WHEN** a user invokes `validate`, `down`, or `destroy` for a declaration that
  contains initialization commands
- **THEN** none of the declared commands is executed

### Requirement: Initialization command execution is literal and observable

Each initialization command SHALL be passed to Docker Sandboxes as its declared
argument vector without evaluation by a host shell. Its standard output and
standard error SHALL be attached to the corresponding CLI streams, its standard
input SHALL be non-interactive, and execution SHALL honor command cancellation
so progress and diagnostics remain visible without delaying lifecycle shutdown.

#### Scenario: Command uses explicit shell evaluation

- **WHEN** a command vector declares a shell executable, an evaluation option,
  and a script argument
- **THEN** those arguments are passed unchanged for execution inside the sandbox
- **AND** the declared shell performs the requested evaluation

#### Scenario: Command contains shell metacharacters without a shell

- **WHEN** an argument contains shell metacharacters but the command vector does
  not invoke a shell
- **THEN** the argument is passed literally without host-shell interpretation

#### Scenario: Command writes output

- **WHEN** an initialization command writes to standard output or standard error
- **THEN** the output is observable on the matching `sbxflow up` CLI stream

#### Scenario: Command attempts to read standard input

- **WHEN** an initialization command attempts to read standard input
- **THEN** it does not consume the CLI input reserved for lifecycle interaction

### Requirement: Initialization failure rolls back the new sandbox

If an initialization command exits unsuccessfully, cannot be executed, or is
cancelled, `up` SHALL stop before executing later commands or entering the agent,
return a non-zero result with context identifying the failed command, and remove
the sandbox created by that invocation plus its declared sandbox-scoped
resources. If cleanup also fails, `up` SHALL report both failures while preserving
the initialization failure as the primary cause. Cleanup does not revert changes
already made to the host-mounted workspace.

#### Scenario: A command fails

- **WHEN** an initialization command returns a non-zero result
- **THEN** its diagnostic output remains visible
- **AND** `up` identifies the failed command and returns a non-zero result
- **AND** does not execute any later command or enter the agent
- **AND** removes the newly created sandbox and its declared scoped resources

#### Scenario: Cleanup also fails

- **WHEN** initialization fails and cleanup of the new sandbox or a declared
  scoped resource also fails
- **THEN** `up` reports the initialization and cleanup failures
- **AND** preserves the initialization failure as the primary cause

#### Scenario: A retry follows initialization rollback

- **WHEN** initialization failed, cleanup succeeded, and the user invokes
  ordinary `up` again
- **THEN** `up` creates a new sandbox and reruns the complete initialization
  sequence

#### Scenario: A failed command changed the host workspace

- **WHEN** an initialization command changes the host-mounted workspace before
  failing
- **THEN** sandbox rollback does not claim to revert that host filesystem change

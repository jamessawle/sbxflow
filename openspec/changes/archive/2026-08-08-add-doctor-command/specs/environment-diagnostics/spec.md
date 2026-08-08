## Purpose

Defines how users assess whether Docker Sandboxes is compatible, healthy, and
safely configured on the local system without inspecting a repository
declaration.

## ADDED Requirements

### Requirement: Doctor is repository-independent

The `doctor` command SHALL assess the local Docker Sandboxes installation and
global configuration without discovering, reading, or validating
`sbxflow.yaml`.

#### Scenario: Run outside a configured repository

- **WHEN** a user runs `sbxflow doctor` from a directory with no
  `sbxflow.yaml`
- **THEN** the command performs its system checks without reporting a missing
  declaration

#### Scenario: Declaration does not affect diagnostics

- **WHEN** a user runs `sbxflow doctor` in a directory containing an invalid
  `sbxflow.yaml`
- **THEN** the command does not read or report errors from that file

### Requirement: Doctor checks sbx compatibility

The `doctor` command SHALL identify whether `sbx` is executable and whether its
version falls within the range supported by sbxflow.

#### Scenario: sbx is unavailable

- **WHEN** `sbx` cannot be found or executed
- **THEN** the command reports the unavailable installation as a failure
- **AND** exits with a non-zero status

#### Scenario: sbx version is unsupported

- **WHEN** the installed `sbx` version is outside sbxflow's supported range
- **THEN** the command reports the installed and supported versions
- **AND** exits with a non-zero status

#### Scenario: sbx version is compatible

- **WHEN** the installed `sbx` version is within sbxflow's supported range
- **THEN** the command reports the installation as compatible

### Requirement: Docker diagnostics are summarized

The `doctor` command SHALL run Docker Sandboxes' machine-readable diagnostics
and report only their aggregate passed, warned, failed, and skipped totals.
Docker's individual check names, details, and remediation messages SHALL NOT be
reproduced by sbxflow.

#### Scenario: Diagnostics complete successfully

- **WHEN** `sbx diagnose` returns a supported summary with no failed checks
- **THEN** the command reports the aggregate totals
- **AND** directs the user to `sbx diagnose` for detailed results

#### Scenario: Docker diagnostics contain failures

- **WHEN** the diagnostic summary contains one or more failed checks
- **THEN** the command reports the aggregate totals without reproducing the
  individual failures
- **AND** exits with a non-zero status

#### Scenario: Diagnostic output cannot be summarized

- **WHEN** `sbx diagnose` cannot be executed or its output does not contain a
  supported summary
- **THEN** the command reports that Docker diagnostics could not be summarized
- **AND** directs the user to run `sbx diagnose`
- **AND** exits with a non-zero status

### Requirement: Global network policy initialization is checked

The `doctor` command SHALL determine whether Docker Sandboxes has an initialized
global network policy without judging a valid policy solely by its selected
preset.

#### Scenario: Global policy is initialized

- **WHEN** Docker Sandboxes reports an initialized global network policy
- **THEN** the command reports the policy check as passed

#### Scenario: Deny-all policy has no network rules

- **WHEN** Docker Sandboxes reports an active global local policy whose
  deny-all preset is represented without explicit network rules
- **THEN** the command reports the policy check as passed
- **AND** does not infer initialization solely from network-rule presence

#### Scenario: Global policy is not initialized

- **WHEN** Docker Sandboxes reports no initialized global network policy
- **THEN** the command reports a warning with non-mutating initialization
  guidance

#### Scenario: Organisation policy is active

- **WHEN** Docker Sandboxes reports organisation-managed policy
- **THEN** the command identifies the managed source without recommending a
  local override

#### Scenario: Global policy cannot be inspected

- **WHEN** the policy query cannot return machine-readable policy state
- **THEN** the command warns that global network policy could not be inspected
- **AND** does not infer daemon health from the aggregate Docker diagnostic
  totals

### Requirement: Unrestricted remote kit sources are reported

The `doctor` command SHALL inspect the effective global `kit.allowedSources`
setting and warn when its value contains an entry exactly equal to `"*"`.

#### Scenario: Remote kit sources are restricted

- **WHEN** every effective allowed-source entry is a restricted publisher
  prefix
- **THEN** the command reports the kit-source configuration as passed

#### Scenario: Remote kit sources contain a wildcard

- **WHEN** any effective allowed-source entry equals `"*"`
- **THEN** the command warns that remote kit sources are unrestricted
- **AND** identifies whether the value came from a local override, environment,
  default, or organisation-managed source when that information is available

#### Scenario: Wildcard is organisation-managed

- **WHEN** the unrestricted value is organisation-managed
- **THEN** remediation directs the user to their Docker Sandboxes administrator
  rather than suggesting a local override

### Requirement: Doctor checks do not modify the system

The `doctor` command SHALL inspect and report system state without changing
Docker Sandboxes settings, initializing policy, starting the daemon, or fixing
reported problems.

#### Scenario: A problem has a known remediation

- **WHEN** a check identifies a problem with a known remediation
- **THEN** the command reports guidance without performing the remediation

### Requirement: Doctor completes independent checks

The `doctor` command SHALL continue checks that do not depend on a failed
prerequisite and SHALL identify dependent checks as skipped.

#### Scenario: sbx is unavailable

- **WHEN** the compatibility check cannot locate `sbx`
- **THEN** checks requiring `sbx` are reported as skipped
- **AND** any independent checks still run

### Requirement: Doctor exit status distinguishes failures from warnings

The `doctor` command SHALL exit successfully when checks pass or only produce
configuration warnings, and SHALL exit with a non-zero status when any required
health or compatibility check fails.

#### Scenario: Only recommendations are reported

- **WHEN** all required checks pass and one or more configuration checks warn
- **THEN** the command exits successfully

#### Scenario: Required check fails

- **WHEN** any required health or compatibility check fails
- **THEN** the command exits with a non-zero status

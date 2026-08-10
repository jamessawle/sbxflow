# sbxflow

> sbxflow applies a repository's declared Docker Sandbox configuration and
> lifecycle.

`sbxflow` is a configuration-driven CLI for creating and entering a repository's
[Docker Sandbox](https://docs.docker.com/ai/sandboxes/). It validates the
declared agent and ordered set of kits before creating a missing sandbox or
interactively entering an existing one.

> [!NOTE]
> sbxflow is in early development. Its foundational CLI and configuration
> validation and the interactive `up` lifecycle command are available. Existing
> sandboxes are not yet reconciled with declaration changes. The
> state-preserving `down` and permanent `destroy` commands are available.

## Configuration

A repository declares its sandbox in `sbxflow.yaml`:

```yaml
version: 1

sandbox:
  name: my-project
  agent: codex

  kits:
    sources:
      community:
        type: git
        repo: https://github.com/docker/sbx-kits-contrib.git
        ref: v0.14.0

    use:
      - source: community
        kit: mise
```

Kit sources are named once and selected through the ordered `use` list. The
source type determines how each selection is resolved:

- Git sources declare a repository and ref; `kit` selects a directory within
  the repository.
- OCI sources declare a registry base; `kit` and `version` select an artifact.
- Local sources declare a root directory; `kit` selects a directory or ZIP file
  beneath it.

`agent` must be one of the identifiers supported across sbxflow's compatible
Docker Sandboxes range: `claude`, `codex`, `copilot`, `cursor`, `docker-agent`,
`droid`, `gemini`, `kiro`, `opencode`, or `shell`. The schema pins this set so
validation remains deterministic and does not require an installed `sbx`;
`shell` supports manual or custom agent setups.

See [`examples`](examples/) for complete configurations, including the
configuration derived from a working `personal-site` sandbox script.

The published [Draft 2020-12 JSON Schema](schema/sbxflow.schema.json) is the
structural contract used by the CLI and can also be used by editors and CI
tooling.

## Foundational CLI

Run the CLI from the repository with Mise:

```text
mise exec -- go run ./cmd/sbxflow
```

The root command displays help when invoked without arguments. The same help is
available through `sbxflow help`, `sbxflow -h`, and `sbxflow --help`.

Use `sbxflow -v` or `sbxflow --version` to display the build identity. Local
builds report an explicit development version. A `version` subcommand is not
available.

Completion and man-page commands are intentionally absent from the foundational
executable. The repository-aware `validate`, `up`, `down`, and `destroy`
commands are available.

### `sbxflow doctor`

Assess whether the local Docker Sandboxes installation is compatible, healthy,
and safely configured:

```text
sbxflow doctor
```

`doctor` can run from any directory. It does not discover or read
`sbxflow.yaml`; an absent or invalid repository declaration does not affect its
results. The initial release supports `sbx` versions from v0.35.0 up to, but not
including, v0.38.0.

The command runs these ordered system checks:

- Confirm that `sbx` is installed and its version is supported.
- Summarize only the aggregate pass, warning, failure, and skipped totals from
  Docker's machine-readable diagnostics. Run `sbx diagnose` for individual
  results and remediation.
- Check whether a global network policy is initialized or
  organisation-managed without grading the selected policy preset.
- Warn when the effective global `kit.allowedSources` value contains the
  unrestricted `"*"` entry, with guidance based on the setting's reported
  source.

The checks are inspection-only: sbxflow does not initialize policy, change
settings, start the daemon, or fix reported problems. Warnings and skipped
advisory checks retain a successful exit status; a failed compatibility or
Docker health check exits non-zero.

### `sbxflow validate`

Validate the nearest repository declaration before sandbox creation or entry
relies on it:

```text
sbxflow validate
```

`validate` searches the current directory and then its ancestors for the
nearest `sbxflow.yaml`. It rejects malformed or multi-document YAML, duplicate
mapping keys, unsupported versions, unknown fields, invalid source shapes,
duplicate selections, unknown source references, and source-specific selection
errors.

Git and OCI references are normalized and checked offline. Validation does not
clone Git repositories, pull OCI artifacts, test credentials, or check remote
availability. Kits selected from local sources are the only artifacts passed to
Docker: after rejecting URI-like, absolute, unavailable, and source-escaping
paths, sbxflow runs `sbx kit validate <absolute-path>` for each selected local
directory or ZIP in declaration order.

Successful output identifies the declaration and reports the least-privilege
settings future lifecycle commands will apply to their own Docker Sandbox
subprocesses:

```text
Declaration: /work/my-project/sbxflow.yaml

Derived State:
  Kit:
    Allowed Sources:
      - docker.io/
      - github.com/docker/sbx-kits-contrib
    Local Kits Allowed: true

Validation:
  State: pass
  Findings: []
```

The command is inspection-only. It does not create or alter a sandbox, resolve
remote kits, or modify global Docker Sandbox settings. Structural, semantic,
path, or local-kit failures are reported on standard error and produce a
non-zero exit status. Configurations without selected local sources do not
require or invoke `sbx`.

The complete report is written to standard output on success or standard error
on failure. If validation stops before trust derivation, it reports
`Derived State: unavailable`; failures appear under `Validation.Findings`. This
YAML-like presentation is intended for people rather than as a stable
machine-readable interface.

### `sbxflow up`

Validate the nearest repository declaration, then create or enter its named
Docker Sandbox and attach the declared agent to the current terminal:

```text
sbxflow up
sbxflow up --recreate
```

`up` searches the current directory and its ancestors for `sbxflow.yaml`, using
the declaration directory as the repository workspace. Complete validation
runs before Docker inspects or changes sandbox state. On success, `up` writes
`Configuration valid: <declaration path>` to standard error before continuing
with Docker Sandbox inspection and lifecycle execution.

When the declared name does not exist, `up` creates it with the declared agent,
workspace, and every selected kit in declaration order. Git and OCI selections
use their linked remote references; local selections use their validated
canonical host paths. The derived kit-source allowlist and local-kit permission
are applied only to the interactive `sbx run` process, without modifying global
Docker Sandbox settings.

When the name already exists, `up` enters it through Docker Sandboxes. Docker
attaches when it is running or restarts it when stopped, and verifies that it
uses the declared agent. Existing sandbox workspace and kits are not inspected
or reconciled: changing their declaration does not update, recreate, or grade
the existing sandbox.

Use `up --recreate` to replace an existing exact-name sandbox from the current
declaration. After complete validation and exact-name lookup succeed, sbxflow
runs `sbx rm --force <declared-name>` without confirmation, then creates and
enters the replacement with the declared workspace, agent, kits, and derived
trust. There is no `-r` shorthand. If the declared sandbox is absent, the flag
follows the ordinary create-and-enter path without attempting removal.

Recreation permanently removes the sandbox's installed tools, Docker images,
agent history, configuration changes, and other persisted state, including for
a stopped sandbox. It does not delete files from the repository's host
workspace. A lookup or removal failure stops before creation or entry. Removal
and replacement are not transactional: if removal succeeds but creation fails,
the old sandbox remains absent and a later ordinary `up` can create it from the
current declaration.

The agent process receives the terminal's standard input, standard output,
standard error, and normal signals directly. `up` imposes no session timeout and
returns the Docker process result.

`up` accepts no positional arguments. Its only command-specific flag is the
long-form `--recreate` option.

### `sbxflow down`

```text
sbxflow down
```

Stop the sandbox declared by the nearest repository configuration without
removing it. `down` resolves only the supported configuration version and exact
non-empty `sandbox.name`; it does not require the agent, kits, local paths,
linked references, or derived trust to pass complete validation. The identity
document must still be safe and unambiguous: malformed or multi-document YAML,
duplicate keys, unsupported or missing versions, and invalid sandbox names are
rejected before Docker is invoked.

After resolving the target, `down` uses Docker Sandboxes' quiet name listing and
matches the declared name exactly. If the sandbox exists, sbxflow runs
`sbx stop <declared-name>`. If it is absent, `down` succeeds without invoking
`sbx stop`; similarly named sandboxes are never selected. Docker handles an
already stopped sandbox idempotently.

Stopping preserves the sandbox's installed tools, Docker images, agent history,
configuration changes, and other persisted state. A later `up` restarts and
enters that existing sandbox. Docker's stop output and diagnostics remain
attached to the terminal, and its exit status is preserved.

`down` accepts no positional arguments or command-specific flags.

### `sbxflow destroy`

```text
sbxflow destroy
sbxflow destroy --force
```

Permanently remove the sandbox declared by the nearest repository
configuration and all of its persisted Docker Sandbox state. Like `down`,
`destroy` resolves only the supported configuration version and exact non-empty
`sandbox.name`; unrelated agent or kit errors do not prevent removal. It lists
Docker Sandboxes by quiet name and removes only an exact match, so similarly
named sandboxes are never selected. If the declared sandbox is absent,
`destroy` succeeds without invoking removal.

By default, `destroy` runs `sbx rm <declared-name>` with the command's input,
output, and error streams attached. Docker owns the confirmation prompt and
active-session protection, and sbxflow preserves its output and exit status. If
the user declines or Docker refuses removal, sbxflow does not retry with force.

Use `--force` or `-f` to run `sbx rm --force <declared-name>`, skipping Docker's
confirmation and permitting removal while the sandbox has an active session.
Force does not broaden the target beyond the exact declared name. `destroy`
accepts no positional arguments or other command-specific flags.

Successful destruction permanently removes the sandbox's installed tools,
Docker images, agent history, configuration changes, and other persisted state.
The repository's host workspace remains intact; Docker-managed worktrees remain
subject to Docker's removal behavior. A later `up` creates a new sandbox from
the current declaration. Use `down` instead when the sandbox and its state
should be preserved for a later session.

## Kit source trust

sbxflow derives Docker's kit-source settings from the kits selected in `use`:

- Docker Hub remains allowed by default.
- Each selected Git or OCI source adds its narrowest required remote prefix.
- Local kits are enabled only when a selected kit uses a local source.

The derived settings apply only to the `sbx` commands run by sbxflow. sbxflow
does not modify the host's global Docker Sandbox settings. Host and organisation
policy can still impose stricter controls.

## Lifecycle scope

The initial lifecycle is deliberately limited to `up`, `down`, and `destroy`.
Configurable lifecycle hooks are not part of the first configuration format.

## Development

Development tools are managed with [Mise](https://mise.jdx.dev/):

```text
mise run setup      Install tools and activate the Git hooks
mise run fmt        Format Markdown and Go files
mise run validate   Run all repository checks
```

The pre-commit hook rejects whitespace errors and Markdown that has not been
formatted. Run `mise run fmt`, then stage the resulting changes before
committing.

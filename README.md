# sbxflow

> sbxflow applies a repository's declared Docker Sandbox configuration and
> lifecycle.

`sbxflow` is a configuration-driven CLI for creating and managing a repository's
[Docker Sandbox](https://docs.docker.com/ai/sandboxes/). It keeps the sandbox
running with the declared agent and ordered set of kits, recreating it when its
build inputs change.

> [!NOTE]
> sbxflow is in early development. Its foundational CLI is available, but the
> lifecycle commands described below are not implemented yet.

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

See [`examples`](examples/) for complete configurations, including the
configuration derived from a working `personal-site` sandbox script.

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

Completion, man-page, and lifecycle commands are intentionally absent from the
foundational executable.

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

## Planned lifecycle commands

```text
sbxflow up
sbxflow down
sbxflow destroy
sbxflow validate
```

### `sbxflow up` (planned)

Bring the declared sandbox up:

- Create it when it does not exist.
- Start it when it is stopped.
- Recreate it when its build inputs are out of date.
- Leave it unchanged when it is current and running.

### `sbxflow down` (planned)

Stop the sandbox while preserving its installed tools, Docker images, agent
history, and other sandbox state.

### `sbxflow destroy` (planned)

Remove the sandbox and its persisted state completely. Repository files remain
on the host.

### `sbxflow validate` (planned)

Validate the configuration's syntax and semantics without changing sandbox
state. Validation also reports the effective kit-source trust derived from the
ordered selections.

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

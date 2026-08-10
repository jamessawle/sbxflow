# sbxflow

> sbxflow applies a repository's declared Docker Sandbox configuration and
> lifecycle.

`sbxflow` is a configuration-driven CLI for creating and managing a repository's
[Docker Sandbox](https://docs.docker.com/ai/sandboxes/). It keeps the sandbox
running with the declared agent and ordered set of kits, recreating it when its
build inputs change.

> [!NOTE]
> sbxflow is in early development. Its foundational CLI and configuration
> validation are available, but the lifecycle commands described below are not
> implemented yet.

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

Completion, man-page, and lifecycle commands are intentionally absent from the
foundational executable. The repository-aware `validate` command is available.

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

Validate the nearest repository declaration before lifecycle commands rely on
it:

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
[VALID] local kit repository/tooling: /work/my-project/kits/tooling
kit.allowedSources: [docker.io/, github.com/docker/sbx-kits-contrib]
kit.allowLocalKits: true
```

The command is inspection-only. It does not create or alter a sandbox, resolve
remote kits, or modify global Docker Sandbox settings. Structural, semantic,
path, or local-kit failures are reported on standard error and produce a
non-zero exit status. Configurations without selected local sources do not
require or invoke `sbx`.

## Planned lifecycle commands

```text
sbxflow up
sbxflow down
sbxflow destroy
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

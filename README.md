# sbxflow

[![Latest release](https://img.shields.io/github/v/release/jamessawle/sbxflow)](https://github.com/jamessawle/sbxflow/releases/latest)
[![License](https://img.shields.io/github/license/jamessawle/sbxflow)](LICENSE)
[![Build status](https://github.com/jamessawle/sbxflow/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/jamessawle/sbxflow/actions/workflows/ci.yml)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=jamessawle_sbxflow&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=jamessawle_sbxflow)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/jamessawle/sbxflow/badge)](https://scorecard.dev/viewer/?uri=github.com/jamessawle/sbxflow)
[![OpenSSF Best Practices](https://www.bestpractices.dev/projects/14024/badge)](https://www.bestpractices.dev/projects/14024)

`sbxflow` applies a repository's declared [Docker Sandbox](https://docs.docker.com/ai/sandboxes/) configuration and lifecycle. It validates the declared agent and ordered set of kits, then creates, enters, stops, or removes the repository's sandbox.

## Configuration

Declare the sandbox in `sbxflow.yaml`:

```yaml
version: 1

sandbox:
  name: my-project
  agent: codex

  network:
    allowedHosts:
      - api.github.com
      - registry.npmjs.org:443
      - "*.githubusercontent.com"

  kits:
    sources:
      community:
        type: git
        repo: https://github.com/docker/sbx-kits-contrib.git
        ref: v0.16.0

    use:
      - source: community
        kit: mise
```

Kit sources can be Git repositories, OCI registries, or local directories. Sources are named once and selected in order through `kits.use`. See the [examples](examples/) for complete configurations and the published [JSON Schema](schema/sbxflow.schema.json) for the structural contract.

## Commands

Run the CLI from this repository with Mise:

```text
mise exec -- go run ./cmd/sbxflow <command>
```

For example, run `mise exec -- go run ./cmd/sbxflow doctor` to execute the
environment checks without installing sbxflow first. Use `sbxflow -v` or
`sbxflow --version` to display the build identity. Root and command help are
available through `--help`.

## Installation

### Homebrew

Install the latest published release on macOS with Homebrew:

```text
brew install --cask jamessawle/tap/sbxflow
```

Verify the installation:

```text
sbxflow --version
sbxflow doctor
```

Upgrade or uninstall it with Homebrew:

```text
brew upgrade --cask sbxflow
brew uninstall --cask sbxflow
```

Linux and Windows users install from the release archives below.

### Direct download

Release archives for macOS, Linux, and Windows are available from the
[GitHub releases page](https://github.com/jamessawle/sbxflow/releases). Download
the archive for your operating system and architecture together with
`checksums.txt`. For example, on Linux amd64:

```sh
version=0.1.5
archive="sbxflow_${version}_linux_amd64.tar.gz"
base_url="https://github.com/jamessawle/sbxflow/releases/download/v${version}"
curl -LO "${base_url}/${archive}"
curl -LO "${base_url}/checksums.txt"
gh attestation verify "${archive}" --repo jamessawle/sbxflow
gh attestation verify checksums.txt --repo jamessawle/sbxflow
grep " ${archive}$" checksums.txt | sha256sum --check
tar -xzf "${archive}"
sudo install -m 0755 sbxflow /usr/local/bin/sbxflow
sbxflow --version
```

Attestation verification requires network access and a current GitHub CLI. It
authenticates the repository and workflow that produced each downloaded file;
the checksum check independently confirms that the archive matches the
published manifest. Each release also includes
`sbxflow-provenance.intoto.jsonl` so automated supply-chain checks can discover
the provenance directly from its assets.

On macOS, select `darwin_arm64` for Apple silicon or `darwin_amd64` for an
Intel Mac and replace the checksum command with:

```sh
grep " ${archive}$" checksums.txt | shasum -a 256 --check
```

Windows amd64 releases use a `.zip` archive. Compare the SHA-256 value reported
by `Get-FileHash <archive> -Algorithm SHA256` with the corresponding line in
`checksums.txt`, extract `sbxflow.exe`, and place it in a directory on `PATH`.

To upgrade a direct installation, verify and install the newer release over the
existing executable. To uninstall it, remove the installed `sbxflow` executable.
These operations do not remove Docker sandboxes; use `sbxflow destroy` first if
you also intend to remove a repository's declared sandbox and persisted state.

### `sbxflow doctor`

Checks whether Docker Sandboxes is installed at a compatible version, summarizes Docker's diagnostics, and reports global network and kit-source policy posture. It does not read `sbxflow.yaml` or change the host configuration.

sbxflow v0.1.5 supports `sbx` versions from v0.35.0 up to, but not including, v0.38.0.

### `sbxflow validate`

Finds the nearest `sbxflow.yaml`, validates its structure and semantics, derives the least-privilege kit-source settings, and validates selected local kits with Docker Sandboxes. Git and OCI references are checked offline; validation does not clone or pull them.

### `sbxflow up`

```text
sbxflow up
sbxflow up --recreate
sbxflow up --recreate --force
```

Validates the declaration, then creates and enters a missing sandbox or enters an existing one. An existing sandbox's workspace and kits are not inspected or reconciled when the declaration changes.

Use `--recreate` to replace an existing sandbox from the current declaration. Recreation permanently removes the sandbox's installed tools, Docker images, agent history, configuration changes, and other persisted state. A running sandbox requires confirmation; a stopped sandbox does not. The repository's host workspace is not deleted.

Use `--force` with `--recreate` to bypass confirmation for a running sandbox. This still permanently removes its persisted state and can terminate other attached terminal sessions. `--force` is not valid without `--recreate`.

### `sbxflow down`

Stops the declared sandbox while preserving its state. If the sandbox does not exist, the command succeeds without attempting to stop anything.

### `sbxflow destroy`

```text
sbxflow destroy
sbxflow destroy --force
```

Permanently removes the exact declared sandbox and its persisted state. Docker owns the default confirmation; `--force` or `-f` skips it and permits removal during an active session. The repository's host workspace is not deleted.

## Kit source trust

sbxflow derives Docker's kit-source settings from the selected kits:

- Docker Hub remains allowed by default.
- Each selected Git or OCI source adds its narrowest required remote prefix.
- Local kits are enabled only when selected.

The derived settings apply only to Docker Sandbox processes started by sbxflow. Host and organisation policy can impose stricter controls.

## Sandbox network access

Use the optional `sandbox.network.allowedHosts` list to declare additional
network resources needed by the sandbox. Entries are unique and remain in
declaration order. Each entry is a host, domain, wildcard subdomain, bracketed
IPv6 literal, or `**` for all hosts, with an optional `:port` suffix from 1 to
65535:

| Entry                     | Matches                       |
| ------------------------- | ----------------------------- |
| `api.github.com`          | that host on port 443         |
| `registry.npmjs.org:443`  | that host on an explicit port |
| `*.githubusercontent.com` | any subdomain                 |
| `[fd00::1]:8443`          | an IPv6 literal on a port     |
| `**`                      | all outbound hosts            |

Docker Sandboxes matches network requests by host and port, so a URL is not a
usable resource. `https://api.github.com` is accepted by `sbx policy allow` but
never matches any request, so sbxflow rejects it during validation instead. For
the same reason it rejects a port outside 1 to 65535 and a bracketed literal
that is not a well-formed IPv6 address.

Docker Sandboxes only accepts a sandbox-scoped rule for a sandbox that already
exists. When `up` creates a missing sandbox it therefore provisions the sandbox,
applies the rule, and only then starts the agent, so the rule is in force before
any agent traffic. If the rule cannot be applied, `up` removes the sandbox it
just created rather than entering it without the declared access.
Organisation-managed policy remains authoritative and can prevent a local rule
from taking effect. An ordinary `up` never reconciles an existing sandbox when
this list changes; use `up --recreate` to remove the currently declared resources
and apply the current declaration to the replacement.

Declared resources are owned by sbxflow. Both `destroy` and recreation remove the
sandbox first, then remove each currently declared resource. Docker Sandboxes
ordinarily discards a sandbox-scoped policy along with its sandbox, so this
cleanup is idempotent and a resource that is already gone is not an error. Avoid
manually modifying overlapping sandbox-scoped resources. If cleanup does fail
after the sandbox was removed, retry the resource reported in the error with:

```text
sbx policy rm network --sandbox <sandbox-name> --resource <resource>
```

## Versioning

sbxflow follows [Semantic Versioning](https://semver.org/). Before v1.0.0, minor releases may make incompatible changes to the CLI or configuration format; patch releases do not intentionally introduce incompatible changes.

## Development

Development tools are managed with [Mise](https://mise.jdx.dev/):

```text
mise run setup
mise run fmt
mise run validate
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for the development and release workflows.
Pull requests also run SonarQube Cloud against changed code. Its quality gate
ratchets quality forward by blocking new reliability, security,
maintainability, duplication, or coverage regressions without retroactively
failing a pull request for issues in unchanged code.

## License

sbxflow is available under the [MIT License](LICENSE).

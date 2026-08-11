# Contributing

Thank you for contributing to sbxflow. Read [`README.md`](README.md) first for
the project's purpose, current interface, and lifecycle scope. See
[`docs/structure.md`](docs/structure.md) for the internal package boundaries and
code placement guidance.

Participation is governed by the [`CODE_OF_CONDUCT.md`](CODE_OF_CONDUCT.md).
Report suspected vulnerabilities privately as described in
[`SECURITY.md`](SECURITY.md), rather than through a public issue or pull request.

## Development setup

Development tools and Git hooks are managed with
[Mise](https://mise.jdx.dev/):

```text
mise run setup
```

The setup installs the pinned Go toolchain, Zig compiler for CGO race tests, and
[OpenSpec](https://openspec.dev/) CLI. The repository also includes Codex skills
for its spec-driven workflow; restart Codex after setup so it discovers them.

Use Mise to run one-off Go commands with the repository's pinned toolchain:

```text
mise exec -- go build ./...
mise exec -- go test ./...
```

The existing project tasks cover the routine workflow:

```text
mise run fmt        Format Markdown and Go files
mise run test:architecture
                    Check production package dependency boundaries
mise run test:release
                    Build a release snapshot and validate its archives and cask
mise run validate   Check Markdown, workflows, specs, Go formatting, vet, tests, and builds
```

Start a change with `$openspec-propose "describe the change"`. The usual flow is
`$openspec-apply-change`, `$openspec-verify-change`, and
`$openspec-archive-change`; use `$openspec-explore` when requirements need more
discussion first. OpenSpec artifacts live in `openspec/` and are checked by
`mise run validate`.

Pull requests that add major new functionality must include automated tests
covering it. Pull requests that fix a bug should include a regression test
where practical. Reviewers may request additional tests before merging.

Create a focused branch from the latest `main`. Before committing, format any
Markdown changes and run the repository checks:

```text
mise run fmt
mise run validate
```

Go changes must be formatted with `gofmt`; `mise run fmt` applies it alongside
the repository's Markdown formatter. The aggregate validation task also runs
`go vet ./...`, `go test ./...`, `mise run test:architecture`, and
`go build ./...`.

The architecture check classifies every production package by the type encoded
in its path and enforces the relationship matrix documented in
[`docs/structure.md`](docs/structure.md). Before adding a package, decide
whether it is an inbound adapter, application workflow, domain concept, port,
outbound adapter, entrypoint, or resource and place it accordingly.
Keep `go test ./...` and `mise run test:architecture` as separate checks so
failures remain clear. Changes to the type rules or relationship matrix are
deliberate architecture decisions and must update the declarative policy and
documentation together; do not broaden policy to make an unrelated change
pass.

The matrix has two capability-specific shortcuts: inbound adapters may consume
the BuildInfo port, and application workflows may consume the Sandbox port.
They avoid pass-through layers that would add no behavior; they do not permit
either source type to consume arbitrary ports.

The pre-commit hook also rejects staged whitespace errors and changes that do
not pass repository validation.

## Pull requests

Pull request titles must follow the
[Conventional Commits](https://www.conventionalcommits.org/) format because
GitHub uses the pull request title as the squash commit title on `main`:

```text
<type>(optional-scope)!: <description>
```

The allowed types are:

- `build`
- `chore`
- `ci`
- `docs`
- `feat`
- `fix`
- `perf`
- `refactor`
- `revert`
- `style`
- `test`

Examples:

```text
feat(cli): add the validate command
fix(config): reject unknown source types
docs: explain kit source trust
feat(config)!: replace the kits declaration format
```

The required `Validate` check enforces this format whenever a pull request is
opened, updated, reopened, or retitled. Pull requests must also pass the
repository's CodeQL checks for Go and GitHub Actions. Intermediate commits on a
branch do not need conventional messages because pull requests are squash
merged.

In the pull request body, summarize what changed and list the validation you
ran. All required checks must pass and conversations must be resolved before a
pull request can merge.

### Review policy

Pull requests require one approval. CodeRabbit provides the routinely
available automated review for this single-maintainer project: it reviews each
new head revision, requests changes when it finds issues, and approves the
revision when its review and pre-merge checks pass. Stale approvals are
dismissed, and the latest push must be approved by someone other than its
author. The repository owner remains the code owner; CodeRabbit is a GitHub App
and is not eligible for CODEOWNERS membership.

Independent human review is encouraged whenever another qualified contributor
is available. As the maintainer group grows, reassess whether to require human
approval in addition to CodeRabbit's automated review.

If CodeRabbit is unavailable and delaying urgent security, release-recovery, or
repository-recovery work, the repository administrator may temporarily set the
required approval count to zero. Required status checks and all other branch
protections remain in force. The pull request must record the reason for the
override, and the administrator must restore the approval requirement
immediately after the emergency merge.

## Releases

Releases are created from tags named as canonical `v`-prefixed Semantic
Versions, such as `v0.1.0`. Follow the complete
[`release runbook`](docs/releasing.md) for preparation, tagging, publication,
verification, and failure recovery. Before creating the tag:

- [ ] Confirm `mise run validate` passes on the release commit.
- [ ] Review the changes that will form the generated release notes and identify
      relevant compatibility or breaking changes.
- [ ] Smoke-test the release commit with an `sbx` version inside the documented
      supported range.
- [ ] Confirm the documented `sbx` range agrees with the range enforced by
      `sbxflow doctor`.

Pushing the tag runs the release workflow. It validates the tag and repository,
then uses the pinned GoReleaser version to build the supported executables,
generate checksums, publish the GitHub release, and open a Homebrew cask update
in `jamessawle/homebrew-tap`. The repository must provide a
`HOMEBREW_TAP_GITHUB_TOKEN` Actions secret whose token can write contents to the
tap repository and create the update pull request. Do not move a published
release tag; follow the runbook's partial-failure guidance and issue a new
version when published artifacts must change.

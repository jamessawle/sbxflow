# Contributing

Thank you for contributing to sbxflow. Read [`README.md`](README.md) first for
the project's purpose, current interface, and lifecycle scope. See
[`docs/structure.md`](docs/structure.md) for the internal package boundaries and
code placement guidance.

## Development setup

Development tools and Git hooks are managed with
[Mise](https://mise.jdx.dev/):

```text
mise run setup
```

The setup installs the pinned Go toolchain and
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
mise run validate   Check Markdown, workflows, specs, Go formatting, vet, tests, and builds
```

Start a change with `$openspec-propose "describe the change"`. The usual flow is
`$openspec-apply-change`, `$openspec-verify-change`, and
`$openspec-archive-change`; use `$openspec-explore` when requirements need more
discussion first. OpenSpec artifacts live in `openspec/` and are checked by
`mise run validate`.

Create a focused branch from the latest `main`. Before committing, format any
Markdown changes and run the repository checks:

```text
mise run fmt
mise run validate
```

Go changes must be formatted with `gofmt`; `mise run fmt` applies it alongside
the repository's Markdown formatter. The aggregate validation task also runs
`go vet ./...`, `go test ./...`, and `go build ./...`.

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
opened, updated, reopened, or retitled. Intermediate commits on a branch do not
need conventional messages because pull requests are squash merged.

In the pull request body, summarize what changed and list the validation you
ran. All required checks must pass and conversations must be resolved before a
pull request can merge.

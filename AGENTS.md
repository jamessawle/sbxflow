# Agent guidance

Before making changes in this repository, read and follow:

- [`README.md`](README.md) for the project's purpose, interface, and scope.
- [`CONTRIBUTING.md`](CONTRIBUTING.md) for the development workflow, validation
  commands, and pull request requirements.
- [`docs/structure.md`](docs/structure.md) for internal package responsibilities,
  dependency direction, and code placement guidance.

Keep implementation and documentation aligned with these files.

The repository-owned architecture test classifies production packages by the
types documented in `docs/structure.md` and enforces their allowed
relationships. Classify new packages before adding them. Do not broaden the
relationship matrix merely to make an unrelated change pass. If requested work
conflicts with a boundary, stop and request an explicit architecture decision.

OpenSpec is managed by Mise. Run OpenSpec CLI commands as
`mise exec -- openspec ...` so repository-local skills use the pinned version
rather than any global installation.

# Internal structure

The `internal/` tree separates command-line presentation, application
workflows, repository configuration, and the external Docker Sandboxes
adapter. This document is the placement guide for new code and refactors.

## Package layout

```text
internal/
├── cli/                  Cobra commands and human-readable rendering
├── application/
│   ├── validation/       Complete repository validation workflow
│   ├── lifecycle/        Up, down, and destroy workflows
│   └── doctor/           Compatibility and system-health workflow
├── configuration/        Discovery and interpretation of sbxflow.yaml
├── sbx/                  Adapter to the external sbx executable
└── buildinfo/            Build version and commit metadata
```

### `cli`

`cli` translates command-line input into application calls and renders their
results. It owns Cobra commands, flags, standard stream selection, and exit
error presentation. It does not parse repository declarations or construct
`sbx` invocations.

### `application/validation`

`validation` coordinates the complete validation pipeline. It asks
`configuration` to discover, parse, link, and safely resolve a declaration,
then asks `sbx` to validate selected local kits. Its report is consumed by the
`validate` command and the `up` lifecycle workflow.

### `application/lifecycle`

`lifecycle` implements the behaviour of `up`, `down`, and `destroy`. It decides
which configuration data and safeguards each operation requires, but delegates
all Docker Sandbox process details to `sbx`.

### `application/doctor`

`doctor` defines the ordered compatibility and health checks, their
prerequisites, grading, summaries, and guidance. It interprets captured Docker
Sandboxes inspection results but does not construct or execute subprocesses.

### `configuration`

`configuration` answers what the nearest repository declares. It owns
`sbxflow.yaml` discovery, YAML and schema parsing, typed models, source linking,
trust derivation, safe local-path resolution, and the minimal target parser used
by teardown operations. It never invokes `sbx`.

### `sbx`

`sbx` is the only package that knows how to invoke the external Docker
Sandboxes executable. It owns executable lookup, command names and arguments,
process timeouts, environment variables, captured output, and attached stream
execution. Application packages decide what an outcome means for sbxflow.

### `buildinfo`

`buildinfo` exposes version and commit metadata injected at build time. It has
no application or infrastructure responsibilities.

## Dependency direction

Dependencies point from presentation and workflows toward configuration and
external adapters:

```text
cmd/sbxflow
    └── cli
        ├── application/validation
        ├── application/lifecycle ── application/validation
        └── application/doctor
                         │
             ┌───────────┴───────────┐
             ▼                       ▼
       configuration                sbx
```

The boundary rules are:

- `configuration` does not import `application`, `cli`, or `sbx`.
- `sbx` does not import `application`, `cli`, or `configuration`.
- Application packages do not import Cobra or construct raw `sbx` arguments.
- `cli` contains presentation logic, not application policy.
- Cross-application dependencies must remain directional. Currently only
  `lifecycle` depends on `validation`.

## Placing new code

Use the question the code answers to select its package:

- How is a flag accepted or a result displayed? Put it in `cli`.
- What does a configuration field mean or how is it resolved? Put it in
  `configuration` and update the public schema when necessary.
- What should validation, lifecycle, or doctor do? Put it in the corresponding
  `application` package.
- How is an operation expressed through the Docker Sandboxes CLI? Put it in
  `sbx`.
- Is it only build identity? Put it in `buildinfo`.

Avoid generic packages such as `common`, `helpers`, or `utils`. When code does
not fit an existing responsibility, first determine the missing concept rather
than creating a miscellaneous package.

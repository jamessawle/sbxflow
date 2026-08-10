# Internal structure

The repository organizes Go packages by architectural type. Package paths make
the type visible to contributors and let the architecture test apply the same
rules to packages that are added later.

## Package layout

```text
cmd/sbxflow/                              startup and dependency wiring

internal/
├── adapters/
│   ├── inbound/
│   │   └── cli/                          Cobra commands and rendering
│   └── outbound/
│       ├── declaration/                  discovery, decoding, filesystem
│       └── sbx/                          Docker Sandboxes subprocesses
├── application/
│   ├── doctor/                           system-health workflow
│   ├── lifecycle/                        up, down, and destroy workflows
│   └── validation/                       repository validation workflow
├── domain/
│   └── configuration/                    linking, trust, semantic rules
├── ports/
│   ├── buildInfo/                        linker-provided build identity
│   ├── declaration/                      declaration boundary and DTOs
│   └── sandbox/                          Docker Sandboxes capabilities

schema/                                   published JSON Schema resource
```

The `adapters/inbound` and `adapters/outbound` directories classify their Go
packages; they are not packages themselves. A package may have any filenames
that make sense in Go. No marker file such as `_main.go` is required.

## Architectural types

### Entry point

Packages under `cmd/` start an executable and form its wiring boundary.
`cmd/sbxflow` constructs concrete adapters, injects them into domain and
application services through their port interfaces, attaches process streams
and build metadata, and maps the returned error to an exit status. Concrete
values are narrowed to ports before injection so deeper architecture analysis
still observes the intended boundary. Business rules do not belong here.

### Inbound adapters

Packages under `internal/adapters/inbound/` translate external input into calls
to application workflows and render their results. The CLI adapter owns Cobra
commands, flags, terminal streams, and human-readable output. It receives its
application dependencies and never constructs or imports domain packages or
outbound adapters. It may consume the BuildInfo port directly because adding an
application pass-through for immutable linker values would add no behavior.

### Application

Packages under `internal/application/` coordinate use cases. `lifecycle`
decides how `up`, `down`, and `destroy` proceed; `doctor` orders and grades
system checks; `validation` coordinates resolved configuration with local-kit
validation through Docker Sandboxes. Application packages do not depend on
peer application packages. They consume domain concepts and may call the
Sandbox port directly; routing Docker Sandbox capabilities through the domain
would misrepresent an external operation as a domain rule.

### Domain

Packages under `internal/domain/` contain reusable sbxflow rules and models.
`configuration` owns declaration linking, trust derivation, semantic rules,
safe configuration resolution, and the resulting configuration validity
model. It consumes declaration ports and never imports adapters, application,
or presentation code.

### Ports

Packages under `internal/ports/` define stable boundaries between consumers and
adapters. Each package is capability-specific: `declaration` defines loading,
local-path resolution, and target resolution; `sandbox` defines kit validation,
inspection, and lifecycle operations; `buildInfo` exposes linker-provided build
identity. A port contains only narrow interfaces and the request/result types
crossing that boundary. Ports contain no business rules, filesystem logic,
subprocess logic, rendering, or first-party imports. BuildInfo's only
implementation is its two linker variables and value accessor; a separate
adapter would only forward immutable process metadata.

### Outbound adapters

Packages under `internal/adapters/outbound/` implement ports using filesystem,
decoding, or external-process operations.
`declaration` discovers and reads `sbxflow.yaml`, decodes YAML against the public
schema, and resolves safe local paths. `sbx` is the only package that constructs
Docker Sandboxes subprocess invocations. Outbound adapters depend on ports, not
on their domain or application callers, and do not import one another.

### Resources

`schema` publishes the JSON Schema and has no first-party dependencies. The
declaration adapter may consume that resource while decoding repository
declarations.

## Allowed dependencies

The architecture check evaluates direct production imports. Each production
package must match exactly one architectural type.

| Source type      | Allowed first-party targets    |
| ---------------- | ------------------------------ |
| Entry point      | Any production package         |
| Inbound adapter  | Application and BuildInfo port |
| Application      | Domain and Sandbox port        |
| Domain           | General ports                  |
| Port             | None                           |
| Outbound adapter | General ports and resources    |
| Resource         | None                           |

Only the entrypoint imports concrete outbound adapters. The BuildInfo port is
the one inbound-to-port shortcut, and the Sandbox port is the one
application-to-port shortcut. These exceptions are capability-specific:
inbound adapters do not gain access to general ports, and application packages
do not gain access to general ports. Domain code consumes general port
interfaces, and outbound adapters implement ports without importing their
callers. Go's structural interfaces keep adapters independent of concrete
consumers while compile-time assertions verify that implementations still
satisfy their ports.

Cobra may be imported only by inbound adapters. Test-only imports are excluded
from architecture enforcement so package tests can use fixtures and test
doubles across a production boundary.

Run the policy directly with:

```text
mise run test:architecture
```

`mise run validate` runs the pinned go-arch-lint policy along with the ordinary
Go tests, vet, build, documentation, workflow, and OpenSpec checks. The policy
is declared in `.go-arch-lint.yml`; its recursive components apply to packages
added beneath the typed roots.

## Placing new code

Classify the concept before creating a package:

- External input or output translation belongs in an inbound adapter.
- A user-visible use case or workflow, including `validate`, belongs in
  application.
- A reusable sbxflow model or rule belongs in domain.
- A capability boundary and its request/result types belong in a narrowly
  named port package.
- Filesystem, decoder, network, or subprocess integration belongs in an
  outbound adapter.
- Executable construction belongs in `cmd`; linker-provided build identity
  belongs in the BuildInfo port.

New packages inherit the relationship matrix from their path. A new top-level
type or relationship is an architecture decision and requires aligned policy,
documentation, and review. Avoid generic packages such as `common`, `contracts`,
`helpers`, or `utils`; identify the missing capability or concept and place it
by responsibility instead.

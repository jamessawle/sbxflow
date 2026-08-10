## 1. Domain Boundaries

- [x] 1.1 Move configuration linking, trust derivation, and semantic rules into
      `internal/domain/configuration` with their existing tests.
- [x] 1.2 Keep configuration resolution and validity state in
      `internal/domain/configuration`, and make `internal/application/validation`
      coordinate that domain with external local-kit validation.
- [x] 1.3 Add capability-scoped declaration, sandbox, and BuildInfo port
      packages with narrow interfaces or values and no first-party
      dependencies.
- [x] 1.4 Make lifecycle consume validation through an interface it owns so it
      does not import the peer validation application package.

## 2. Adapter Boundaries

- [x] 2.1 Move declaration discovery, YAML and JSON Schema decoding, lifecycle
      target loading, and local filesystem resolution into
      `internal/adapters/outbound/declaration`.
- [x] 2.2 Move the Docker Sandboxes integration into
      `internal/adapters/outbound/sbx` and preserve its process-level tests.
- [x] 2.3 Move Cobra commands and rendering into
      `internal/adapters/inbound/cli` while retaining injected workflow
      interfaces.

## 3. Application and Entrypoint Wiring

- [x] 3.1 Update doctor and lifecycle workflows to consume domain services and
      ports without importing concrete outbound adapters.
- [x] 3.2 Move concrete dependency construction from CLI into `cmd/sbxflow`,
      wiring declaration and sandbox ports, workflows, streams, and the
      BuildInfo port at startup.
- [x] 3.3 Remove obsolete package paths and update all production and test
      imports without changing CLI behavior.

## 4. Type-Based Architecture Enforcement

- [x] 4.1 Add a pinned go-arch-lint component policy that classifies every
      production package by architectural type and rejects unmatched files.
- [x] 4.2 Encode the allowed type relationship matrix, prohibit Cobra outside
      inbound adapters, make ports first-party leaves, allow only the
      entrypoint to import concrete outbound adapters, and constrain the two
      shortcuts to inbound-to-BuildInfo and application-to-Sandbox.
- [x] 4.3 Exclude test files and review the resolved component mapping for all
      architectural types.
- [x] 4.4 Point `mise run test:architecture` at the replacement checker and
      confirm both it and `go test ./...` remain required by
      `mise run validate`.
- [x] 4.5 Remove `arch-go.yml` and the Arch-Go Mise pin after go-arch-lint
      passes the reorganized repository.

## 5. Documentation and Validation

- [x] 5.1 Rewrite `docs/structure.md` around architectural types, ports,
      responsibilities, allowed relationships, and entrypoint wiring.
- [x] 5.2 Update `CONTRIBUTING.md` and `AGENTS.md` so contributors classify new
      packages by type and do not broaden relationships to make unrelated
      changes pass.
- [x] 5.3 Run `mise run fmt` and review the package moves, artifacts, and
      formatted documentation.
- [x] 5.4 Run `mise run test:architecture` and `mise run validate`, confirming
      the architecture policy, documentation, OpenSpec, Go tests, vet, and
      build all pass.

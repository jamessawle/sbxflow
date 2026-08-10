## 1. Architecture Documentation

- [ ] 1.1 Update `docs/structure.md` to show the complete intentional package
      graph, distinguish mechanically enforced import rules from review-only
      responsibility rules, and explain how the architecture check runs.
- [ ] 1.2 Update `CONTRIBUTING.md` with the dedicated architecture command and
      add `AGENTS.md` guidance that prohibits weakening architecture policy to
      make unrelated changes pass.

## 2. Dependency Enforcement

- [ ] 2.1 Pin Arch-Go v2.1.2 as a Mise-managed development tool and add
      `arch-go.yml` with 100 percent compliance and package coverage thresholds.
- [ ] 2.2 Encode the exact allowed first-party dependency graph and prohibit
      Cobra imports from application packages in `arch-go.yml`.
- [ ] 2.3 Add a dedicated `mise run test:architecture` task, keep it separate
      from `go test ./...`, and require both checks from `mise run validate`.
- [ ] 2.4 Run the dedicated check and confirm it accepts the current production
      graph while ignoring broader test-only imports such as the CLI test
      dependency on `internal/configuration`.

## 3. Validation

- [ ] 3.1 Run `mise run fmt` and review the formatted implementation and
      artifacts.
- [ ] 3.2 Run `mise run validate` and confirm documentation, OpenSpec, Go tests,
      vet, and build checks pass.

## 1. Destroy Lifecycle Execution

- [x] 1.1 Add a destroy lifecycle runner that reuses target resolution and
      exact-name lookup, treats an absent target as success, and invokes
      attached `sbx rm <name>` or `sbx rm --force <name>` for an existing target.
- [x] 1.2 Add focused lifecycle tests for resolution and executable failures,
      absent and similar names, confirmed and forced argument construction,
      attached input/output/error streams, context propagation, list failures,
      and single-rendered removal failures.

## 2. CLI and Executable Integration

- [x] 2.1 Add an injected Cobra `destroy` command with no positional arguments,
      `--force`/`-f`, repository working-directory discovery, all three stream
      bindings, contextual help, and attached-error suppression.
- [x] 2.2 Register the destroy dependency in root construction and update CLI
      helpers and tests for root discovery, help without lifecycle work, valid
      force forms, and rejection of arguments, `--all`, `--name`, and unknown
      flags before lifecycle work.
- [x] 2.3 Add executable-level coverage proving target-only resolution,
      exact-name lookup, confirmation input forwarding, force forwarding,
      Docker output and exit-code preservation, and absent-target idempotency.

## 3. Documentation

- [x] 3.1 Update README availability, command usage, confirmation and force
      behavior, exact targeting, permanent sandbox-state removal, host-workspace
      preservation, and the distinction between `down` and `destroy`.

## 4. Validation

- [x] 4.1 Run `mise run fmt` and inspect the resulting Go, Markdown, and OpenSpec
      artifact changes.
- [x] 4.2 Run focused Go tests for lifecycle, CLI, and executable behavior, then
      run `mise run validate` and resolve every failure.

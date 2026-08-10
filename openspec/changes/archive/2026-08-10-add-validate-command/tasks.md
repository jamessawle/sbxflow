## 1. Configuration Schema and Model

- [x] 1.1 Add the YAML parsing and Draft 2020-12 JSON Schema validation
      dependencies and a checked-in, runtime-embedded schema for version 1.
- [x] 1.2 Define typed configuration and source models that preserve the ordered
      kit selections after schema validation.
- [x] 1.3 Implement single-document YAML parsing with duplicate-key rejection,
      schema error reporting, and typed decoding.
- [x] 1.4 Add schema and loader tests covering every repository example,
      unsupported versions, source unions, unknown fields, malformed YAML,
      duplicate keys, multiple documents, and exact duplicate selections.

## 2. Discovery, Linking, and Trust

- [x] 2.1 Implement nearest-ancestor `sbxflow.yaml` discovery with deterministic
      missing-file errors and tests for direct, nested, nearest, and absent
      declarations.
- [x] 2.2 Link each ordered selection to its named source and validate unknown
      references plus OCI-only `version` rules, with field-oriented errors and
      unit tests.
- [x] 2.3 Normalize selected Git and OCI references without network access and
      derive deterministic `kit.allowedSources` and `kit.allowLocalKits` values.
- [x] 2.4 Add trust-derivation tests for Docker Hub retention, exact remote
      prefixes, first-selection ordering, deduplication, local-kit enablement,
      and exclusion of unused sources.

## 3. Local Kit Validation

- [x] 3.1 Move the generic subprocess runner and output types into a neutral
      internal package, adapt doctor to use it, and preserve doctor test coverage.
- [x] 3.2 Require local roots and selections to be host filesystem references,
      reject URI-like roots plus URI-like or absolute kit paths, resolve roots
      from the declaration directory, and reject unavailable roots or targets
      that escape through traversal or symbolic links.
- [x] 3.3 Implement ordered `sbx kit validate <absolute-path>` execution with a
      bounded timeout, captured diagnostics, continued independent checks, and
      no executable lookup when no local kit is selected.
- [x] 3.4 Add local-validator tests proving that source provenance, rather than
      directory or ZIP packaging, controls invocation; cover URI rejection,
      unavailable roots, valid filesystem artifacts, Docker validation failures,
      unavailable or timed-out executables, multiple selections, and proof that
      Git and OCI selections never invoke Docker.

## 4. Validate Command

- [x] 4.1 Add a validation runner that gates structural, semantic, path, local-kit,
      and trust phases and returns an ordered report suitable for CLI rendering.
- [x] 4.2 Register `validate` in the injected Cobra command tree with contextual
      help, compact success and failure output, and correct exit status.
- [x] 4.3 Add command-level tests for help discovery, declaration and local-kit
      reports, derived trust output, actionable errors, and stdout/stderr routing.
- [x] 4.4 Add built-executable tests using temporary repositories and a fake
      `sbx` executable to cover discovery, successful validation, local-kit
      failure, missing configuration, exit status, and non-mutating invocation.

## 5. Documentation and Verification

- [x] 5.1 Update the README and examples documentation to make `validate`
      available, link the public schema, explain offline remote checks and local
      Docker kit validation, and show the derived trust report.
- [x] 5.2 Add regression coverage proving validation never invokes remote Git or
      OCI resolution, lifecycle commands, or Docker settings mutations.
- [x] 5.3 Run `mise run fmt` and `mise run validate`, then address all reported
      formatting, schema, vet, test, and build failures.
- [x] 5.4 Enumerate the supported Docker Sandbox agents in the public schema,
      document the offline compatibility policy, and test every accepted value
      plus rejection of unsupported names.
- [x] 5.5 Add a root `sbxflow.yaml` for `codex` using the pinned community `mise`
      kit and `kits/mise/network-go` from `jamessawle/sbx-kits`, and cover it as
      a repository validation fixture.
- [x] 5.6 Run `mise run fmt`, validate the root declaration through the built
      command, and run `mise run validate` after the review changes.
- [x] 5.7 End successful `validate` output with `Validation: passed`, update the
      documented output, and cover the terminal result in command and executable
      tests without emitting it on failure.
- [x] 5.8 Run `mise run fmt`, exercise the root declaration, and run
      `mise run validate` after the output clarification.
- [x] 5.9 Replace the flat validation output with a cohesive structured report
      containing declaration, derived kit state, validation state, and findings;
      route the whole report to stdout on pass or stderr on fail and cover early
      and post-derivation failures.
- [x] 5.10 Update the documented output, run `mise run fmt`, exercise the root
      declaration, and run `mise run validate` after the report redesign.

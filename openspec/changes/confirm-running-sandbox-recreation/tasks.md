## 1. Sandbox State Inspection

- [ ] 1.1 Define the narrow absent, stopped, and running sandbox state contract
      in the Sandbox port and update `up` dependency wiring without broadening
      architecture relationships.
- [ ] 1.2 Implement exact-name machine-readable state inspection in the `sbx`
      adapter, including actionable failures for command errors, malformed output,
      duplicate matches, and unrecognized states.
- [ ] 1.3 Add adapter and lifecycle fixtures covering absent, stopped, running,
      malformed, ambiguous, and failed inspection results.

## 2. Safe Recreation Confirmation

- [ ] 2.1 Add a narrow lifecycle confirmation dependency and branch `up` so only
      running recreation requests confirmation, while absent and stopped behavior
      remains unchanged.
- [ ] 2.2 Implement the CLI warning and default-negative response parser on the
      command input and error streams, including affirmative, negative, empty,
      malformed, EOF, and read-error cases.
- [ ] 2.3 Preserve forced exact-name removal after approval, stop before mutation
      on every non-approval, and retain attached Docker diagnostics and failure
      ordering.
- [ ] 2.4 Update unit and built-executable tests for prompts, stream routing,
      non-zero cancellation, inspection failures, removal ordering, and successful
      replacement.

## 3. Documentation and Validation

- [ ] 3.1 Update `up --recreate` help and README lifecycle documentation with
      conditional confirmation, cross-terminal risk, non-interactive failure, and
      the inspection-removal race.
- [ ] 3.2 Run `mise run fmt`, `mise run test:architecture`, and
      `mise run validate`, resolving only issues within this change's scope.

## 1. Validation Status Behavior

- [x] 1.1 Add lifecycle tests proving successful validation writes the
      declaration path to standard error before Docker lookup, while failed
      validation writes no success status.
- [x] 1.2 Emit the plain configuration-valid status from lifecycle orchestration
      immediately after complete validation succeeds.

## 2. Interface Coverage and Documentation

- [x] 2.1 Update CLI and executable-level `up` tests for the new standard-error
      output while preserving attached agent output and failure behavior.
- [x] 2.2 Document the successful `up` validation status in README.md.

## 3. Validation

- [x] 3.1 Run repository formatting and the complete Mise validation suite.

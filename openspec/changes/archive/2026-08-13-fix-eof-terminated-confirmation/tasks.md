## 1. Confirmation Reader

- [x] 1.1 Update the inbound CLI confirmation reader to return buffered text
      when EOF terminates a non-empty response while preserving immediate EOF,
      non-EOF errors, and no-progress failures as errors.
- [x] 1.2 Extend focused confirmation tests for newline-terminated responses,
      EOF-terminated affirmative, negative, and malformed responses, immediate
      EOF, and a non-EOF error after buffered bytes.

## 2. Executable Regression Coverage

- [x] 2.1 Add built-command tests proving EOF-terminated affirmative input
      recreates a running sandbox while EOF-terminated negative and malformed
      responses cancel without removal, creation, or entry.
- [x] 2.2 Add built-command coverage proving immediate EOF reports unavailable
      confirmation and performs no removal, creation, or entry.

## 3. Validation

- [x] 3.1 Run repository formatting and the focused CLI and executable test
      suites.
- [x] 3.2 Run `mise run validate` and confirm the OpenSpec change and repository
      checks pass.

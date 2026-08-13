## Why

`up --recreate` currently rejects an otherwise complete confirmation response when standard input closes without a trailing newline. This makes non-interactive input such as `printf yes | sbxflow up --recreate` fail differently from the equivalent newline-terminated response.

## What Changes

- Treat EOF as a valid terminator when the confirmation reader has already buffered response bytes.
- Continue to require an explicit normalized `y` or `yes` response before approving destructive recreation.
- Keep immediate EOF and non-EOF read failures unavailable, failing closed without changing sandbox state.
- Add unit and executable-level regression coverage for newline-terminated, EOF-terminated, empty, negative, malformed, and failing-reader input.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `cli-interface`: Distinguish a complete EOF-terminated confirmation response from unavailable confirmation input.

## Impact

- Affects confirmation input handling in `internal/adapters/inbound/cli` and its tests.
- Updates the CLI confirmation contract without changing flags, output formats, dependencies, or architecture boundaries.
- Resolves GitHub issue #70.

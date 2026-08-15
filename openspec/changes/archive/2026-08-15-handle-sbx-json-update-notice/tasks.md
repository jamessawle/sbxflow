## 1. Regression Coverage

- [x] 1.1 Add state-inspection cases based on the captured box-drawn Docker
      update notice, including presentation variation around its stable title,
      proving a notice after a valid listing preserves the decoded sandbox
      state.
- [x] 1.2 Add negative cases for leading contamination, malformed JSON before a
      notice, multiple JSON documents, invalid UTF-8 in the trailing bytes and
      within the listing's own name and status values, a near-match notice
      title, and unknown trailing output.

## 2. State Output Decoding

- [x] 2.1 Replace whole-buffer unmarshalling with a focused sandbox-list decoder
      that isolates one leading JSON document and its trailing bytes.
- [x] 2.2 Add narrow update-notice classification using Docker's stable title
      marker while preserving collection, duplicate-name, and lifecycle-state
      validation.
- [x] 2.4 Reject invalid UTF-8 within the decoded listing so replacement
      characters cannot satisfy an exact name match or misreport a lifecycle
      state.
- [x] 2.3 Return distinct, bounded errors for multiple JSON documents and unknown
      trailing output without including sandbox-list contents.

## 3. Validation

- [x] 3.1 Run the targeted lifecycle and outbound-adapter tests and resolve any
      regression.
- [x] 3.2 Run `mise run fmt`, resolving all formatting failures.
- [x] 3.3 Run `mise run validate`, resolving all specification, architecture,
      vet, test, and build failures. `mise` was unavailable in the authoring
      environment, so `go build ./...`, `go vet ./...`, and `go test ./...` were
      run locally; `mise run validate` itself ran green in CI through the
      `Validate` job.

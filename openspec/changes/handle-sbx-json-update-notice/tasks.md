## 1. Regression Coverage

- [ ] 1.1 Add state-inspection cases based on the captured box-drawn Docker
      update notice, including presentation variation around its stable title,
      proving a notice after a valid listing preserves the decoded sandbox
      state.
- [ ] 1.2 Add negative cases for leading contamination, malformed JSON before a
      notice, multiple JSON documents, invalid UTF-8, a near-match notice title,
      and unknown trailing output.

## 2. State Output Decoding

- [ ] 2.1 Replace whole-buffer unmarshalling with a focused sandbox-list decoder
      that isolates one leading JSON document and its trailing bytes.
- [ ] 2.2 Add narrow update-notice classification using Docker's stable title
      marker while preserving collection, duplicate-name, and lifecycle-state
      validation.
- [ ] 2.3 Return distinct, bounded errors for multiple JSON documents and unknown
      trailing output without including sandbox-list contents.

## 3. Validation

- [ ] 3.1 Run the targeted lifecycle and outbound-adapter tests and resolve any
      regression.
- [ ] 3.2 Run `mise run fmt` and `mise run validate`, resolving all formatting,
      specification, architecture, vet, test, and build failures.

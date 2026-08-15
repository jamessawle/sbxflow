## Why

Docker Sandboxes v0.35.0 can append its asynchronous, human-readable update
notice to an otherwise valid `sbx ls --json` response. The once-per-day notice
currently makes `sbxflow up` fail while decoding sandbox state even though the
machine-readable document itself is complete and valid.

## What Changes

- Recognize Docker Sandboxes' known trailing update notice after a complete
  sandbox-list JSON document.
- Continue lifecycle processing from the decoded document when that notice is
  present.
- Preserve strict failures for leading contamination, malformed or incomplete
  JSON, multiple JSON documents, and unrecognized trailing output.
- Add regression coverage for the observed update-notice shape and the strict
  rejection boundaries.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `sandbox-lifecycle`: Define reliable state inspection when Docker appends its
  update notice to a successful machine-readable listing.

## Impact

The change is limited to Docker Sandboxes output decoding in
`internal/adapters/outbound/sbx` and its lifecycle regression tests. It does not
change the CLI interface, sandbox state, package boundaries, dependencies, or
supported Docker Sandboxes version range.

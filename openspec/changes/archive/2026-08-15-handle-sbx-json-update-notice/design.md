## Context

See `proposal.md` for the motivation and the sandbox-lifecycle delta for the
behavioral contract. `Client.InspectSandbox` currently passes all captured
standard output from `sbx ls --json` to `json.Unmarshal`, which requires the
buffer to contain exactly one JSON value plus whitespace.

The diagnostic spike found:

- Forty consecutive, separately captured invocations returned valid JSON and
  no standard-error output, confirming that the failure is intermittent and
  gated by cached update-check state.
- Docker Sandboxes v0.35.0 stores a 24-hour update-check interval and recorded
  v0.38.0 as the latest version at the time of the reported failure.
- The installed binary contains `shouldCheckForUpdates`, `startUpdateCheck`,
  `asyncCheckUpdate`, and the literal `Docker Sandboxes Update Available`.
- After temporarily aging the update-check timestamp, an instrumented
  invocation reproduced the failure shape with exit status zero and empty
  standard error. Standard output contained 1,519 bytes forming one valid JSON
  document followed by a 1,626-byte valid UTF-8 update banner.
- The exact suffix starts with two line feeds and the box-drawing character `╭`
  (`0a 0a e2 95 ad`), contains `Docker Sandboxes Update Available`, the current
  and available versions, a release-notes URL, and an upgrade command, and ends
  with the matching box border.

The forced capture confirms that Docker's asynchronous update notice is the
source of the sporadic corruption. The original update-state file was restored
and verified against its pre-spike SHA-256 checksum.

## Goals / Non-Goals

**Goals:**

- Decode the complete leading sandbox-list document independently from a
  recognized Docker update notice.
- Keep the acceptance rule narrow and fail closed for every other unexpected
  output shape.
- Preserve existing state validation and lifecycle safety behavior.
- Provide failure messages that distinguish malformed JSON, multiple JSON
  values, and unknown trailing output without echoing the sandbox listing.

**Non-Goals:**

- Repair or suppress Docker Sandboxes' update checker.
- Accept arbitrary log, warning, or progress output around machine-readable
  responses.
- Surface the ignored update notice through sbxflow; Docker's latest release can
  be outside sbxflow's supported version range.
- Change the supported Docker Sandboxes versions or lifecycle ports.

## Decisions

### Decode one leading JSON document, then classify the remainder

State inspection will use a streaming JSON decoder for the first value and use
its consumed input offset to isolate the remaining bytes. The first decode
continues to reject leading non-whitespace content and malformed or incomplete
JSON. Existing checks for the `sandboxes` collection, exact-name uniqueness,
and recognized lifecycle states remain unchanged.

After trimming surrounding whitespace, an empty remainder succeeds. A remainder
that begins with another decodable JSON value is rejected as multiple
machine-readable documents. Other non-empty remainders proceed to the
update-notice classifier.

Continuing to use `json.Unmarshal` was rejected because it cannot distinguish a
valid leading document from the known auxiliary suffix. Decoding the first
value and ignoring all remaining bytes was rejected because it would silently
weaken the machine-readable contract.

### Recognize the notice by Docker's stable title marker

The classifier will accept only valid UTF-8 trailing text containing the exact
title `Docker Sandboxes Update Available`. Matching the semantic title rather
than terminal borders, ANSI styling, current/latest version numbers, or upgrade
instructions keeps the workaround stable across color and layout changes. The
title is specific to the verified Docker update path; all other suffixes fail.

Matching the complete captured bytes was rejected because terminal width,
versions, release URLs, upgrade instructions, and ANSI styling can legitimately
vary. Regression coverage will retain the captured box-drawn shape while
varying presentation around the stable title marker. Treating any valid UTF-8
suffix as human-readable diagnostics was rejected because unrelated
successful-output contamination would become invisible.

### Ignore the recognized notice only for state decoding

The workaround will stay in the Docker Sandboxes outbound adapter's state-list
decoder. It will not change the generic process runner, the sandbox port, or
other commands. A generic output sanitizer would hide command-specific protocol
violations, while changing the port to return auxiliary diagnostics would widen
the application contract for a notice that sbxflow cannot safely recommend: the
advertised Docker release may be unsupported by the installed sbxflow version.

### Keep diagnostics bounded and free of listing contents

Unknown trailing output will produce an error that identifies unexpected bytes
after the machine-readable listing and reports their size or classification. It
will not embed the full stdout buffer, which can include sandbox names and host
workspace paths. Existing subprocess failures will continue to report Docker's
stderr/stdout diagnostics through the established captured-error path.

## Risks / Trade-offs

- **Docker changes the notice title** → The command fails closed with an
  unexpected-trailing-output error; update the narrow classifier from a new
  captured fixture.
- **Unrelated output contains the exact update title** → Leading contamination
  and additional JSON are still rejected, and the suffix originates from the
  trusted `sbx` child; keep the classifier local and cover negative near-matches.
- **The notice is emitted before rather than after JSON in a future release** →
  The first decode rejects it, preserving the machine-readable contract rather
  than guessing where JSON begins.
- **The upstream bug is fixed** → Ordinary JSON continues through the empty
  remainder path, leaving the compatibility code dormant.

## Migration Plan

No sandbox or configuration migration is required. Ship the adapter and
regression tests in a patch release. Rollback restores strict whole-buffer
decoding and does not alter sandbox data.

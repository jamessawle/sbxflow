## Context

See `proposal.md` for motivation. The CLI adapter currently owns the `--recreate` flag and interactive confirmer, while the lifecycle workflow owns state inspection and decides whether confirmation is required. Recreation already uses the shared exact-name forced-removal path after approval or for a stopped sandbox.

The architecture permits the inbound CLI adapter to call the lifecycle application package. Destructive lifecycle decisions remain in the application layer; Cobra flag parsing and usage errors remain in the inbound adapter.

## Goals / Non-Goals

**Goals:**

- Represent forced recreation explicitly from CLI parsing through lifecycle decision-making.
- Reject the invalid force-without-recreate combination before validation or external operations, including for direct lifecycle callers.
- Reuse the existing removal, cleanup, replacement, and attachment paths.
- Preserve the interactive path unchanged when force is absent.

**Non-Goals:**

- Adding `-f` or another shorthand to `up`.
- Changing `destroy --force`, Docker's removal interface, or ordinary `up` reconciliation behavior.
- Introducing a general non-interactive mode or changing confirmation response parsing.

## Decisions

### Carry force as an explicit lifecycle option

Add a force value alongside recreation in the up workflow options. When the inspected sandbox is running and recreation is requested, the lifecycle workflow calls the confirmer only when force is false. All existing sandbox removals continue through the same exact-name removal helper with Docker force enabled.

This keeps the destructive decision visible at the application boundary. Encoding force by passing a nil confirmer was rejected because nil currently means confirmation is unavailable, obscures intent, and could accidentally convert missing wiring into approval.

### Validate the flag relationship at both boundaries

The CLI rejects `--force` without `--recreate` before obtaining the working directory or invoking its runner, producing a direct usage error. The lifecycle runner also rejects the equivalent invalid option combination before configuration validation.

The application guard is deliberate defense in depth for tests and future non-CLI adapters. Relying only on Cobra flag validation was rejected because `UpRunner` is a separately callable application boundary.

### Keep force long-only and scoped to confirmation bypass

Register `--force` without shorthand. It changes only the approval step for a running recreation: stopped and absent states retain their existing behavior, and force never turns an ordinary `up` into removal.

Reusing `-f` was rejected to keep the issue's proposed interface explicit and avoid implying that `up` and `destroy` have identical semantics.

### Preserve existing safety and cleanup mechanics

After approval is established either interactively or by force, the existing recreation flow remains responsible for exact-name removal, declared network-resource cleanup, replacement creation, and agent entry. No new outbound capability or subprocess invocation is required.

## Risks / Trade-offs

- [Users can terminate other attached sessions without a prompt] → Make the option explicit, require `--recreate`, keep it long-only, and disclose the consequence in command help and README.
- [CLI and application validation could drift] → Cover both boundaries with tests using the same force-requires-recreate rule.
- [A refactor could accidentally bypass confirmation for ordinary recreation] → Retain executable and lifecycle tests asserting that unforced running recreation calls the confirmer and cancellation prevents mutations.

## Migration Plan

This is an additive CLI change with no stored-data or configuration migration. Release it with updated documentation and tests. Rollback consists of reverting the flag and lifecycle option; existing invocations without `--force` remain compatible throughout.

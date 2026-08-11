## 1. Declaration Contract

- [x] 1.1 Add the optional network model and strict `allowedHosts` schema,
      including decoding and lifecycle-target coverage for ordered resources,
      invalid entries, duplicates, unknown fields, and resources that carry a
      scheme or path and so could never match a Docker network request.
- [x] 1.2 Thread declared network resources through validated lifecycle plans
      and port request values with unit tests that preserve order and absence.

## 2. Docker Network Policy Adapter

- [x] 2.1 Add captured sandbox-scoped allow and remove operations using Docker's
      policy CLI, with tests for the single comma-separated resources argument,
      empty input, ordered resources, diagnostics, partial failures, and
      idempotent removal of an already-absent resource or scoped policy.
- [x] 2.2 Split provisioning from attachment with a `Creator` capability so the
      scoped rule can be applied to an existing sandbox before its agent starts,
      and add compensation that removes the just-created sandbox when the rule is
      rejected, reporting any compensation failure without hiding the original
      diagnostic.

## 3. Shared Lifecycle Removal

- [x] 3.1 Extract removal sequencing inside the lifecycle application package so
      destroy and recreation share sandbox removal followed by declared network
      cleanup.
- [x] 3.2 Add lifecycle tests for successful cleanup, removal failure without
      cleanup, cleanup failure after removal, recreation cleanup and reapplication,
      and unchanged behavior without allowed hosts.

## 4. Public Documentation and Validation

- [x] 4.1 Document `sandbox.network.allowedHosts`, ownership, recreation,
      cleanup, and organisation-policy precedence in the README and a complete
      repository example.
- [x] 4.2 Format the repository and run the focused Go tests, architecture test,
      OpenSpec validation, and full repository validation.

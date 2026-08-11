## 1. Declaration Contract

- [ ] 1.1 Add the optional network model and strict `allowedHosts` schema,
      including decoding and lifecycle-target coverage for ordered resources,
      invalid entries, duplicates, and unknown fields.
- [ ] 1.2 Thread declared network resources through validated lifecycle plans
      and port request values with unit tests that preserve order and absence.

## 2. Docker Network Policy Adapter

- [ ] 2.1 Add captured sandbox-scoped allow and remove operations using Docker's
      policy CLI, with tests for exact arguments, empty input, ordered resources,
      diagnostics, and partial failures.
- [ ] 2.2 Add creation-failure compensation that removes the just-added rule and
      reports any compensation failure without hiding the original diagnostic.

## 3. Shared Lifecycle Removal

- [ ] 3.1 Extract removal sequencing inside the lifecycle application package so
      destroy and recreation share sandbox removal followed by declared network
      cleanup.
- [ ] 3.2 Add lifecycle tests for successful cleanup, removal failure without
      cleanup, cleanup failure after removal, recreation cleanup and reapplication,
      and unchanged behavior without allowed hosts.

## 4. Public Documentation and Validation

- [ ] 4.1 Document `sandbox.network.allowedHosts`, ownership, recreation,
      cleanup, and organisation-policy precedence in the README and a complete
      repository example.
- [ ] 4.2 Format the repository and run the focused Go tests, architecture test,
      OpenSpec validation, and full repository validation.

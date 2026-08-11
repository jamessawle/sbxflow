## ADDED Requirements

### Requirement: Sandbox network allow resources are strictly declared

The `validate` command SHALL accept an optional
`sandbox.network.allowedHosts` array of unique, non-empty strings and SHALL
reject unknown network fields. Each entry SHALL be a host, domain, wildcard
subdomain, bracketed IPv6 literal, or `**`, each with an optional `:port` suffix
from 1 to 65535, and SHALL be retained exactly as declared. A bracketed literal
SHALL be a well-formed IPv6 address. Docker Sandboxes matches network requests
by host and port, so a resource carrying a scheme or path, a port outside the
matchable range, or a malformed literal is accepted by its policy CLI but can
never match a request; validation SHALL reject such an entry rather than let it
appear applied.

#### Scenario: Network allow resources are declared

- **WHEN** a version 1 declaration contains unique, non-empty entries under
  `sandbox.network.allowedHosts`
- **THEN** structural validation succeeds
- **AND** the entries remain in declaration order

#### Scenario: Network allow resource is empty

- **WHEN** `sandbox.network.allowedHosts` contains an empty string
- **THEN** structural validation reports the invalid entry
- **AND** exits with a non-zero status

#### Scenario: Network allow resource is repeated

- **WHEN** `sandbox.network.allowedHosts` contains the same string more than
  once
- **THEN** structural validation reports the duplicate entry
- **AND** exits with a non-zero status

#### Scenario: Network allow resource is not a host

- **WHEN** `sandbox.network.allowedHosts` contains a URL, a path, or any other
  value that is not a host, domain, wildcard subdomain, IP literal, or `**` with
  an optional port
- **THEN** structural validation reports the invalid entry
- **AND** exits with a non-zero status

#### Scenario: Network allow resource cannot match a request

- **WHEN** `sandbox.network.allowedHosts` contains a port outside 1 to 65535, or
  a bracketed literal that is not a well-formed IPv6 address
- **THEN** structural validation reports the invalid entry
- **AND** exits with a non-zero status

#### Scenario: Network declaration contains an unknown field

- **WHEN** `sandbox.network` contains a field other than `allowedHosts`
- **THEN** structural validation reports the unknown field
- **AND** exits with a non-zero status

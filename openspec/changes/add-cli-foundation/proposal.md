## Why

sbxflow needs an executable command-line foundation before its declared lifecycle
commands can be implemented. Establishing the root experience now gives later
commands a consistent, testable interface for help, version reporting, errors,
and process execution.

## What Changes

- Introduce a Go executable and module for `sbxflow`.
- Build the command tree with Cobra and present help and errors with Fang.
- Provide root help through the no-argument invocation, `help`, `-h`, and
  `--help`.
- Provide build-aware version output through `-v` and `--version`, without a
  `version` subcommand.
- Keep Fang's completion and man-page commands disabled until they are added as
  intentional public capabilities.
- Add automated tests, Mise development tasks, and documentation for the initial
  CLI behavior.

## Capabilities

### New Capabilities

- `cli-interface`: Defines the initial executable's root help, version flags,
  error behavior, and intentionally limited command surface.

### Modified Capabilities

None.

## Impact

- Adds Go as the application implementation language and pinned development
  toolchain.
- Adds Cobra and Fang as runtime dependencies.
- Introduces the executable entry point, CLI package structure, and CLI-focused
  tests.
- Extends Mise setup and validation to build, format, vet, and test Go code.
- Updates the README and contributor guidance to describe the available CLI and
  development commands while leaving lifecycle commands unimplemented.

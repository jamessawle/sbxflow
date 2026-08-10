package doctor

import "github.com/jamessawle/sbxflow/internal/command"

// Compatibility aliases keep the doctor package's existing test fixtures
// focused while process execution is shared from the neutral command package.
type CommandOutput = command.Output
type CommandRunner = command.Runner
type ExecCommandRunner = command.ExecRunner

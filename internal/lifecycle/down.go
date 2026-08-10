package lifecycle

import (
	"context"
	"fmt"

	"github.com/jamessawle/sbxflow/internal/command"
)

// DownRunner resolves and stops an existing repository sandbox.
type DownRunner struct {
	Targets     TargetResolver
	Commands    command.Runner
	Interactive command.InteractiveRunner
}

// NewDefaultDownRunner constructs production down lifecycle dependencies.
func NewDefaultDownRunner() DownRunner {
	return DownRunner{
		Targets:     NewDefaultTargetResolver(),
		Commands:    command.ExecRunner{Timeout: defaultSandboxLookupTimeout},
		Interactive: command.InteractiveExecRunner{},
	}
}

// Run resolves the nearest target and idempotently stops its exact sandbox.
func (r DownRunner) Run(ctx context.Context, start string, streams Streams) error {
	target, err := r.Targets.Resolve(start)
	if err != nil {
		return err
	}
	executable, err := r.Commands.LookPath("sbx")
	if err != nil {
		return fmt.Errorf("locate sbx for sandbox lookup: %w", err)
	}
	exists, err := sandboxExists(ctx, r.Commands, executable, target.Name)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	err = r.Interactive.Run(ctx, command.InteractiveInvocation{
		Executable: executable,
		Args:       []string{"stop", target.Name},
		Stdout:     streams.Out,
		Stderr:     streams.Err,
	})
	if err != nil {
		return AttachedProcessError{Err: err}
	}
	return nil
}

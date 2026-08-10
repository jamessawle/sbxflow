package lifecycle

import (
	"context"
	"fmt"

	"github.com/jamessawle/sbxflow/internal/command"
)

// DestroyRunner resolves and removes an existing repository sandbox.
type DestroyRunner struct {
	Targets     TargetResolver
	Commands    command.Runner
	Interactive command.InteractiveRunner
}

// NewDefaultDestroyRunner constructs production destroy lifecycle dependencies.
func NewDefaultDestroyRunner() DestroyRunner {
	return DestroyRunner{
		Targets:     NewDefaultTargetResolver(),
		Commands:    command.ExecRunner{Timeout: defaultSandboxLookupTimeout},
		Interactive: command.InteractiveExecRunner{},
	}
}

// Run resolves the nearest target and idempotently removes its exact sandbox.
func (r DestroyRunner) Run(ctx context.Context, start string, force bool, streams Streams) error {
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

	args := []string{"rm", target.Name}
	if force {
		args = []string{"rm", "--force", target.Name}
	}
	err = r.Interactive.Run(ctx, command.InteractiveInvocation{
		Executable: executable,
		Args:       args,
		Stdin:      streams.In,
		Stdout:     streams.Out,
		Stderr:     streams.Err,
	})
	if err != nil {
		return AttachedProcessError{Err: err}
	}
	return nil
}

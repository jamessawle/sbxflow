package lifecycle

import (
	"context"

	sandboxport "github.com/jamessawle/sbxflow/internal/ports/sandbox"
)

// DownRunner resolves and stops an existing repository sandbox.
type DownRunner struct {
	Targets   TargetResolver
	Sandboxes interface {
		sandboxport.ExistenceChecker
		sandboxport.Stopper
	}
}

// Run resolves the nearest target and idempotently stops its exact sandbox.
func (r DownRunner) Run(ctx context.Context, start string, streams Streams) error {
	target, err := r.Targets.Resolve(start)
	if err != nil {
		return err
	}
	exists, err := r.Sandboxes.Exists(ctx, target.Name)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	err = r.Sandboxes.StopSandbox(ctx, target.Name, sandboxport.Streams{Out: streams.Out, Err: streams.Err})
	if err != nil {
		return AttachedProcessError{Err: err}
	}
	return nil
}

package lifecycle

import (
	"context"

	sandboxport "github.com/jamessawle/sbxflow/internal/ports/sandbox"
)

// DestroyRunner resolves and removes an existing repository sandbox.
type DestroyRunner struct {
	Targets   TargetResolver
	Sandboxes interface {
		sandboxport.Lookup
		sandboxport.Remover
	}
}

// Run resolves the nearest target and idempotently removes its exact sandbox.
func (r DestroyRunner) Run(ctx context.Context, start string, force bool, streams Streams) error {
	target, err := r.Targets.Resolve(start)
	if err != nil {
		return err
	}
	exists, err := r.Sandboxes.SandboxExists(ctx, target.Name)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	err = r.Sandboxes.RemoveSandbox(ctx, sandboxport.RemoveRequest{
		Name:  target.Name,
		Force: force,
		Streams: sandboxport.Streams{
			In: streams.In, Out: streams.Out, Err: streams.Err,
		},
	})
	if err != nil {
		return AttachedProcessError{Err: err}
	}
	return nil
}

package lifecycle

import (
	"context"

	sandboxport "github.com/jamessawle/sbxflow/internal/ports/sandbox"
)

// DestroyRunner resolves and removes an existing repository sandbox.
type DestroyRunner struct {
	Targets   TargetResolver
	Sandboxes interface {
		sandboxport.ExistenceChecker
		removalSandbox
	}
}

// Run resolves the nearest target and idempotently removes its exact sandbox.
func (r DestroyRunner) Run(ctx context.Context, start string, force bool, streams Streams) error {
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

	return removeSandbox(ctx, r.Sandboxes, target.Name, target.AllowedHosts, force, streams)
}

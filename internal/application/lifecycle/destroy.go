package lifecycle

import (
	"context"

	"github.com/jamessawle/sbxflow/internal/sbx"
)

// DestroyRunner resolves and removes an existing repository sandbox.
type DestroyRunner struct {
	Targets   TargetResolver
	Sandboxes sbx.Client
}

// NewDefaultDestroyRunner constructs production destroy lifecycle dependencies.
func NewDefaultDestroyRunner() DestroyRunner {
	return DestroyRunner{
		Targets:   NewDefaultTargetResolver(),
		Sandboxes: sbx.NewClient(defaultSandboxLookupTimeout),
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

	err = r.Sandboxes.RemoveSandbox(ctx, sbx.RemoveRequest{
		Name:  target.Name,
		Force: force,
		Streams: sbx.Streams{
			In: streams.In, Out: streams.Out, Err: streams.Err,
		},
	})
	if err != nil {
		return AttachedProcessError{Err: err}
	}
	return nil
}

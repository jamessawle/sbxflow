package lifecycle

import (
	"context"

	sandboxport "github.com/jamessawle/sbxflow/internal/ports/sandbox"
)

type removalSandbox interface {
	sandboxport.Remover
}

// removeSandbox removes the exact environment and lets Docker Sandboxes discard
// all resources scoped to it.
func removeSandbox(ctx context.Context, sandboxes removalSandbox, name string, force bool, streams Streams) error {
	if err := sandboxes.RemoveSandbox(ctx, sandboxport.RemoveRequest{
		Name: name, Force: force,
		Streams: sandboxport.Streams{In: streams.In, Out: streams.Out, Err: streams.Err},
	}); err != nil {
		return AttachedProcessError{Err: err}
	}
	return nil
}

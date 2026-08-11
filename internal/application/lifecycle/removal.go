package lifecycle

import (
	"context"
	"fmt"

	sandboxport "github.com/jamessawle/sbxflow/internal/ports/sandbox"
)

type removalSandbox interface {
	sandboxport.Remover
	sandboxport.NetworkPolicy
}

// removeSandbox removes the exact sandbox and then discards each declared
// network resource. Docker Sandboxes discards a sandbox-scoped policy along with
// its sandbox, so cleanup is ordinarily a no-op; it still runs to cover a policy
// that outlived the sandbox it was scoped to. Each resource is removed
// independently because Docker Sandboxes rejects a whole removal request when
// any one of its resources is already absent.
func removeSandbox(ctx context.Context, sandboxes removalSandbox, name string, allowedHosts []string, force bool, streams Streams) error {
	if err := sandboxes.RemoveSandbox(ctx, sandboxport.RemoveRequest{
		Name: name, Force: force,
		Streams: sandboxport.Streams{In: streams.In, Out: streams.Out, Err: streams.Err},
	}); err != nil {
		return AttachedProcessError{Err: err}
	}
	for _, resource := range allowedHosts {
		if err := sandboxes.RemoveNetworkResource(ctx, sandboxport.NetworkRemoveRequest{Name: name, Resource: resource}); err != nil {
			return fmt.Errorf("sandbox %q was removed but network cleanup was incomplete for resource %q: %w", name, resource, err)
		}
	}
	return nil
}

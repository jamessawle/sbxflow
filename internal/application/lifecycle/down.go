package lifecycle

import (
	"context"

	"github.com/jamessawle/sbxflow/internal/sbx"
)

// DownRunner resolves and stops an existing repository sandbox.
type DownRunner struct {
	Targets   TargetResolver
	Sandboxes sbx.Client
}

// NewDefaultDownRunner constructs production down lifecycle dependencies.
func NewDefaultDownRunner() DownRunner {
	return DownRunner{
		Targets:   NewDefaultTargetResolver(),
		Sandboxes: sbx.NewClient(defaultSandboxLookupTimeout),
	}
}

// Run resolves the nearest target and idempotently stops its exact sandbox.
func (r DownRunner) Run(ctx context.Context, start string, streams Streams) error {
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

	err = r.Sandboxes.StopSandbox(ctx, target.Name, sbx.Streams{Out: streams.Out, Err: streams.Err})
	if err != nil {
		return AttachedProcessError{Err: err}
	}
	return nil
}

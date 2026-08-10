package doctor

import (
	"context"

	"github.com/jamessawle/sbxflow/internal/sbx"
)

// CommandOutput is the captured output from one Docker Sandboxes inspection.
type CommandOutput = sbx.Output

// Inspector is the subset of the Docker Sandboxes adapter used by doctor.
type Inspector interface {
	Locate() (string, error)
	Version(context.Context, string) CommandOutput
	Diagnose(context.Context, string) CommandOutput
	ListPolicies(context.Context, string) CommandOutput
	GetKitAllowedSources(context.Context, string) CommandOutput
}

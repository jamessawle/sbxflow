package lifecycle

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/jamessawle/sbxflow/internal/domain/configuration"
	sandboxport "github.com/jamessawle/sbxflow/internal/ports/sandbox"
)

// ErrValidationFailed identifies a lifecycle request stopped by repository
// validation.
var ErrValidationFailed = errors.New("configuration validation failed")

// ValidationReport is the validated domain state returned by up.
type ValidationReport = configuration.Validation

// Validator runs the complete repository validation pipeline.
type Validator interface {
	Run(ctx context.Context, start string) configuration.Validation
}

// Streams are attached to the interactive Docker agent process.
type Streams struct {
	In  io.Reader
	Out io.Writer
	Err io.Writer
}

// UpRunner validates, inspects, and enters the declared sandbox.
type UpRunner struct {
	Validation Validator
	Sandboxes  interface {
		sandboxport.Lookup
		sandboxport.Runner
	}
}

// AttachedProcessError identifies a failure already rendered by the attached
// Docker process on the CLI error stream.
type AttachedProcessError struct {
	Err error
}

func (e AttachedProcessError) Error() string { return e.Err.Error() }
func (e AttachedProcessError) Unwrap() error { return e.Err }

// Run validates the repository, chooses the creation or existing-sandbox
// invocation, and remains attached until Docker exits.
func (r UpRunner) Run(ctx context.Context, start string, streams Streams) (ValidationReport, error) {
	report := r.Validation.Run(ctx, start)
	if !report.Valid() {
		return report, ErrValidationFailed
	}
	if streams.Err != nil {
		fmt.Fprintf(streams.Err, "Configuration valid: %s\n", report.Declaration)
	}

	plan, err := NewPlan(report)
	if err != nil {
		return report, err
	}
	exists, err := r.Sandboxes.SandboxExists(ctx, plan.Name)
	if err != nil {
		return report, err
	}
	err = r.Sandboxes.RunSandbox(ctx, sandboxport.RunRequest{
		Name:           plan.Name,
		Agent:          plan.Agent,
		Workspace:      plan.Workspace,
		Kits:           plan.Kits,
		AllowedSources: plan.Trust.AllowedSources,
		AllowLocalKits: plan.Trust.AllowLocalKits,
		Exists:         exists,
	}, sandboxport.Streams{
		In: streams.In, Out: streams.Out, Err: streams.Err,
	})
	if err != nil {
		return report, AttachedProcessError{Err: err}
	}
	return report, nil
}

// Package validation runs the gated repository declaration validation
// pipeline.
package validation

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jamessawle/sbxflow/internal/configuration"
	"github.com/jamessawle/sbxflow/internal/sbx"
)

const defaultLocalKitTimeout = 30 * time.Second

// Report contains the ordered, human-renderable validation result.
type Report struct {
	Declaration string
	Linked      configuration.LinkedConfiguration
	LocalKits   []LocalKitResult
	Errors      []error
}

// Valid reports whether every validation phase succeeded.
func (r Report) Valid() bool { return len(r.Errors) == 0 }

// Runner executes repository validation using injected subprocess execution.
type Runner struct {
	Sandboxes sbx.Client
}

// NewDefaultRunner constructs the production validation runner.
func NewDefaultRunner() Runner {
	return Runner{Sandboxes: sbx.NewClient(defaultLocalKitTimeout)}
}

// Run discovers and validates the nearest declaration. Each unsafe phase gates
// later phases; independent Docker local-kit failures are accumulated.
func (r Runner) Run(ctx context.Context, start string) Report {
	report := Report{}
	declaration, err := configuration.Discover(start)
	if err != nil {
		report.Errors = append(report.Errors, err)
		return report
	}
	report.Declaration = declaration

	data, err := os.ReadFile(declaration)
	if err != nil {
		report.Errors = append(report.Errors, fmt.Errorf("read repository declaration %q: %w", declaration, err))
		return report
	}
	document, err := configuration.Load(data)
	if err != nil {
		report.Errors = append(report.Errors, err)
		return report
	}
	linked, err := configuration.Link(document)
	if err != nil {
		report.Errors = append(report.Errors, err)
		return report
	}
	report.Linked = linked

	targets, err := configuration.ResolveLocalKits(declaration, linked)
	if err != nil {
		report.Errors = append(report.Errors, err)
		return report
	}
	report.LocalKits = validateLocalKits(ctx, targets, r.Sandboxes)
	for _, result := range report.LocalKits {
		if result.Err != nil {
			report.Errors = append(report.Errors, result.Err)
		}
	}
	return report
}

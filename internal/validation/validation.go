// Package validation runs the gated repository declaration validation
// pipeline.
package validation

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jamessawle/sbxflow/internal/command"
	"github.com/jamessawle/sbxflow/internal/config"
	"github.com/jamessawle/sbxflow/internal/localkit"
)

const defaultLocalKitTimeout = 30 * time.Second

// Report contains the ordered, human-renderable validation result.
type Report struct {
	Declaration string
	Linked      config.LinkedConfiguration
	LocalKits   []localkit.Result
	Errors      []error
}

// Valid reports whether every validation phase succeeded.
func (r Report) Valid() bool { return len(r.Errors) == 0 }

// Runner executes repository validation using injected subprocess execution.
type Runner struct {
	Commands command.Runner
}

// NewDefaultRunner constructs the production validation runner.
func NewDefaultRunner() Runner {
	return Runner{Commands: command.ExecRunner{Timeout: defaultLocalKitTimeout}}
}

// Run discovers and validates the nearest declaration. Each unsafe phase gates
// later phases; independent Docker local-kit failures are accumulated.
func (r Runner) Run(ctx context.Context, start string) Report {
	report := Report{}
	declaration, err := config.Discover(start)
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
	configuration, err := config.Load(data)
	if err != nil {
		report.Errors = append(report.Errors, err)
		return report
	}
	linked, err := config.Link(configuration)
	if err != nil {
		report.Errors = append(report.Errors, err)
		return report
	}
	report.Linked = linked

	targets, err := localkit.Resolve(declaration, linked)
	if err != nil {
		report.Errors = append(report.Errors, err)
		return report
	}
	report.LocalKits = localkit.Validate(ctx, targets, r.Commands)
	for _, result := range report.LocalKits {
		if result.Err != nil {
			report.Errors = append(report.Errors, result.Err)
		}
	}
	return report
}

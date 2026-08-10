// Package doctor inspects the local Docker Sandboxes installation and global
// configuration without modifying system or repository state.
package doctor

import (
	"context"
	"fmt"
	"strings"
)

// Status is the outcome of a diagnostic check.
type Status string

const (
	StatusPass Status = "pass"
	StatusWarn Status = "warn"
	StatusFail Status = "fail"
	StatusSkip Status = "skip"
)

// Grade determines whether a failed result affects the process exit status.
type Grade string

const (
	GradeRequired Grade = "required"
	GradeAdvisory Grade = "advisory"
)

// Fact identifies a prerequisite supplied by an earlier check.
type Fact string

const (
	FactSbxExecutable Fact = "sbx executable"
	FactSbxCompatible Fact = "compatible sbx version"
)

// Result is the stable output of one diagnostic check.
type Result struct {
	ID       string
	Status   Status
	Grade    Grade
	Summary  string
	Guidance string
	Provides map[Fact]string
}

// Environment contains shared, read-only diagnostic dependencies and facts.
type Environment struct {
	Sandboxes Inspector
	Facts     map[Fact]string
}

// Check performs one diagnostic inspection.
type Check interface {
	ID() string
	Grade() Grade
	Requires() []Fact
	Run(ctx context.Context, environment Environment) Result
}

// Report contains results in deterministic check registration order.
type Report struct {
	Results []Result
}

// Failed reports whether a required diagnostic failed.
func (r Report) Failed() bool {
	for _, result := range r.Results {
		if result.Grade == GradeRequired && result.Status == StatusFail {
			return true
		}
	}
	return false
}

// Runner executes diagnostic checks sequentially.
type Runner struct {
	sandboxes Inspector
	checks    []Check
}

// NewRunner constructs a runner using the checks in the supplied order.
func NewRunner(sandboxes Inspector, checks ...Check) Runner {
	return Runner{sandboxes: sandboxes, checks: checks}
}

// Run executes every check whose prerequisites are available. Missing
// prerequisites skip only the dependent check.
func (r Runner) Run(ctx context.Context) Report {
	facts := make(map[Fact]string)
	report := Report{Results: make([]Result, 0, len(r.checks))}

	for _, check := range r.checks {
		missing := missingFacts(check.Requires(), facts)
		if len(missing) > 0 {
			report.Results = append(report.Results, Result{
				ID:      check.ID(),
				Status:  StatusSkip,
				Grade:   check.Grade(),
				Summary: fmt.Sprintf("prerequisite unavailable: %s", strings.Join(missing, ", ")),
			})
			continue
		}

		result := check.Run(ctx, Environment{Sandboxes: r.sandboxes, Facts: facts})
		result.ID = check.ID()
		result.Grade = check.Grade()
		report.Results = append(report.Results, result)
		for fact, value := range result.Provides {
			facts[fact] = value
		}
	}

	return report
}

func missingFacts(required []Fact, facts map[Fact]string) []string {
	missing := make([]string, 0, len(required))
	for _, fact := range required {
		if _, ok := facts[fact]; !ok {
			missing = append(missing, string(fact))
		}
	}
	return missing
}

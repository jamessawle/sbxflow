package doctor

import (
	"context"
	"reflect"
	"testing"
)

func TestRunnerOrdersContinuesAndSkipsChecks(t *testing.T) {
	var invoked []string
	checks := []Check{
		fakeCheck{id: "compatibility", grade: GradeRequired, run: func(Environment) Result {
			invoked = append(invoked, "compatibility")
			return Result{Status: StatusFail, Summary: "unsupported"}
		}},
		fakeCheck{id: "diagnostics", grade: GradeRequired, requires: []Fact{FactSbxCompatible}, run: func(Environment) Result {
			invoked = append(invoked, "diagnostics")
			return Result{Status: StatusPass}
		}},
		fakeCheck{id: "independent", grade: GradeAdvisory, run: func(Environment) Result {
			invoked = append(invoked, "independent")
			return Result{Status: StatusWarn, Summary: "recommendation"}
		}},
	}

	report := NewRunner(nil, checks...).Run(context.Background())

	if want := []string{"compatibility", "independent"}; !reflect.DeepEqual(invoked, want) {
		t.Fatalf("invoked = %v, want %v", invoked, want)
	}
	if want := []Status{StatusFail, StatusSkip, StatusWarn}; !reflect.DeepEqual(statuses(report), want) {
		t.Fatalf("statuses = %v, want %v", statuses(report), want)
	}
	if !report.Failed() {
		t.Fatal("Failed() = false, want true for required failure")
	}
}

func TestRunnerSharesFactsAndWarningOnlySucceeds(t *testing.T) {
	checks := []Check{
		fakeCheck{id: "compatibility", grade: GradeRequired, run: func(Environment) Result {
			return Result{
				Status: StatusPass,
				Provides: map[Fact]string{
					FactSbxExecutable: "/usr/local/bin/sbx",
					FactSbxCompatible: "0.35.0",
				},
			}
		}},
		fakeCheck{id: "configuration", grade: GradeAdvisory, requires: []Fact{FactSbxCompatible}, run: func(environment Environment) Result {
			if got := environment.Facts[FactSbxExecutable]; got != "/usr/local/bin/sbx" {
				t.Fatalf("executable fact = %q", got)
			}
			return Result{Status: StatusWarn}
		}},
	}

	report := NewRunner(nil, checks...).Run(context.Background())
	if report.Failed() {
		t.Fatal("Failed() = true, want false for advisory warning")
	}
	if want := []Status{StatusPass, StatusWarn}; !reflect.DeepEqual(statuses(report), want) {
		t.Fatalf("statuses = %v, want %v", statuses(report), want)
	}
}

type fakeCheck struct {
	id       string
	grade    Grade
	requires []Fact
	run      func(Environment) Result
}

func (s fakeCheck) ID() string       { return s.id }
func (s fakeCheck) Grade() Grade     { return s.grade }
func (s fakeCheck) Requires() []Fact { return s.requires }
func (s fakeCheck) Run(_ context.Context, environment Environment) Result {
	return s.run(environment)
}

func statuses(report Report) []Status {
	statuses := make([]Status, 0, len(report.Results))
	for _, result := range report.Results {
		statuses = append(statuses, result.Status)
	}
	return statuses
}

package cli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jamessawle/sbxflow/internal/lifecycle"
	"github.com/jamessawle/sbxflow/internal/validation"
)

func TestUpHelpDoesNotInvokeRunner(t *testing.T) {
	for name, args := range map[string][]string{
		"flag":         {"up", "--help"},
		"help command": {"help", "up"},
	} {
		t.Run(name, func(t *testing.T) {
			runner := &fakeUpRunner{}
			stdout, stderr, err := executeWithUp(args, runner)
			if err != nil || stderr != "" || runner.calls != 0 {
				t.Fatalf("up help error = %v, stderr = %q, calls = %d", err, stderr, runner.calls)
			}
			for _, want := range []string{"sbxflow up", "interactively create or enter", "without reconciling"} {
				if !strings.Contains(stdout, want) {
					t.Errorf("stdout does not contain %q:\n%s", want, stdout)
				}
			}
		})
	}
}

func TestUpRejectsArgumentsAndFlagsWithoutInvokingRunner(t *testing.T) {
	for name, args := range map[string][]string{
		"argument": {"up", "extra"},
		"flag":     {"up", "--agent", "other"},
	} {
		t.Run(name, func(t *testing.T) {
			runner := &fakeUpRunner{}
			stdout, stderr, err := executeWithUp(args, runner)
			if err == nil || stdout != "" || stderr == "" || runner.calls != 0 {
				t.Fatalf("up invalid error = %v, stdout = %q, stderr = %q, calls = %d", err, stdout, stderr, runner.calls)
			}
		})
	}
}

func TestUpInjectsWorkingDirectoryAndStreams(t *testing.T) {
	runner := &fakeUpRunner{}
	stdout, stderr, err := executeWithUp([]string{"up"}, runner)
	if err != nil || stdout != "" || stderr != "" || runner.calls != 1 || runner.start == "" {
		t.Fatalf("up error = %v, stdout = %q, stderr = %q, runner = %#v", err, stdout, stderr, runner)
	}
	if runner.streams.In == nil || runner.streams.Out == nil || runner.streams.Err == nil {
		t.Fatalf("streams = %#v", runner.streams)
	}
}

func TestUpRendersValidationFailureOnce(t *testing.T) {
	runner := &fakeUpRunner{
		report: validation.Report{Errors: []error{errors.New("invalid declaration")}},
		err:    lifecycle.ErrValidationFailed,
	}
	stdout, stderr, err := executeWithUp([]string{"up"}, runner)
	if !errors.Is(err, errValidationFailed) || stdout != "" {
		t.Fatalf("up error = %v, stdout = %q", err, stdout)
	}
	if strings.Count(stderr, "invalid declaration") != 1 || strings.Contains(stderr, errValidationFailed.Error()) {
		t.Fatalf("stderr = %q, want one cohesive validation report", stderr)
	}
}

func TestUpDoesNotAppendAttachedProcessError(t *testing.T) {
	runner := &fakeUpRunner{err: lifecycle.AttachedProcessError{Err: errors.New("exit status 7")}}
	stdout, stderr, err := executeWithUp([]string{"up"}, runner)
	if err == nil || stdout != "" || stderr != "" {
		t.Fatalf("up error = %v, stdout = %q, stderr = %q", err, stdout, stderr)
	}
}

func executeWithUp(args []string, runner UpRunner) (string, string, error) {
	var stdout, stderr bytes.Buffer
	root := NewRootCommand(
		Streams{In: strings.NewReader("input"), Out: &stdout, Err: &stderr},
		fakeDoctorRunner{},
		fakeValidateRunner{},
		runner,
	)
	root.SetArgs(args)
	err := root.ExecuteContext(context.Background())
	return stdout.String(), stderr.String(), err
}

type fakeUpRunner struct {
	report  validation.Report
	err     error
	calls   int
	start   string
	streams lifecycle.Streams
}

func (r *fakeUpRunner) Run(_ context.Context, start string, streams lifecycle.Streams) (validation.Report, error) {
	r.calls++
	r.start = start
	r.streams = streams
	return r.report, r.err
}

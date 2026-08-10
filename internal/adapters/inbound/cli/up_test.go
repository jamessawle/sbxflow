package cli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jamessawle/sbxflow/internal/application/lifecycle"
	"github.com/jamessawle/sbxflow/internal/domain/configuration"
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
			for _, want := range []string{"sbxflow up", "interactively create or enter", "without reconciling", "--recreate", "force-removed", "persisted state"} {
				if !strings.Contains(stdout, want) {
					t.Errorf("stdout does not contain %q:\n%s", want, stdout)
				}
			}
		})
	}
}

func TestUpRejectsArgumentsAndFlagsWithoutInvokingRunner(t *testing.T) {
	for name, args := range map[string][]string{
		"argument":            {"up", "extra"},
		"short recreate flag": {"up", "-r"},
		"unrelated flag":      {"up", "--agent", "other"},
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

func TestUpInjectsWorkingDirectoryOptionsAndStreams(t *testing.T) {
	for _, test := range []struct {
		name         string
		args         []string
		wantRecreate bool
	}{
		{name: "default", args: []string{"up"}},
		{name: "recreate", args: []string{"up", "--recreate"}, wantRecreate: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			runner := &fakeUpRunner{stdout: "agent output\n", stderr: "Configuration valid: /repo/sbxflow.yaml\n"}
			stdout, stderr, err := executeWithUp(test.args, runner)
			if err != nil || stdout != "agent output\n" || stderr != "Configuration valid: /repo/sbxflow.yaml\n" || runner.calls != 1 || runner.start == "" {
				t.Fatalf("up error = %v, stdout = %q, stderr = %q, runner = %#v", err, stdout, stderr, runner)
			}
			if runner.options.Recreate != test.wantRecreate {
				t.Fatalf("options = %#v, want recreate %v", runner.options, test.wantRecreate)
			}
			if runner.streams.In == nil || runner.streams.Out == nil || runner.streams.Err == nil {
				t.Fatalf("streams = %#v", runner.streams)
			}
		})
	}
}

func TestUpRendersValidationFailureOnce(t *testing.T) {
	runner := &fakeUpRunner{
		report: configuration.Validation{Errors: []error{errors.New("invalid declaration")}},
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
		&fakeDownRunner{},
		&fakeDestroyRunner{},
	)
	root.SetArgs(args)
	err := root.ExecuteContext(context.Background())
	return stdout.String(), stderr.String(), err
}

type fakeUpRunner struct {
	report  configuration.Validation
	err     error
	stdout  string
	stderr  string
	calls   int
	start   string
	options lifecycle.UpOptions
	streams lifecycle.Streams
}

func (r *fakeUpRunner) Run(_ context.Context, start string, options lifecycle.UpOptions, streams lifecycle.Streams) (configuration.Validation, error) {
	r.calls++
	r.start = start
	r.options = options
	r.streams = streams
	_, _ = streams.Out.Write([]byte(r.stdout))
	_, _ = streams.Err.Write([]byte(r.stderr))
	return r.report, r.err
}

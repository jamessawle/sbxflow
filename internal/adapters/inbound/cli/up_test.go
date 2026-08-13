package cli

import (
	"bytes"
	"context"
	"errors"
	"io"
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
			for _, want := range []string{"sbxflow up", "interactively create or enter", "without reconciling", "--recreate", "--force", "requires confirmation", "bypasses", "permanent state loss", "attached sessions"} {
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
		"short force flag":    {"up", "-f"},
		"unrelated flag":      {"up", "--agent", "other"},
		"force alone":         {"up", "--force"},
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
		wantForce    bool
	}{
		{name: "default", args: []string{"up"}},
		{name: "recreate", args: []string{"up", "--recreate"}, wantRecreate: true},
		{name: "forced recreate", args: []string{"up", "--recreate", "--force"}, wantRecreate: true, wantForce: true},
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
			if runner.options.Force != test.wantForce {
				t.Fatalf("options = %#v, want force %v", runner.options, test.wantForce)
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

func TestRunningRecreationConfirmationUsesCommandStreamsAndDefaultsNegative(t *testing.T) {
	for _, test := range []struct {
		name      string
		input     string
		want      bool
		wantError bool
	}{
		{name: "lowercase yes", input: "yes\n", want: true},
		{name: "uppercase short", input: "Y\n", want: true},
		{name: "negative", input: "no\n"},
		{name: "empty", input: "\n"},
		{name: "malformed", input: "approve\n"},
		{name: "EOF-terminated affirmative", input: "yes", want: true},
		{name: "EOF-terminated negative", input: "no"},
		{name: "EOF-terminated malformed", input: "approve"},
		{name: "immediate EOF", input: "", wantError: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			var stderr bytes.Buffer
			got, err := (recreationConfirmer{}).ConfirmRunningSandboxRecreation("project", lifecycle.Streams{In: strings.NewReader(test.input), Err: &stderr})
			if got != test.want || (err != nil) != test.wantError {
				t.Fatalf("confirmation = %v, %v", got, err)
			}
			for _, want := range []string{"project", "permanently removes", "other attached terminal sessions", "[y/N]"} {
				if !strings.Contains(stderr.String(), want) {
					t.Errorf("warning does not contain %q: %s", want, stderr.String())
				}
			}
		})
	}
	var stderr bytes.Buffer
	_, err := (recreationConfirmer{}).ConfirmRunningSandboxRecreation("project", lifecycle.Streams{In: errorReader{}, Err: &stderr})
	if err == nil || !strings.Contains(err.Error(), "read confirmation") {
		t.Fatalf("read error = %v", err)
	}
	stderr.Reset()
	_, err = (recreationConfirmer{}).ConfirmRunningSandboxRecreation("project", lifecycle.Streams{In: &bufferedErrorReader{input: "yes"}, Err: &stderr})
	if err == nil || !strings.Contains(err.Error(), "read confirmation") {
		t.Fatalf("buffered read error = %v", err)
	}
	stderr.Reset()
	got, err := (recreationConfirmer{}).ConfirmRunningSandboxRecreation("project", lifecycle.Streams{In: newlineEOFReader{}, Err: &stderr})
	if got || err != nil {
		t.Fatalf("newline with EOF confirmation = %v, %v", got, err)
	}
	for _, test := range []struct {
		name  string
		input string
		want  bool
	}{
		{name: "affirmative", input: "yes\n", want: true},
		{name: "negative", input: "no"},
	} {
		t.Run("same-read EOF "+test.name, func(t *testing.T) {
			var stderr bytes.Buffer
			got, err := (recreationConfirmer{}).ConfirmRunningSandboxRecreation("project", lifecycle.Streams{In: &eofWithDataReader{input: test.input}, Err: &stderr})
			if got != test.want || err != nil {
				t.Fatalf("confirmation = %v, %v", got, err)
			}
		})
	}
}

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }

type newlineEOFReader struct{}

func (newlineEOFReader) Read(buffer []byte) (int, error) {
	buffer[0] = '\n'
	return 1, io.EOF
}

type eofWithDataReader struct {
	input string
	index int
}

func (reader *eofWithDataReader) Read(buffer []byte) (int, error) {
	buffer[0] = reader.input[reader.index]
	reader.index++
	if reader.index == len(reader.input) {
		return 1, io.EOF
	}
	return 1, nil
}

type bufferedErrorReader struct {
	input string
	index int
}

func (reader *bufferedErrorReader) Read(buffer []byte) (int, error) {
	if reader.index == len(reader.input) {
		return 0, io.ErrUnexpectedEOF
	}
	buffer[0] = reader.input[reader.index]
	reader.index++
	return 1, nil
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

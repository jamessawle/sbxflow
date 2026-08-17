package cli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jamessawle/sbxflow/internal/application/lifecycle"
)

func TestDownHelpDoesNotInvokeRunner(t *testing.T) {
	for name, args := range map[string][]string{
		"flag":         {"down", "--help"},
		"help command": {"help", "down"},
	} {
		t.Run(name, func(t *testing.T) {
			runner := &fakeDownRunner{}
			stdout, stderr, err := executeWithDown(args, runner)
			if err != nil || stderr != "" || runner.calls != 0 {
				t.Fatalf("down help error = %v, stderr = %q, calls = %d", err, stderr, runner.calls)
			}
			for _, want := range []string{"sbxflow down", "without removing", "persisted sandbox state", "later up"} {
				if !strings.Contains(stdout, want) {
					t.Errorf("stdout does not contain %q:\n%s", want, stdout)
				}
			}
		})
	}
}

func TestDownRejectsArgumentsAndFlagsWithoutInvokingRunner(t *testing.T) {
	for name, args := range map[string][]string{
		"argument": {"down", "extra"},
		"flag":     {"down", "--name", "other"},
	} {
		t.Run(name, func(t *testing.T) {
			runner := &fakeDownRunner{}
			stdout, stderr, err := executeWithDown(args, runner)
			if err == nil || stdout != "" || stderr == "" || runner.calls != 0 {
				t.Fatalf("down invalid error = %v, stdout = %q, stderr = %q, calls = %d", err, stdout, stderr, runner.calls)
			}
		})
	}
}

func TestDownInjectsWorkingDirectoryAndOutputStreams(t *testing.T) {
	runner := &fakeDownRunner{stdout: "stopped\n", stderr: "docker diagnostic\n"}
	stdout, stderr, err := executeWithDown([]string{"down"}, runner)
	if err != nil || stdout != "stopped\n" || stderr != "docker diagnostic\n" || runner.calls != 1 || runner.start == "" {
		t.Fatalf("down error = %v, stdout = %q, stderr = %q, runner = %#v", err, stdout, stderr, runner)
	}
	if runner.streams.In != nil || runner.streams.Out == nil || runner.streams.Err == nil {
		t.Fatalf("streams = %#v", runner.streams)
	}
}

func TestDownRendersTargetFailureOnce(t *testing.T) {
	runner := &fakeDownRunner{err: errors.New("no sbxflow.yaml found")}
	stdout, stderr, err := executeWithDown([]string{"down"}, runner)
	if err == nil || stdout != "" || strings.Count(stderr, "no sbxflow.yaml found") != 1 {
		t.Fatalf("down error = %v, stdout = %q, stderr = %q", err, stdout, stderr)
	}
}

func TestDownDoesNotAppendAttachedProcessError(t *testing.T) {
	runner := &fakeDownRunner{
		stderr: "docker stop diagnostic\n",
		err:    lifecycle.AttachedProcessError{Err: errors.New("exit status 7")},
	}
	stdout, stderr, err := executeWithDown([]string{"down"}, runner)
	if err == nil || stdout != "" || stderr != "docker stop diagnostic\n" || strings.Contains(stderr, "exit status 7") {
		t.Fatalf("down error = %v, stdout = %q, stderr = %q", err, stdout, stderr)
	}
}

func executeWithDown(args []string, runner DownRunner) (string, string, error) {
	var stdout, stderr bytes.Buffer
	root := NewRootCommand(
		Streams{In: strings.NewReader("input"), Out: &stdout, Err: &stderr},
		NewDownCommand(runner),
	)
	root.SetArgs(args)
	err := root.ExecuteContext(context.Background())
	return stdout.String(), stderr.String(), err
}

type fakeDownRunner struct {
	err     error
	stdout  string
	stderr  string
	calls   int
	start   string
	streams lifecycle.Streams
}

func (r *fakeDownRunner) Run(_ context.Context, start string, streams lifecycle.Streams) error {
	r.calls++
	r.start = start
	r.streams = streams
	_, _ = streams.Out.Write([]byte(r.stdout))
	_, _ = streams.Err.Write([]byte(r.stderr))
	return r.err
}

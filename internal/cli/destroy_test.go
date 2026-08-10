package cli

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/jamessawle/sbxflow/internal/lifecycle"
)

func TestDestroyHelpDoesNotInvokeRunner(t *testing.T) {
	for name, args := range map[string][]string{
		"flag":         {"destroy", "--help"},
		"help command": {"help", "destroy"},
	} {
		t.Run(name, func(t *testing.T) {
			runner := &fakeDestroyRunner{}
			stdout, stderr, err := executeWithDestroy(args, "input", runner)
			if err != nil || stderr != "" || runner.calls != 0 {
				t.Fatalf("destroy help error = %v, stderr = %q, calls = %d", err, stderr, runner.calls)
			}
			for _, want := range []string{"sbxflow destroy", "permanently remove", "Unlike down", "--force", "active session"} {
				if !strings.Contains(stdout, want) {
					t.Errorf("stdout does not contain %q:\n%s", want, stdout)
				}
			}
		})
	}
}

func TestDestroyRejectsArgumentsAndUnsupportedFlagsWithoutInvokingRunner(t *testing.T) {
	for name, args := range map[string][]string{
		"argument":     {"destroy", "other"},
		"all":          {"destroy", "--all"},
		"name":         {"destroy", "--name", "other"},
		"unknown flag": {"destroy", "--unknown"},
	} {
		t.Run(name, func(t *testing.T) {
			runner := &fakeDestroyRunner{}
			stdout, stderr, err := executeWithDestroy(args, "input", runner)
			if err == nil || stdout != "" || stderr == "" || runner.calls != 0 {
				t.Fatalf("destroy invalid error = %v, stdout = %q, stderr = %q, calls = %d", err, stdout, stderr, runner.calls)
			}
		})
	}
}

func TestDestroyInjectsWorkingDirectoryForceAndAllStreams(t *testing.T) {
	for name, test := range map[string]struct {
		args  []string
		force bool
	}{
		"confirmed":   {args: []string{"destroy"}},
		"long force":  {args: []string{"destroy", "--force"}, force: true},
		"short force": {args: []string{"destroy", "-f"}, force: true},
	} {
		t.Run(name, func(t *testing.T) {
			runner := &fakeDestroyRunner{stdout: "removed\n", stderr: "docker diagnostic\n"}
			stdout, stderr, err := executeWithDestroy(test.args, "yes\n", runner)
			if err != nil || stdout != "removed\n" || stderr != "docker diagnostic\n" || runner.calls != 1 || runner.start == "" || runner.force != test.force {
				t.Fatalf("destroy error = %v, stdout = %q, stderr = %q, runner = %#v", err, stdout, stderr, runner)
			}
			if got, _ := io.ReadAll(runner.streams.In); string(got) != "yes\n" || runner.streams.Out == nil || runner.streams.Err == nil {
				t.Fatalf("streams = %#v, stdin = %q", runner.streams, got)
			}
		})
	}
}

func TestDestroyRendersTargetFailureOnce(t *testing.T) {
	runner := &fakeDestroyRunner{err: errors.New("no sbxflow.yaml found")}
	stdout, stderr, err := executeWithDestroy([]string{"destroy"}, "", runner)
	if err == nil || stdout != "" || strings.Count(stderr, "no sbxflow.yaml found") != 1 {
		t.Fatalf("destroy error = %v, stdout = %q, stderr = %q", err, stdout, stderr)
	}
}

func TestDestroyDoesNotAppendAttachedProcessError(t *testing.T) {
	runner := &fakeDestroyRunner{
		stderr: "docker remove diagnostic\n",
		err:    lifecycle.AttachedProcessError{Err: errors.New("exit status 7")},
	}
	stdout, stderr, err := executeWithDestroy([]string{"destroy"}, "", runner)
	if err == nil || stdout != "" || stderr != "docker remove diagnostic\n" || strings.Contains(stderr, "exit status 7") {
		t.Fatalf("destroy error = %v, stdout = %q, stderr = %q", err, stdout, stderr)
	}
}

func executeWithDestroy(args []string, input string, runner DestroyRunner) (string, string, error) {
	var stdout, stderr bytes.Buffer
	root := NewRootCommand(
		Streams{In: strings.NewReader(input), Out: &stdout, Err: &stderr},
		fakeDoctorRunner{},
		fakeValidateRunner{},
		&fakeUpRunner{},
		&fakeDownRunner{},
		runner,
	)
	root.SetArgs(args)
	err := root.ExecuteContext(context.Background())
	return stdout.String(), stderr.String(), err
}

type fakeDestroyRunner struct {
	err     error
	stdout  string
	stderr  string
	calls   int
	start   string
	force   bool
	streams lifecycle.Streams
}

func (r *fakeDestroyRunner) Run(_ context.Context, start string, force bool, streams lifecycle.Streams) error {
	r.calls++
	r.start = start
	r.force = force
	r.streams = streams
	_, _ = streams.Out.Write([]byte(r.stdout))
	_, _ = streams.Err.Write([]byte(r.stderr))
	return r.err
}

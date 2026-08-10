package lifecycle

import (
	"bytes"
	"context"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/jamessawle/sbxflow/internal/command"
	"github.com/jamessawle/sbxflow/internal/config"
	"github.com/jamessawle/sbxflow/internal/validation"
)

func TestSandboxExistsUsesExactNonemptyLines(t *testing.T) {
	for _, test := range []struct {
		name   string
		output string
		want   bool
	}{
		{name: "exact", output: "other\nproject\n", want: true},
		{name: "prefix", output: "project-extra\n", want: false},
		{name: "surrounding whitespace", output: " project \n", want: false},
		{name: "empty", output: "\n\n", want: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			commands := &fakeCommandRunner{output: command.Output{Stdout: []byte(test.output)}}
			got, err := sandboxExists(context.Background(), commands, "/bin/sbx", "project")
			if err != nil || got != test.want {
				t.Fatalf("sandboxExists() = %v, %v, want %v, nil", got, err, test.want)
			}
			if !reflect.DeepEqual(commands.args, []string{"ls", "--quiet"}) {
				t.Fatalf("args = %#v", commands.args)
			}
		})
	}
}

func TestSandboxExistsReportsDockerDiagnostics(t *testing.T) {
	commands := &fakeCommandRunner{output: command.Output{Stderr: []byte("daemon unavailable\n"), ExitCode: 7, Err: errors.New("exit 7")}}
	_, err := sandboxExists(context.Background(), commands, "/bin/sbx", "project")
	if err == nil || !strings.Contains(err.Error(), "sbx ls --quiet") || !strings.Contains(err.Error(), "daemon unavailable") {
		t.Fatalf("sandboxExists() error = %v", err)
	}
}

func TestRunnerValidationGatesLifecycleLookup(t *testing.T) {
	commands := &fakeCommandRunner{}
	var stderr bytes.Buffer
	runner := Runner{
		Validation:  fakeValidator{report: validation.Report{Errors: []error{errors.New("invalid")}}},
		Commands:    commands,
		Interactive: &fakeInteractiveRunner{},
	}
	_, err := runner.Run(context.Background(), "/repo", Streams{Err: &stderr})
	if !errors.Is(err, ErrValidationFailed) || commands.lookups != 0 || commands.runs != 0 {
		t.Fatalf("Run() error = %v, lookups = %d, runs = %d", err, commands.lookups, commands.runs)
	}
	if stderr.String() != "" {
		t.Fatalf("stderr = %q, want no validation success status", stderr.String())
	}
}

func TestRunnerReportsValidationSuccessBeforeLifecycleLookup(t *testing.T) {
	var stderr bytes.Buffer
	commands := &fakeCommandRunner{
		path: "/bin/sbx",
		onLookup: func() {
			if got, want := stderr.String(), "Configuration valid: /repo/sbxflow.yaml\n"; got != want {
				t.Fatalf("stderr before lookup = %q, want %q", got, want)
			}
		},
	}
	runner := Runner{
		Validation:  fakeValidator{report: validReport()},
		Commands:    commands,
		Interactive: &fakeInteractiveRunner{},
	}
	_, err := runner.Run(context.Background(), "/repo", Streams{Err: &stderr})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if got, want := stderr.String(), "Configuration valid: /repo/sbxflow.yaml\n"; got != want {
		t.Fatalf("stderr = %q, want %q", got, want)
	}
}

func TestRunnerLookupFailureDoesNotRunAgent(t *testing.T) {
	interactive := &fakeInteractiveRunner{}
	commands := &fakeCommandRunner{path: "/bin/sbx", output: command.Output{Err: errors.New("lookup failed")}}
	runner := Runner{Validation: fakeValidator{report: validReport()}, Commands: commands, Interactive: interactive}
	_, err := runner.Run(context.Background(), "/repo", Streams{})
	if err == nil || interactive.calls != 0 {
		t.Fatalf("Run() error = %v, interactive calls = %d", err, interactive.calls)
	}
}

func TestRunnerSelectsExactMissingAndExistingArguments(t *testing.T) {
	for _, test := range []struct {
		name       string
		listOutput string
		wantArgs   []string
	}{
		{
			name:       "missing",
			listOutput: "other\n",
			wantArgs:   []string{"run", "--name", "project", "--kit", "git+https://github.com/example/kits.git#ref=v1&dir=tooling", "codex", "/repo"},
		},
		{
			name:       "existing",
			listOutput: "project\n",
			wantArgs:   []string{"run", "codex", "--name", "project"},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			commands := &fakeCommandRunner{path: "/bin/sbx", output: command.Output{Stdout: []byte(test.listOutput)}}
			interactive := &fakeInteractiveRunner{}
			var stdout, stderr bytes.Buffer
			stdin := strings.NewReader("input")
			runner := Runner{Validation: fakeValidator{report: validReport()}, Commands: commands, Interactive: interactive}
			_, err := runner.Run(context.Background(), "/repo/nested", Streams{In: stdin, Out: &stdout, Err: &stderr})
			if err != nil {
				t.Fatalf("Run() error = %v", err)
			}
			if !reflect.DeepEqual(interactive.invocation.Args, test.wantArgs) {
				t.Fatalf("args = %#v, want %#v", interactive.invocation.Args, test.wantArgs)
			}
			if interactive.invocation.Executable != "/bin/sbx" || interactive.invocation.Stdin != stdin || interactive.invocation.Stdout != &stdout || interactive.invocation.Stderr != &stderr {
				t.Fatalf("interactive invocation = %#v", interactive.invocation)
			}
			if got := interactive.invocation.Environment[allowedSourcesEnvironment]; got != `["docker.io/","github.com/example/kits"]` {
				t.Fatalf("allowed sources = %q", got)
			}
		})
	}
}

func TestRunnerMarksAttachedProcessFailureAsRendered(t *testing.T) {
	interactive := &fakeInteractiveRunner{err: errors.New("exit 9")}
	runner := Runner{
		Validation:  fakeValidator{report: validReport()},
		Commands:    &fakeCommandRunner{path: "/bin/sbx", output: command.Output{}},
		Interactive: interactive,
	}
	_, err := runner.Run(context.Background(), "/repo", Streams{Err: io.Discard})
	var attached AttachedProcessError
	if !errors.As(err, &attached) || !strings.Contains(err.Error(), "exit 9") {
		t.Fatalf("Run() error = %T %v", err, err)
	}
}

func validReport() validation.Report {
	return validation.Report{
		Declaration: "/repo/sbxflow.yaml",
		Linked: config.LinkedConfiguration{
			Configuration: config.Configuration{Sandbox: config.Sandbox{Name: "project", Agent: "codex"}},
			Selections: []config.LinkedSelection{{
				Index:           0,
				Source:          config.Source{Type: config.SourceGit},
				RemoteReference: "git+https://github.com/example/kits.git#ref=v1&dir=tooling",
			}},
			Trust: config.Trust{AllowedSources: []string{"docker.io/", "github.com/example/kits"}},
		},
	}
}

type fakeValidator struct{ report validation.Report }

func (v fakeValidator) Run(context.Context, string) validation.Report { return v.report }

type fakeCommandRunner struct {
	path     string
	pathErr  error
	output   command.Output
	args     []string
	onLookup func()
	lookups  int
	runs     int
}

func (r *fakeCommandRunner) LookPath(string) (string, error) {
	r.lookups++
	if r.onLookup != nil {
		r.onLookup()
	}
	return r.path, r.pathErr
}

func (r *fakeCommandRunner) Run(_ context.Context, _ string, args ...string) command.Output {
	r.runs++
	r.args = append([]string(nil), args...)
	return r.output
}

type fakeInteractiveRunner struct {
	calls      int
	invocation command.InteractiveInvocation
	err        error
}

func (r *fakeInteractiveRunner) Run(_ context.Context, invocation command.InteractiveInvocation) error {
	r.calls++
	r.invocation = invocation
	return r.err
}

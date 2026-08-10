package lifecycle

import (
	"bytes"
	"context"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/jamessawle/sbxflow/internal/adapters/outbound/sbx"
)

func TestDownRunnerResolutionFailurePreventsDockerLookup(t *testing.T) {
	commands := &fakeCommandRunner{}
	runner := DownRunner{
		Targets:   fakeTargetResolver{err: errors.New("no sbxflow.yaml found")},
		Sandboxes: sbx.Client{Commands: commands, Interactive: &fakeInteractiveRunner{}},
	}
	err := runner.Run(context.Background(), "/repo", Streams{})
	if err == nil || commands.lookups != 0 || commands.runs != 0 {
		t.Fatalf("Run() error = %v, lookups = %d, runs = %d", err, commands.lookups, commands.runs)
	}
}

func TestDownRunnerReportsExecutableLookupFailure(t *testing.T) {
	want := errors.New("not installed")
	runner := DownRunner{
		Targets:   fakeTargetResolver{target: Target{Name: "project"}},
		Sandboxes: sbx.Client{Commands: &fakeCommandRunner{pathErr: want}, Interactive: &fakeInteractiveRunner{}},
	}
	err := runner.Run(context.Background(), "/repo", Streams{})
	if !errors.Is(err, want) || !strings.Contains(err.Error(), "locate sbx") {
		t.Fatalf("Run() error = %v, want contextual executable failure", err)
	}
}

func TestDownRunnerAbsentAndSimilarNamesAreNoOps(t *testing.T) {
	for _, output := range []string{"", "other\n", "project-extra\nprefix-project\n"} {
		commands := &fakeCommandRunner{path: "/bin/sbx", output: sbx.Output{Stdout: []byte(output)}}
		interactive := &fakeInteractiveRunner{}
		runner := DownRunner{
			Targets:   fakeTargetResolver{target: Target{Name: "project"}},
			Sandboxes: sbx.Client{Commands: commands, Interactive: interactive},
		}
		if err := runner.Run(context.Background(), "/repo", Streams{}); err != nil {
			t.Fatalf("Run() error = %v for list output %q", err, output)
		}
		if commands.runs != 1 || !reflect.DeepEqual(commands.args, []string{"ls", "--quiet"}) || interactive.calls != 0 {
			t.Fatalf("list output %q: commands = %#v, interactive calls = %d", output, commands, interactive.calls)
		}
	}
}

func TestDownRunnerStopsExactExistingSandboxWithAttachedOutput(t *testing.T) {
	commands := &fakeCommandRunner{path: "/bin/sbx", output: sbx.Output{Stdout: []byte("other\nproject\n")}}
	interactive := &fakeInteractiveRunner{}
	var stdout, stderr bytes.Buffer
	runner := DownRunner{
		Targets:   fakeTargetResolver{target: Target{Declaration: "/repo/sbxflow.yaml", Name: "project"}},
		Sandboxes: sbx.Client{Commands: commands, Interactive: interactive},
	}
	ctx := context.WithValue(context.Background(), downContextKey{}, "preserved")
	if err := runner.Run(ctx, "/repo/nested", Streams{In: strings.NewReader("unused"), Out: &stdout, Err: &stderr}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	want := sbx.InteractiveInvocation{
		Executable: "/bin/sbx",
		Args:       []string{"stop", "project"},
		Stdout:     &stdout,
		Stderr:     &stderr,
	}
	if !reflect.DeepEqual(interactive.invocation, want) {
		t.Fatalf("invocation = %#v, want %#v", interactive.invocation, want)
	}
	if interactive.ctx.Value(downContextKey{}) != "preserved" {
		t.Fatal("stop invocation did not preserve lifecycle context")
	}
}

func TestDownRunnerAlreadyStoppedReliesOnSuccessfulDockerStop(t *testing.T) {
	interactive := &fakeInteractiveRunner{}
	runner := DownRunner{
		Targets:   fakeTargetResolver{target: Target{Name: "project"}},
		Sandboxes: sbx.Client{Commands: &fakeCommandRunner{path: "/bin/sbx", output: sbx.Output{Stdout: []byte("project\n")}}, Interactive: interactive},
	}
	if err := runner.Run(context.Background(), "/repo", Streams{}); err != nil || interactive.calls != 1 {
		t.Fatalf("Run() error = %v, stop calls = %d", err, interactive.calls)
	}
}

func TestDownRunnerPreservesListFailureAndSkipsStop(t *testing.T) {
	interactive := &fakeInteractiveRunner{}
	runner := DownRunner{
		Targets: fakeTargetResolver{target: Target{Name: "project"}},
		Sandboxes: sbx.Client{
			Commands: &fakeCommandRunner{path: "/bin/sbx", output: sbx.Output{
				Stderr: []byte("daemon unavailable\n"), ExitCode: 7, Err: errors.New("exit 7"),
			}},
			Interactive: interactive,
		},
	}
	err := runner.Run(context.Background(), "/repo", Streams{})
	if err == nil || !strings.Contains(err.Error(), "daemon unavailable") || interactive.calls != 0 {
		t.Fatalf("Run() error = %v, stop calls = %d", err, interactive.calls)
	}
}

func TestDownRunnerMarksAttachedStopFailureAsRendered(t *testing.T) {
	want := errors.New("exit status 9")
	runner := DownRunner{
		Targets:   fakeTargetResolver{target: Target{Name: "project"}},
		Sandboxes: sbx.Client{Commands: &fakeCommandRunner{path: "/bin/sbx", output: sbx.Output{Stdout: []byte("project\n")}}, Interactive: &fakeInteractiveRunner{err: want}},
	}
	err := runner.Run(context.Background(), "/repo", Streams{Err: io.Discard})
	var attached AttachedProcessError
	if !errors.As(err, &attached) || !errors.Is(err, want) {
		t.Fatalf("Run() error = %T %v, want attached process failure", err, err)
	}
}

type fakeTargetResolver struct {
	target Target
	err    error
}

func (r fakeTargetResolver) Resolve(string) (Target, error) { return r.target, r.err }

type downContextKey struct{}

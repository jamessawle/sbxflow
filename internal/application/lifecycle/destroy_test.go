package lifecycle

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/jamessawle/sbxflow/internal/adapters/outbound/sbx"
)

func TestDestroyRunnerResolutionFailurePreventsDockerLookup(t *testing.T) {
	commands := &fakeCommandRunner{}
	runner := DestroyRunner{
		Targets:   fakeTargetResolver{err: errors.New("no sbxflow.yaml found")},
		Sandboxes: sbx.Client{Commands: commands, Interactive: &fakeInteractiveRunner{}},
	}
	err := runner.Run(context.Background(), "/repo", false, Streams{})
	if err == nil || commands.lookups != 0 || commands.runs != 0 {
		t.Fatalf("Run() error = %v, lookups = %d, runs = %d", err, commands.lookups, commands.runs)
	}
}

func TestDestroyRunnerReportsExecutableLookupFailure(t *testing.T) {
	want := errors.New("not installed")
	runner := DestroyRunner{
		Targets:   fakeTargetResolver{target: Target{Name: "project"}},
		Sandboxes: sbx.Client{Commands: &fakeCommandRunner{pathErr: want}, Interactive: &fakeInteractiveRunner{}},
	}
	err := runner.Run(context.Background(), "/repo", false, Streams{})
	if !errors.Is(err, want) || !strings.Contains(err.Error(), "locate sbx") {
		t.Fatalf("Run() error = %v, want contextual executable failure", err)
	}
}

func TestDestroyRunnerAbsentAndSimilarNamesAreNoOps(t *testing.T) {
	for _, output := range []string{"", "other\n", "project-extra\nprefix-project\n"} {
		commands := &fakeCommandRunner{path: "/bin/sbx", output: sbx.Output{Stdout: []byte(output)}}
		interactive := &fakeInteractiveRunner{}
		runner := DestroyRunner{
			Targets:   fakeTargetResolver{target: Target{Name: "project"}},
			Sandboxes: sbx.Client{Commands: commands, Interactive: interactive},
		}
		if err := runner.Run(context.Background(), "/repo", true, Streams{}); err != nil {
			t.Fatalf("Run() error = %v for list output %q", err, output)
		}
		if commands.runs != 1 || !reflect.DeepEqual(commands.args, []string{"ls", "--quiet"}) || interactive.calls != 0 {
			t.Fatalf("list output %q: commands = %#v, interactive calls = %d", output, commands, interactive.calls)
		}
	}
}

func TestDestroyRunnerRemovesExactSandboxWithSelectedSafeguardsAndAttachedStreams(t *testing.T) {
	for _, test := range []struct {
		name  string
		force bool
		args  []string
	}{
		{name: "confirmed", args: []string{"rm", "project"}},
		{name: "forced", force: true, args: []string{"rm", "--force", "project"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			commands := &fakeCommandRunner{path: "/bin/sbx", output: sbx.Output{Stdout: []byte("other\nproject\n")}}
			interactive := &fakeInteractiveRunner{checkContext: func(ctx context.Context) {
				if ctx.Value(destroyContextKey{}) != "preserved" {
					t.Fatal("remove invocation did not preserve lifecycle context")
				}
			}}
			var stdout, stderr bytes.Buffer
			stdin := strings.NewReader("yes\n")
			runner := DestroyRunner{
				Targets:   fakeTargetResolver{target: Target{Declaration: "/repo/sbxflow.yaml", Name: "project"}},
				Sandboxes: sbx.Client{Commands: commands, Interactive: interactive},
			}
			ctx := context.WithValue(context.Background(), destroyContextKey{}, "preserved")
			if err := runner.Run(ctx, "/repo/nested", test.force, Streams{In: stdin, Out: &stdout, Err: &stderr}); err != nil {
				t.Fatalf("Run() error = %v", err)
			}
			want := sbx.InteractiveInvocation{
				Executable: "/bin/sbx",
				Args:       test.args,
				Stdin:      stdin,
				Stdout:     &stdout,
				Stderr:     &stderr,
			}
			if !reflect.DeepEqual(interactive.invocation, want) {
				t.Fatalf("invocation = %#v, want %#v", interactive.invocation, want)
			}
		})
	}
}

func TestDestroyRunnerPreservesListFailureAndSkipsRemoval(t *testing.T) {
	interactive := &fakeInteractiveRunner{}
	runner := DestroyRunner{
		Targets: fakeTargetResolver{target: Target{Name: "project"}},
		Sandboxes: sbx.Client{
			Commands: &fakeCommandRunner{path: "/bin/sbx", output: sbx.Output{
				Stderr: []byte("daemon unavailable\n"), ExitCode: 7, Err: errors.New("exit 7"),
			}},
			Interactive: interactive,
		},
	}
	err := runner.Run(context.Background(), "/repo", false, Streams{})
	if err == nil || !strings.Contains(err.Error(), "daemon unavailable") || interactive.calls != 0 {
		t.Fatalf("Run() error = %v, remove calls = %d", err, interactive.calls)
	}
}

func TestDestroyRunnerMarksAttachedRemovalFailureAsRendered(t *testing.T) {
	want := errors.New("exit status 9")
	runner := DestroyRunner{
		Targets:   fakeTargetResolver{target: Target{Name: "project"}},
		Sandboxes: sbx.Client{Commands: &fakeCommandRunner{path: "/bin/sbx", output: sbx.Output{Stdout: []byte("project\n")}}, Interactive: &fakeInteractiveRunner{err: want}},
	}
	err := runner.Run(context.Background(), "/repo", false, Streams{})
	var attached AttachedProcessError
	if !errors.As(err, &attached) || !errors.Is(err, want) {
		t.Fatalf("Run() error = %T %v, want attached process failure", err, err)
	}
}

type destroyContextKey struct{}

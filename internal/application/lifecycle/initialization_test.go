package lifecycle

import (
	"bytes"
	"context"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/jamessawle/sbxflow/internal/adapters/outbound/sbx"
	"github.com/jamessawle/sbxflow/internal/domain/configuration"
	sandboxport "github.com/jamessawle/sbxflow/internal/ports/sandbox"
)

func TestSandboxClientExecutesLiteralCommandWithMatchingOutput(t *testing.T) {
	commands := &fakeCommandRunner{path: "/bin/sbx"}
	interactive := &fakeInteractiveRunner{err: errors.New("exit 9")}
	var stdout, stderr bytes.Buffer
	command := []string{"printf", "%s", "$HOME; literal"}
	err := (sbx.Client{Commands: commands, Interactive: interactive}).ExecuteCommand(context.Background(), sandboxport.CommandRequest{
		Environment: sandboxport.Environment{Name: "exact-sandbox", Agent: "codex", Workspace: t.TempDir()}, Command: command,
	}, sandboxport.Streams{Out: &stdout, Err: &stderr})
	if err == nil || interactive.calls != 1 {
		t.Fatalf("ExecuteCommand() error = %v, calls = %d", err, interactive.calls)
	}
	invocation := interactive.invocation
	if len(invocation.Args) != 7 || !reflect.DeepEqual(invocation.Args[:2], []string{"env", "exec"}) || invocation.Args[3] != "--" || !reflect.DeepEqual(invocation.Args[4:], command) || invocation.Stdin != nil || invocation.Stdout != &stdout || invocation.Stderr != &stderr {
		t.Fatalf("invocation = %#v, want literal non-interactive invocation", invocation)
	}
	if _, err := os.Stat(invocation.Args[2]); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("temporary environment still exists or stat failed: %v", err)
	}
}

func initializationReport() configuration.Validation {
	report := networkReport()
	report.Linked.Configuration.Sandbox.Hooks.Initialize = []configuration.Command{
		{Command: []string{"bash", "-c", "printf 'one' && printf 'err' >&2"}},
		{Command: []string{"printf", "%s", "$HOME; literal"}},
	}
	return report
}

func TestUpInitializesCreatedSandboxInOrderBeforeAttachment(t *testing.T) {
	sandboxes := &fakeUpSandboxes{}
	var stdout, stderr bytes.Buffer
	stdin := strings.NewReader("agent input")
	runner := UpRunner{Validation: fakeValidator{report: initializationReport()}, Sandboxes: sandboxes}
	if _, err := runner.Run(context.Background(), "/repo", UpOptions{}, Streams{In: stdin, Out: &stdout, Err: &stderr}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	wantCalls := []string{"lookup project", "create project", "allow project", "execute project", "execute project", "run project"}
	if !reflect.DeepEqual(sandboxes.calls, wantCalls) {
		t.Fatalf("calls = %#v, want %#v", sandboxes.calls, wantCalls)
	}
	wantRequests := []sandboxport.CommandRequest{
		{Environment: expectedEnvironment(), Command: []string{"bash", "-c", "printf 'one' && printf 'err' >&2"}},
		{Environment: expectedEnvironment(), Command: []string{"printf", "%s", "$HOME; literal"}},
	}
	if !reflect.DeepEqual(sandboxes.executeRequests, wantRequests) {
		t.Fatalf("execute requests = %#v, want %#v", sandboxes.executeRequests, wantRequests)
	}
	for _, streams := range sandboxes.executeStreams {
		if streams.In != nil || streams.Out != &stdout || streams.Err != &stderr {
			t.Fatalf("execute streams = %#v, want non-interactive matching output", streams)
		}
	}
}

func TestUpSkipsInitializationForExistingSandbox(t *testing.T) {
	for _, state := range []sandboxport.State{sandboxport.StateRunning, sandboxport.StateStopped} {
		sandboxes := &fakeUpSandboxes{state: state}
		runner := UpRunner{Validation: fakeValidator{report: initializationReport()}, Sandboxes: sandboxes}
		if _, err := runner.Run(context.Background(), "/repo", UpOptions{}, Streams{}); err != nil {
			t.Fatalf("Run(%s) error = %v", state, err)
		}
		if !reflect.DeepEqual(sandboxes.calls, []string{"lookup project", "run project"}) {
			t.Fatalf("Run(%s) calls = %#v", state, sandboxes.calls)
		}
	}
}

func TestInitializationFailureStopsAndRollsBack(t *testing.T) {
	want := errors.New("exit 7")
	sandboxes := &fakeUpSandboxes{executeErr: want}
	runner := UpRunner{Validation: fakeValidator{report: initializationReport()}, Sandboxes: sandboxes}
	_, err := runner.Run(context.Background(), "/repo", UpOptions{}, Streams{})
	if !errors.Is(err, want) || !strings.Contains(err.Error(), "initialize command 1") || !strings.Contains(err.Error(), "bash") {
		t.Fatalf("Run() error = %v", err)
	}
	wantCalls := []string{"lookup project", "create project", "allow project", "execute project", "remove project"}
	if !reflect.DeepEqual(sandboxes.calls, wantCalls) || len(sandboxes.executeRequests) != 1 {
		t.Fatalf("calls = %#v, requests = %#v", sandboxes.calls, sandboxes.executeRequests)
	}
}

func TestInitializationAndCleanupFailuresAreReportedTogether(t *testing.T) {
	initializationErr := errors.New("initialization failed")
	sandboxes := &fakeUpSandboxes{executeErr: initializationErr, removeErr: errors.New("cleanup failed")}
	runner := UpRunner{Validation: fakeValidator{report: initializationReport()}, Sandboxes: sandboxes}
	_, err := runner.Run(context.Background(), "/repo", UpOptions{}, Streams{})
	if !errors.Is(err, initializationErr) || !strings.Contains(err.Error(), "cleanup failed") {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestCancelledInitializationStillRollsBack(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	sandboxes := &fakeUpSandboxes{executeErr: context.Canceled}
	runner := UpRunner{Validation: fakeValidator{report: initializationReport()}, Sandboxes: sandboxes}
	_, err := runner.Run(ctx, "/repo", UpOptions{}, Streams{})
	if !errors.Is(err, context.Canceled) || sandboxes.removeContextErr != nil {
		t.Fatalf("Run() error = %v, cleanup context error = %v", err, sandboxes.removeContextErr)
	}
}

func TestRetryAfterInitializationFailureRunsCompleteSequence(t *testing.T) {
	sandboxes := &fakeUpSandboxes{executeErr: errors.New("partial setup")}
	runner := UpRunner{Validation: fakeValidator{report: initializationReport()}, Sandboxes: sandboxes}
	_, _ = runner.Run(context.Background(), "/repo", UpOptions{}, Streams{})
	sandboxes.calls = nil
	sandboxes.executeRequests = nil
	sandboxes.executeErr = nil
	if _, err := runner.Run(context.Background(), "/repo", UpOptions{}, Streams{}); err != nil {
		t.Fatalf("retry Run() error = %v", err)
	}
	want := []string{"lookup project", "create project", "allow project", "execute project", "execute project", "run project"}
	if !reflect.DeepEqual(sandboxes.calls, want) {
		t.Fatalf("retry calls = %#v, want %#v", sandboxes.calls, want)
	}
}

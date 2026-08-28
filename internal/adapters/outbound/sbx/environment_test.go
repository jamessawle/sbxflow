package sbx

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	sandboxport "github.com/jamessawle/sbxflow/internal/ports/sandbox"
)

func TestRenderEnvironmentWritesOrderedPrivateDocumentOutsideWorkspace(t *testing.T) {
	workspace := t.TempDir()
	temporaryRoot := t.TempDir()
	client := Client{TemporaryDirectory: temporaryRoot}
	environment := sandboxport.Environment{
		Name:      "project",
		Agent:     "codex",
		Workspace: workspace,
		Kits:      []string{"git+https://example.test/kits.git#ref=v1&dir=one", "/local kit"},
	}

	rendered, err := client.renderEnvironment(environment)
	if err != nil {
		t.Fatalf("renderEnvironment() error = %v", err)
	}
	t.Cleanup(func() { _ = rendered.cleanup() })

	data, err := os.ReadFile(rendered.path)
	if err != nil {
		t.Fatalf("read environment: %v", err)
	}
	want := "schemaVersion: \"1\"\nname: project\nagent: codex\nworkspace: " + workspace + "\nkits:\n- \"git+https://example.test/kits.git#ref=v1&dir=one\"\n- /local kit\n"
	if string(data) != want {
		t.Fatalf("environment document =\n%s\nwant:\n%s", data, want)
	}
	assertOwnerOnly(t, rendered.directory, 0o700)
	assertOwnerOnly(t, rendered.path, 0o600)
	inside, err := pathWithin(workspace, rendered.path)
	if err != nil || inside {
		t.Fatalf("environment path within workspace = %v, %v", inside, err)
	}
}

func TestRenderRemovalEnvironmentContainsOnlyIdentityAndInertInputs(t *testing.T) {
	rendered, err := (Client{TemporaryDirectory: t.TempDir()}).renderRemovalEnvironment("exact-name")
	if err != nil {
		t.Fatalf("renderRemovalEnvironment() error = %v", err)
	}
	t.Cleanup(func() { _ = rendered.cleanup() })

	data, err := os.ReadFile(rendered.path)
	if err != nil {
		t.Fatalf("read removal environment: %v", err)
	}
	wantPrefix := "schemaVersion: \"1\"\nname: exact-name\nagent: shell\nworkspace: "
	if !strings.HasPrefix(string(data), wantPrefix) || strings.Contains(string(data), "kits:") {
		t.Fatalf("removal environment =\n%s\nwant identity-only document", data)
	}
	workspace := strings.TrimSpace(strings.TrimPrefix(string(data), wantPrefix))
	if filepath.Dir(workspace) != rendered.directory || filepath.Base(workspace) != "workspace" {
		t.Fatalf("removal workspace = %q, want private sibling of document", workspace)
	}
	assertOwnerOnly(t, workspace, 0o700)
	inside, err := pathWithin(workspace, rendered.path)
	if err != nil || inside {
		t.Fatalf("removal document within inert workspace = %v, %v", inside, err)
	}
}

func TestRenderEnvironmentRejectsTemporaryDirectoryWithinWorkspaceAndCleansIt(t *testing.T) {
	workspace := t.TempDir()
	_, err := (Client{TemporaryDirectory: workspace}).renderEnvironment(sandboxport.Environment{
		Name: "project", Agent: "codex", Workspace: workspace,
	})
	if err == nil || !strings.Contains(err.Error(), "within mounted workspace") {
		t.Fatalf("renderEnvironment() error = %v", err)
	}
	entries, readErr := os.ReadDir(workspace)
	if readErr != nil || len(entries) != 0 {
		t.Fatalf("workspace entries after rejected render = %#v, %v", entries, readErr)
	}
}

func TestEnvironmentCommandsCleanTemporaryMaterialAfterEveryResult(t *testing.T) {
	for _, test := range []struct {
		name      string
		result    error
		cancelled bool
	}{
		{name: "success"},
		{name: "failure", result: errors.New("exit 9")},
		{name: "cancellation", cancelled: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			workspace := t.TempDir()
			interactive := &recordingInteractiveRunner{result: test.result}
			client := Client{Commands: fixedRunner{}, Interactive: interactive, TemporaryDirectory: t.TempDir()}
			ctx := context.Background()
			if test.cancelled {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
				interactive.useContextError = true
			}
			err := client.CreateSandbox(ctx, sandboxport.CreateRequest{Environment: sandboxport.Environment{
				Name: "project", Agent: "codex", Workspace: workspace,
			}}, sandboxport.Streams{})
			if test.cancelled {
				if !errors.Is(err, context.Canceled) {
					t.Fatalf("CreateSandbox() error = %v, want cancellation", err)
				}
			} else if !errors.Is(err, test.result) {
				t.Fatalf("CreateSandbox() error = %v, want %v", err, test.result)
			}
			if _, statErr := os.Stat(interactive.paths[0]); !errors.Is(statErr, os.ErrNotExist) {
				t.Fatalf("temporary environment still exists or stat failed: %v", statErr)
			}
		})
	}
}

func TestEnvironmentCommandsUseExactArgumentsStreamsAndScopedTrust(t *testing.T) {
	workspace := t.TempDir()
	interactive := &recordingInteractiveRunner{}
	client := Client{Commands: fixedRunner{}, Interactive: interactive, TemporaryDirectory: t.TempDir()}
	environment := sandboxport.Environment{
		Name: "project", Agent: "codex", Workspace: workspace, Kits: []string{"first", "second"},
		AllowedSources: []string{"docker.io/", "github.com/example/"}, AllowLocalKits: true,
	}
	stdin := strings.NewReader("input")
	var stdout, stderr bytes.Buffer
	streams := sandboxport.Streams{In: stdin, Out: &stdout, Err: &stderr}

	if err := client.CreateSandbox(context.Background(), sandboxport.CreateRequest{Environment: environment}, streams); err != nil {
		t.Fatal(err)
	}
	if err := client.RunSandbox(context.Background(), sandboxport.RunRequest{Environment: environment}, streams); err != nil {
		t.Fatal(err)
	}
	command := []string{"printf", "%s", "$HOME; literal"}
	if err := client.ExecuteCommand(context.Background(), sandboxport.CommandRequest{Environment: environment, Command: command}, streams); err != nil {
		t.Fatal(err)
	}
	if err := client.RemoveSandbox(context.Background(), sandboxport.RemoveRequest{Name: "project", Force: true, Streams: streams}); err != nil {
		t.Fatal(err)
	}

	wantArgs := [][]string{
		{"env", "create", "<file>"},
		{"env", "run", "<file>"},
		{"env", "exec", "<file>", "--", "printf", "%s", "$HOME; literal"},
		{"env", "rm", "--force", "<file>"},
	}
	gotArgs := make([][]string, len(interactive.invocations))
	for index, invocation := range interactive.invocations {
		gotArgs[index] = append([]string(nil), invocation.Args...)
		gotArgs[index][environmentPathIndex(invocation.Args)] = "<file>"
	}
	if !reflect.DeepEqual(gotArgs, wantArgs) {
		t.Fatalf("arguments = %#v, want %#v", gotArgs, wantArgs)
	}
	for index, invocation := range interactive.invocations[:3] {
		if invocation.Environment[allowedSourcesEnvironment] != `["docker.io/","github.com/example/"]` || invocation.Environment[allowLocalEnvironment] != "true" {
			t.Fatalf("invocation %d trust = %#v", index, invocation.Environment)
		}
	}
	if interactive.invocations[3].Environment != nil {
		t.Fatalf("removal trust = %#v, want none", interactive.invocations[3].Environment)
	}
	if interactive.invocations[0].Stdin != stdin || interactive.invocations[1].Stdin != stdin || interactive.invocations[2].Stdin != nil || interactive.invocations[3].Stdin != stdin {
		t.Fatalf("stdin attachments = %#v", interactive.invocations)
	}
	for index, invocation := range interactive.invocations {
		if invocation.Stdout != &stdout || invocation.Stderr != &stderr {
			t.Fatalf("invocation %d output streams = %#v", index, invocation)
		}
		if _, err := os.Stat(interactive.paths[index]); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("temporary path %d still exists or stat failed: %v", index, err)
		}
	}
	if !bytes.Contains(interactive.documents[0], []byte("kits:\n- first\n- second\n")) || !bytes.Equal(interactive.documents[0], interactive.documents[1]) || !bytes.Equal(interactive.documents[1], interactive.documents[2]) {
		t.Fatalf("full environment documents differ or lost kit order: %q", interactive.documents)
	}
	if bytes.Contains(interactive.documents[3], []byte("kits:")) || !bytes.Contains(interactive.documents[3], []byte("name: project\n")) {
		t.Fatalf("removal document is not identity-only: %s", interactive.documents[3])
	}
}

func assertOwnerOnly(t *testing.T, path string, want os.FileMode) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %q: %v", path, err)
	}
	if got := info.Mode().Perm(); got != want {
		t.Fatalf("mode for %q = %o, want %o", path, got, want)
	}
}

type fixedRunner struct{}

func (fixedRunner) LookPath(string) (string, error) { return "/bin/sbx", nil }
func (fixedRunner) Run(context.Context, string, ...string) Output {
	return Output{}
}

type recordingInteractiveRunner struct {
	result          error
	useContextError bool
	invocations     []InteractiveInvocation
	paths           []string
	documents       [][]byte
}

func (runner *recordingInteractiveRunner) Run(ctx context.Context, invocation InteractiveInvocation) error {
	path := invocation.Args[environmentPathIndex(invocation.Args)]
	document, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	runner.invocations = append(runner.invocations, invocation)
	runner.paths = append(runner.paths, path)
	runner.documents = append(runner.documents, document)
	if runner.useContextError {
		return ctx.Err()
	}
	return runner.result
}

func environmentPathIndex(args []string) int {
	if len(args) >= 4 && args[1] == "rm" && args[2] == "--force" {
		return 3
	}
	return 2
}

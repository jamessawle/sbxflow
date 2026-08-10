package localkit

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/jamessawle/sbxflow/internal/command"
	"github.com/jamessawle/sbxflow/internal/config"
)

func TestResolveLocalTargets(t *testing.T) {
	repository := t.TempDir()
	root := filepath.Join(repository, "kits")
	directory := filepath.Join(root, "directory")
	zip := filepath.Join(root, "packaged.zip")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(zip, []byte("fixture"), 0o600); err != nil {
		t.Fatal(err)
	}
	linked := linkedLocal(t, "./kits", "directory", "packaged.zip")

	targets, err := Resolve(filepath.Join(repository, config.Filename), linked)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	want := []string{directory, zip}
	for index, target := range targets {
		canonicalWant, err := filepath.EvalSymlinks(want[index])
		if err != nil {
			t.Fatal(err)
		}
		if target.Path != canonicalWant {
			t.Errorf("target %d path = %q, want %q", index, target.Path, canonicalWant)
		}
	}
}

func TestResolveRejectsUnsafeLocalReferences(t *testing.T) {
	repository := t.TempDir()
	root := filepath.Join(repository, "kits")
	outside := filepath.Join(repository, "outside")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(outside, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "escape")); err != nil {
		t.Fatal(err)
	}
	declaration := filepath.Join(repository, config.Filename)

	tests := map[string]struct {
		root string
		kit  string
		want string
	}{
		"URI root":           {root: "https://example.com/kits", kit: "one", want: "host filesystem path"},
		"unavailable root":   {root: "./missing", kit: "one", want: "unavailable"},
		"absolute kit":       {root: "./kits", kit: outside, want: "relative host filesystem path"},
		"URI kit":            {root: "./kits", kit: "oci://example/kit", want: "relative host filesystem path"},
		"traversal":          {root: "./kits", kit: "../outside", want: "escapes"},
		"symbolic link":      {root: "./kits", kit: "escape", want: "escapes"},
		"unavailable target": {root: "./kits", kit: "missing", want: "unavailable"},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := Resolve(declaration, linkedLocal(t, test.root, test.kit))
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Resolve() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestValidateUsesLocalProvenanceAndContinues(t *testing.T) {
	runner := &fakeRunner{
		path: "/fake/sbx",
		outputs: []command.Output{
			{Err: errors.New("exit status 1"), Stderr: []byte("bad kit"), ExitCode: 1},
			{Stdout: []byte("valid")},
		},
	}
	targets := []Target{
		{Index: 1, Source: "local", Kit: "kit.zip", Path: "/tmp/kits/kit.zip"},
		{Index: 3, Source: "local", Kit: "directory", Path: "/tmp/kits/directory"},
	}
	results := Validate(context.Background(), targets, runner)
	if len(results) != 2 || results[0].Valid || !results[1].Valid || results[0].Diagnostics != "bad kit" {
		t.Fatalf("Validate() = %#v", results)
	}
	wantCalls := [][]string{{"kit", "validate", targets[0].Path}, {"kit", "validate", targets[1].Path}}
	if !reflect.DeepEqual(runner.calls, wantCalls) {
		t.Fatalf("calls = %#v, want %#v", runner.calls, wantCalls)
	}
}

func TestValidateNoTargetsDoesNotLookUpSbx(t *testing.T) {
	runner := &fakeRunner{lookupErr: errors.New("must not be called")}
	if results := Validate(context.Background(), nil, runner); results != nil || runner.lookups != 0 {
		t.Fatalf("Validate() = %#v, lookups = %d", results, runner.lookups)
	}
}

func TestValidateReportsUnavailableAndTimeout(t *testing.T) {
	target := []Target{{Source: "local", Kit: "kit", Path: "/tmp/kit"}}
	unavailable := &fakeRunner{lookupErr: errors.New("not found")}
	if results := Validate(context.Background(), target, unavailable); len(results) != 1 || !strings.Contains(results[0].Err.Error(), "unavailable") {
		t.Fatalf("unavailable results = %#v", results)
	}
	timedOut := &fakeRunner{path: "/fake/sbx", outputs: []command.Output{{Err: context.DeadlineExceeded, ExitCode: -1}}}
	if results := Validate(context.Background(), target, timedOut); len(results) != 1 || !strings.Contains(results[0].Err.Error(), "timed out") {
		t.Fatalf("timeout results = %#v", results)
	}
}

func linkedLocal(t *testing.T, root string, kits ...string) config.LinkedConfiguration {
	t.Helper()
	configuration := config.Configuration{Version: 1, Sandbox: config.Sandbox{Name: "demo", Agent: "codex", Kits: config.Kits{
		Sources: map[string]config.Source{"local": {Type: config.SourceLocal, Root: root}},
	}}}
	for _, kit := range kits {
		configuration.Sandbox.Kits.Use = append(configuration.Sandbox.Kits.Use, config.Selection{Source: "local", Kit: kit})
	}
	linked, err := config.Link(configuration)
	if err != nil {
		t.Fatalf("Link() error = %v", err)
	}
	return linked
}

type fakeRunner struct {
	path      string
	lookupErr error
	outputs   []command.Output
	lookups   int
	calls     [][]string
}

func (r *fakeRunner) LookPath(string) (string, error) {
	r.lookups++
	return r.path, r.lookupErr
}

func (r *fakeRunner) Run(_ context.Context, _ string, args ...string) command.Output {
	r.calls = append(r.calls, append([]string(nil), args...))
	if len(r.outputs) == 0 {
		return command.Output{Err: errors.New("unexpected call"), ExitCode: -1}
	}
	output := r.outputs[0]
	r.outputs = r.outputs[1:]
	return output
}

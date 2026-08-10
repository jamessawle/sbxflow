package validation

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/jamessawle/sbxflow/internal/command"
)

func TestRunnerGatesPhasesAndReportsTrust(t *testing.T) {
	repository := t.TempDir()
	writeDeclaration(t, repository, `version: 1
sandbox:
  name: demo
  agent: codex
  kits:
    sources:
      remote:
        type: git
        repo: https://github.com/example/kits.git
        ref: v1
    use:
      - source: remote
        kit: tooling
`)
	commands := &recordingRunner{lookupErr: errors.New("must not look up sbx")}
	report := (Runner{Commands: commands}).Run(context.Background(), repository)
	if !report.Valid() {
		t.Fatalf("Run() errors = %v", report.Errors)
	}
	if commands.lookups != 0 || len(commands.calls) != 0 {
		t.Fatalf("remote validation invoked commands: lookups=%d calls=%#v", commands.lookups, commands.calls)
	}
	want := []string{"docker.io/", "github.com/example/kits"}
	if !reflect.DeepEqual(report.Linked.Trust.AllowedSources, want) {
		t.Fatalf("AllowedSources = %#v, want %#v", report.Linked.Trust.AllowedSources, want)
	}
}

func TestRunnerValidatesLocalKitsInOrderAndCollectsFailures(t *testing.T) {
	repository := t.TempDir()
	for _, name := range []string{"one", "two.zip"} {
		path := filepath.Join(repository, "kits", name)
		if filepath.Ext(name) == ".zip" {
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte("zip"), 0o600); err != nil {
				t.Fatal(err)
			}
		} else if err := os.MkdirAll(path, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	writeDeclaration(t, repository, `version: 1
sandbox:
  name: demo
  agent: codex
  kits:
    sources:
      local:
        type: local
        root: ./kits
    use:
      - source: local
        kit: one
      - source: local
        kit: two.zip
`)
	commands := &recordingRunner{path: "/fake/sbx", outputs: []command.Output{
		{Err: errors.New("invalid"), Stderr: []byte("bad one"), ExitCode: 1},
		{},
	}}
	report := (Runner{Commands: commands}).Run(context.Background(), repository)
	if report.Valid() || len(report.Errors) != 1 || len(report.LocalKits) != 2 || !report.LocalKits[1].Valid {
		t.Fatalf("Run() report = %#v", report)
	}
	if !report.Linked.Trust.AllowLocalKits || len(commands.calls) != 2 {
		t.Fatalf("trust/calls = %#v / %#v", report.Linked.Trust, commands.calls)
	}
}

func TestRunnerStopsBeforeSubprocessOnUnsafeInput(t *testing.T) {
	repository := t.TempDir()
	writeDeclaration(t, repository, `version: 1
sandbox:
  name: demo
  agent: codex
  kits:
    sources:
      local:
        type: local
        root: https://example.com/kits
    use:
      - source: local
        kit: one
`)
	commands := &recordingRunner{}
	report := (Runner{Commands: commands}).Run(context.Background(), repository)
	if report.Valid() || !strings.Contains(report.Errors[0].Error(), "host filesystem path") {
		t.Fatalf("Run() report = %#v", report)
	}
	if commands.lookups != 0 || len(commands.calls) != 0 {
		t.Fatalf("unsafe input invoked subprocess: %#v", commands)
	}
}

func writeDeclaration(t *testing.T, directory, contents string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(directory, "sbxflow.yaml"), []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
}

type recordingRunner struct {
	path      string
	lookupErr error
	outputs   []command.Output
	lookups   int
	calls     [][]string
}

func (r *recordingRunner) LookPath(string) (string, error) {
	r.lookups++
	return r.path, r.lookupErr
}

func (r *recordingRunner) Run(_ context.Context, _ string, args ...string) command.Output {
	r.calls = append(r.calls, append([]string(nil), args...))
	output := r.outputs[0]
	r.outputs = r.outputs[1:]
	return output
}

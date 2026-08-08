package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/jamessawle/sbxflow/internal/buildinfo"
)

func TestRootHelp(t *testing.T) {
	invocations := map[string][]string{
		"no arguments": nil,
		"help command": {"help"},
		"short flag":   {"-h"},
		"long flag":    {"--help"},
	}

	for name, args := range invocations {
		t.Run(name, func(t *testing.T) {
			stdout, stderr, err := execute(args)
			if err != nil {
				t.Fatalf("Execute() error = %v; stderr = %q", err, stderr)
			}
			if stderr != "" {
				t.Fatalf("stderr = %q, want empty", stderr)
			}

			for _, want := range []string{
				"Apply a repository's Docker Sandbox configuration and lifecycle",
				"Usage:",
				"sbxflow",
				"Available Commands:",
				"help ",
				"Flags:",
				"-h, --help",
				"-v, --version",
			} {
				if !strings.Contains(stdout, want) {
					t.Errorf("stdout does not contain %q:\n%s", want, stdout)
				}
			}

			for _, unavailable := range []string{"completion", "man", "up", "down", "destroy", "validate", "doctor"} {
				if strings.Contains(stdout, "\n  "+unavailable+" ") {
					t.Errorf("stdout advertises unavailable command %q:\n%s", unavailable, stdout)
				}
			}
		})
	}
}

func TestVersionFlags(t *testing.T) {
	shortOut, shortErr, err := execute([]string{"-v"})
	if err != nil {
		t.Fatalf("Execute(-v) error = %v; stderr = %q", err, shortErr)
	}
	longOut, longErr, err := execute([]string{"--version"})
	if err != nil {
		t.Fatalf("Execute(--version) error = %v; stderr = %q", err, longErr)
	}

	if shortErr != "" || longErr != "" {
		t.Fatalf("version stderr = %q, %q; want both empty", shortErr, longErr)
	}
	if shortOut == "" || shortOut != longOut {
		t.Fatalf("version output = %q, %q; want the same non-empty output", shortOut, longOut)
	}
	if !strings.Contains(shortOut, "development") {
		t.Fatalf("version output = %q, want development identity", shortOut)
	}
}

func TestVersionIncludesShortCommit(t *testing.T) {
	stdout, stderr, err := executeWithInfo(
		[]string{"--version"},
		buildinfo.Info{Version: "v1.2.3", Commit: "0123456789abcdef"},
	)
	if err != nil {
		t.Fatalf("Execute(--version) error = %v; stderr = %q", err, stderr)
	}
	if want := "sbxflow version v1.2.3 (0123456)\n"; stdout != want {
		t.Fatalf("stdout = %q, want %q", stdout, want)
	}
}

func TestUnavailableCommands(t *testing.T) {
	for _, name := range []string{"version", "completion", "man", "up", "down", "destroy", "validate", "doctor", "unknown"} {
		t.Run(name, func(t *testing.T) {
			stdout, stderr, err := execute([]string{name})
			if err == nil {
				t.Fatal("Execute() error = nil, want non-nil")
			}
			if stdout != "" {
				t.Fatalf("stdout = %q, want empty", stdout)
			}
			if !strings.Contains(strings.ToLower(stderr), "unknown command") || !strings.Contains(stderr, name) {
				t.Fatalf("stderr does not identify unknown command %q: %s", name, stderr)
			}
		})
	}
}

func execute(args []string) (string, string, error) {
	return executeWithInfo(args, buildinfo.Info{Version: "development"})
}

func executeWithInfo(args []string, info buildinfo.Info) (string, string, error) {
	var stdout, stderr bytes.Buffer
	err := Execute(
		context.Background(),
		args,
		Streams{In: strings.NewReader(""), Out: &stdout, Err: &stderr},
		info,
	)
	return stdout.String(), stderr.String(), err
}

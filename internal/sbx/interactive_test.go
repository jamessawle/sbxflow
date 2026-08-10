package sbx

import (
	"bytes"
	"context"
	"os"
	"runtime"
	"strings"
	"testing"
)

func TestInteractiveExecRunnerPreservesStreamsEnvironmentAndResult(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell fixture uses POSIX sh")
	}
	t.Setenv("SBXFLOW_INTERACTIVE_REPLACE", "old")
	t.Setenv("SBXFLOW_INTERACTIVE_KEEP", "kept")

	var stdout, stderr bytes.Buffer
	err := (InteractiveExecRunner{}).Run(context.Background(), InteractiveInvocation{
		Executable:  "sh",
		Args:        []string{"-c", `read value; printf 'in=%s replace=%s keep=%s' "$value" "$SBXFLOW_INTERACTIVE_REPLACE" "$SBXFLOW_INTERACTIVE_KEEP"; printf 'diagnostic' >&2`},
		Environment: map[string]string{"SBXFLOW_INTERACTIVE_REPLACE": "new"},
		Stdin:       strings.NewReader("hello\n"),
		Stdout:      &stdout,
		Stderr:      &stderr,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if got, want := stdout.String(), "in=hello replace=new keep=kept"; got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
	if stderr.String() != "diagnostic" {
		t.Fatalf("stderr = %q, want diagnostic", stderr.String())
	}

	err = (InteractiveExecRunner{}).Run(context.Background(), InteractiveInvocation{
		Executable: "sh",
		Args:       []string{"-c", "exit 7"},
		Stdin:      strings.NewReader(""),
		Stdout:     &stdout,
		Stderr:     &stderr,
	})
	if err == nil {
		t.Fatal("Run() error = nil, want unsuccessful process result")
	}
}

func TestMergeEnvironmentReplacesOnlyControlledValues(t *testing.T) {
	got := mergeEnvironment(
		[]string{"FIRST=one", "REPLACE=old", "LAST=three", "REPLACE=duplicate"},
		map[string]string{"REPLACE": "new", "ADDED": "value"},
	)
	want := []string{"FIRST=one", "LAST=three", "ADDED=value", "REPLACE=new"}
	if strings.Join(got, "\x00") != strings.Join(want, "\x00") {
		t.Fatalf("mergeEnvironment() = %#v, want %#v", got, want)
	}
}

func TestInteractiveRunnerDoesNotNeedShellForExecutable(t *testing.T) {
	var stdout bytes.Buffer
	err := (InteractiveExecRunner{}).Run(context.Background(), InteractiveInvocation{
		Executable: os.Args[0],
		Args:       []string{"-test.run=^$"},
		Stdout:     &stdout,
		Stderr:     &stdout,
	})
	if err != nil {
		t.Fatalf("Run(test executable) error = %v", err)
	}
}

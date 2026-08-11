package sbx

import (
	"context"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"
)

// InteractiveInvocation describes a subprocess attached directly to the CLI
// streams with process-local environment overrides.
type InteractiveInvocation struct {
	Executable  string
	Args        []string
	Environment map[string]string
	Stdin       io.Reader
	Stdout      io.Writer
	Stderr      io.Writer
}

// InteractiveRunner runs a subprocess without capturing its terminal streams.
type InteractiveRunner interface {
	Run(ctx context.Context, invocation InteractiveInvocation) error
}

// InteractiveExecRunner invokes attached subprocesses directly without a
// shell or lifecycle timeout.
type InteractiveExecRunner struct{}

// Run starts the invocation with inherited environment and terminal streams.
func (InteractiveExecRunner) Run(ctx context.Context, invocation InteractiveInvocation) error {
	process := exec.CommandContext(ctx, invocation.Executable, invocation.Args...) //nolint:gosec // executable and args come from the validated sandbox declaration; running them is this adapter's purpose.
	process.Env = mergeEnvironment(os.Environ(), invocation.Environment)
	process.Stdin = invocation.Stdin
	process.Stdout = invocation.Stdout
	process.Stderr = invocation.Stderr
	return process.Run()
}

func mergeEnvironment(base []string, overrides map[string]string) []string {
	keys := make([]string, 0, len(overrides))
	for key := range overrides {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	merged := make([]string, 0, len(base)+len(keys))
	for _, entry := range base {
		key, _, _ := strings.Cut(entry, "=")
		if _, replaced := overrides[key]; !replaced {
			merged = append(merged, entry)
		}
	}
	for _, key := range keys {
		merged = append(merged, key+"="+overrides[key])
	}
	return merged
}

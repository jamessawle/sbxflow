package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestExecutableStreamsAndExitStatuses(t *testing.T) {
	binaryName := "sbxflow"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	binary := filepath.Join(t.TempDir(), binaryName)

	build := exec.Command("go", "build", "-o", binary, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build executable: %v\n%s", err, output)
	}

	t.Run("help succeeds on stdout", func(t *testing.T) {
		stdout, stderr, exitCode := runBinary(t, binary)
		if exitCode != 0 {
			t.Fatalf("exit code = %d, want 0; stderr = %q", exitCode, stderr)
		}
		if !strings.Contains(stdout, "Usage:") {
			t.Fatalf("stdout does not contain help usage: %s", stdout)
		}
		if stderr != "" {
			t.Fatalf("stderr = %q, want empty", stderr)
		}
	})

	t.Run("version succeeds on stdout", func(t *testing.T) {
		stdout, stderr, exitCode := runBinary(t, binary, "--version")
		if exitCode != 0 {
			t.Fatalf("exit code = %d, want 0; stderr = %q", exitCode, stderr)
		}
		if !strings.Contains(stdout, "development") {
			t.Fatalf("stdout does not contain development build identity: %s", stdout)
		}
		if stderr != "" {
			t.Fatalf("stderr = %q, want empty", stderr)
		}
	})

	t.Run("errors fail on stderr", func(t *testing.T) {
		stdout, stderr, exitCode := runBinary(t, binary, "unknown")
		if exitCode == 0 {
			t.Fatal("exit code = 0, want non-zero")
		}
		if stdout != "" {
			t.Fatalf("stdout = %q, want empty", stdout)
		}
		if !strings.Contains(strings.ToLower(stderr), "unknown command") || !strings.Contains(stderr, "unknown") {
			t.Fatalf("stderr does not identify the unknown command: %s", stderr)
		}
	})
}

func runBinary(t *testing.T, binary string, args ...string) (string, string, int) {
	t.Helper()

	var stdout, stderr bytes.Buffer
	command := exec.Command(binary, args...)
	command.Stdout = &stdout
	command.Stderr = &stderr
	command.Env = append(os.Environ(), "TERM=dumb")

	err := command.Run()
	if err == nil {
		return stdout.String(), stderr.String(), 0
	}

	var exitError *exec.ExitError
	if !errors.As(err, &exitError) {
		t.Fatalf("run executable: %v", err)
	}
	return stdout.String(), stderr.String(), exitError.ExitCode()
}

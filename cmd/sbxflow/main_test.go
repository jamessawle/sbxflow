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

	t.Run("doctor reports checks and required failure", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("fake sbx fixture uses a POSIX shell script")
		}

		fakeDirectory := t.TempDir()
		fakeSbx := filepath.Join(fakeDirectory, "sbx")
		script := `#!/bin/sh
case "$1" in
  version)
    echo 'sbx version: v0.35.0 fake'
    ;;
  diagnose)
    echo '{"version":"1.0","summary":{"pass":2,"warn":0,"fail":1,"skip":1}}'
    exit 1
    ;;
  policy)
    echo '{"rules":[]}'
    ;;
  settings)
    echo '{"key":"kit.allowedSources","value":["*"],"type":"json","source":"env"}'
    ;;
  *)
    echo 'unexpected fake sbx invocation' >&2
    exit 2
    ;;
esac
`
		if err := os.WriteFile(fakeSbx, []byte(script), 0o700); err != nil {
			t.Fatalf("write fake sbx: %v", err)
		}

		stdout, stderr, exitCode := runBinaryWithEnv(
			t,
			binary,
			[]string{"doctor"},
			[]string{"PATH=" + fakeDirectory},
		)
		if exitCode == 0 {
			t.Fatalf("exit code = 0, want non-zero; stdout = %q", stdout)
		}
		for _, want := range []string{"[PASS] sbx-compatibility", "[FAIL] docker-diagnostics", "[WARN] network-policy", "[WARN] kit-sources"} {
			if !strings.Contains(stdout, want) {
				t.Errorf("stdout does not contain %q:\n%s", want, stdout)
			}
		}
		if !strings.Contains(stderr, "required doctor checks failed") {
			t.Fatalf("stderr does not report required failure: %q", stderr)
		}
	})

	t.Run("validate discovers configuration and preserves system state", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("fake sbx fixture uses a POSIX shell script")
		}

		repository := t.TempDir()
		localKit := filepath.Join(repository, "kits", "tooling")
		if err := os.MkdirAll(localKit, 0o755); err != nil {
			t.Fatalf("create local kit: %v", err)
		}
		configuration := `version: 1
sandbox:
  name: executable-test
  agent: codex
  kits:
    sources:
      remote:
        type: git
        repo: https://github.com/example/kits.git
        ref: v1
      local:
        type: local
        root: ./kits
    use:
      - source: remote
        kit: remote-tooling
      - source: local
        kit: tooling
`
		declaration := filepath.Join(repository, "sbxflow.yaml")
		if err := os.WriteFile(declaration, []byte(configuration), 0o600); err != nil {
			t.Fatalf("write declaration: %v", err)
		}
		nested := filepath.Join(repository, "nested", "work")
		if err := os.MkdirAll(nested, 0o755); err != nil {
			t.Fatalf("create nested work directory: %v", err)
		}

		fakeDirectory := t.TempDir()
		logPath := filepath.Join(fakeDirectory, "calls.log")
		fakeSbx := filepath.Join(fakeDirectory, "sbx")
		script := `#!/bin/sh
printf '%s\n' "$*" >> "$SBX_TEST_LOG"
if [ "$1 $2" != "kit validate" ]; then
  echo 'state-mutating or remote invocation attempted' >&2
  exit 9
fi
if [ "${SBX_TEST_REJECT:-}" = "1" ]; then
  echo 'fixture rejected local kit' >&2
  exit 1
fi
echo 'fixture accepted local kit'
`
		if err := os.WriteFile(fakeSbx, []byte(script), 0o700); err != nil {
			t.Fatalf("write fake sbx: %v", err)
		}

		stdout, stderr, exitCode := runBinaryInDirectory(
			t,
			binary,
			nested,
			[]string{"validate"},
			[]string{"PATH=" + fakeDirectory, "SBX_TEST_LOG=" + logPath},
		)
		if exitCode != 0 || stderr != "" {
			t.Fatalf("validate exit = %d, stderr = %q, stdout = %q", exitCode, stderr, stdout)
		}
		for _, want := range []string{declaration, "[VALID] local kit local/tooling", "github.com/example/kits", "kit.allowLocalKits: true"} {
			if !strings.Contains(stdout, want) {
				t.Errorf("stdout does not contain %q:\n%s", want, stdout)
			}
		}
		calls, err := os.ReadFile(logPath)
		if err != nil {
			t.Fatalf("read fake sbx calls: %v", err)
		}
		canonicalLocalKit, err := filepath.EvalSymlinks(localKit)
		if err != nil {
			t.Fatalf("canonicalize local kit: %v", err)
		}
		if got := strings.TrimSpace(string(calls)); got != "kit validate "+canonicalLocalKit {
			t.Fatalf("sbx calls = %q, want only local kit validation", got)
		}

		_, stderr, exitCode = runBinaryInDirectory(
			t,
			binary,
			nested,
			[]string{"validate"},
			[]string{"PATH=" + fakeDirectory, "SBX_TEST_LOG=" + logPath, "SBX_TEST_REJECT=1"},
		)
		if exitCode == 0 || !strings.Contains(stderr, "fixture rejected local kit") {
			t.Fatalf("rejected validate exit = %d, stderr = %q", exitCode, stderr)
		}
	})

	t.Run("validate reports missing configuration", func(t *testing.T) {
		directory := t.TempDir()
		stdout, stderr, exitCode := runBinaryInDirectory(t, binary, directory, []string{"validate"}, nil)
		if exitCode == 0 || stdout != "" || !strings.Contains(stderr, "no sbxflow.yaml found") {
			t.Fatalf("exit = %d, stdout = %q, stderr = %q", exitCode, stdout, stderr)
		}
	})
}

func runBinary(t *testing.T, binary string, args ...string) (string, string, int) {
	return runBinaryWithEnv(t, binary, args, nil)
}

func runBinaryWithEnv(t *testing.T, binary string, args, environment []string) (string, string, int) {
	return runBinaryInDirectory(t, binary, "", args, environment)
}

func runBinaryInDirectory(t *testing.T, binary, directory string, args, environment []string) (string, string, int) {
	t.Helper()

	var stdout, stderr bytes.Buffer
	command := exec.Command(binary, args...)
	command.Dir = directory
	command.Stdout = &stdout
	command.Stderr = &stderr
	command.Env = append(os.Environ(), "TERM=dumb")
	command.Env = append(command.Env, environment...)

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

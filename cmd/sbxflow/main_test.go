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
		for _, want := range []string{declaration, "Derived State:", "github.com/example/kits", "Local Kits Allowed: true", "State: pass", "Findings: []"} {
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

		stdout, stderr, exitCode = runBinaryInDirectory(
			t,
			binary,
			nested,
			[]string{"validate"},
			[]string{"PATH=" + fakeDirectory, "SBX_TEST_LOG=" + logPath, "SBX_TEST_REJECT=1"},
		)
		if exitCode == 0 || stdout != "" || !strings.Contains(stderr, "fixture rejected local kit") {
			t.Fatalf("rejected validate exit = %d, stderr = %q", exitCode, stderr)
		}
		for _, want := range []string{"Derived State:", "State: fail", "Findings:"} {
			if !strings.Contains(stderr, want) {
				t.Errorf("failure stderr does not contain %q:\n%s", want, stderr)
			}
		}
	})

	t.Run("validate reports missing configuration", func(t *testing.T) {
		directory := t.TempDir()
		stdout, stderr, exitCode := runBinaryInDirectory(t, binary, directory, []string{"validate"}, nil)
		if exitCode == 0 || stdout != "" || !strings.Contains(stderr, "no sbxflow.yaml found") {
			t.Fatalf("exit = %d, stdout = %q, stderr = %q", exitCode, stdout, stderr)
		}
	})

	t.Run("up creates or enters interactively with scoped trust", func(t *testing.T) {
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
  name: executable-up
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
		fakeSbx := filepath.Join(fakeDirectory, "sbx")
		script := `#!/bin/sh
printf '%s\n' "$*" >> "$SBX_TEST_LOG"
case "$1 $2" in
  "kit validate")
    exit 0
    ;;
  "ls --quiet")
    if [ -n "${SBX_TEST_EXISTING:-}" ]; then
      printf '%s\n' "$SBX_TEST_EXISTING"
    fi
    ;;
  *)
    if [ "$1" != "run" ]; then
      echo 'unexpected fake sbx invocation' >&2
      exit 8
    fi
    printf 'allowed=%s\nlocal=%s\n' "$DOCKER_SANDBOXES_KIT_ALLOWED_SOURCES" "$DOCKER_SANDBOXES_KIT_ALLOW_LOCAL" >> "$SBX_TEST_ENV_LOG"
    read input
    printf 'agent received: %s\n' "$input"
    echo 'docker run diagnostic' >&2
    exit "${SBX_TEST_RUN_EXIT:-0}"
    ;;
esac
`
		if err := os.WriteFile(fakeSbx, []byte(script), 0o700); err != nil {
			t.Fatalf("write fake sbx: %v", err)
		}
		canonicalLocalKit, err := filepath.EvalSymlinks(localKit)
		if err != nil {
			t.Fatalf("canonicalize local kit: %v", err)
		}
		canonicalRepository, err := filepath.EvalSymlinks(repository)
		if err != nil {
			t.Fatalf("canonicalize repository: %v", err)
		}
		validationStatus := "Configuration valid: " + filepath.Join(canonicalRepository, "sbxflow.yaml") + "\n"

		t.Run("missing sandbox", func(t *testing.T) {
			logPath := filepath.Join(fakeDirectory, "missing-calls.log")
			environmentPath := filepath.Join(fakeDirectory, "missing-env.log")
			stdout, stderr, exitCode := runBinaryInDirectoryWithInput(
				t,
				binary,
				nested,
				[]string{"up"},
				[]string{"PATH=" + fakeDirectory, "SBX_TEST_LOG=" + logPath, "SBX_TEST_ENV_LOG=" + environmentPath},
				"hello creation\n",
			)
			if exitCode != 0 || !strings.Contains(stdout, "agent received: hello creation") || stderr != validationStatus+"docker run diagnostic\n" {
				t.Fatalf("up exit = %d, stdout = %q, stderr = %q", exitCode, stdout, stderr)
			}
			calls, err := os.ReadFile(logPath)
			if err != nil {
				t.Fatalf("read calls: %v", err)
			}
			wantRun := "run --name executable-up --kit git+https://github.com/example/kits.git#ref=v1&dir=remote-tooling --kit " + canonicalLocalKit + " codex " + canonicalRepository
			for _, want := range []string{"kit validate " + canonicalLocalKit, "ls --quiet", wantRun} {
				if !strings.Contains(string(calls), want+"\n") {
					t.Errorf("calls do not contain %q:\n%s", want, calls)
				}
			}
			environment, err := os.ReadFile(environmentPath)
			if err != nil {
				t.Fatalf("read environment: %v", err)
			}
			wantEnvironment := "allowed=[\"docker.io/\",\"github.com/example/kits\"]\nlocal=true\n"
			if string(environment) != wantEnvironment {
				t.Fatalf("environment = %q, want %q", environment, wantEnvironment)
			}
		})

		t.Run("existing sandbox and failed session", func(t *testing.T) {
			logPath := filepath.Join(fakeDirectory, "existing-calls.log")
			environmentPath := filepath.Join(fakeDirectory, "existing-env.log")
			stdout, stderr, exitCode := runBinaryInDirectoryWithInput(
				t,
				binary,
				nested,
				[]string{"up"},
				[]string{
					"PATH=" + fakeDirectory,
					"SBX_TEST_LOG=" + logPath,
					"SBX_TEST_ENV_LOG=" + environmentPath,
					"SBX_TEST_EXISTING=executable-up",
					"SBX_TEST_RUN_EXIT=7",
				},
				"hello existing\n",
			)
			if exitCode != 7 || !strings.Contains(stdout, "agent received: hello existing") {
				t.Fatalf("up exit = %d, stdout = %q, stderr = %q", exitCode, stdout, stderr)
			}
			if stderr != validationStatus+"docker run diagnostic\n" {
				t.Fatalf("stderr = %q, want validation status followed by Docker diagnostic", stderr)
			}
			calls, err := os.ReadFile(logPath)
			if err != nil {
				t.Fatalf("read calls: %v", err)
			}
			if !strings.Contains(string(calls), "run codex --name executable-up\n") {
				t.Fatalf("calls do not contain existing invocation:\n%s", calls)
			}
			if strings.Contains(string(calls), "run codex --name executable-up --kit") || strings.Contains(string(calls), "run codex --name executable-up "+canonicalRepository) {
				t.Fatalf("existing invocation contains creation inputs:\n%s", calls)
			}
		})
	})

	t.Run("down resolves identity only and preserves Docker stop failure", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("fake sbx fixture uses a POSIX shell script")
		}

		repository := t.TempDir()
		configuration := `version: 1
sandbox:
  name: executable-down
  agent: unsupported
  kits:
    sources:
      local:
        type: local
        root: https://unavailable.example/kits
    use:
      - source: missing
        kit: ../unsafe
`
		if err := os.WriteFile(filepath.Join(repository, "sbxflow.yaml"), []byte(configuration), 0o600); err != nil {
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
case "$1 $2" in
  "ls --quiet")
    echo 'executable-down'
    ;;
  "stop executable-down")
    echo 'stopping executable-down'
    echo 'docker stop diagnostic' >&2
    exit 7
    ;;
  *)
    echo 'unexpected fake sbx invocation' >&2
    exit 8
    ;;
esac
`
		if err := os.WriteFile(fakeSbx, []byte(script), 0o700); err != nil {
			t.Fatalf("write fake sbx: %v", err)
		}

		stdout, stderr, exitCode := runBinaryInDirectory(
			t,
			binary,
			nested,
			[]string{"down"},
			[]string{"PATH=" + fakeDirectory, "SBX_TEST_LOG=" + logPath},
		)
		if exitCode != 7 || stdout != "stopping executable-down\n" || stderr != "docker stop diagnostic\n" {
			t.Fatalf("down exit = %d, stdout = %q, stderr = %q", exitCode, stdout, stderr)
		}
		calls, err := os.ReadFile(logPath)
		if err != nil {
			t.Fatalf("read fake sbx calls: %v", err)
		}
		if got, want := string(calls), "ls --quiet\nstop executable-down\n"; got != want {
			t.Fatalf("sbx calls = %q, want %q", got, want)
		}
	})

	t.Run("destroy resolves the exact identity and preserves removal behavior", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("fake sbx fixture uses a POSIX shell script")
		}

		repository := t.TempDir()
		configuration := `version: 1
sandbox:
  name: executable-destroy
  agent: unsupported
  kits:
    sources:
      local:
        type: local
        root: https://unavailable.example/kits
    use:
      - source: missing
        kit: ../unsafe
`
		if err := os.WriteFile(filepath.Join(repository, "sbxflow.yaml"), []byte(configuration), 0o600); err != nil {
			t.Fatalf("write declaration: %v", err)
		}
		nested := filepath.Join(repository, "nested", "work")
		if err := os.MkdirAll(nested, 0o755); err != nil {
			t.Fatalf("create nested work directory: %v", err)
		}

		fakeDirectory := t.TempDir()
		fakeSbx := filepath.Join(fakeDirectory, "sbx")
		script := `#!/bin/sh
printf '%s\n' "$*" >> "$SBX_TEST_LOG"
case "$1 $2" in
  "ls --quiet")
    if [ -n "${SBX_TEST_EXISTING:-}" ]; then
      printf '%s\n' "$SBX_TEST_EXISTING"
    fi
    ;;
  "rm executable-destroy")
    printf 'confirm removal: '
    IFS= read -r answer
    printf 'answer=%s\n' "$answer"
    echo 'docker remove diagnostic' >&2
    exit "${SBX_TEST_REMOVE_EXIT:-0}"
    ;;
  "rm --force")
    if [ "$3" != "executable-destroy" ]; then
      echo 'unexpected force target' >&2
      exit 8
    fi
    echo 'forced removal'
    ;;
  *)
    echo 'unexpected fake sbx invocation' >&2
    exit 8
    ;;
esac
`
		if err := os.WriteFile(fakeSbx, []byte(script), 0o700); err != nil {
			t.Fatalf("write fake sbx: %v", err)
		}

		t.Run("confirmed removal forwards input output and failure", func(t *testing.T) {
			logPath := filepath.Join(fakeDirectory, "confirmed-calls.log")
			stdout, stderr, exitCode := runBinaryInDirectoryWithInput(
				t,
				binary,
				nested,
				[]string{"destroy"},
				[]string{
					"PATH=" + fakeDirectory,
					"SBX_TEST_LOG=" + logPath,
					"SBX_TEST_EXISTING=other\nexecutable-destroy\nexecutable-destroy-extra",
					"SBX_TEST_REMOVE_EXIT=7",
				},
				"yes\n",
			)
			if exitCode != 7 || stdout != "confirm removal: answer=yes\n" || stderr != "docker remove diagnostic\n" {
				t.Fatalf("destroy exit = %d, stdout = %q, stderr = %q", exitCode, stdout, stderr)
			}
			calls, err := os.ReadFile(logPath)
			if err != nil {
				t.Fatalf("read fake sbx calls: %v", err)
			}
			if got, want := string(calls), "ls --quiet\nrm executable-destroy\n"; got != want {
				t.Fatalf("sbx calls = %q, want %q", got, want)
			}
		})

		t.Run("force forwards Docker force for only the declared target", func(t *testing.T) {
			logPath := filepath.Join(fakeDirectory, "forced-calls.log")
			stdout, stderr, exitCode := runBinaryInDirectory(
				t,
				binary,
				nested,
				[]string{"destroy", "-f"},
				[]string{"PATH=" + fakeDirectory, "SBX_TEST_LOG=" + logPath, "SBX_TEST_EXISTING=executable-destroy"},
			)
			if exitCode != 0 || stdout != "forced removal\n" || stderr != "" {
				t.Fatalf("destroy --force exit = %d, stdout = %q, stderr = %q", exitCode, stdout, stderr)
			}
			calls, err := os.ReadFile(logPath)
			if err != nil {
				t.Fatalf("read fake sbx calls: %v", err)
			}
			if got, want := string(calls), "ls --quiet\nrm --force executable-destroy\n"; got != want {
				t.Fatalf("sbx calls = %q, want %q", got, want)
			}
		})

		t.Run("absent exact target is idempotent", func(t *testing.T) {
			logPath := filepath.Join(fakeDirectory, "absent-calls.log")
			stdout, stderr, exitCode := runBinaryInDirectory(
				t,
				binary,
				nested,
				[]string{"destroy", "--force"},
				[]string{"PATH=" + fakeDirectory, "SBX_TEST_LOG=" + logPath, "SBX_TEST_EXISTING=executable-destroy-extra"},
			)
			if exitCode != 0 || stdout != "" || stderr != "" {
				t.Fatalf("absent destroy exit = %d, stdout = %q, stderr = %q", exitCode, stdout, stderr)
			}
			calls, err := os.ReadFile(logPath)
			if err != nil {
				t.Fatalf("read fake sbx calls: %v", err)
			}
			if got, want := string(calls), "ls --quiet\n"; got != want {
				t.Fatalf("sbx calls = %q, want %q", got, want)
			}
		})
	})
}

func runBinary(t *testing.T, binary string, args ...string) (string, string, int) {
	return runBinaryWithEnv(t, binary, args, nil)
}

func runBinaryWithEnv(t *testing.T, binary string, args, environment []string) (string, string, int) {
	return runBinaryInDirectory(t, binary, "", args, environment)
}

func runBinaryInDirectory(t *testing.T, binary, directory string, args, environment []string) (string, string, int) {
	return runBinaryInDirectoryWithInput(t, binary, directory, args, environment, "")
}

func runBinaryInDirectoryWithInput(t *testing.T, binary, directory string, args, environment []string, input string) (string, string, int) {
	t.Helper()

	var stdout, stderr bytes.Buffer
	command := exec.Command(binary, args...)
	command.Dir = directory
	command.Stdin = strings.NewReader(input)
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

package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
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

	runExecutableTest(t, "help succeeds on stdout", binary, executableHelpTest)
	runExecutableTest(t, "up help advertises recreation", binary, executableUpHelpTest)
	runExecutableTest(t, "version succeeds on stdout", binary, executableVersionTest)
	runExecutableTest(t, "errors fail on stderr", binary, executableUnknownCommandTest)
	runExecutableTest(t, "doctor reports checks and required failure", binary, executableDoctorTest)
	runExecutableTest(t, "validate discovers configuration and preserves system state", binary, executableValidateTest)
	runExecutableTest(t, "validate reports missing configuration", binary, executableMissingConfigurationTest)
	runExecutableTest(t, "up creates or enters interactively with scoped trust", binary, executableUpTest)
	runExecutableTest(t, "down resolves identity only and preserves Docker stop failure", binary, executableDownTest)
	runExecutableTest(t, "destroy resolves the exact identity and preserves removal behavior", binary, executableDestroyTest)
}

func runExecutableTest(t *testing.T, name, binary string, test func(*testing.T, string)) {
	t.Helper()
	t.Run(name, func(t *testing.T) { test(t, binary) })
}

func executableHelpTest(t *testing.T, binary string) {
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
}

func executableUpHelpTest(t *testing.T, binary string) {
	stdout, stderr, exitCode := runBinary(t, binary, "up", "--help")
	if exitCode != 0 || stderr != "" {
		t.Fatalf("exit code = %d, stderr = %q", exitCode, stderr)
	}
	for _, want := range []string{"--recreate", "--force", "bypasses", "permanent state loss", "attached sessions"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("stdout does not contain %q:\n%s", want, stdout)
		}
	}
}

func executableVersionTest(t *testing.T, binary string) {
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
}

func executableUnknownCommandTest(t *testing.T, binary string) {
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
}

func executableDoctorTest(t *testing.T, binary string) {
	if runtime.GOOS == "windows" {
		t.Skip("fake sbx fixture uses a POSIX shell script")
	}

	fakeDirectory := t.TempDir()
	fakeSbx := filepath.Join(fakeDirectory, "sbx")
	script := `#!/bin/sh
case "$1" in
  version)
    echo 'sbx version: v0.39.0 fake'
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
}

func executableValidateTest(t *testing.T, binary string) {
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
  hooks:
    initialize:
      - command: [must-not-run]
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
	assertValidateSuccess(t, stdout, stderr, exitCode, declaration, logPath, localKit)

	stdout, stderr, exitCode = runBinaryInDirectory(
		t,
		binary,
		nested,
		[]string{"validate"},
		[]string{"PATH=" + fakeDirectory, "SBX_TEST_LOG=" + logPath, "SBX_TEST_REJECT=1"},
	)
	assertValidateFailure(t, stdout, stderr, exitCode)
}

func assertValidateSuccess(t *testing.T, stdout, stderr string, exitCode int, declaration, logPath, localKit string) {
	t.Helper()
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
}

func assertValidateFailure(t *testing.T, stdout, stderr string, exitCode int) {
	t.Helper()
	if exitCode == 0 || stdout != "" || !strings.Contains(stderr, "fixture rejected local kit") {
		t.Fatalf("rejected validate exit = %d, stderr = %q", exitCode, stderr)
	}
	for _, want := range []string{"Derived State:", "State: fail", "Findings:"} {
		if !strings.Contains(stderr, want) {
			t.Errorf("failure stderr does not contain %q:\n%s", want, stderr)
		}
	}
}

func executableMissingConfigurationTest(t *testing.T, binary string) {
	directory := t.TempDir()
	stdout, stderr, exitCode := runBinaryInDirectory(t, binary, directory, []string{"validate"}, nil)
	if exitCode == 0 || stdout != "" || !strings.Contains(stderr, "no sbxflow.yaml found") {
		t.Fatalf("exit = %d, stdout = %q, stderr = %q", exitCode, stdout, stderr)
	}
}

func executableUpTest(t *testing.T, binary string) {
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
  network:
    allowedHosts:
      - api.example.com
  hooks:
    initialize:
      - command: [printf, hook-output]
      - command: [sh, -c, "printf shell-hook"]
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
record_environment() {
  operation=$1
  environment_file=$2
  if [ -n "${SBX_TEST_DOCUMENT_LOG:-}" ]; then
    printf '=== %s ===\n' "$operation" >> "$SBX_TEST_DOCUMENT_LOG"
    while IFS= read -r line; do printf '%s\n' "$line" >> "$SBX_TEST_DOCUMENT_LOG"; done < "$environment_file"
  fi
  if [ -n "${SBX_TEST_PATH_LOG:-}" ]; then
    printf '%s\n' "$environment_file" >> "$SBX_TEST_PATH_LOG"
  fi
}
case "$1" in
  env) ;;
  *) printf '%s\n' "$*" >> "$SBX_TEST_LOG" ;;
esac
case "$1 $2" in
  "kit validate")
    exit 0
    ;;
  "ls --json")
    if [ -n "${SBX_TEST_LS_FAIL:-}" ]; then
      echo 'state unavailable' >&2
      exit 6
    fi
    if [ -n "${SBX_TEST_EXISTING:-}" ]; then
      printf '{"sandboxes":[{"name":"%s","status":"%s"}]}\n' "$SBX_TEST_EXISTING" "${SBX_TEST_STATUS:-stopped}"
    else
      printf '{"sandboxes":[]}\n'
    fi
    ;;
	"policy allow")
		exit 0
		;;
	"env create")
		printf 'env create <env>\n' >> "$SBX_TEST_LOG"
		record_environment create "$3"
		printf 'allowed=%s\nlocal=%s\n' "$DOCKER_SANDBOXES_KIT_ALLOWED_SOURCES" "$DOCKER_SANDBOXES_KIT_ALLOW_LOCAL" >> "$SBX_TEST_ENV_LOG"
		echo 'docker create diagnostic' >&2
		exit "${SBX_TEST_CREATE_EXIT:-0}"
		;;
	"env run")
		printf 'env run <env>\n' >> "$SBX_TEST_LOG"
		record_environment run "$3"
		printf 'allowed=%s\nlocal=%s\n' "$DOCKER_SANDBOXES_KIT_ALLOWED_SOURCES" "$DOCKER_SANDBOXES_KIT_ALLOW_LOCAL" >> "$SBX_TEST_ENV_LOG"
		read input
		printf 'agent received: %s\n' "$input"
		echo 'docker run diagnostic' >&2
		exit "${SBX_TEST_RUN_EXIT:-0}"
		;;
	"env exec")
		environment_file=$3
		printf 'env exec <env> --' >> "$SBX_TEST_LOG"
		shift 4
		printf ' %s' "$@" >> "$SBX_TEST_LOG"
		printf '\n' >> "$SBX_TEST_LOG"
		record_environment exec "$environment_file"
		printf 'allowed=%s\nlocal=%s\n' "$DOCKER_SANDBOXES_KIT_ALLOWED_SOURCES" "$DOCKER_SANDBOXES_KIT_ALLOW_LOCAL" >> "$SBX_TEST_ENV_LOG"
		if [ -n "${SBX_TEST_EXEC_EXIT:-}" ]; then
			echo 'hook failed' >&2
			exit "$SBX_TEST_EXEC_EXIT"
		fi
		case "$1" in
			printf) printf '%s' "$2" ;;
			sh) printf shell-hook ;;
		esac
		;;
	"env rm")
		if [ "$3" = "--force" ]; then
			environment_file=$4
			printf 'env rm --force <env>\n' >> "$SBX_TEST_LOG"
		else
			environment_file=$3
			printf 'env rm <env>\n' >> "$SBX_TEST_LOG"
		fi
		record_environment rm "$environment_file"
		name=
		while IFS= read -r line; do
			case "$line" in name:\ *) name=${line#name: } ;; esac
		done < "$environment_file"
		if [ "$name" != "executable-up" ]; then
			echo 'unexpected sandbox removal target' >&2
			exit 8
		fi
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
	canonicalLocalKit, err := filepath.EvalSymlinks(localKit)
	if err != nil {
		t.Fatalf("canonicalize local kit: %v", err)
	}
	canonicalRepository, err := filepath.EvalSymlinks(repository)
	if err != nil {
		t.Fatalf("canonicalize repository: %v", err)
	}
	validationStatus := "Configuration valid: " + filepath.Join(canonicalRepository, "sbxflow.yaml") + "\n"

	fixture := executableUpFixture{binary, nested, fakeDirectory, canonicalLocalKit, canonicalRepository, validationStatus}
	t.Run("force without recreate stops before validation and Docker", fixture.forceWithoutRecreate)
	t.Run("missing sandbox", fixture.missingSandbox)
	t.Run("existing sandbox and failed session", fixture.existingSandbox)
	t.Run("existing sandbox is force-recreated before creation", fixture.recreateExisting)
	t.Run("running sandbox requires confirmation and preserves later input", fixture.confirmRunning)
	t.Run("forced running recreation skips confirmation", fixture.forceRunning)
	fixture.runEOFConfirmationCases(t)
	t.Run("running sandbox decline stops before mutation", fixture.declineRunning)
	t.Run("inspection failure stops before mutation", fixture.inspectionFailure)
	t.Run("malformed hooks stop before Docker access", fixture.malformedHooks)

}

type executableUpFixture struct {
	binary, nested, fakeDirectory                            string
	canonicalLocalKit, canonicalRepository, validationStatus string
}

func (fixture executableUpFixture) malformedHooks(t *testing.T) {
	declaration := filepath.Join(fixture.canonicalRepository, "sbxflow.yaml")
	original, err := os.ReadFile(declaration)
	if err != nil {
		t.Fatal(err)
	}
	malformed := strings.Replace(string(original), "command: [printf, hook-output]", "command: []", 1)
	if err := os.WriteFile(declaration, []byte(malformed), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.WriteFile(declaration, original, 0o600) })
	logPath := filepath.Join(fixture.fakeDirectory, "malformed-hooks-calls.log")
	_, stderr, exitCode := runBinaryInDirectory(t, fixture.binary, fixture.nested, []string{"up"}, []string{"PATH=" + fixture.fakeDirectory, "SBX_TEST_LOG=" + logPath})
	if exitCode == 0 || !strings.Contains(stderr, "command") {
		t.Fatalf("up exit = %d, stderr = %q", exitCode, stderr)
	}
	if _, err := os.Stat(logPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Docker call log exists or stat failed: %v", err)
	}
}

func (fixture executableUpFixture) forceWithoutRecreate(t *testing.T) {
	binary, nested, fakeDirectory := fixture.binary, fixture.nested, fixture.fakeDirectory
	canonicalLocalKit, canonicalRepository, validationStatus := fixture.canonicalLocalKit, fixture.canonicalRepository, fixture.validationStatus
	_, _, _, _, _, _ = binary, nested, fakeDirectory, canonicalLocalKit, canonicalRepository, validationStatus
	missing := t.TempDir()
	logPath := filepath.Join(fakeDirectory, "invalid-force-calls.log")
	_, stderr, exitCode := runBinaryInDirectoryWithInput(
		t, binary, missing, []string{"up", "--force"},
		[]string{"PATH=" + fakeDirectory, "SBX_TEST_LOG=" + logPath}, "",
	)
	if exitCode == 0 || !strings.Contains(stderr, "--force requires --recreate") {
		t.Fatalf("up --force exit = %d, stderr = %q", exitCode, stderr)
	}
	if _, err := os.Stat(logPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Docker call log exists or stat failed: %v", err)
	}
}

func (fixture executableUpFixture) missingSandbox(t *testing.T) {
	binary, nested, fakeDirectory := fixture.binary, fixture.nested, fixture.fakeDirectory
	canonicalLocalKit, canonicalRepository, validationStatus := fixture.canonicalLocalKit, fixture.canonicalRepository, fixture.validationStatus
	_, _, _, _, _, _ = binary, nested, fakeDirectory, canonicalLocalKit, canonicalRepository, validationStatus
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
	if exitCode != 0 || !strings.Contains(stdout, "hook-outputshell-hookagent received: hello creation") || stderr != validationStatus+"docker create diagnostic\ndocker run diagnostic\n" {
		t.Fatalf("up exit = %d, stdout = %q, stderr = %q", exitCode, stdout, stderr)
	}
	calls, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read calls: %v", err)
	}
	wantCalls := []string{
		"kit validate " + canonicalLocalKit,
		"ls --json",
		"env create <env>",
		"policy allow network --sandbox executable-up api.example.com",
		"env exec <env> -- printf hook-output",
		"env exec <env> -- sh -c printf shell-hook",
		"env run <env>",
	}
	if got := strings.Split(strings.TrimSpace(string(calls)), "\n"); !reflect.DeepEqual(got, wantCalls) {
		t.Fatalf("calls = %#v, want %#v", got, wantCalls)
	}
	environment, err := os.ReadFile(environmentPath)
	if err != nil {
		t.Fatalf("read environment: %v", err)
	}
	trust := "allowed=[\"docker.io/\",\"github.com/example/kits\"]\nlocal=true\n"
	if string(environment) != trust+trust+trust+trust {
		t.Fatalf("environment = %q, want scoped trust on creation, both hooks, and attachment", environment)
	}
}

func (fixture executableUpFixture) existingSandbox(t *testing.T) {
	binary, nested, fakeDirectory := fixture.binary, fixture.nested, fixture.fakeDirectory
	canonicalLocalKit, canonicalRepository, validationStatus := fixture.canonicalLocalKit, fixture.canonicalRepository, fixture.validationStatus
	_, _, _, _, _, _ = binary, nested, fakeDirectory, canonicalLocalKit, canonicalRepository, validationStatus
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
	if !strings.Contains(string(calls), "env run <env>\n") {
		t.Fatalf("calls do not contain existing invocation:\n%s", calls)
	}
	if strings.Contains(string(calls), "env create") || strings.Contains(string(calls), "env exec") {
		t.Fatalf("existing invocation contains creation inputs:\n%s", calls)
	}
}

func (fixture executableUpFixture) recreateExisting(t *testing.T) {
	binary, nested, fakeDirectory := fixture.binary, fixture.nested, fixture.fakeDirectory
	canonicalLocalKit, canonicalRepository, validationStatus := fixture.canonicalLocalKit, fixture.canonicalRepository, fixture.validationStatus
	_, _, _, _, _, _ = binary, nested, fakeDirectory, canonicalLocalKit, canonicalRepository, validationStatus
	logPath := filepath.Join(fakeDirectory, "recreate-calls.log")
	environmentPath := filepath.Join(fakeDirectory, "recreate-env.log")
	stdout, stderr, exitCode := runBinaryInDirectoryWithInput(
		t,
		binary,
		nested,
		[]string{"up", "--recreate"},
		[]string{
			"PATH=" + fakeDirectory,
			"SBX_TEST_LOG=" + logPath,
			"SBX_TEST_ENV_LOG=" + environmentPath,
			"SBX_TEST_EXISTING=executable-up",
		},
		"hello replacement\n",
	)
	if exitCode != 0 || !strings.Contains(stdout, "agent received: hello replacement") || stderr != validationStatus+"docker create diagnostic\ndocker run diagnostic\n" {
		t.Fatalf("up --recreate exit = %d, stdout = %q, stderr = %q", exitCode, stdout, stderr)
	}
	calls, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read calls: %v", err)
	}
	wantCalls := []string{
		"kit validate " + canonicalLocalKit,
		"ls --json",
		"env rm --force <env>",
		"env create <env>",
		"policy allow network --sandbox executable-up api.example.com",
		"env exec <env> -- printf hook-output",
		"env exec <env> -- sh -c printf shell-hook",
		"env run <env>",
	}
	if got := strings.Split(strings.TrimSpace(string(calls)), "\n"); !reflect.DeepEqual(got, wantCalls) {
		t.Fatalf("calls = %#v, want %#v", got, wantCalls)
	}
}

func (fixture executableUpFixture) confirmRunning(t *testing.T) {
	binary, nested, fakeDirectory := fixture.binary, fixture.nested, fixture.fakeDirectory
	canonicalLocalKit, canonicalRepository, validationStatus := fixture.canonicalLocalKit, fixture.canonicalRepository, fixture.validationStatus
	_, _, _, _, _, _ = binary, nested, fakeDirectory, canonicalLocalKit, canonicalRepository, validationStatus
	logPath := filepath.Join(fakeDirectory, "running-approved-calls.log")
	environmentPath := filepath.Join(fakeDirectory, "running-approved-env.log")
	stdout, stderr, exitCode := runBinaryInDirectoryWithInput(
		t, binary, nested, []string{"up", "--recreate"},
		[]string{"PATH=" + fakeDirectory, "SBX_TEST_LOG=" + logPath, "SBX_TEST_ENV_LOG=" + environmentPath, "SBX_TEST_EXISTING=executable-up", "SBX_TEST_STATUS=running"},
		"yes\nhello approved replacement\n",
	)
	if exitCode != 0 || !strings.Contains(stdout, "agent received: hello approved replacement") {
		t.Fatalf("up --recreate exit = %d, stdout = %q, stderr = %q", exitCode, stdout, stderr)
	}
	for _, want := range []string{validationStatus, "running sandbox \"executable-up\"", "other attached terminal sessions", "[y/N]", "docker run diagnostic"} {
		if !strings.Contains(stderr, want) {
			t.Errorf("stderr does not contain %q: %q", want, stderr)
		}
	}
}

func (fixture executableUpFixture) forceRunning(t *testing.T) {
	binary, nested, fakeDirectory := fixture.binary, fixture.nested, fixture.fakeDirectory
	canonicalLocalKit, canonicalRepository, validationStatus := fixture.canonicalLocalKit, fixture.canonicalRepository, fixture.validationStatus
	_, _, _, _, _, _ = binary, nested, fakeDirectory, canonicalLocalKit, canonicalRepository, validationStatus
	logPath := filepath.Join(fakeDirectory, "running-forced-calls.log")
	environmentPath := filepath.Join(fakeDirectory, "running-forced-env.log")
	stdout, stderr, exitCode := runBinaryInDirectoryWithInput(
		t, binary, nested, []string{"up", "--recreate", "--force"},
		[]string{"PATH=" + fakeDirectory, "SBX_TEST_LOG=" + logPath, "SBX_TEST_ENV_LOG=" + environmentPath, "SBX_TEST_EXISTING=executable-up", "SBX_TEST_STATUS=running"},
		"hello forced replacement\n",
	)
	if exitCode != 0 || !strings.Contains(stdout, "agent received: hello forced replacement") {
		t.Fatalf("forced up exit = %d, stdout = %q, stderr = %q", exitCode, stdout, stderr)
	}
	for _, unwanted := range []string{"running sandbox", "other attached terminal sessions", "[y/N]"} {
		if strings.Contains(stderr, unwanted) {
			t.Errorf("stderr contains confirmation marker %q: %q", unwanted, stderr)
		}
	}
	calls, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read calls: %v", err)
	}
	wantCalls := []string{
		"kit validate " + canonicalLocalKit,
		"ls --json",
		"env rm --force <env>",
		"env create <env>",
		"policy allow network --sandbox executable-up api.example.com",
		"env exec <env> -- printf hook-output",
		"env exec <env> -- sh -c printf shell-hook",
		"env run <env>",
	}
	if got := strings.Split(strings.TrimSpace(string(calls)), "\n"); !reflect.DeepEqual(got, wantCalls) {
		t.Fatalf("calls = %#v, want %#v", got, wantCalls)
	}
}

func (fixture executableUpFixture) runEOFConfirmationCases(t *testing.T) {
	binary, nested, fakeDirectory := fixture.binary, fixture.nested, fixture.fakeDirectory
	_, _, _ = binary, nested, fakeDirectory
	for _, test := range []struct {
		name       string
		input      string
		wantError  string
		wantMutate bool
	}{
		{name: "affirmative", input: "yes", wantMutate: true},
		{name: "negative", input: "no", wantError: "recreation cancelled"},
		{name: "malformed", input: "approve", wantError: "recreation cancelled"},
		{name: "immediate EOF", wantError: "read confirmation response"},
	} {
		t.Run("running sandbox with "+test.name+" EOF-terminated confirmation", func(t *testing.T) {
			logPath := filepath.Join(fakeDirectory, "running-eof-"+test.name+"-calls.log")
			environmentPath := filepath.Join(fakeDirectory, "running-eof-"+test.name+"-env.log")
			_, stderr, exitCode := runBinaryInDirectoryWithInput(
				t, binary, nested, []string{"up", "--recreate"},
				[]string{"PATH=" + fakeDirectory, "SBX_TEST_LOG=" + logPath, "SBX_TEST_ENV_LOG=" + environmentPath, "SBX_TEST_EXISTING=executable-up", "SBX_TEST_STATUS=running"},
				test.input,
			)
			if (exitCode != 0) != (test.wantError != "") || !strings.Contains(stderr, test.wantError) {
				t.Fatalf("up --recreate exit = %d, stderr = %q", exitCode, stderr)
			}
			calls, err := os.ReadFile(logPath)
			if err != nil {
				t.Fatalf("read calls: %v", err)
			}
			callLog := string(calls)
			mutations := []string{"env rm --force <env>\n", "env create <env>\n", "env run <env>\n"}
			for _, mutation := range mutations {
				if strings.Contains(callLog, mutation) != test.wantMutate {
					t.Fatalf("call %q present = %v, want mutation %v; calls:\n%s", mutation, strings.Contains(callLog, mutation), test.wantMutate, calls)
				}
			}
		})
	}
}

func (fixture executableUpFixture) declineRunning(t *testing.T) {
	binary, nested, fakeDirectory := fixture.binary, fixture.nested, fixture.fakeDirectory
	canonicalLocalKit, canonicalRepository, validationStatus := fixture.canonicalLocalKit, fixture.canonicalRepository, fixture.validationStatus
	_, _, _, _, _, _ = binary, nested, fakeDirectory, canonicalLocalKit, canonicalRepository, validationStatus
	logPath := filepath.Join(fakeDirectory, "running-declined-calls.log")
	environmentPath := filepath.Join(fakeDirectory, "running-declined-env.log")
	_, stderr, exitCode := runBinaryInDirectoryWithInput(
		t, binary, nested, []string{"up", "--recreate"},
		[]string{"PATH=" + fakeDirectory, "SBX_TEST_LOG=" + logPath, "SBX_TEST_ENV_LOG=" + environmentPath, "SBX_TEST_EXISTING=executable-up", "SBX_TEST_STATUS=running"},
		"no\n",
	)
	if exitCode == 0 || !strings.Contains(stderr, "recreation cancelled") {
		t.Fatalf("up --recreate exit = %d, stderr = %q", exitCode, stderr)
	}
	calls, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read calls: %v", err)
	}
	if strings.Contains(string(calls), "env rm") || strings.Contains(string(calls), "env create") || strings.Contains(string(calls), "env run") {
		t.Fatalf("declined calls contain mutation: %s", calls)
	}
}

func (fixture executableUpFixture) inspectionFailure(t *testing.T) {
	binary, nested, fakeDirectory := fixture.binary, fixture.nested, fixture.fakeDirectory
	canonicalLocalKit, canonicalRepository, validationStatus := fixture.canonicalLocalKit, fixture.canonicalRepository, fixture.validationStatus
	_, _, _, _, _, _ = binary, nested, fakeDirectory, canonicalLocalKit, canonicalRepository, validationStatus
	logPath := filepath.Join(fakeDirectory, "inspection-failed-calls.log")
	environmentPath := filepath.Join(fakeDirectory, "inspection-failed-env.log")
	_, stderr, exitCode := runBinaryInDirectoryWithInput(
		t, binary, nested, []string{"up", "--recreate"},
		[]string{"PATH=" + fakeDirectory, "SBX_TEST_LOG=" + logPath, "SBX_TEST_ENV_LOG=" + environmentPath, "SBX_TEST_LS_FAIL=1"}, "",
	)
	if exitCode == 0 || !strings.Contains(stderr, "state unavailable") {
		t.Fatalf("up --recreate exit = %d, stderr = %q", exitCode, stderr)
	}
	calls, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read calls: %v", err)
	}
	if strings.Contains(string(calls), "env rm") || strings.Contains(string(calls), "env create") || strings.Contains(string(calls), "env run") {
		t.Fatalf("failed inspection calls contain mutation: %s", calls)
	}
}

func executableDownTest(t *testing.T, binary string) {
	if runtime.GOOS == "windows" {
		t.Skip("fake sbx fixture uses a POSIX shell script")
	}

	repository := t.TempDir()
	configuration := `version: 1
sandbox:
  name: executable-down
  agent: unsupported
  hooks:
    initialize:
      - command: []
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
}

func executableDestroyTest(t *testing.T, binary string) {
	if runtime.GOOS == "windows" {
		t.Skip("fake sbx fixture uses a POSIX shell script")
	}

	repository := t.TempDir()
	configuration := `version: 1
sandbox:
  name: executable-destroy
  agent: unsupported
  hooks:
    initialize:
      - command: []
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
case "$1" in
  env) ;;
  *) printf '%s\n' "$*" >> "$SBX_TEST_LOG" ;;
esac
case "$1 $2" in
  "ls --quiet")
    if [ -n "${SBX_TEST_EXISTING:-}" ]; then
      printf '%s\n' "$SBX_TEST_EXISTING"
    fi
    ;;
  "env rm")
    force=false
    environment_file=$3
    if [ "$3" = "--force" ]; then
      force=true
      environment_file=$4
      printf 'env rm --force <env>\n' >> "$SBX_TEST_LOG"
    else
      printf 'env rm <env>\n' >> "$SBX_TEST_LOG"
    fi
    name=
    has_kits=false
    while IFS= read -r line; do
      case "$line" in
        name:\ *) name=${line#name: } ;;
        kits:*) has_kits=true ;;
      esac
    done < "$environment_file"
    if [ "$name" != "executable-destroy" ] || [ "$has_kits" = true ]; then
      echo 'unexpected removal environment' >&2
      exit 8
    fi
    if [ "$force" = true ]; then
      echo 'forced removal'
    else
      printf 'confirm removal: '
      IFS= read -r answer
      printf 'answer=%s\n' "$answer"
      echo 'docker remove diagnostic' >&2
      exit "${SBX_TEST_REMOVE_EXIT:-0}"
    fi
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

	destroyFixture := executableDestroyFixture{binary, nested, fakeDirectory}
	t.Run("confirmed removal forwards input output and failure", destroyFixture.confirmedRemoval)
	t.Run("force forwards Docker force for only the declared target", destroyFixture.forcedRemoval)
	t.Run("absent exact target is idempotent", destroyFixture.absentRemoval)
}

type executableDestroyFixture struct{ binary, nested, fakeDirectory string }

func (fixture executableDestroyFixture) confirmedRemoval(t *testing.T) {
	binary, nested, fakeDirectory := fixture.binary, fixture.nested, fixture.fakeDirectory
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
	if got, want := string(calls), "ls --quiet\nenv rm <env>\n"; got != want {
		t.Fatalf("sbx calls = %q, want %q", got, want)
	}
}

func (fixture executableDestroyFixture) forcedRemoval(t *testing.T) {
	binary, nested, fakeDirectory := fixture.binary, fixture.nested, fixture.fakeDirectory
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
	if got, want := string(calls), "ls --quiet\nenv rm --force <env>\n"; got != want {
		t.Fatalf("sbx calls = %q, want %q", got, want)
	}
}

func (fixture executableDestroyFixture) absentRemoval(t *testing.T) {
	binary, nested, fakeDirectory := fixture.binary, fixture.nested, fixture.fakeDirectory
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

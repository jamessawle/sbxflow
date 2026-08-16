package lifecycle

import (
	"bytes"
	"context"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/jamessawle/sbxflow/internal/adapters/outbound/sbx"
	"github.com/jamessawle/sbxflow/internal/domain/configuration"
	sandboxport "github.com/jamessawle/sbxflow/internal/ports/sandbox"
)

func TestInspectUsesExactMachineReadableState(t *testing.T) {
	for _, test := range []struct {
		name   string
		output string
		want   sandboxport.State
	}{
		{name: "running", output: `{"sandboxes":[{"name":"other","status":"stopped"},{"name":"project","status":"running"}]}`, want: sandboxport.StateRunning},
		{name: "stopped", output: `{"sandboxes":[{"name":"project","status":"stopped"}]}`, want: sandboxport.StateStopped},
		{name: "exact name", output: `{"sandboxes":[{"name":"project-extra","status":"running"}]}`, want: sandboxport.StateAbsent},
		{name: "absent", output: `{"sandboxes":[]}`, want: sandboxport.StateAbsent},
		{
			name:   "running before box-drawn update notice",
			output: `{"sandboxes":[{"name":"project","status":"running"}]}` + dockerUpdateNotice,
			want:   sandboxport.StateRunning,
		},
		{
			name:   "stopped before styled update notice",
			output: `{"sandboxes":[{"name":"project","status":"stopped"}]}` + "\n\n\x1b[1;36mDocker Sandboxes Update Available\x1b[0m\nRun your preferred upgrade command.\n",
			want:   sandboxport.StateStopped,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			commands := &fakeCommandRunner{output: sbx.Output{Stdout: []byte(test.output)}}
			got, err := (sbx.Client{Commands: commands}).Inspect(context.Background(), "project")
			if err != nil || got != test.want {
				t.Fatalf("Inspect() = %v, %v, want %v, nil", got, err, test.want)
			}
			if !reflect.DeepEqual(commands.args, []string{"ls", "--json"}) {
				t.Fatalf("args = %#v", commands.args)
			}
		})
	}
}

func TestInspectRejectsInvalidResults(t *testing.T) {
	for _, test := range []struct {
		name   string
		output []byte
		want   string
	}{
		{name: "malformed", output: []byte(`{`), want: "decode"},
		{name: "malformed before notice", output: []byte(`{"sandboxes":[}` + dockerUpdateNotice), want: "decode"},
		{name: "leading contamination", output: []byte("checking for updates\n" + `{"sandboxes":[]}`), want: "decode"},
		{name: "multiple documents", output: []byte(`{"sandboxes":[]}` + "\n" + `{"sandboxes":[]}`), want: "multiple JSON documents"},
		{name: "multiple documents before notice", output: []byte(`{"sandboxes":[]}` + "\n{}" + dockerUpdateNotice), want: "multiple JSON documents"},
		{name: "invalid UTF-8", output: append([]byte(`{"sandboxes":[]}`), 0xff), want: "unexpected trailing output"},
		{name: "invalid UTF-8 in name", output: []byte("{\"sandboxes\":[{\"name\":\"proj\xffct\",\"status\":\"running\"}]}"), want: "invalid UTF-8"},
		{name: "invalid UTF-8 in status", output: []byte("{\"sandboxes\":[{\"name\":\"project\",\"status\":\"run\xffning\"}]}"), want: "invalid UTF-8"},
		{name: "invalid UTF-8 in name before notice", output: []byte("{\"sandboxes\":[{\"name\":\"proj\xffct\",\"status\":\"running\"}]}" + dockerUpdateNotice), want: "invalid UTF-8"},
		{name: "near-match notice title", output: []byte(`{"sandboxes":[]}` + "\nDocker Sandbox Update Available\n"), want: "unexpected trailing output"},
		{name: "unknown trailing output", output: []byte(`{"sandboxes":[]}` + "\nunrelated warning\n"), want: "unexpected trailing output"},
		{name: "missing collection", output: []byte(`{}`), want: "missing sandboxes collection"},
		{name: "duplicate", output: []byte(`{"sandboxes":[{"name":"project","status":"running"},{"name":"project","status":"running"}]}`), want: "2 exact-name matches"},
		{name: "unknown", output: []byte(`{"sandboxes":[{"name":"project","status":"starting"}]}`), want: "unrecognized lifecycle state"},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := (sbx.Client{Commands: &fakeCommandRunner{output: sbx.Output{Stdout: test.output}}}).Inspect(context.Background(), "project")
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Inspect() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestInspectBoundsUnexpectedTrailingOutputError(t *testing.T) {
	const privateListing = `{"sandboxes":[{"name":"private-project-at-/Users/example/secret","status":"running"}]}`
	trailing := strings.Repeat("unrecognized output ", 100)
	_, err := (sbx.Client{Commands: &fakeCommandRunner{output: sbx.Output{Stdout: []byte(privateListing + trailing)}}}).Inspect(context.Background(), "project")
	if err == nil || !strings.Contains(err.Error(), "unexpected trailing output") || !strings.Contains(err.Error(), "1999 bytes") {
		t.Fatalf("Inspect() error = %v, want bounded trailing-output classification", err)
	}
	if strings.Contains(err.Error(), privateListing) || strings.Contains(err.Error(), "private-project") || len(err.Error()) > 200 {
		t.Fatalf("Inspect() error exposes listing contents or is unbounded: %q", err)
	}
}

const dockerUpdateNotice = `

╭──────────────────────────────────────────────────────────────╮
│ Docker Sandboxes Update Available                            │
│                                                              │
│ Current version: 0.35.0                                      │
│ Available version: 0.38.0                                    │
│ Release notes: https://docs.docker.com/ai/sandboxes/          │
│ Upgrade with: docker desktop update                          │
╰──────────────────────────────────────────────────────────────╯`

func TestInspectReportsDockerDiagnostics(t *testing.T) {
	commands := &fakeCommandRunner{output: sbx.Output{Stderr: []byte("daemon unavailable\n"), ExitCode: 7, Err: errors.New("exit 7")}}
	_, err := (sbx.Client{Commands: commands}).Inspect(context.Background(), "project")
	if err == nil || !strings.Contains(err.Error(), "sbx ls --json") || !strings.Contains(err.Error(), "daemon unavailable") {
		t.Fatalf("Inspect() error = %v", err)
	}
}

func TestRunnerValidationGatesLifecycleLookup(t *testing.T) {
	commands := &fakeCommandRunner{}
	var stderr bytes.Buffer
	runner := UpRunner{
		Validation: fakeValidator{report: configuration.Validation{Errors: []error{errors.New("invalid")}}},
		Sandboxes:  sbx.Client{Commands: commands, Interactive: &fakeInteractiveRunner{}},
	}
	_, err := runner.Run(context.Background(), "/repo", UpOptions{}, Streams{Err: &stderr})
	if !errors.Is(err, ErrValidationFailed) || commands.lookups != 0 || commands.runs != 0 {
		t.Fatalf("Run() error = %v, lookups = %d, runs = %d", err, commands.lookups, commands.runs)
	}
	if stderr.String() != "" {
		t.Fatalf("stderr = %q, want no validation success status", stderr.String())
	}
}

func TestRunnerReportsValidationSuccessBeforeLifecycleLookup(t *testing.T) {
	var stderr bytes.Buffer
	commands := &fakeCommandRunner{
		path:   "/bin/sbx",
		output: sbx.Output{Stdout: []byte(`{"sandboxes":[]}`)},
		onLookup: func() {
			if got, want := stderr.String(), "Configuration valid: /repo/sbxflow.yaml\n"; got != want {
				t.Fatalf("stderr before lookup = %q, want %q", got, want)
			}
		},
	}
	runner := UpRunner{
		Validation: fakeValidator{report: validReport()},
		Sandboxes:  sbx.Client{Commands: commands, Interactive: &fakeInteractiveRunner{}},
	}
	_, err := runner.Run(context.Background(), "/repo", UpOptions{}, Streams{Err: &stderr})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if got, want := stderr.String(), "Configuration valid: /repo/sbxflow.yaml\n"; got != want {
		t.Fatalf("stderr = %q, want %q", got, want)
	}
}

func TestRunnerLookupFailureDoesNotRunAgent(t *testing.T) {
	interactive := &fakeInteractiveRunner{}
	commands := &fakeCommandRunner{path: "/bin/sbx", output: sbx.Output{Err: errors.New("lookup failed")}}
	runner := UpRunner{Validation: fakeValidator{report: validReport()}, Sandboxes: sbx.Client{Commands: commands, Interactive: interactive}}
	_, err := runner.Run(context.Background(), "/repo", UpOptions{}, Streams{})
	if err == nil || interactive.calls != 0 {
		t.Fatalf("Run() error = %v, interactive calls = %d", err, interactive.calls)
	}
}

func TestRunnerSelectsExactMissingAndExistingArguments(t *testing.T) {
	attachArgs := []string{"run", "codex", "--name", "project"}
	for _, test := range []struct {
		name       string
		listOutput string
		wantArgs   [][]string
	}{
		{
			name:       "missing",
			listOutput: `{"sandboxes":[{"name":"other","status":"running"}]}`,
			wantArgs: [][]string{
				{"create", "--name", "project", "--kit", "git+https://github.com/example/kits.git#ref=v1&dir=tooling", "codex", "/repo"},
				attachArgs,
			},
		},
		{
			name:       "existing",
			listOutput: `{"sandboxes":[{"name":"project","status":"running"}]}`,
			wantArgs:   [][]string{attachArgs},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			assertRunnerInvocations(t, test.listOutput, test.wantArgs)
		})
	}
}

func assertRunnerInvocations(t *testing.T, listOutput string, wantArgs [][]string) {
	t.Helper()
	commands := &fakeCommandRunner{path: "/bin/sbx", output: sbx.Output{Stdout: []byte(listOutput)}}
	interactive := &fakeInteractiveRunner{}
	var stdout, stderr bytes.Buffer
	stdin := strings.NewReader("input")
	runner := UpRunner{Validation: fakeValidator{report: validReport()}, Sandboxes: sbx.Client{Commands: commands, Interactive: interactive}}
	_, err := runner.Run(context.Background(), "/repo/nested", UpOptions{}, Streams{In: stdin, Out: &stdout, Err: &stderr})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	assertInvocationArguments(t, interactive.invocations, wantArgs)
	assertInvocationStreamsAndTrust(t, interactive.invocations, stdin, &stdout, &stderr)
}

func assertInvocationArguments(t *testing.T, invocations []sbx.InteractiveInvocation, want [][]string) {
	t.Helper()
	got := make([][]string, 0, len(invocations))
	for _, invocation := range invocations {
		got = append(got, invocation.Args)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("args = %#v, want %#v", got, want)
	}
}

func assertInvocationStreamsAndTrust(t *testing.T, invocations []sbx.InteractiveInvocation, stdin io.Reader, stdout, stderr io.Writer) {
	t.Helper()
	for index, invocation := range invocations {
		if invocation.Executable != "/bin/sbx" || invocation.Stdin != stdin || invocation.Stdout != stdout || invocation.Stderr != stderr {
			t.Fatalf("interactive invocation %d = %#v", index, invocation)
		}
		if got := invocation.Environment["DOCKER_SANDBOXES_KIT_ALLOWED_SOURCES"]; got != `["docker.io/","github.com/example/kits"]` {
			t.Fatalf("invocation %d allowed sources = %q", index, got)
		}
	}
}

func TestRunnerRecreateLeavesAbsentSandboxOnCreatePath(t *testing.T) {
	sandboxes := &fakeUpSandboxes{}
	runner := UpRunner{Validation: fakeValidator{report: validReport()}, Sandboxes: sandboxes}
	_, err := runner.Run(context.Background(), "/repo", UpOptions{Recreate: true}, Streams{})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !reflect.DeepEqual(sandboxes.calls, []string{"lookup project", "create project", "run project"}) {
		t.Fatalf("calls = %#v, want lookup then creation without removal", sandboxes.calls)
	}
}

func TestRunnerRejectsForceWithoutRecreateBeforeValidation(t *testing.T) {
	validator := &countingValidator{report: validReport()}
	runner := UpRunner{Validation: validator, Sandboxes: &fakeUpSandboxes{}}
	_, err := runner.Run(context.Background(), "/repo", UpOptions{Force: true}, Streams{})
	if !errors.Is(err, ErrForceRequiresRecreate) || validator.calls != 0 {
		t.Fatalf("Run() error = %v, validation calls = %d", err, validator.calls)
	}
}

func TestRunnerRecreatesExactExistingSandboxBeforeCreatingFromPlan(t *testing.T) {
	sandboxes := &fakeUpSandboxes{state: sandboxport.StateStopped}
	var stdout, stderr bytes.Buffer
	stdin := strings.NewReader("input")
	streams := Streams{In: stdin, Out: &stdout, Err: &stderr}
	runner := UpRunner{Validation: fakeValidator{report: validReport()}, Sandboxes: sandboxes}

	_, err := runner.Run(context.Background(), "/repo/nested", UpOptions{Recreate: true}, streams)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !reflect.DeepEqual(sandboxes.calls, []string{"lookup project", "remove project", "create project", "run project"}) {
		t.Fatalf("calls = %#v, want lookup, removal, creation, then attachment", sandboxes.calls)
	}
	wantRemove := sandboxport.RemoveRequest{
		Name:  "project",
		Force: true,
		Streams: sandboxport.Streams{
			In: stdin, Out: &stdout, Err: &stderr,
		},
	}
	if !reflect.DeepEqual(sandboxes.removeRequest, wantRemove) {
		t.Fatalf("remove request = %#v, want %#v", sandboxes.removeRequest, wantRemove)
	}
	wantCreate := sandboxport.CreateRequest{
		Name:           "project",
		Agent:          "codex",
		Workspace:      "/repo",
		Kits:           []string{"git+https://github.com/example/kits.git#ref=v1&dir=tooling"},
		AllowedSources: []string{"docker.io/", "github.com/example/kits"},
	}
	if !reflect.DeepEqual(sandboxes.createRequest, wantCreate) {
		t.Fatalf("create request = %#v, want %#v", sandboxes.createRequest, wantCreate)
	}
	wantRun := sandboxport.RunRequest{
		Name:           "project",
		Agent:          "codex",
		AllowedSources: []string{"docker.io/", "github.com/example/kits"},
	}
	if !reflect.DeepEqual(sandboxes.runRequest, wantRun) {
		t.Fatalf("run request = %#v, want %#v", sandboxes.runRequest, wantRun)
	}
	if !reflect.DeepEqual(sandboxes.createStreams, wantRemove.Streams) {
		t.Fatalf("create streams = %#v, want %#v", sandboxes.createStreams, wantRemove.Streams)
	}
	if !reflect.DeepEqual(sandboxes.runStreams, wantRemove.Streams) {
		t.Fatalf("run streams = %#v, want %#v", sandboxes.runStreams, wantRemove.Streams)
	}
}

func TestRunnerConfirmsOnlyRunningRecreation(t *testing.T) {
	for _, test := range []struct {
		name       string
		state      sandboxport.State
		recreate   bool
		approved   bool
		force      bool
		confirmErr error
		wantCalls  []string
		wantErr    error
	}{
		{name: "running approved", state: sandboxport.StateRunning, recreate: true, approved: true, wantCalls: []string{"lookup project", "confirm project", "remove project", "create project", "run project"}},
		{name: "running forced", state: sandboxport.StateRunning, recreate: true, force: true, wantCalls: []string{"lookup project", "remove project", "create project", "run project"}},
		{name: "running declined", state: sandboxport.StateRunning, recreate: true, wantCalls: []string{"lookup project", "confirm project"}, wantErr: ErrRecreationCancelled},
		{name: "running input failure", state: sandboxport.StateRunning, recreate: true, confirmErr: io.ErrUnexpectedEOF, wantCalls: []string{"lookup project", "confirm project"}, wantErr: io.ErrUnexpectedEOF},
		{name: "stopped", state: sandboxport.StateStopped, recreate: true, wantCalls: []string{"lookup project", "remove project", "create project", "run project"}},
		{name: "stopped forced", state: sandboxport.StateStopped, recreate: true, force: true, wantCalls: []string{"lookup project", "remove project", "create project", "run project"}},
		{name: "absent forced", state: sandboxport.StateAbsent, recreate: true, force: true, wantCalls: []string{"lookup project", "create project", "run project"}},
		{name: "running ordinary up", state: sandboxport.StateRunning, wantCalls: []string{"lookup project", "run project"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			sandboxes := &fakeUpSandboxes{state: test.state}
			confirmer := &fakeConfirmer{approved: test.approved, err: test.confirmErr, calls: &sandboxes.calls}
			runner := UpRunner{Validation: fakeValidator{report: validReport()}, Sandboxes: sandboxes}
			_, err := runner.Run(context.Background(), "/repo", UpOptions{Recreate: test.recreate, Force: test.force, Confirmer: confirmer}, Streams{})
			if test.wantErr == nil && err != nil || test.wantErr != nil && !errors.Is(err, test.wantErr) {
				t.Fatalf("Run() error = %v, want %v", err, test.wantErr)
			}
			if !reflect.DeepEqual(sandboxes.calls, test.wantCalls) {
				t.Fatalf("calls = %#v, want %#v", sandboxes.calls, test.wantCalls)
			}
		})
	}
}

func TestRunnerRecreateStopsAtOperationFailures(t *testing.T) {
	lookupFailure := errors.New("lookup failed")
	removalFailure := errors.New("removal failed")
	createFailure := errors.New("creation failed")
	runFailure := errors.New("replacement failed")
	for _, test := range []struct {
		name         string
		sandboxes    *fakeUpSandboxes
		wantCalls    []string
		wantErr      error
		wantAttached bool
	}{
		{
			name:      "lookup",
			sandboxes: &fakeUpSandboxes{lookupErr: lookupFailure},
			wantCalls: []string{"lookup project"},
			wantErr:   lookupFailure,
		},
		{
			name:         "removal",
			sandboxes:    &fakeUpSandboxes{state: sandboxport.StateStopped, removeErr: removalFailure},
			wantCalls:    []string{"lookup project", "remove project"},
			wantErr:      removalFailure,
			wantAttached: true,
		},
		{
			name:         "replacement creation",
			sandboxes:    &fakeUpSandboxes{state: sandboxport.StateStopped, createErr: createFailure},
			wantCalls:    []string{"lookup project", "remove project", "create project"},
			wantErr:      createFailure,
			wantAttached: true,
		},
		{
			name:         "replacement execution",
			sandboxes:    &fakeUpSandboxes{state: sandboxport.StateStopped, runErr: runFailure},
			wantCalls:    []string{"lookup project", "remove project", "create project", "run project"},
			wantErr:      runFailure,
			wantAttached: true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			runner := UpRunner{Validation: fakeValidator{report: validReport()}, Sandboxes: test.sandboxes}
			_, err := runner.Run(context.Background(), "/repo", UpOptions{Recreate: true}, Streams{})
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("Run() error = %v, want %v", err, test.wantErr)
			}
			var attached AttachedProcessError
			if errors.As(err, &attached) != test.wantAttached {
				t.Fatalf("Run() attached error = %v, want %v", errors.As(err, &attached), test.wantAttached)
			}
			if !reflect.DeepEqual(test.sandboxes.calls, test.wantCalls) {
				t.Fatalf("calls = %#v, want %#v", test.sandboxes.calls, test.wantCalls)
			}
		})
	}
}

func TestRunnerMarksAttachedProcessFailureAsRendered(t *testing.T) {
	interactive := &fakeInteractiveRunner{err: errors.New("exit 9")}
	runner := UpRunner{
		Validation: fakeValidator{report: validReport()},
		Sandboxes:  sbx.Client{Commands: &fakeCommandRunner{path: "/bin/sbx", output: sbx.Output{Stdout: []byte(`{"sandboxes":[]}`)}}, Interactive: interactive},
	}
	_, err := runner.Run(context.Background(), "/repo", UpOptions{}, Streams{Err: io.Discard})
	var attached AttachedProcessError
	if !errors.As(err, &attached) || !strings.Contains(err.Error(), "exit 9") {
		t.Fatalf("Run() error = %T %v", err, err)
	}
}

func validReport() configuration.Validation {
	return configuration.Validation{
		Declaration: "/repo/sbxflow.yaml",
		Linked: configuration.LinkedConfiguration{
			Configuration: configuration.Configuration{Sandbox: configuration.Sandbox{Name: "project", Agent: "codex"}},
			Selections: []configuration.LinkedSelection{{
				Index:           0,
				Source:          configuration.Source{Type: configuration.SourceGit},
				RemoteReference: "git+https://github.com/example/kits.git#ref=v1&dir=tooling",
			}},
			Trust: configuration.Trust{AllowedSources: []string{"docker.io/", "github.com/example/kits"}},
		},
	}
}

type fakeValidator struct{ report configuration.Validation }

func (v fakeValidator) Run(context.Context, string) configuration.Validation { return v.report }

type countingValidator struct {
	report configuration.Validation
	calls  int
}

func (v *countingValidator) Run(context.Context, string) configuration.Validation {
	v.calls++
	return v.report
}

type fakeConfirmer struct {
	approved bool
	err      error
	calls    *[]string
}

func (c *fakeConfirmer) ConfirmRunningSandboxRecreation(name string, _ Streams) (bool, error) {
	*c.calls = append(*c.calls, "confirm "+name)
	return c.approved, c.err
}

type fakeCommandRunner struct {
	path     string
	pathErr  error
	output   sbx.Output
	args     []string
	onLookup func()
	lookups  int
	runs     int
}

func (r *fakeCommandRunner) LookPath(string) (string, error) {
	r.lookups++
	if r.onLookup != nil {
		r.onLookup()
	}
	return r.path, r.pathErr
}

func (r *fakeCommandRunner) Run(_ context.Context, _ string, args ...string) sbx.Output {
	r.runs++
	r.args = append([]string(nil), args...)
	return r.output
}

type fakeInteractiveRunner struct {
	calls        int
	invocation   sbx.InteractiveInvocation
	invocations  []sbx.InteractiveInvocation
	err          error
	checkContext func(context.Context)
}

func (r *fakeInteractiveRunner) Run(ctx context.Context, invocation sbx.InteractiveInvocation) error {
	r.calls++
	if ctx == nil {
		panic("nil context")
	}
	if r.checkContext != nil {
		r.checkContext(ctx)
	}
	r.invocation = invocation
	r.invocations = append(r.invocations, invocation)
	return r.err
}

type fakeUpSandboxes struct {
	state           sandboxport.State
	lookupErr       error
	removeErr       error
	createErr       error
	runErr          error
	calls           []string
	removeRequest   sandboxport.RemoveRequest
	createRequest   sandboxport.CreateRequest
	createStreams   sandboxport.Streams
	runRequest      sandboxport.RunRequest
	runStreams      sandboxport.Streams
	allowErr        error
	cleanupErr      error
	allowRequest    sandboxport.NetworkAllowRequest
	cleanupRequests []sandboxport.NetworkRemoveRequest
}

func (s *fakeUpSandboxes) CreateSandbox(_ context.Context, request sandboxport.CreateRequest, streams sandboxport.Streams) error {
	s.calls = append(s.calls, "create "+request.Name)
	s.createRequest = request
	s.createStreams = streams
	return s.createErr
}

func (s *fakeUpSandboxes) AllowNetwork(_ context.Context, request sandboxport.NetworkAllowRequest) error {
	s.calls = append(s.calls, "allow "+request.Name)
	s.allowRequest = request
	return s.allowErr
}

func (s *fakeUpSandboxes) RemoveNetworkResource(_ context.Context, request sandboxport.NetworkRemoveRequest) error {
	s.calls = append(s.calls, "cleanup "+request.Resource)
	s.cleanupRequests = append(s.cleanupRequests, request)
	return s.cleanupErr
}

func (s *fakeUpSandboxes) Inspect(_ context.Context, name string) (sandboxport.State, error) {
	s.calls = append(s.calls, "lookup "+name)
	if s.state == "" {
		s.state = sandboxport.StateAbsent
	}
	return s.state, s.lookupErr
}

func (s *fakeUpSandboxes) RemoveSandbox(_ context.Context, request sandboxport.RemoveRequest) error {
	s.calls = append(s.calls, "remove "+request.Name)
	s.removeRequest = request
	return s.removeErr
}

func (s *fakeUpSandboxes) RunSandbox(_ context.Context, request sandboxport.RunRequest, streams sandboxport.Streams) error {
	s.calls = append(s.calls, "run "+request.Name)
	s.runRequest = request
	s.runStreams = streams
	return s.runErr
}

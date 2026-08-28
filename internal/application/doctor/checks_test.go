package doctor

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestCompatibilityCheck(t *testing.T) {
	tests := []struct {
		name       string
		lookupPath string
		lookupErr  error
		output     CommandOutput
		wantStatus Status
		wantText   string
	}{
		{name: "unavailable", lookupErr: errors.New("not found"), wantStatus: StatusFail, wantText: "not installed"},
		{name: "unexecutable", lookupPath: "/fake/sbx", output: CommandOutput{ExitCode: -1, Err: errors.New("permission denied")}, wantStatus: StatusFail, wantText: "could not be determined"},
		{name: "malformed", lookupPath: "/fake/sbx", output: successfulOutput("unexpected"), wantStatus: StatusFail, wantText: "unrecognized"},
		{name: "below lower boundary", lookupPath: "/fake/sbx", output: successfulOutput("sbx version: v0.38.99 abc\n"), wantStatus: StatusFail, wantText: "unsupported"},
		{name: "lower boundary", lookupPath: "/fake/sbx", output: successfulOutput("sbx version: v0.39.0 abc\n"), wantStatus: StatusPass, wantText: "compatible"},
		{name: "upper supported", lookupPath: "/fake/sbx", output: successfulOutput("sbx version: v0.39.99 abc\n"), wantStatus: StatusPass, wantText: "compatible"},
		{name: "exclusive upper boundary", lookupPath: "/fake/sbx", output: successfulOutput("sbx version: v0.40.0 abc\n"), wantStatus: StatusFail, wantText: "unsupported"},
		{name: "pre-release at lower boundary", lookupPath: "/fake/sbx", output: successfulOutput("sbx version: v0.39.0-rc.1 abc\n"), wantStatus: StatusFail, wantText: "pre-release"},
		{name: "build metadata is released", lookupPath: "/fake/sbx", output: successfulOutput("sbx version: v0.39.0+build.7 abc\n"), wantStatus: StatusPass, wantText: "compatible"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			commands := &fakeCommandRunner{path: test.lookupPath, lookupErr: test.lookupErr, outputs: []CommandOutput{test.output}}
			result := compatibilityCheck{}.Run(context.Background(), Environment{Sandboxes: commands})
			if result.Status != test.wantStatus || !strings.Contains(result.Summary, test.wantText) {
				t.Fatalf("result = %#v, want status %s containing %q", result, test.wantStatus, test.wantText)
			}
			if result.Status == StatusPass && (result.Provides[FactSbxExecutable] == "" || result.Provides[FactSbxCompatible] == "") {
				t.Fatalf("pass result does not provide compatibility facts: %#v", result.Provides)
			}
		})
	}
}

func TestDiagnosticsCheckSummaries(t *testing.T) {
	tests := []struct {
		name       string
		stdout     string
		err        error
		wantStatus Status
		wantText   string
	}{
		{
			name:       "passing unknown fields ignored",
			stdout:     `{"version":"1.0","unknown":true,"checks":[{"name":"private"}],"summary":{"pass":4,"warn":0,"fail":0,"skip":1,"future":9}}`,
			wantStatus: StatusPass,
			wantText:   "4 passed, 0 warned, 0 failed, 1 skipped",
		},
		{
			name:       "warnings",
			stdout:     `{"version":"1.0","summary":{"pass":3,"warn":1,"fail":0,"skip":0}}`,
			wantStatus: StatusWarn,
			wantText:   "1 warned",
		},
		{
			name:       "valid summary despite diagnostic exit",
			stdout:     `{"version":"1.0","summary":{"pass":2,"warn":0,"fail":1,"skip":1}}`,
			err:        errors.New("exit status 1"),
			wantStatus: StatusFail,
			wantText:   "1 failed",
		},
		{name: "malformed", stdout: `{`, wantStatus: StatusFail, wantText: "could not be summarized"},
		{name: "unsupported envelope", stdout: `{"version":"2.0","summary":{"pass":1,"warn":0,"fail":0,"skip":0}}`, wantStatus: StatusFail, wantText: "could not be summarized"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			commands := &fakeCommandRunner{outputs: []CommandOutput{{Stdout: []byte(test.stdout), Err: test.err}}}
			result := diagnosticsCheck{}.Run(context.Background(), compatibleEnvironment(commands))
			if result.Status != test.wantStatus || !strings.Contains(result.Summary, test.wantText) {
				t.Fatalf("result = %#v, want status %s containing %q", result, test.wantStatus, test.wantText)
			}
			if strings.Contains(result.Summary, "private") {
				t.Fatalf("summary reproduces individual check: %q", result.Summary)
			}
		})
	}
}

func TestNetworkPolicyCheck(t *testing.T) {
	tests := []struct {
		name       string
		output     CommandOutput
		wantStatus Status
		wantText   string
	}{
		{name: "initialized local", output: successfulOutput(`{"rules":[{"policy_id":"local-policy","scope":"global","resource_type":"network","origin":"local","status":"active"}]}`), wantStatus: StatusPass, wantText: "initialized"},
		{name: "deny-all has only global filesystem rules", output: successfulOutput(`{"rules":[{"policy_id":"local-policy","scope":"global","resource_type":"filesystem:read","origin":"local","status":"active"},{"policy_id":"local-policy","scope":"global","resource_type":"filesystem:write","origin":"local","status":"active"}]}`), wantStatus: StatusPass, wantText: "initialized"},
		{name: "active organisation", output: successfulOutput(`{"rules":[{"policy_id":"managed-policy","scope":"global","resource_type":"filesystem:read","origin":"org","status":"active"}]}`), wantStatus: StatusPass, wantText: "organisation-managed"},
		{name: "scoped organisation rule is not global", output: successfulOutput(`{"rules":[{"policy_id":"managed-policy","scope":"sandbox:example","resource_type":"network","origin":"org","status":"active"}]}`), wantStatus: StatusWarn, wantText: "not initialized"},
		{name: "absent", output: successfulOutput(`{"rules":[]}`), wantStatus: StatusWarn, wantText: "not initialized"},
		{name: "sandbox-only policies do not initialize global policy", output: successfulOutput(`{"rules":[{"policy_id":"sandbox-policy","scope":"sandbox:example","resource_type":"network","origin":"local","status":"active"}]}`), wantStatus: StatusWarn, wantText: "not initialized"},
		{name: "inactive does not initialize", output: successfulOutput(`{"rules":[{"policy_id":"local-policy","scope":"global","resource_type":"network","origin":"local","status":"inactive"}]}`), wantStatus: StatusWarn, wantText: "not initialized"},
		{name: "unavailable", output: CommandOutput{Err: errors.New("daemon unavailable")}, wantStatus: StatusWarn, wantText: "could not be inspected"},
		{name: "malformed", output: successfulOutput(`{"policies":[]}`), wantStatus: StatusWarn, wantText: "could not be inspected"},
		{name: "null rules", output: successfulOutput(`{"rules":null}`), wantStatus: StatusWarn, wantText: "could not be inspected"},
		{name: "incomplete rule", output: successfulOutput(`{"rules":[{"scope":"global","resource_type":"network","origin":"local","status":"active"}]}`), wantStatus: StatusWarn, wantText: "could not be inspected"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			commands := &fakeCommandRunner{outputs: []CommandOutput{test.output}}
			result := networkPolicyCheck{}.Run(context.Background(), compatibleEnvironment(commands))
			if result.Status != test.wantStatus || !strings.Contains(result.Summary, test.wantText) {
				t.Fatalf("result = %#v, want status %s containing %q", result, test.wantStatus, test.wantText)
			}
			if strings.Contains(strings.ToLower(result.Guidance), "override") && strings.Contains(test.name, "organisation") {
				t.Fatalf("organisation guidance recommends override: %q", result.Guidance)
			}
		})
	}
}

func TestKitSourcesCheck(t *testing.T) {
	tests := []struct {
		name         string
		output       CommandOutput
		wantStatus   Status
		wantText     string
		wantGuidance string
	}{
		{name: "restricted prefixes", output: kitSetting(`["docker.io/","github.com/docker/"]`, "default"), wantStatus: StatusPass, wantText: "restricted"},
		{name: "wildcard mixed with prefixes", output: kitSetting(`["docker.io/","*"]`, "default"), wantStatus: StatusWarn, wantText: "unrestricted"},
		{name: "local override", output: kitSetting(`["*"]`, "override"), wantStatus: StatusWarn, wantText: "local override", wantGuidance: "local"},
		{name: "environment", output: kitSetting(`["*"]`, "env"), wantStatus: StatusWarn, wantText: "environment", wantGuidance: "DOCKER_SANDBOXES_KIT_ALLOWED_SOURCES"},
		{name: "organisation managed", output: kitSetting(`["*"]`, "org"), wantStatus: StatusWarn, wantText: "organisation-managed", wantGuidance: "administrator"},
		{name: "similar value is not wildcard", output: kitSetting(`["github.com/*"]`, "override"), wantStatus: StatusPass, wantText: "restricted"},
		{name: "unreadable", output: CommandOutput{Err: errors.New("unavailable")}, wantStatus: StatusWarn, wantText: "could not be inspected"},
		{name: "malformed", output: successfulOutput(`{"key":"kit.allowedSources","value":"*","source":"default"}`), wantStatus: StatusWarn, wantText: "could not be inspected"},
		{name: "null value", output: successfulOutput(`{"key":"kit.allowedSources","value":null,"source":"default"}`), wantStatus: StatusWarn, wantText: "could not be inspected"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			commands := &fakeCommandRunner{outputs: []CommandOutput{test.output}}
			result := kitSourcesCheck{}.Run(context.Background(), compatibleEnvironment(commands))
			if result.Status != test.wantStatus || !strings.Contains(result.Summary, test.wantText) || !strings.Contains(result.Guidance, test.wantGuidance) {
				t.Fatalf("result = %#v, want status %s, summary %q, guidance %q", result, test.wantStatus, test.wantText, test.wantGuidance)
			}
		})
	}
}

func TestDefaultChecksAreInspectionOnly(t *testing.T) {
	commands := &fakeCommandRunner{
		path: "/fake/sbx",
		outputs: []CommandOutput{
			successfulOutput("sbx version: v0.39.0 abc\n"),
			successfulOutput(`{"version":"1.0","summary":{"pass":1,"warn":0,"fail":0,"skip":0}}`),
			successfulOutput(`{"rules":[]}`),
			kitSetting(`["docker.io/"]`, "default"),
		},
	}

	report := NewRunner(commands, compatibilityCheck{}, diagnosticsCheck{}, networkPolicyCheck{}, kitSourcesCheck{}).Run(context.Background())
	if report.Failed() {
		t.Fatalf("report unexpectedly failed: %#v", report)
	}

	want := [][]string{
		{"version"},
		{"diagnose", "--output", "json"},
		{"policy", "ls", "--json", "--include-inactive"},
		{"settings", "get", "--json", "kit.allowedSources"},
	}
	if !reflect.DeepEqual(commands.calls, want) {
		t.Fatalf("calls = %#v, want inspection-only calls %#v", commands.calls, want)
	}
	for _, call := range commands.calls {
		joined := strings.Join(call, " ")
		for _, mutation := range []string{" set ", " init ", " reset ", " start ", " setup ", " allow ", " deny "} {
			if strings.Contains(" "+joined+" ", mutation) {
				t.Fatalf("check invoked mutation command %q", joined)
			}
		}
	}
}

type fakeCommandRunner struct {
	path      string
	lookupErr error
	outputs   []CommandOutput
	calls     [][]string
}

func (r *fakeCommandRunner) Locate() (string, error) { return r.path, r.lookupErr }
func (r *fakeCommandRunner) Version(context.Context, string) CommandOutput {
	return r.run("version")
}
func (r *fakeCommandRunner) Diagnose(context.Context, string) CommandOutput {
	return r.run("diagnose", "--output", "json")
}
func (r *fakeCommandRunner) ListPolicies(context.Context, string) CommandOutput {
	return r.run("policy", "ls", "--json", "--include-inactive")
}
func (r *fakeCommandRunner) GetKitAllowedSources(context.Context, string) CommandOutput {
	return r.run("settings", "get", "--json", "kit.allowedSources")
}
func (r *fakeCommandRunner) run(args ...string) CommandOutput {
	r.calls = append(r.calls, append([]string(nil), args...))
	if len(r.outputs) == 0 {
		return CommandOutput{ExitCode: -1, Err: errors.New("unexpected command")}
	}
	output := r.outputs[0]
	r.outputs = r.outputs[1:]
	return output
}

func successfulOutput(stdout string) CommandOutput {
	return CommandOutput{Stdout: []byte(stdout), ExitCode: 0}
}

func compatibleEnvironment(sandboxes Inspector) Environment {
	return Environment{
		Sandboxes: sandboxes,
		Facts: map[Fact]string{
			FactSbxExecutable: "/fake/sbx",
			FactSbxCompatible: SupportedSbxMinimum,
		},
	}
}

func kitSetting(value, source string) CommandOutput {
	return successfulOutput(fmt.Sprintf(`{"key":"kit.allowedSources","value":%s,"type":"json","source":%q}`, value, source))
}

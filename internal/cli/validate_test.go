package cli

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamessawle/sbxflow/internal/buildinfo"
	"github.com/jamessawle/sbxflow/internal/config"
	"github.com/jamessawle/sbxflow/internal/localkit"
	"github.com/jamessawle/sbxflow/internal/validation"
)

func TestValidateHelp(t *testing.T) {
	for _, args := range [][]string{{"validate", "--help"}, {"help", "validate"}} {
		stdout, stderr, err := execute(args)
		if err != nil || stderr != "" {
			t.Fatalf("validate help error = %v, stderr = %q", err, stderr)
		}
		for _, want := range []string{"sbxflow validate", "nearest sbxflow.yaml", "checked offline", "validated by Docker Sandboxes"} {
			if !strings.Contains(stdout, want) {
				t.Errorf("stdout does not contain %q:\n%s", want, stdout)
			}
		}
	}
}

func TestValidateRendersSuccessReport(t *testing.T) {
	declaration := filepath.Join("repo", "sbxflow.yaml")
	report := validation.Report{
		Declaration: declaration,
		Linked: config.LinkedConfiguration{
			Configuration: config.Configuration{Version: 1},
			Trust:         config.Trust{AllowedSources: []string{"docker.io/", "github.com/example/kits"}, AllowLocalKits: true},
		},
		LocalKits: []localkit.Result{{Target: localkit.Target{Source: "local", Kit: "tooling", Path: "/repo/kits/tooling"}, Valid: true}},
	}
	stdout, stderr, err := executeWithValidate([]string{"validate"}, fakeValidateRunner{report: report})
	if err != nil || stderr != "" {
		t.Fatalf("validate error = %v, stderr = %q", err, stderr)
	}
	want := "Declaration: " + declaration + "\n\n" +
		"Derived State:\n" +
		"  Kit:\n" +
		"    Allowed Sources:\n" +
		"      - docker.io/\n" +
		"      - github.com/example/kits\n" +
		"    Local Kits Allowed: true\n\n" +
		"Validation:\n" +
		"  State: pass\n" +
		"  Findings: []\n"
	if stdout != want {
		t.Fatalf("stdout = %q, want %q", stdout, want)
	}
}

func TestValidateRendersActionableFailure(t *testing.T) {
	report := validation.Report{
		Declaration: "/repo/sbxflow.yaml",
		Linked:      config.LinkedConfiguration{Configuration: config.Configuration{Version: 1}, Trust: config.Trust{AllowedSources: []string{"docker.io/"}, AllowLocalKits: true}},
		LocalKits:   []localkit.Result{{Target: localkit.Target{Source: "local", Kit: "bad", Path: "/repo/kits/bad"}, Diagnostics: "missing metadata", Err: errors.New("rejected")}},
		Errors:      []error{errors.New("local kit local/bad was rejected by sbx")},
	}
	stdout, stderr, err := executeWithValidate([]string{"validate"}, fakeValidateRunner{report: report})
	if !errors.Is(err, errValidationFailed) {
		t.Fatalf("error = %v, want validation failure", err)
	}
	if stdout != "" {
		t.Fatalf("stdout = %q, want empty cohesive failure output", stdout)
	}
	for _, want := range []string{"Declaration: /repo/sbxflow.yaml", "Derived State:", "State: fail", "Findings:", "local kit local/bad was rejected by sbx: missing metadata"} {
		if !strings.Contains(stderr, want) {
			t.Errorf("stderr does not contain %q:\n%s", want, stderr)
		}
	}
	if strings.Contains(stderr, errValidationFailed.Error()) {
		t.Fatalf("failure appends an unstructured Cobra error: %q", stderr)
	}
}

func TestValidateRendersUnavailableDerivedState(t *testing.T) {
	report := validation.Report{Errors: []error{errors.New("no sbxflow.yaml found")}}
	stdout, stderr, err := executeWithValidate([]string{"validate"}, fakeValidateRunner{report: report})
	if !errors.Is(err, errValidationFailed) || stdout != "" {
		t.Fatalf("error = %v, stdout = %q", err, stdout)
	}
	want := "Declaration: unavailable\n\n" +
		"Derived State: unavailable\n\n" +
		"Validation:\n" +
		"  State: fail\n" +
		"  Findings:\n" +
		"    - no sbxflow.yaml found\n"
	if stderr != want {
		t.Fatalf("stderr = %q, want %q", stderr, want)
	}
}

func executeWithValidate(args []string, runner ValidateRunner) (string, string, error) {
	var output strings.Builder
	var errorOutput strings.Builder
	root := NewRootCommand(
		Streams{In: strings.NewReader(""), Out: &output, Err: &errorOutput},
		fakeDoctorRunner{},
		runner,
		&fakeUpRunner{},
		&fakeDownRunner{},
		&fakeDestroyRunner{},
	)
	root.Version = formatVersion(buildinfo.Info{Version: "development"})
	root.SetArgs(args)
	err := root.ExecuteContext(context.Background())
	return output.String(), errorOutput.String(), err
}

type fakeValidateRunner struct{ report validation.Report }

func (r fakeValidateRunner) Run(context.Context, string) validation.Report { return r.report }

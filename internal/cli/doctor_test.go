package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamessawle/sbxflow/internal/buildinfo"
	"github.com/jamessawle/sbxflow/internal/doctor"
)

func TestDoctorHelp(t *testing.T) {
	for name, args := range map[string][]string{
		"flag":         {"doctor", "--help"},
		"help command": {"help", "doctor"},
	} {
		t.Run(name, func(t *testing.T) {
			stdout, stderr, err := execute(args)
			if err != nil {
				t.Fatalf("doctor help error = %v; stderr = %q", err, stderr)
			}
			if stderr != "" {
				t.Fatalf("stderr = %q, want empty", stderr)
			}
			for _, want := range []string{
				"sbxflow doctor",
				"compatible, healthy, and safely configured",
				"without reading sbxflow.yaml or changing the system",
			} {
				if !strings.Contains(stdout, want) {
					t.Errorf("stdout does not contain %q:\n%s", want, stdout)
				}
			}
		})
	}
}

func TestDoctorRendersDeterministicResults(t *testing.T) {
	report := doctor.Report{Results: []doctor.Result{
		{ID: "sbx-compatibility", Status: doctor.StatusPass, Grade: doctor.GradeRequired, Summary: "sbx v0.35.0 is compatible"},
		{ID: "docker-diagnostics", Status: doctor.StatusWarn, Grade: doctor.GradeRequired, Summary: "3 passed, 1 warned, 0 failed, 0 skipped", Guidance: "Run `sbx diagnose` for detailed results."},
		{ID: "network-policy", Status: doctor.StatusSkip, Grade: doctor.GradeAdvisory, Summary: "prerequisite unavailable: daemon"},
	}}

	stdout, stderr, err := executeWithDoctor(
		[]string{"doctor"},
		buildinfo.Info{Version: "development"},
		fakeDoctorRunner{report: report},
	)
	if err != nil {
		t.Fatalf("doctor error = %v; stderr = %q", err, stderr)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
	want := "[PASS] sbx-compatibility: sbx v0.35.0 is compatible\n" +
		"[WARN] docker-diagnostics: 3 passed, 1 warned, 0 failed, 0 skipped\n" +
		"  Run `sbx diagnose` for detailed results.\n" +
		"[SKIP] network-policy: prerequisite unavailable: daemon\n"
	if stdout != want {
		t.Fatalf("stdout = %q, want %q", stdout, want)
	}
}

func TestDoctorWarningOnlySucceedsAndRequiredFailureFails(t *testing.T) {
	tests := []struct {
		name      string
		result    doctor.Result
		wantError bool
	}{
		{
			name:   "advisory warning",
			result: doctor.Result{ID: "kit-sources", Status: doctor.StatusWarn, Grade: doctor.GradeAdvisory, Summary: "unrestricted"},
		},
		{
			name:      "required failure",
			result:    doctor.Result{ID: "docker-diagnostics", Status: doctor.StatusFail, Grade: doctor.GradeRequired, Summary: "failed"},
			wantError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stdout, stderr, err := executeWithDoctor(
				[]string{"doctor"},
				buildinfo.Info{Version: "development"},
				fakeDoctorRunner{report: doctor.Report{Results: []doctor.Result{test.result}}},
			)
			if (err != nil) != test.wantError {
				t.Fatalf("error = %v, wantError %v", err, test.wantError)
			}
			if !strings.Contains(stdout, strings.ToUpper(string(test.result.Status))) {
				t.Fatalf("stdout does not render result: %q", stdout)
			}
			if test.wantError && !strings.Contains(stderr, errDoctorFailed.Error()) {
				t.Fatalf("stderr = %q, want doctor failure", stderr)
			}
			if !test.wantError && stderr != "" {
				t.Fatalf("stderr = %q, want empty", stderr)
			}
		})
	}
}

func TestDoctorDoesNotReadRepositoryDeclaration(t *testing.T) {
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "sbxflow.yaml"), []byte("not: [valid"), 0o600); err != nil {
		t.Fatalf("write invalid declaration: %v", err)
	}
	t.Chdir(directory)

	stdout, stderr, err := executeWithDoctor(
		[]string{"doctor"},
		buildinfo.Info{Version: "development"},
		fakeDoctorRunner{report: doctor.Report{Results: []doctor.Result{{
			ID: "sbx-compatibility", Status: doctor.StatusPass, Grade: doctor.GradeRequired, Summary: "compatible",
		}}}},
	)
	if err != nil || stderr != "" {
		t.Fatalf("doctor error = %v; stderr = %q", err, stderr)
	}
	if strings.Contains(stdout, "sbxflow.yaml") || strings.Contains(stdout, "not: [valid") {
		t.Fatalf("doctor output references declaration: %q", stdout)
	}
}

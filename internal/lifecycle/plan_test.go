package lifecycle

import (
	"reflect"
	"testing"

	"github.com/jamessawle/sbxflow/internal/config"
	"github.com/jamessawle/sbxflow/internal/localkit"
	"github.com/jamessawle/sbxflow/internal/validation"
)

func TestNewPlanPreservesKitOrderAndTrust(t *testing.T) {
	report := validation.Report{
		Declaration: "/repo/sbxflow.yaml",
		Linked: config.LinkedConfiguration{
			Configuration: config.Configuration{Sandbox: config.Sandbox{Name: "project", Agent: "codex"}},
			Selections: []config.LinkedSelection{
				{Index: 0, Source: config.Source{Type: config.SourceGit}, RemoteReference: "git+https://github.com/example/kits.git#ref=v1&dir=git-kit"},
				{Index: 1, Source: config.Source{Type: config.SourceLocal}},
				{Index: 2, Source: config.Source{Type: config.SourceOCI}, RemoteReference: "ghcr.io/example/kits/oci-kit:v2"},
			},
			Trust: config.Trust{
				AllowedSources: []string{"docker.io/", "github.com/example/kits", "ghcr.io/example/kits/oci-kit"},
				AllowLocalKits: true,
			},
		},
		LocalKits: []localkit.Result{{
			Target: localkit.Target{Index: 1, Path: "/canonical/repo/kits/local-kit"},
			Valid:  true,
		}},
	}

	plan, err := NewPlan(report)
	if err != nil {
		t.Fatalf("NewPlan() error = %v", err)
	}
	if plan.Name != "project" || plan.Agent != "codex" || plan.Workspace != "/repo" {
		t.Fatalf("NewPlan() identity = %#v", plan)
	}
	wantKits := []string{
		"git+https://github.com/example/kits.git#ref=v1&dir=git-kit",
		"/canonical/repo/kits/local-kit",
		"ghcr.io/example/kits/oci-kit:v2",
	}
	if !reflect.DeepEqual(plan.Kits, wantKits) {
		t.Fatalf("Kits = %#v, want %#v", plan.Kits, wantKits)
	}
	wantEnvironment := map[string]string{
		"DOCKER_SANDBOXES_KIT_ALLOWED_SOURCES": `["docker.io/","github.com/example/kits","ghcr.io/example/kits/oci-kit"]`,
		"DOCKER_SANDBOXES_KIT_ALLOW_LOCAL":     "true",
	}
	if !reflect.DeepEqual(plan.Environment, wantEnvironment) {
		t.Fatalf("Environment = %#v, want %#v", plan.Environment, wantEnvironment)
	}
}

func TestNewPlanRejectsMissingLocalValidationResult(t *testing.T) {
	report := validation.Report{
		Declaration: "/repo/sbxflow.yaml",
		Linked: config.LinkedConfiguration{
			Configuration: config.Configuration{Sandbox: config.Sandbox{Name: "project", Agent: "codex"}},
			Selections:    []config.LinkedSelection{{Index: 3, Source: config.Source{Type: config.SourceLocal}}},
		},
	}
	if _, err := NewPlan(report); err == nil {
		t.Fatal("NewPlan() error = nil, want missing local result")
	}
}

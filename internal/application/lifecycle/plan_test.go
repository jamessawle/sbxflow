package lifecycle

import (
	"reflect"
	"testing"

	"github.com/jamessawle/sbxflow/internal/domain/configuration"
)

func TestNewPlanPreservesKitOrderAndTrust(t *testing.T) {
	report := configuration.Validation{
		Declaration: "/repo/sbxflow.yaml",
		Linked: configuration.LinkedConfiguration{
			Configuration: configuration.Configuration{Sandbox: configuration.Sandbox{Name: "project", Agent: "codex", Network: configuration.Network{AllowedHosts: []string{"first.example", "second.example"}}}},
			Selections: []configuration.LinkedSelection{
				{Index: 0, Source: configuration.Source{Type: configuration.SourceGit}, RemoteReference: "git+https://github.com/example/kits.git#ref=v1&dir=git-kit"},
				{Index: 1, Source: configuration.Source{Type: configuration.SourceLocal}},
				{Index: 2, Source: configuration.Source{Type: configuration.SourceOCI}, RemoteReference: "ghcr.io/example/kits/oci-kit:v2"},
			},
			Trust: configuration.Trust{
				AllowedSources: []string{"docker.io/", "github.com/example/kits", "ghcr.io/example/kits/oci-kit"},
				AllowLocalKits: true,
			},
		},
		LocalKits: []configuration.LocalKitValidation{{
			Target: configuration.LocalKit{Index: 1, Path: "/canonical/repo/kits/local-kit"},
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
	if !reflect.DeepEqual(plan.Trust, report.Linked.Trust) {
		t.Fatalf("Trust = %#v, want %#v", plan.Trust, report.Linked.Trust)
	}
	if !reflect.DeepEqual(plan.AllowedHosts, []string{"first.example", "second.example"}) {
		t.Fatalf("AllowedHosts = %#v", plan.AllowedHosts)
	}
}

func TestNewPlanRejectsMissingLocalValidationResult(t *testing.T) {
	report := configuration.Validation{
		Declaration: "/repo/sbxflow.yaml",
		Linked: configuration.LinkedConfiguration{
			Configuration: configuration.Configuration{Sandbox: configuration.Sandbox{Name: "project", Agent: "codex"}},
			Selections:    []configuration.LinkedSelection{{Index: 3, Source: configuration.Source{Type: configuration.SourceLocal}}},
		},
	}
	if _, err := NewPlan(report); err == nil {
		t.Fatal("NewPlan() error = nil, want missing local result")
	}
}

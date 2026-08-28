package lifecycle

import (
	"reflect"
	"testing"

	"github.com/jamessawle/sbxflow/internal/domain/configuration"
	sandboxport "github.com/jamessawle/sbxflow/internal/ports/sandbox"
)

func TestNewPlanPreservesKitOrderAndTrust(t *testing.T) {
	report := configuration.Validation{
		Declaration: "/repo/sbxflow.yaml",
		Linked: configuration.LinkedConfiguration{
			Configuration: configuration.Configuration{Sandbox: configuration.Sandbox{Name: "project", Agent: "codex", Network: configuration.Network{AllowedHosts: []string{"first.example", "second.example"}}, Hooks: configuration.Hooks{Initialize: []configuration.Command{{Command: []string{"npm", "ci"}}}}}},
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
	if plan.Environment.Name != "project" || plan.Environment.Agent != "codex" || plan.Environment.Workspace != "/repo" {
		t.Fatalf("NewPlan() identity = %#v", plan)
	}
	wantKits := []string{
		"git+https://github.com/example/kits.git#ref=v1&dir=git-kit",
		"/canonical/repo/kits/local-kit",
		"ghcr.io/example/kits/oci-kit:v2",
	}
	if !reflect.DeepEqual(plan.Environment.Kits, wantKits) {
		t.Fatalf("Kits = %#v, want %#v", plan.Environment.Kits, wantKits)
	}
	if !reflect.DeepEqual(plan.Environment.AllowedSources, report.Linked.Trust.AllowedSources) || plan.Environment.AllowLocalKits != report.Linked.Trust.AllowLocalKits {
		t.Fatalf("Environment trust = %#v, want %#v", plan.Environment, report.Linked.Trust)
	}
	if !reflect.DeepEqual(plan.AllowedHosts, []string{"first.example", "second.example"}) {
		t.Fatalf("AllowedHosts = %#v", plan.AllowedHosts)
	}
	if !reflect.DeepEqual(plan.Initialize, [][]string{{"npm", "ci"}}) {
		t.Fatalf("Initialize = %#v", plan.Initialize)
	}
	report.Linked.Configuration.Sandbox.Hooks.Initialize[0].Command[0] = "changed"
	report.Linked.Trust.AllowedSources[0] = "changed"
	if plan.Initialize[0][0] != "npm" {
		t.Fatalf("Initialize shares declaration storage: %#v", plan.Initialize)
	}
	if plan.Environment.AllowedSources[0] != "docker.io/" {
		t.Fatalf("AllowedSources shares declaration storage: %#v", plan.Environment.AllowedSources)
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

func TestNewPlanMapsWorkspaceModes(t *testing.T) {
	for _, test := range []struct {
		name      string
		workspace *configuration.Workspace
		want      sandboxport.WorkspaceMode
	}{
		{name: "omitted", want: sandboxport.WorkspaceModeDirect},
		{name: "explicit direct", workspace: &configuration.Workspace{Mode: configuration.WorkspaceModeDirect}, want: sandboxport.WorkspaceModeDirect},
		{name: "clone", workspace: &configuration.Workspace{Mode: configuration.WorkspaceModeClone}, want: sandboxport.WorkspaceModeClone},
	} {
		t.Run(test.name, func(t *testing.T) {
			report := configuration.Validation{
				Declaration: "/repo/sbxflow.yaml",
				Linked: configuration.LinkedConfiguration{Configuration: configuration.Configuration{Sandbox: configuration.Sandbox{
					Name: "project", Agent: "codex", Workspace: test.workspace,
				}}},
			}
			plan, err := NewPlan(report)
			if err != nil || plan.Environment.WorkspaceMode != test.want {
				t.Fatalf("NewPlan() mode = %q, %v, want %q", plan.Environment.WorkspaceMode, err, test.want)
			}
			if plan.Environment.Workspace != "/repo" {
				t.Fatalf("NewPlan() workspace = %q, want declaration directory", plan.Environment.Workspace)
			}
		})
	}
}

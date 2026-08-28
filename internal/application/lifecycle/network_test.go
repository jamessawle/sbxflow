package lifecycle

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/jamessawle/sbxflow/internal/domain/configuration"
	sandboxport "github.com/jamessawle/sbxflow/internal/ports/sandbox"
)

func networkReport() configuration.Validation {
	report := validReport()
	report.Linked.Configuration.Sandbox.Network.AllowedHosts = []string{"first.example", "second.example"}
	return report
}

func TestUpAppliesNetworkBetweenCreationAndAttachment(t *testing.T) {
	sandboxes := &fakeUpSandboxes{}
	runner := UpRunner{Validation: fakeValidator{report: networkReport()}, Sandboxes: sandboxes}
	if _, err := runner.Run(context.Background(), "/repo", UpOptions{}, Streams{}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	// Docker Sandboxes scopes a rule only to an existing sandbox, so the rule
	// lands after creation but before the agent starts.
	wantCalls := []string{"lookup project", "create project", "allow project", "run project"}
	if !reflect.DeepEqual(sandboxes.calls, wantCalls) {
		t.Fatalf("calls = %#v, want %#v", sandboxes.calls, wantCalls)
	}
	wantAllow := sandboxport.NetworkAllowRequest{Name: "project", Resources: []string{"first.example", "second.example"}}
	if !reflect.DeepEqual(sandboxes.allowRequest, wantAllow) {
		t.Fatalf("allow request = %#v, want %#v", sandboxes.allowRequest, wantAllow)
	}
}

func TestUpSkipsNetworkForExistingSandboxAndEmptyDeclaration(t *testing.T) {
	existing := &fakeUpSandboxes{state: sandboxport.StateRunning}
	runner := UpRunner{Validation: fakeValidator{report: networkReport()}, Sandboxes: existing}
	if _, err := runner.Run(context.Background(), "/repo", UpOptions{}, Streams{}); err != nil {
		t.Fatalf("existing Run() error = %v", err)
	}
	if !reflect.DeepEqual(existing.calls, []string{"lookup project", "run project"}) {
		t.Fatalf("existing calls = %#v, want attachment without reconciliation", existing.calls)
	}

	withoutHosts := &fakeUpSandboxes{}
	runner = UpRunner{Validation: fakeValidator{report: validReport()}, Sandboxes: withoutHosts}
	if _, err := runner.Run(context.Background(), "/repo", UpOptions{}, Streams{}); err != nil {
		t.Fatalf("empty declaration Run() error = %v", err)
	}
	if !reflect.DeepEqual(withoutHosts.calls, []string{"lookup project", "create project", "run project"}) {
		t.Fatalf("empty declaration calls = %#v, want no network request", withoutHosts.calls)
	}
}

func TestUpRemovesCreatedSandboxWhenNetworkAllowFails(t *testing.T) {
	sandboxes := &fakeUpSandboxes{allowErr: errors.New("allow failed")}
	runner := UpRunner{Validation: fakeValidator{report: networkReport()}, Sandboxes: sandboxes}
	_, err := runner.Run(context.Background(), "/repo", UpOptions{}, Streams{})
	if !errors.Is(err, sandboxes.allowErr) {
		t.Fatalf("Run() error = %v, want the allow failure", err)
	}
	// The sandbox exists by the time the rule is rejected, so it is removed again
	// rather than entered without its declared access.
	wantCalls := []string{"lookup project", "create project", "allow project", "remove project"}
	if !reflect.DeepEqual(sandboxes.calls, wantCalls) {
		t.Fatalf("calls = %#v, want %#v", sandboxes.calls, wantCalls)
	}
	if !sandboxes.removeRequest.Force {
		t.Fatalf("remove request = %#v, want a forced removal of the just-created sandbox", sandboxes.removeRequest)
	}
}

func TestUpReportsAllowAndRollbackFailuresTogether(t *testing.T) {
	sandboxes := &fakeUpSandboxes{allowErr: errors.New("allow failed"), removeErr: errors.New("rollback failed")}
	runner := UpRunner{Validation: fakeValidator{report: networkReport()}, Sandboxes: sandboxes}
	_, err := runner.Run(context.Background(), "/repo", UpOptions{}, Streams{})
	if err == nil || !strings.Contains(err.Error(), "allow failed") || !strings.Contains(err.Error(), "rollback failed") {
		t.Fatalf("Run() error = %v, want both diagnostics", err)
	}
	if !errors.Is(err, sandboxes.allowErr) {
		t.Fatalf("Run() error = %v, want the allow failure preserved for unwrapping", err)
	}
}

func TestRecreationCleansThenReappliesNetwork(t *testing.T) {
	sandboxes := &fakeUpSandboxes{state: sandboxport.StateStopped}
	runner := UpRunner{Validation: fakeValidator{report: networkReport()}, Sandboxes: sandboxes}
	_, err := runner.Run(context.Background(), "/repo", UpOptions{Recreate: true}, Streams{})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	want := []string{"lookup project", "remove project", "create project", "allow project", "run project"}
	if !reflect.DeepEqual(sandboxes.calls, want) {
		t.Fatalf("calls = %#v, want %#v", sandboxes.calls, want)
	}
}

func TestRemovalDelegatesScopedResourceCleanupToEnvironmentRemoval(t *testing.T) {
	sandboxes := &fakeUpSandboxes{removeErr: errors.New("remove failed")}
	err := removeSandbox(context.Background(), sandboxes, "project", true, Streams{})
	if !errors.Is(err, sandboxes.removeErr) || !reflect.DeepEqual(sandboxes.calls, []string{"remove project"}) {
		t.Fatalf("removeSandbox() = %v, calls %#v", err, sandboxes.calls)
	}
	sandboxes = &fakeUpSandboxes{}
	err = removeSandbox(context.Background(), sandboxes, "project", true, Streams{})
	if err != nil || !reflect.DeepEqual(sandboxes.calls, []string{"remove project"}) || len(sandboxes.cleanupRequests) != 0 {
		t.Fatalf("removeSandbox() = %v, calls %#v", err, sandboxes.calls)
	}
}

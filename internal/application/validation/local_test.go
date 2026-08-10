// Local-kit validation tests exercise application orchestration at its
// Sandbox port boundary.
package validation

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jamessawle/sbxflow/internal/domain/configuration"
	sandboxport "github.com/jamessawle/sbxflow/internal/ports/sandbox"
)

func TestValidateLocalKitsUsesProvenanceAndContinues(t *testing.T) {
	validator := &fakeKitValidator{
		outputs: []sandboxport.Output{
			{Err: errors.New("exit status 1"), Stderr: []byte("bad kit")},
			{Stdout: []byte("valid")},
		},
	}
	targets := []configuration.LocalKit{
		{Index: 1, Source: "local", Kit: "kit.zip", Path: "/tmp/kits/kit.zip"},
		{Index: 3, Source: "local", Kit: "directory", Path: "/tmp/kits/directory"},
	}
	results := validateLocalKits(context.Background(), targets, validator)
	if len(results) != 2 || results[0].Valid || !results[1].Valid || results[0].Diagnostics != "bad kit" {
		t.Fatalf("validateLocalKits() = %#v", results)
	}
	if len(validator.paths) != 2 || validator.paths[0] != targets[0].Path || validator.paths[1] != targets[1].Path {
		t.Fatalf("paths = %#v", validator.paths)
	}
}

func TestValidateLocalKitsNoTargetsDoesNotLookUpSbx(t *testing.T) {
	validator := &fakeKitValidator{err: errors.New("must not be called")}
	if results := validateLocalKits(context.Background(), nil, validator); results != nil || validator.calls != 0 {
		t.Fatalf("validateLocalKits() = %#v, calls = %d", results, validator.calls)
	}
}

func TestValidateLocalKitsReportsUnavailableAndTimeout(t *testing.T) {
	target := []configuration.LocalKit{{Source: "local", Kit: "kit", Path: "/tmp/kit"}}
	unavailable := &fakeKitValidator{err: errors.New("not found")}
	if results := validateLocalKits(context.Background(), target, unavailable); len(results) != 1 || !strings.Contains(results[0].Err.Error(), "unavailable") {
		t.Fatalf("unavailable results = %#v", results)
	}
	timedOut := &fakeKitValidator{outputs: []sandboxport.Output{{Err: context.DeadlineExceeded}}}
	if results := validateLocalKits(context.Background(), target, timedOut); len(results) != 1 || !strings.Contains(results[0].Err.Error(), "timed out") {
		t.Fatalf("timeout results = %#v", results)
	}
}

type fakeKitValidator struct {
	outputs []sandboxport.Output
	err     error
	paths   []string
	calls   int
}

func (v *fakeKitValidator) ValidateKits(_ context.Context, paths []string) ([]sandboxport.Output, error) {
	v.calls++
	v.paths = append([]string(nil), paths...)
	return v.outputs, v.err
}

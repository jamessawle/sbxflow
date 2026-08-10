package validation

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/jamessawle/sbxflow/internal/configuration"
	"github.com/jamessawle/sbxflow/internal/sbx"
)

func TestValidateLocalKitsUsesProvenanceAndContinues(t *testing.T) {
	runner := &recordingRunner{
		path: "/fake/sbx",
		outputs: []sbx.Output{
			{Err: errors.New("exit status 1"), Stderr: []byte("bad kit"), ExitCode: 1},
			{Stdout: []byte("valid")},
		},
	}
	targets := []configuration.LocalKit{
		{Index: 1, Source: "local", Kit: "kit.zip", Path: "/tmp/kits/kit.zip"},
		{Index: 3, Source: "local", Kit: "directory", Path: "/tmp/kits/directory"},
	}
	results := validateLocalKits(context.Background(), targets, sbx.Client{Commands: runner})
	if len(results) != 2 || results[0].Valid || !results[1].Valid || results[0].Diagnostics != "bad kit" {
		t.Fatalf("validateLocalKits() = %#v", results)
	}
	wantCalls := [][]string{{"kit", "validate", targets[0].Path}, {"kit", "validate", targets[1].Path}}
	if !reflect.DeepEqual(runner.calls, wantCalls) {
		t.Fatalf("calls = %#v, want %#v", runner.calls, wantCalls)
	}
}

func TestValidateLocalKitsNoTargetsDoesNotLookUpSbx(t *testing.T) {
	runner := &recordingRunner{lookupErr: errors.New("must not be called")}
	if results := validateLocalKits(context.Background(), nil, sbx.Client{Commands: runner}); results != nil || runner.lookups != 0 {
		t.Fatalf("validateLocalKits() = %#v, lookups = %d", results, runner.lookups)
	}
}

func TestValidateLocalKitsReportsUnavailableAndTimeout(t *testing.T) {
	target := []configuration.LocalKit{{Source: "local", Kit: "kit", Path: "/tmp/kit"}}
	unavailable := &recordingRunner{lookupErr: errors.New("not found")}
	if results := validateLocalKits(context.Background(), target, sbx.Client{Commands: unavailable}); len(results) != 1 || !strings.Contains(results[0].Err.Error(), "unavailable") {
		t.Fatalf("unavailable results = %#v", results)
	}
	timedOut := &recordingRunner{path: "/fake/sbx", outputs: []sbx.Output{{Err: context.DeadlineExceeded, ExitCode: -1}}}
	if results := validateLocalKits(context.Background(), target, sbx.Client{Commands: timedOut}); len(results) != 1 || !strings.Contains(results[0].Err.Error(), "timed out") {
		t.Fatalf("timeout results = %#v", results)
	}
}

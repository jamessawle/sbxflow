package sbx

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
)

type networkCall struct{ args []string }
type networkRunner struct {
	path    string
	err     error
	outputs []Output
	calls   []networkCall
}

func (r *networkRunner) LookPath(string) (string, error) { return r.path, r.err }
func (r *networkRunner) Run(_ context.Context, _ string, args ...string) Output {
	r.calls = append(r.calls, networkCall{args: append([]string(nil), args...)})
	if len(r.outputs) == 0 {
		return Output{}
	}
	output := r.outputs[0]
	r.outputs = r.outputs[1:]
	return output
}

func TestAllowNetworkJoinsOrderedResourcesAndSkipsEmptyInput(t *testing.T) {
	runner := &networkRunner{path: "/bin/sbx"}
	client := Client{Commands: runner}
	if err := client.AllowNetwork(context.Background(), NetworkAllowRequest{Name: "project"}); err != nil {
		t.Fatalf("empty AllowNetwork() error = %v", err)
	}
	if len(runner.calls) != 0 {
		t.Fatalf("empty input calls = %#v", runner.calls)
	}
	request := NetworkAllowRequest{Name: "project", Resources: []string{"first.example", "second.example:443"}}
	if err := client.AllowNetwork(context.Background(), request); err != nil {
		t.Fatalf("AllowNetwork() error = %v", err)
	}
	// Docker Sandboxes takes one comma-separated RESOURCES argument and rejects a
	// second positional as a mistaken sandbox name.
	want := []string{"policy", "allow", "network", "--sandbox", "project", "first.example,second.example:443"}
	if !reflect.DeepEqual(runner.calls[0].args, want) {
		t.Fatalf("args = %#v, want %#v", runner.calls[0].args, want)
	}
}

func TestAllowNetworkFailurePreservesDiagnostics(t *testing.T) {
	runner := &networkRunner{path: "/bin/sbx", outputs: []Output{{Stderr: []byte("resource rejected\n"), Err: errors.New("exit 7")}}}
	client := Client{Commands: runner}
	err := client.AllowNetwork(context.Background(), NetworkAllowRequest{Name: "project", Resources: []string{"bad"}})
	if err == nil || !strings.Contains(err.Error(), "resource rejected") || !strings.Contains(err.Error(), "project") {
		t.Fatalf("AllowNetwork() error = %v", err)
	}
}

// Package sbx adapts Docker Sandboxes subprocesses to sbxflow workflows.
package sbx

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	sandboxport "github.com/jamessawle/sbxflow/internal/ports/sandbox"
)

const (
	allowedSourcesEnvironment = "DOCKER_SANDBOXES_KIT_ALLOWED_SOURCES"
	allowLocalEnvironment     = "DOCKER_SANDBOXES_KIT_ALLOW_LOCAL"
)

type Streams = sandboxport.Streams

// Client invokes Docker Sandboxes through injected captured and interactive
// process runners.
type Client struct {
	Commands    Runner
	Interactive InteractiveRunner
}

var (
	_ sandboxport.KitValidator     = Client{}
	_ sandboxport.Inspector        = Client{}
	_ sandboxport.ExistenceChecker = Client{}
	_ sandboxport.StateInspector   = Client{}
	_ sandboxport.Creator          = Client{}
	_ sandboxport.Runner           = Client{}
	_ sandboxport.Stopper          = Client{}
	_ sandboxport.Remover          = Client{}
	_ sandboxport.NetworkPolicy    = Client{}
)

// NewClient constructs a production client with the supplied timeout for
// captured inspection commands. Interactive commands have no client timeout.
func NewClient(timeout time.Duration) Client {
	return Client{
		Commands:    ExecRunner{Timeout: timeout},
		Interactive: InteractiveExecRunner{},
	}
}

type CreateRequest = sandboxport.CreateRequest
type RunRequest = sandboxport.RunRequest
type RemoveRequest = sandboxport.RemoveRequest
type NetworkAllowRequest = sandboxport.NetworkAllowRequest
type NetworkRemoveRequest = sandboxport.NetworkRemoveRequest

// absentPolicyDiagnostics are the diagnostics Docker Sandboxes emits when there
// is nothing left to remove, either because the sandbox has no scoped policy or
// because the resource is not part of it.
var absentPolicyDiagnostics = []string{"no scoped policy found", "rule not found"}

// AllowNetwork adds ordered resources to a local rule scoped to one sandbox.
// Docker Sandboxes takes the resources as one comma-separated argument and
// requires the sandbox to already exist.
func (c Client) AllowNetwork(ctx context.Context, request NetworkAllowRequest) error {
	if len(request.Resources) == 0 {
		return nil
	}
	executable, err := c.Commands.LookPath("sbx")
	if err != nil {
		return fmt.Errorf("locate sbx for sandbox network allow: %w", err)
	}
	output := c.Commands.Run(ctx, executable, "policy", "allow", "network", "--sandbox", request.Name, strings.Join(request.Resources, ","))
	return capturedError(output, "allow network resources for Docker Sandbox %q", request.Name)
}

// RemoveNetworkResource removes one resource from a sandbox-scoped local rule.
// Removing a resource that is already absent succeeds, so cleanup stays
// idempotent when Docker Sandboxes has already discarded the scoped policy
// along with its sandbox.
func (c Client) RemoveNetworkResource(ctx context.Context, request NetworkRemoveRequest) error {
	executable, err := c.Commands.LookPath("sbx")
	if err != nil {
		return fmt.Errorf("locate sbx for sandbox network cleanup: %w", err)
	}
	output := c.Commands.Run(ctx, executable, "policy", "rm", "network", "--sandbox", request.Name, "--resource", request.Resource)
	if output.Err != nil && reportsAbsentPolicy(output) {
		return nil
	}
	return capturedError(output, "remove network resource %q for Docker Sandbox %q", request.Resource, request.Name)
}

// reportsAbsentPolicy reports whether command output indicates that the network policy or resource is absent.
func reportsAbsentPolicy(output sandboxport.Output) bool {
	diagnostics := strings.ToLower(string(output.Stderr) + string(output.Stdout))
	for _, absent := range absentPolicyDiagnostics {
		if strings.Contains(diagnostics, absent) {
			return true
		}
	}
	return false
}

// capturedError formats a command error with the operation context and available diagnostics.
func capturedError(output sandboxport.Output, operation string, values ...any) error {
	if output.Err == nil {
		return nil
	}
	prefix := fmt.Sprintf(operation, values...)
	diagnostics := strings.TrimSpace(strings.Join(nonempty(string(output.Stderr), string(output.Stdout)), ": "))
	if diagnostics != "" {
		return fmt.Errorf("%s: %s: %w", prefix, diagnostics, output.Err)
	}
	return fmt.Errorf("%s: %w", prefix, output.Err)
}

// ValidateKits validates local kit paths sequentially.
func (c Client) ValidateKits(ctx context.Context, paths []string) ([]sandboxport.Output, error) {
	if len(paths) == 0 {
		return nil, nil
	}
	executable, err := c.Commands.LookPath("sbx")
	if err != nil {
		return nil, err
	}
	outputs := make([]sandboxport.Output, 0, len(paths))
	for _, path := range paths {
		outputs = append(outputs, c.Commands.Run(ctx, executable, "kit", "validate", path))
	}
	return outputs, nil
}

// Exists reports whether the exact sandbox name exists.
func (c Client) Exists(ctx context.Context, name string) (bool, error) {
	executable, err := c.Commands.LookPath("sbx")
	if err != nil {
		return false, fmt.Errorf("locate sbx for sandbox lookup: %w", err)
	}
	output := c.Commands.Run(ctx, executable, "ls", "--quiet")
	if output.Err != nil {
		diagnostics := strings.TrimSpace(strings.Join(nonempty(string(output.Stderr), string(output.Stdout)), ": "))
		if diagnostics != "" {
			return false, fmt.Errorf("list Docker Sandboxes with `sbx ls --quiet`: %s: %w", diagnostics, output.Err)
		}
		return false, fmt.Errorf("list Docker Sandboxes with `sbx ls --quiet`: %w", output.Err)
	}
	for _, existing := range strings.Split(string(output.Stdout), "\n") {
		if strings.TrimSpace(existing) != "" && existing == name {
			return true, nil
		}
	}
	return false, nil
}

// Inspect reports the normalized lifecycle state of an exact sandbox name from
// Docker's machine-readable listing.
func (c Client) Inspect(ctx context.Context, name string) (sandboxport.State, error) {
	executable, err := c.Commands.LookPath("sbx")
	if err != nil {
		return "", fmt.Errorf("locate sbx for sandbox state inspection: %w", err)
	}
	output := c.Commands.Run(ctx, executable, "ls", "--json")
	if output.Err != nil {
		diagnostics := strings.TrimSpace(strings.Join(nonempty(string(output.Stderr), string(output.Stdout)), ": "))
		if diagnostics != "" {
			return "", fmt.Errorf("inspect Docker Sandboxes with `sbx ls --json`: %s: %w", diagnostics, output.Err)
		}
		return "", fmt.Errorf("inspect Docker Sandboxes with `sbx ls --json`: %w", output.Err)
	}
	listing, err := decodeSandboxListing(output.Stdout)
	if err != nil {
		return "", fmt.Errorf("decode Docker Sandbox state from `sbx ls --json`: %w", err)
	}
	if listing.Sandboxes == nil {
		return "", errors.New("decode Docker Sandbox state from `sbx ls --json`: missing sandboxes collection")
	}
	state := sandboxport.StateAbsent
	matches := 0
	for _, candidate := range *listing.Sandboxes {
		if candidate.Name != name {
			continue
		}
		matches++
		switch strings.ToLower(candidate.Status) {
		case string(sandboxport.StateRunning):
			state = sandboxport.StateRunning
		case string(sandboxport.StateStopped):
			state = sandboxport.StateStopped
		default:
			return "", fmt.Errorf("inspect Docker Sandbox %q: unrecognized lifecycle state %q", name, candidate.Status)
		}
	}
	if matches > 1 {
		return "", fmt.Errorf("inspect Docker Sandbox %q: machine-readable listing contained %d exact-name matches", name, matches)
	}
	return state, nil
}

// CreateSandbox provisions a sandbox without attaching to its agent, so a
// caller can scope policy to the new sandbox before entering it.
func (c Client) CreateSandbox(ctx context.Context, request CreateRequest, streams Streams) error {
	executable, err := c.Commands.LookPath("sbx")
	if err != nil {
		return fmt.Errorf("locate sbx for sandbox creation: %w", err)
	}
	trust, err := kitTrustEnvironment(request.AllowedSources, request.AllowLocalKits)
	if err != nil {
		return err
	}

	args := []string{"create", "--name", request.Name}
	for _, kit := range request.Kits {
		args = append(args, "--kit", kit)
	}
	args = append(args, request.Agent, request.Workspace)

	return c.Interactive.Run(ctx, InteractiveInvocation{
		Executable:  executable,
		Args:        args,
		Environment: trust,
		Stdin:       streams.In,
		Stdout:      streams.Out,
		Stderr:      streams.Err,
	})
}

// RunSandbox enters an existing sandbox and remains attached until sbx exits.
func (c Client) RunSandbox(ctx context.Context, request RunRequest, streams Streams) error {
	executable, err := c.Commands.LookPath("sbx")
	if err != nil {
		return fmt.Errorf("locate sbx for sandbox execution: %w", err)
	}
	trust, err := kitTrustEnvironment(request.AllowedSources, request.AllowLocalKits)
	if err != nil {
		return err
	}

	return c.Interactive.Run(ctx, InteractiveInvocation{
		Executable:  executable,
		Args:        []string{"run", request.Agent, "--name", request.Name},
		Environment: trust,
		Stdin:       streams.In,
		Stdout:      streams.Out,
		Stderr:      streams.Err,
	})
}

// kitTrustEnvironment builds environment variables that configure trusted Docker kit sources and local-kit access.
// It returns an error if the allowed sources cannot be encoded.
func kitTrustEnvironment(allowedSources []string, allowLocalKits bool) (map[string]string, error) {
	encoded, err := json.Marshal(allowedSources)
	if err != nil {
		return nil, fmt.Errorf("encode Docker kit source trust: %w", err)
	}
	return map[string]string{
		allowedSourcesEnvironment: string(encoded),
		allowLocalEnvironment:     strconv.FormatBool(allowLocalKits),
	}, nil
}

// StopSandbox stops a sandbox with its output attached to the supplied streams.
func (c Client) StopSandbox(ctx context.Context, name string, streams Streams) error {
	executable, err := c.Commands.LookPath("sbx")
	if err != nil {
		return fmt.Errorf("locate sbx for sandbox stop: %w", err)
	}
	return c.Interactive.Run(ctx, InteractiveInvocation{
		Executable: executable,
		Args:       []string{"stop", name},
		Stdout:     streams.Out,
		Stderr:     streams.Err,
	})
}

// RemoveSandbox removes a sandbox with Docker's confirmation safeguards unless
// force is requested.
func (c Client) RemoveSandbox(ctx context.Context, request RemoveRequest) error {
	executable, err := c.Commands.LookPath("sbx")
	if err != nil {
		return fmt.Errorf("locate sbx for sandbox removal: %w", err)
	}
	args := []string{"rm", request.Name}
	if request.Force {
		args = []string{"rm", "--force", request.Name}
	}
	return c.Interactive.Run(ctx, InteractiveInvocation{
		Executable: executable,
		Args:       args,
		Stdin:      request.Streams.In,
		Stdout:     request.Streams.Out,
		Stderr:     request.Streams.Err,
	})
}

func nonempty(values ...string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			result = append(result, strings.TrimSpace(value))
		}
	}
	return result
}

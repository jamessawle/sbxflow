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
	_ sandboxport.KitValidator = Client{}
	_ sandboxport.Inspector    = Client{}
	_ sandboxport.Lookup       = Client{}
	_ sandboxport.StateLookup  = Client{}
	_ sandboxport.Runner       = Client{}
	_ sandboxport.Stopper      = Client{}
	_ sandboxport.Remover      = Client{}
)

// NewClient constructs a production client with the supplied timeout for
// captured inspection commands. Interactive commands have no client timeout.
func NewClient(timeout time.Duration) Client {
	return Client{
		Commands:    ExecRunner{Timeout: timeout},
		Interactive: InteractiveExecRunner{},
	}
}

type RunRequest = sandboxport.RunRequest
type RemoveRequest = sandboxport.RemoveRequest

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

// SandboxExists reports whether the exact sandbox name exists.
func (c Client) SandboxExists(ctx context.Context, name string) (bool, error) {
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

// InspectSandbox reports the normalized lifecycle state of an exact sandbox
// name from Docker's machine-readable listing.
func (c Client) InspectSandbox(ctx context.Context, name string) (sandboxport.State, error) {
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
	var listing struct {
		Sandboxes *[]struct {
			Name   string `json:"name"`
			Status string `json:"status"`
		} `json:"sandboxes"`
	}
	if err := json.Unmarshal(output.Stdout, &listing); err != nil {
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

// RunSandbox creates or enters a sandbox and remains attached until sbx exits.
func (c Client) RunSandbox(ctx context.Context, request RunRequest, streams Streams) error {
	executable, err := c.Commands.LookPath("sbx")
	if err != nil {
		return fmt.Errorf("locate sbx for sandbox execution: %w", err)
	}
	allowedSources, err := json.Marshal(request.AllowedSources)
	if err != nil {
		return fmt.Errorf("encode Docker kit source trust: %w", err)
	}

	args := []string{"run", request.Agent, "--name", request.Name}
	if !request.Exists {
		args = []string{"run", "--name", request.Name}
		for _, kit := range request.Kits {
			args = append(args, "--kit", kit)
		}
		args = append(args, request.Agent, request.Workspace)
	}

	return c.Interactive.Run(ctx, InteractiveInvocation{
		Executable: executable,
		Args:       args,
		Environment: map[string]string{
			allowedSourcesEnvironment: string(allowedSources),
			allowLocalEnvironment:     strconv.FormatBool(request.AllowLocalKits),
		},
		Stdin:  streams.In,
		Stdout: streams.Out,
		Stderr: streams.Err,
	})
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

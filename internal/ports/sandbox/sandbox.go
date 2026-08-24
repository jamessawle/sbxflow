// Package sandbox defines narrow ports for Docker Sandboxes capabilities.
package sandbox

import (
	"context"
	"io"
)

// Output preserves the observable result of one captured invocation.
type Output struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int
	Err      error
}

// Streams are attached directly to an interactive sandbox process.
type Streams struct {
	In  io.Reader
	Out io.Writer
	Err io.Writer
}

// CreateRequest contains the inputs needed to create a sandbox without
// attaching to its agent session.
type CreateRequest struct {
	Name           string
	Agent          string
	Workspace      string
	Kits           []string
	AllowedSources []string
	AllowLocalKits bool
}

// RunRequest contains the inputs needed to enter an existing sandbox.
type RunRequest struct {
	Name           string
	Agent          string
	AllowedSources []string
	AllowLocalKits bool
}

// CommandRequest executes one literal command vector inside a sandbox workspace.
type CommandRequest struct {
	Name      string
	Workspace string
	Command   []string
}

// RemoveRequest contains the safeguards and streams for sandbox removal.
type RemoveRequest struct {
	Name    string
	Force   bool
	Streams Streams
}

// NetworkAllowRequest adds ordered resources to one sandbox-scoped local rule.
type NetworkAllowRequest struct {
	Name      string
	Resources []string
}

// NetworkRemoveRequest removes one resource from a sandbox-scoped local rule.
type NetworkRemoveRequest struct {
	Name     string
	Resource string
}

// State is the normalized lifecycle state of an exact sandbox name.
type State string

const (
	StateAbsent  State = "absent"
	StateStopped State = "stopped"
	StateRunning State = "running"
)

// KitValidator validates canonical local-kit paths.
type KitValidator interface {
	ValidateKits(ctx context.Context, paths []string) ([]Output, error)
}

// Inspector exposes read-only system diagnostic operations.
type Inspector interface {
	Locate() (string, error)
	Version(context.Context, string) Output
	Diagnose(context.Context, string) Output
	ListPolicies(context.Context, string) Output
	GetKitAllowedSources(context.Context, string) Output
}

// ExistenceChecker checks whether an exact sandbox exists.
type ExistenceChecker interface {
	Exists(ctx context.Context, name string) (bool, error)
}

// StateInspector inspects the lifecycle state of an exact sandbox name.
type StateInspector interface {
	Inspect(ctx context.Context, name string) (State, error)
}

// Creator creates a sandbox without attaching to its agent session, so that
// sandbox-scoped policy can be applied before the agent starts.
type Creator interface {
	CreateSandbox(ctx context.Context, request CreateRequest, streams Streams) error
}

// Runner enters an existing sandbox.
type Runner interface {
	RunSandbox(ctx context.Context, request RunRequest, streams Streams) error
}

// CommandExecutor executes a non-interactive command with attached output.
type CommandExecutor interface {
	ExecuteCommand(ctx context.Context, request CommandRequest, streams Streams) error
}

// Stopper stops a sandbox without removing it.
type Stopper interface {
	StopSandbox(ctx context.Context, name string, streams Streams) error
}

// Remover permanently removes a sandbox.
type Remover interface {
	RemoveSandbox(ctx context.Context, request RemoveRequest) error
}

// NetworkPolicy manages sandbox-scoped local network allow resources. The
// sandbox must already exist before resources can be scoped to it.
// RemoveNetworkResource is idempotent: a resource that is already absent, or a
// sandbox that has no scoped policy at all, is not an error.
type NetworkPolicy interface {
	AllowNetwork(context.Context, NetworkAllowRequest) error
	RemoveNetworkResource(context.Context, NetworkRemoveRequest) error
}

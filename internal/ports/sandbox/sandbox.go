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

// RunRequest contains the inputs needed to create or enter a sandbox.
type RunRequest struct {
	Name           string
	Agent          string
	Workspace      string
	Kits           []string
	AllowedSources []string
	AllowLocalKits bool
	Exists         bool
}

// RemoveRequest contains the safeguards and streams for sandbox removal.
type RemoveRequest struct {
	Name    string
	Force   bool
	Streams Streams
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

// Lookup checks whether an exact sandbox exists.
type Lookup interface {
	SandboxExists(ctx context.Context, name string) (bool, error)
}

// StateLookup inspects the lifecycle state of an exact sandbox name.
type StateLookup interface {
	InspectSandbox(ctx context.Context, name string) (State, error)
}

// Runner creates or enters a sandbox.
type Runner interface {
	RunSandbox(ctx context.Context, request RunRequest, streams Streams) error
}

// Stopper stops a sandbox without removing it.
type Stopper interface {
	StopSandbox(ctx context.Context, name string, streams Streams) error
}

// Remover permanently removes a sandbox.
type Remover interface {
	RemoveSandbox(ctx context.Context, request RemoveRequest) error
}

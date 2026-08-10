package lifecycle

import (
	"fmt"
	"os"

	"github.com/jamessawle/sbxflow/internal/config"
)

// Target identifies the repository declaration and exact sandbox name needed
// by teardown lifecycle operations.
type Target struct {
	Declaration string
	Name        string
}

// TargetResolver resolves the nearest repository lifecycle target.
type TargetResolver interface {
	Resolve(start string) (Target, error)
}

// RepositoryTargetResolver discovers, reads, and identity-loads a repository
// declaration through injectable filesystem boundaries.
type RepositoryTargetResolver struct {
	Discover func(string) (string, error)
	ReadFile func(string) ([]byte, error)
}

// NewDefaultTargetResolver constructs the production repository target
// resolver.
func NewDefaultTargetResolver() RepositoryTargetResolver {
	return RepositoryTargetResolver{
		Discover: config.Discover,
		ReadFile: os.ReadFile,
	}
}

// Resolve returns the nearest safely interpreted lifecycle target.
func (r RepositoryTargetResolver) Resolve(start string) (Target, error) {
	declaration, err := r.Discover(start)
	if err != nil {
		return Target{}, err
	}
	data, err := r.ReadFile(declaration)
	if err != nil {
		return Target{}, fmt.Errorf("read repository declaration %q: %w", declaration, err)
	}
	target, err := config.LoadLifecycleTarget(data)
	if err != nil {
		return Target{}, fmt.Errorf("resolve lifecycle target from %q: %w", declaration, err)
	}
	return Target{Declaration: declaration, Name: target.Name}, nil
}

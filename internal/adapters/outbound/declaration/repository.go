package declaration

import (
	"fmt"
	"os"

	declarationport "github.com/jamessawle/sbxflow/internal/ports/declaration"
)

// Repository loads declarations from the local filesystem.
type Repository struct {
	Discover func(string) (string, error)
	ReadFile func(string) ([]byte, error)
}

var (
	_ declarationport.Loader           = Repository{}
	_ declarationport.LocalKitResolver = Repository{}
	_ declarationport.TargetResolver   = Repository{}
)

// NewRepository constructs the production declaration adapter.
func NewRepository() Repository {
	return Repository{Discover: Discover, ReadFile: os.ReadFile}
}

// Load discovers, reads, and decodes a declaration.
func (r Repository) Load(start string) (declarationport.Loaded, error) {
	path, err := r.Discover(start)
	if err != nil {
		return declarationport.Loaded{}, err
	}
	data, err := r.ReadFile(path)
	if err != nil {
		return declarationport.Loaded{Path: path}, fmt.Errorf("read repository declaration %q: %w", path, err)
	}
	document, err := Load(data)
	if err != nil {
		return declarationport.Loaded{Path: path}, err
	}
	return declarationport.Loaded{Path: path, Configuration: document}, nil
}

// Resolve returns the nearest declaration identity needed by teardown.
func (r Repository) Resolve(start string) (declarationport.LifecycleTarget, error) {
	path, err := r.Discover(start)
	if err != nil {
		return declarationport.LifecycleTarget{}, err
	}
	data, err := r.ReadFile(path)
	if err != nil {
		return declarationport.LifecycleTarget{}, fmt.Errorf("read repository declaration %q: %w", path, err)
	}
	target, err := LoadLifecycleTarget(data)
	if err != nil {
		return declarationport.LifecycleTarget{}, fmt.Errorf("resolve lifecycle target from %q: %w", path, err)
	}
	target.Declaration = path
	return target, nil
}

// ResolveLocalKits safely resolves local kit requests through the filesystem.
func (Repository) ResolveLocalKits(declarationPath string, requests []declarationport.LocalKitRequest) ([]declarationport.LocalKit, error) {
	return ResolveLocalKits(declarationPath, requests)
}

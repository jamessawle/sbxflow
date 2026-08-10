package lifecycle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamessawle/sbxflow/internal/config"
)

func TestRepositoryTargetResolverSelectsNearestAncestor(t *testing.T) {
	root := t.TempDir()
	writeTargetDeclaration(t, root, "outer")
	nested := filepath.Join(root, "nested", "work")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	near := filepath.Join(root, "nested")
	writeTargetDeclaration(t, near, "exact-inner")

	target, err := NewDefaultTargetResolver().Resolve(nested)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	wantDeclaration := filepath.Join(near, config.Filename)
	if target.Declaration != wantDeclaration || target.Name != "exact-inner" {
		t.Fatalf("Resolve() = %#v, want declaration %q and exact name", target, wantDeclaration)
	}
}

func TestRepositoryTargetResolverReportsDiscoveryFailure(t *testing.T) {
	_, err := NewDefaultTargetResolver().Resolve(t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "no sbxflow.yaml found") {
		t.Fatalf("Resolve() error = %v, want absent declaration", err)
	}
}

func TestRepositoryTargetResolverReportsReadFailure(t *testing.T) {
	want := errors.New("permission denied")
	resolver := RepositoryTargetResolver{
		Discover: func(string) (string, error) { return "/repo/sbxflow.yaml", nil },
		ReadFile: func(string) ([]byte, error) { return nil, want },
	}
	_, err := resolver.Resolve("/repo")
	if !errors.Is(err, want) || !strings.Contains(err.Error(), "/repo/sbxflow.yaml") {
		t.Fatalf("Resolve() error = %v, want contextual read failure", err)
	}
}

func TestRepositoryTargetResolverReportsIdentityFailure(t *testing.T) {
	resolver := RepositoryTargetResolver{
		Discover: func(string) (string, error) { return "/repo/sbxflow.yaml", nil },
		ReadFile: func(string) ([]byte, error) {
			return []byte("version: 2\nsandbox:\n  name: project\n"), nil
		},
	}
	_, err := resolver.Resolve("/repo")
	if err == nil || !strings.Contains(err.Error(), "unsupported configuration version") || !strings.Contains(err.Error(), "/repo/sbxflow.yaml") {
		t.Fatalf("Resolve() error = %v, want contextual identity failure", err)
	}
}

func writeTargetDeclaration(t *testing.T, directory, name string) {
	t.Helper()
	contents := "version: 1\nsandbox:\n  name: " + name + "\n"
	if err := os.WriteFile(filepath.Join(directory, config.Filename), []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
}

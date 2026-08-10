package configuration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveLocalTargets(t *testing.T) {
	repository := t.TempDir()
	root := filepath.Join(repository, "kits")
	directory := filepath.Join(root, "directory")
	zip := filepath.Join(root, "packaged.zip")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(zip, []byte("fixture"), 0o600); err != nil {
		t.Fatal(err)
	}
	linked := linkedLocal(t, "./kits", "directory", "packaged.zip")

	targets, err := ResolveLocalKits(filepath.Join(repository, Filename), linked)
	if err != nil {
		t.Fatalf("ResolveLocalKits() error = %v", err)
	}
	want := []string{directory, zip}
	for index, target := range targets {
		canonicalWant, err := filepath.EvalSymlinks(want[index])
		if err != nil {
			t.Fatal(err)
		}
		if target.Path != canonicalWant {
			t.Errorf("target %d path = %q, want %q", index, target.Path, canonicalWant)
		}
	}
}

func TestResolveRejectsUnsafeLocalReferences(t *testing.T) {
	repository := t.TempDir()
	root := filepath.Join(repository, "kits")
	outside := filepath.Join(repository, "outside")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(outside, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "escape")); err != nil {
		t.Fatal(err)
	}
	declaration := filepath.Join(repository, Filename)

	tests := map[string]struct {
		root string
		kit  string
		want string
	}{
		"URI root":           {root: "https://example.com/kits", kit: "one", want: "host filesystem path"},
		"unavailable root":   {root: "./missing", kit: "one", want: "unavailable"},
		"absolute kit":       {root: "./kits", kit: outside, want: "relative host filesystem path"},
		"URI kit":            {root: "./kits", kit: "oci://example/kit", want: "relative host filesystem path"},
		"traversal":          {root: "./kits", kit: "../outside", want: "escapes"},
		"symbolic link":      {root: "./kits", kit: "escape", want: "escapes"},
		"unavailable target": {root: "./kits", kit: "missing", want: "unavailable"},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := ResolveLocalKits(declaration, linkedLocal(t, test.root, test.kit))
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("ResolveLocalKits() error = %v, want %q", err, test.want)
			}
		})
	}
}

func linkedLocal(t *testing.T, root string, kits ...string) LinkedConfiguration {
	t.Helper()
	document := Configuration{Version: 1, Sandbox: Sandbox{Name: "demo", Agent: "codex", Kits: Kits{
		Sources: map[string]Source{"local": {Type: SourceLocal, Root: root}},
	}}}
	for _, kit := range kits {
		document.Sandbox.Kits.Use = append(document.Sandbox.Kits.Use, Selection{Source: "local", Kit: kit})
	}
	linked, err := Link(document)
	if err != nil {
		t.Fatalf("Link() error = %v", err)
	}
	return linked
}

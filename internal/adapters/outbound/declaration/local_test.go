package declaration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	declarationport "github.com/jamessawle/sbxflow/internal/ports/declaration"
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
	requests := localRequests("./kits", "directory", "packaged.zip")

	targets, err := ResolveLocalKits(filepath.Join(repository, Filename), requests)
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
			_, err := ResolveLocalKits(declaration, localRequests(test.root, test.kit))
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("ResolveLocalKits() error = %v, want %q", err, test.want)
			}
		})
	}
}

func localRequests(root string, kits ...string) []declarationport.LocalKitRequest {
	requests := make([]declarationport.LocalKitRequest, 0, len(kits))
	for index, kit := range kits {
		requests = append(requests, declarationport.LocalKitRequest{Index: index, Source: "local", Root: root, Kit: kit})
	}
	return requests
}

package declaration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiscover(t *testing.T) {
	root := t.TempDir()
	outer := filepath.Join(root, Filename)
	writeTestFile(t, outer, "version: 1\n")
	nested := filepath.Join(root, "one", "two")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("create nested directory: %v", err)
	}

	t.Run("direct", func(t *testing.T) {
		path, err := Discover(root)
		if err != nil || path != outer {
			t.Fatalf("Discover() = %q, %v; want %q", path, err, outer)
		}
	})

	t.Run("nested", func(t *testing.T) {
		path, err := Discover(nested)
		if err != nil || path != outer {
			t.Fatalf("Discover() = %q, %v; want %q", path, err, outer)
		}
	})

	t.Run("nearest", func(t *testing.T) {
		near := filepath.Join(root, "one", Filename)
		writeTestFile(t, near, "version: 1\n")
		path, err := Discover(nested)
		if err != nil || path != near {
			t.Fatalf("Discover() = %q, %v; want %q", path, err, near)
		}
	})

	t.Run("file start", func(t *testing.T) {
		path, err := Discover(filepath.Join(nested, "missing.txt"))
		if err == nil || path != "" {
			t.Fatalf("Discover() = %q, %v; want starting path error", path, err)
		}
	})
}

func TestDiscoverAbsent(t *testing.T) {
	_, err := Discover(t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "no sbxflow.yaml found") {
		t.Fatalf("Discover() error = %v, want missing declaration", err)
	}
}

func writeTestFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

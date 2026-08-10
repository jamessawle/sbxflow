package configuration

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const Filename = "sbxflow.yaml"

// Discover returns the nearest sbxflow.yaml at or above start.
func Discover(start string) (string, error) {
	directory, err := filepath.Abs(start)
	if err != nil {
		return "", fmt.Errorf("resolve starting directory: %w", err)
	}
	info, err := os.Stat(directory)
	if err != nil {
		return "", fmt.Errorf("inspect starting directory %q: %w", directory, err)
	}
	if !info.IsDir() {
		directory = filepath.Dir(directory)
	}

	for {
		candidate := filepath.Join(directory, Filename)
		info, err := os.Stat(candidate)
		switch {
		case err == nil && !info.IsDir():
			return candidate, nil
		case err == nil:
			return "", fmt.Errorf("repository declaration %q is not a file", candidate)
		case !errors.Is(err, os.ErrNotExist):
			return "", fmt.Errorf("inspect repository declaration %q: %w", candidate, err)
		}

		parent := filepath.Dir(directory)
		if parent == directory {
			return "", fmt.Errorf("no %s found in %q or its ancestors", Filename, start)
		}
		directory = parent
	}
}

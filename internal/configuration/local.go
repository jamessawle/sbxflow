package configuration

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// LocalKit identifies one safely resolved local kit selection.
type LocalKit struct {
	Index  int
	Source string
	Kit    string
	Path   string
}

// ResolveLocalKits checks selected local filesystem references and returns
// their canonical targets in declaration order.
func ResolveLocalKits(declarationPath string, linked LinkedConfiguration) ([]LocalKit, error) {
	declarationDirectory := filepath.Dir(declarationPath)
	canonicalRoots := make(map[string]string)
	targets := make([]LocalKit, 0)

	for _, selection := range linked.Selections {
		if selection.Source.Type != SourceLocal {
			continue
		}

		root, ok := canonicalRoots[selection.Selection.Source]
		if !ok {
			var err error
			root, err = resolveLocalRoot(declarationDirectory, selection.Selection.Source, selection.Source.Root)
			if err != nil {
				return nil, err
			}
			canonicalRoots[selection.Selection.Source] = root
		}

		kit := selection.Selection.Kit
		if filepath.IsAbs(kit) || filepath.VolumeName(kit) != "" || isURIReference(kit) {
			return nil, fmt.Errorf("sandbox.kits.use[%d].kit for local source %q must be a relative host filesystem path", selection.Index, selection.Selection.Source)
		}
		candidate := filepath.Join(root, filepath.FromSlash(kit))
		canonicalTarget, err := filepath.EvalSymlinks(candidate)
		if err != nil {
			return nil, fmt.Errorf("sandbox.kits.use[%d].kit %q is unavailable: %w", selection.Index, kit, err)
		}
		canonicalTarget, err = filepath.Abs(canonicalTarget)
		if err != nil {
			return nil, fmt.Errorf("resolve sandbox.kits.use[%d].kit %q: %w", selection.Index, kit, err)
		}
		if !containedBy(root, canonicalTarget) {
			return nil, fmt.Errorf("sandbox.kits.use[%d].kit %q escapes local source root %q", selection.Index, kit, selection.Source.Root)
		}
		targets = append(targets, LocalKit{Index: selection.Index, Source: selection.Selection.Source, Kit: kit, Path: canonicalTarget})
	}

	return targets, nil
}

func resolveLocalRoot(declarationDirectory, sourceName, value string) (string, error) {
	if isURIReference(value) {
		return "", fmt.Errorf("sandbox.kits.sources.%s.root must be a host filesystem path", sourceName)
	}
	root := filepath.FromSlash(value)
	if !filepath.IsAbs(root) {
		root = filepath.Join(declarationDirectory, root)
	}
	canonicalRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", fmt.Errorf("sandbox.kits.sources.%s.root %q is unavailable: %w", sourceName, value, err)
	}
	canonicalRoot, err = filepath.Abs(canonicalRoot)
	if err != nil {
		return "", fmt.Errorf("resolve sandbox.kits.sources.%s.root %q: %w", sourceName, value, err)
	}
	info, err := os.Stat(canonicalRoot)
	if err != nil || !info.IsDir() {
		return "", fmt.Errorf("sandbox.kits.sources.%s.root %q is unavailable or is not a directory", sourceName, value)
	}
	return canonicalRoot, nil
}

func isURIReference(value string) bool {
	if filepath.IsAbs(value) || filepath.VolumeName(value) != "" {
		return false
	}
	if strings.Contains(value, "://") || strings.HasPrefix(value, "git@") {
		return true
	}
	parsed, err := url.Parse(value)
	return err == nil && parsed.Scheme != ""
}

func containedBy(root, target string) bool {
	relative, err := filepath.Rel(root, target)
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}

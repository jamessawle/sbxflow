// Package localkit safely resolves and validates kits selected from local
// sources.
package localkit

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamessawle/sbxflow/internal/command"
	"github.com/jamessawle/sbxflow/internal/config"
)

// Target identifies one safely resolved local kit selection.
type Target struct {
	Index  int
	Source string
	Kit    string
	Path   string
}

// Result reports Docker's validation outcome for one local target.
type Result struct {
	Target      Target
	Valid       bool
	Diagnostics string
	Err         error
}

// Resolve checks all selected local filesystem references and returns their
// canonical targets in declaration order. It performs no subprocess calls.
func Resolve(declarationPath string, linked config.LinkedConfiguration) ([]Target, error) {
	declarationDirectory := filepath.Dir(declarationPath)
	canonicalRoots := make(map[string]string)
	targets := make([]Target, 0)

	for _, selection := range linked.Selections {
		if selection.Source.Type != config.SourceLocal {
			continue
		}

		root, ok := canonicalRoots[selection.Selection.Source]
		if !ok {
			var err error
			root, err = resolveRoot(declarationDirectory, selection.Selection.Source, selection.Source.Root)
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
		targets = append(targets, Target{Index: selection.Index, Source: selection.Selection.Source, Kit: kit, Path: canonicalTarget})
	}

	return targets, nil
}

func resolveRoot(declarationDirectory, sourceName, value string) (string, error) {
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

// Validate invokes Docker's local kit validator sequentially. Independent
// failures do not prevent later targets from being checked.
func Validate(ctx context.Context, targets []Target, runner command.Runner) []Result {
	if len(targets) == 0 {
		return nil
	}
	executable, err := runner.LookPath("sbx")
	if err != nil {
		return []Result{{Err: fmt.Errorf("local kit validation could not be completed: sbx is unavailable: %w", err)}}
	}

	results := make([]Result, 0, len(targets))
	for _, target := range targets {
		output := runner.Run(ctx, executable, "kit", "validate", target.Path)
		diagnostics := strings.TrimSpace(strings.Join(nonempty(string(output.Stderr), string(output.Stdout)), "\n"))
		result := Result{Target: target, Valid: output.Err == nil, Diagnostics: diagnostics}
		if output.Err != nil {
			if errors.Is(output.Err, context.DeadlineExceeded) {
				result.Err = fmt.Errorf("local kit %s/%s validation timed out: %w", target.Source, target.Kit, output.Err)
			} else {
				result.Err = fmt.Errorf("local kit %s/%s was rejected by sbx: %w", target.Source, target.Kit, output.Err)
			}
		}
		results = append(results, result)
	}
	return results
}

func nonempty(values ...string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			result = append(result, strings.TrimSpace(value))
		}
	}
	return result
}

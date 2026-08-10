// Package lifecycle plans and runs repository-aware Docker Sandbox lifecycle
// operations.
package lifecycle

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/jamessawle/sbxflow/internal/config"
	"github.com/jamessawle/sbxflow/internal/validation"
)

const (
	allowedSourcesEnvironment = "DOCKER_SANDBOXES_KIT_ALLOWED_SOURCES"
	allowLocalEnvironment     = "DOCKER_SANDBOXES_KIT_ALLOW_LOCAL"
)

// Plan contains the validated inputs needed to create or enter a sandbox.
type Plan struct {
	Name        string
	Agent       string
	Workspace   string
	Kits        []string
	Environment map[string]string
}

// NewPlan converts a successful validation report into ordered Docker
// execution inputs without resolving any source a second time.
func NewPlan(report validation.Report) (Plan, error) {
	if !report.Valid() {
		return Plan{}, fmt.Errorf("cannot plan lifecycle from an invalid declaration")
	}

	localPaths := make(map[int]string, len(report.LocalKits))
	for _, result := range report.LocalKits {
		if !result.Valid || result.Target.Path == "" {
			return Plan{}, fmt.Errorf("local kit selection %d has no validated path", result.Target.Index)
		}
		localPaths[result.Target.Index] = result.Target.Path
	}

	kits := make([]string, 0, len(report.Linked.Selections))
	for _, selection := range report.Linked.Selections {
		if selection.Source.Type == config.SourceLocal {
			localPath, ok := localPaths[selection.Index]
			if !ok {
				return Plan{}, fmt.Errorf("local kit selection %d has no validation result", selection.Index)
			}
			kits = append(kits, localPath)
			continue
		}
		if selection.RemoteReference == "" {
			return Plan{}, fmt.Errorf("remote kit selection %d has no execution reference", selection.Index)
		}
		kits = append(kits, selection.RemoteReference)
	}

	allowedSources, err := json.Marshal(report.Linked.Trust.AllowedSources)
	if err != nil {
		return Plan{}, fmt.Errorf("encode Docker kit source trust: %w", err)
	}

	return Plan{
		Name:      report.Linked.Configuration.Sandbox.Name,
		Agent:     report.Linked.Configuration.Sandbox.Agent,
		Workspace: filepath.Dir(report.Declaration),
		Kits:      kits,
		Environment: map[string]string{
			allowedSourcesEnvironment: string(allowedSources),
			allowLocalEnvironment:     strconv.FormatBool(report.Linked.Trust.AllowLocalKits),
		},
	}, nil
}

// Package lifecycle plans and runs repository-aware Docker Sandbox lifecycle
// operations.
package lifecycle

import (
	"fmt"
	"path/filepath"

	"github.com/jamessawle/sbxflow/internal/domain/configuration"
)

// Plan contains the validated inputs needed to create or enter a sandbox.
type Plan struct {
	Name         string
	Agent        string
	Workspace    string
	Kits         []string
	Trust        configuration.Trust
	AllowedHosts []string
	Initialize   [][]string
}

// NewPlan converts a successful validation report into ordered Docker
// execution inputs without resolving any source a second time.
func NewPlan(report configuration.Validation) (Plan, error) {
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
		if selection.Source.Type == configuration.SourceLocal {
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

	initialize := make([][]string, len(report.Linked.Configuration.Sandbox.Hooks.Initialize))
	for index, command := range report.Linked.Configuration.Sandbox.Hooks.Initialize {
		initialize[index] = append([]string(nil), command.Command...)
	}
	return Plan{
		Name:         report.Linked.Configuration.Sandbox.Name,
		Agent:        report.Linked.Configuration.Sandbox.Agent,
		Workspace:    filepath.Dir(report.Declaration),
		Kits:         kits,
		Trust:        report.Linked.Trust,
		AllowedHosts: append([]string(nil), report.Linked.Configuration.Sandbox.Network.AllowedHosts...),
		Initialize:   initialize,
	}, nil
}

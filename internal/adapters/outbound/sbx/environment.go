package sbx

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
	sandboxport "github.com/jamessawle/sbxflow/internal/ports/sandbox"
)

const (
	environmentFilename = ".sbxenv.yaml"
	removalAgent        = "shell"
)

type environmentDocument struct {
	SchemaVersion string   `yaml:"schemaVersion"`
	Name          string   `yaml:"name"`
	Agent         string   `yaml:"agent"`
	Workspace     any      `yaml:"workspace"`
	Kits          []string `yaml:"kits,omitempty"`
}

type cloneWorkspaceDocument struct {
	Path  string `yaml:"path"`
	Clone bool   `yaml:"clone"`
}

func fullEnvironmentDocument(environment sandboxport.Environment) environmentDocument {
	workspace := any(environment.Workspace)
	if environment.WorkspaceMode == sandboxport.WorkspaceModeClone {
		workspace = cloneWorkspaceDocument{Path: environment.Workspace, Clone: true}
	}
	return environmentDocument{
		SchemaVersion: "1",
		Name:          environment.Name,
		Agent:         environment.Agent,
		Workspace:     workspace,
		Kits:          append([]string(nil), environment.Kits...),
	}
}

func removalEnvironmentDocument(name, workspace string) environmentDocument {
	return environmentDocument{
		SchemaVersion: "1",
		Name:          name,
		Agent:         removalAgent,
		Workspace:     workspace,
	}
}

type renderedEnvironment struct {
	path      string
	directory string
}

func (c Client) renderEnvironment(environment sandboxport.Environment) (renderedEnvironment, error) {
	return c.renderDocument(fullEnvironmentDocument(environment), environment.Workspace, false)
}

func (c Client) renderRemovalEnvironment(name string) (renderedEnvironment, error) {
	return c.renderDocument(environmentDocument{Name: name}, "", true)
}

func (c Client) renderDocument(document environmentDocument, mountedWorkspace string, identityOnly bool) (renderedEnvironment, error) {
	directory, err := os.MkdirTemp(c.TemporaryDirectory, "sbxflow-environment-")
	if err != nil {
		return renderedEnvironment{}, fmt.Errorf("create private Docker Sandbox environment directory: %w", err)
	}
	rendered := renderedEnvironment{directory: directory, path: filepath.Join(directory, environmentFilename)}
	fail := func(cause error) (renderedEnvironment, error) {
		return renderedEnvironment{}, errors.Join(cause, rendered.cleanup())
	}

	if identityOnly {
		workspace := filepath.Join(directory, "workspace")
		if err := os.Mkdir(workspace, 0o700); err != nil {
			return fail(fmt.Errorf("create private Docker Sandbox removal workspace: %w", err))
		}
		document = removalEnvironmentDocument(document.Name, workspace)
	} else {
		inside, err := pathWithin(mountedWorkspace, directory)
		if err != nil {
			return fail(fmt.Errorf("verify Docker Sandbox environment placement: %w", err))
		}
		if inside {
			return fail(fmt.Errorf("refuse to render Docker Sandbox environment within mounted workspace %q", mountedWorkspace))
		}
	}

	data, err := yaml.Marshal(document)
	if err != nil {
		return fail(fmt.Errorf("serialize Docker Sandbox environment: %w", err))
	}
	if err := os.WriteFile(rendered.path, data, 0o600); err != nil {
		return fail(fmt.Errorf("write private Docker Sandbox environment: %w", err))
	}
	return rendered, nil
}

func (environment renderedEnvironment) cleanup() error {
	if environment.directory == "" {
		return nil
	}
	if err := os.RemoveAll(environment.directory); err != nil {
		return fmt.Errorf("remove private Docker Sandbox environment directory: %w", err)
	}
	return nil
}

func pathWithin(parent, candidate string) (bool, error) {
	canonicalParent, err := canonicalPath(parent)
	if err != nil {
		return false, fmt.Errorf("canonicalize mounted workspace %q: %w", parent, err)
	}
	canonicalCandidate, err := canonicalPath(candidate)
	if err != nil {
		return false, fmt.Errorf("canonicalize temporary environment directory %q: %w", candidate, err)
	}
	relative, err := filepath.Rel(canonicalParent, canonicalCandidate)
	if err != nil {
		return false, err
	}
	return relative == "." || relative != ".." && !filepath.IsAbs(relative) && !startsWithParent(relative), nil
}

func canonicalPath(path string) (string, error) {
	canonical, err := filepath.EvalSymlinks(path)
	if err == nil {
		return canonical, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	return filepath.Abs(path)
}

func startsWithParent(path string) bool {
	separator := string(filepath.Separator)
	return len(path) > 3 && path[:3] == ".."+separator
}

func withRenderedEnvironment(render func() (renderedEnvironment, error), invoke func(string) error) (err error) {
	environment, err := render()
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, environment.cleanup())
	}()
	return invoke(environment.path)
}

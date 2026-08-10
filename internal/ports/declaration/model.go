// Package declaration defines the port used to load and resolve repository
// declarations without coupling adapters to domain services.
package declaration

// SourceType identifies how a kit source is addressed.
type SourceType string

const (
	SourceGit   SourceType = "git"
	SourceOCI   SourceType = "oci"
	SourceLocal SourceType = "local"
)

// Configuration is the boundary representation of a version 1 declaration.
type Configuration struct {
	Version int     `json:"version" yaml:"version"`
	Sandbox Sandbox `json:"sandbox" yaml:"sandbox"`
}

// LifecycleTarget is the declaration identity needed by teardown operations.
type LifecycleTarget struct {
	Declaration string
	Name        string
}

// LocalKit identifies one safely resolved local kit selection.
type LocalKit struct {
	Index  int
	Source string
	Kit    string
	Path   string
}

// LocalKitRequest describes one local kit path that an adapter must resolve.
type LocalKitRequest struct {
	Index  int
	Source string
	Root   string
	Kit    string
}

// Sandbox declares the sandbox identity and ordered kit configuration.
type Sandbox struct {
	Name  string `json:"name" yaml:"name"`
	Agent string `json:"agent" yaml:"agent"`
	Kits  Kits   `json:"kits" yaml:"kits"`
}

// Kits contains reusable named sources and their ordered selections.
type Kits struct {
	Sources map[string]Source `json:"sources" yaml:"sources"`
	Use     []Selection       `json:"use" yaml:"use"`
}

// Source is one schema-discriminated Git, OCI, or local source.
type Source struct {
	Type SourceType `json:"type" yaml:"type"`
	Repo string     `json:"repo,omitempty" yaml:"repo,omitempty"`
	Ref  string     `json:"ref,omitempty" yaml:"ref,omitempty"`
	Base string     `json:"base,omitempty" yaml:"base,omitempty"`
	Root string     `json:"root,omitempty" yaml:"root,omitempty"`
}

// Selection identifies a kit from a named source. Version is used by OCI.
type Selection struct {
	Source  string `json:"source" yaml:"source"`
	Kit     string `json:"kit" yaml:"kit"`
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
}

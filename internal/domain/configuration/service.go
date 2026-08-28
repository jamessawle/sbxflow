package configuration

import declarationport "github.com/jamessawle/sbxflow/internal/ports/declaration"

// Resolution is the domain state produced from a repository declaration
// before external local-kit validation is applied.
type Resolution struct {
	Declaration string
	Linked      LinkedConfiguration
	LocalKits   []LocalKit
	Errors      []error
	Warnings    []string
}

// Valid reports whether declaration discovery, linking, and path resolution
// succeeded.
func (r Resolution) Valid() bool { return len(r.Errors) == 0 }

// LocalKitValidation records the validation result for one resolved local kit.
type LocalKitValidation struct {
	Target      LocalKit
	Valid       bool
	Diagnostics string
	Err         error
}

// Validation is the complete validity state of a repository configuration.
type Validation struct {
	Declaration string
	Linked      LinkedConfiguration
	LocalKits   []LocalKitValidation
	Errors      []error
	Warnings    []string
}

// Valid reports whether every validation phase succeeded.
func (v Validation) Valid() bool { return len(v.Errors) == 0 }

// ConfigurationResolver supplies resolved configuration domain state to
// application use cases.
type ConfigurationResolver interface {
	Resolve(start string) Resolution
}

// Resolver resolves repository declarations through injected ports and
// applies configuration-domain rules.
type Resolver struct {
	Declarations declarationport.Loader
	LocalPaths   declarationport.LocalKitResolver
}

// Resolve discovers, links, and safely resolves the nearest declaration.
// Each unsafe phase gates the phases that follow it.
func (r Resolver) Resolve(start string) Resolution {
	resolution := Resolution{}
	declaration, err := r.Declarations.Load(start)
	if err != nil {
		resolution.Declaration = declaration.Path
		resolution.Errors = append(resolution.Errors, err)
		return resolution
	}
	resolution.Declaration = declaration.Path
	if declaration.Configuration.Sandbox.Workspace == nil {
		resolution.Warnings = append(resolution.Warnings, "sandbox.workspace.mode is omitted; the effective mode is currently direct. Declare direct or clone explicitly because the implicit mode will change to clone in a future pre-1.0 release")
		declaration.Configuration.Sandbox.Workspace = &Workspace{Mode: WorkspaceModeDirect}
	}
	linked, err := Link(declaration.Configuration)
	if err != nil {
		resolution.Errors = append(resolution.Errors, err)
		return resolution
	}
	resolution.Linked = linked
	resolution.LocalKits, err = r.LocalPaths.ResolveLocalKits(declaration.Path, LocalKitRequests(linked))
	if err != nil {
		resolution.Errors = append(resolution.Errors, err)
	}
	return resolution
}

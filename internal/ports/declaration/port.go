package declaration

// Loaded is a decoded repository declaration and its discovered path.
type Loaded struct {
	Path          string
	Configuration Configuration
}

// Loader discovers, reads, and decodes a repository declaration.
type Loader interface {
	Load(start string) (Loaded, error)
}

// LocalKitResolver safely resolves selected local filesystem paths.
type LocalKitResolver interface {
	ResolveLocalKits(declarationPath string, requests []LocalKitRequest) ([]LocalKit, error)
}

// TargetResolver resolves the minimal identity required by teardown.
type TargetResolver interface {
	Resolve(start string) (LifecycleTarget, error)
}

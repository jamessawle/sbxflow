// Package buildinfo exposes the identity of the current sbxflow build.
//
// Build identity is an intentionally self-contained port: linker-injected
// values need no outbound adapter, and inbound adapters may read them directly.
package buildinfo

// Version and Commit may be replaced at build time with -ldflags -X values.
var (
	Version = "development"
	Commit  = ""
)

// Info describes an sbxflow build.
type Info struct {
	Version string
	Commit  string
}

// Current returns the build identity, preserving an explicit development
// version when release metadata was not injected.
func Current() Info {
	version := Version
	if version == "" {
		version = "development"
	}

	return Info{
		Version: version,
		Commit:  Commit,
	}
}

package buildinfo

import "testing"

func TestCurrent(t *testing.T) {
	originalVersion, originalCommit := Version, Commit
	t.Cleanup(func() {
		Version, Commit = originalVersion, originalCommit
	})

	Version = "v1.2.3"
	Commit = "0123456789abcdef"

	got := Current()
	if got.Version != Version || got.Commit != Commit {
		t.Fatalf("Current() = %#v, want version %q and commit %q", got, Version, Commit)
	}
}

func TestCurrentFallsBackToDevelopmentVersion(t *testing.T) {
	originalVersion := Version
	t.Cleanup(func() { Version = originalVersion })

	Version = ""
	if got := Current().Version; got != "development" {
		t.Fatalf("Current().Version = %q, want %q", got, "development")
	}
}

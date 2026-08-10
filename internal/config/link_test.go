package config

import (
	"reflect"
	"strings"
	"testing"
)

func TestLinkValidatesReferencesAndVersionRules(t *testing.T) {
	configuration := testConfiguration()

	tests := map[string]struct {
		mutate func(*Configuration)
		want   string
	}{
		"unknown source": {
			mutate: func(c *Configuration) { c.Sandbox.Kits.Use[0].Source = "missing" },
			want:   "use[0].source references unknown source",
		},
		"OCI version required": {
			mutate: func(c *Configuration) { c.Sandbox.Kits.Use[1].Version = "" },
			want:   "use[1].version is required",
		},
		"Git version rejected": {
			mutate: func(c *Configuration) { c.Sandbox.Kits.Use[0].Version = "v2" },
			want:   "use[0].version is not valid for Git",
		},
		"local version rejected": {
			mutate: func(c *Configuration) { c.Sandbox.Kits.Use[2].Version = "v2" },
			want:   "use[2].version is not valid for local",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			candidate := cloneConfiguration(configuration)
			test.mutate(&candidate)
			_, err := Link(candidate)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Link() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestLinkPreservesOrderAndDerivesTrust(t *testing.T) {
	configuration := testConfiguration()
	configuration.Sandbox.Kits.Use = append(configuration.Sandbox.Kits.Use,
		Selection{Source: "git", Kit: "second"},
		Selection{Source: "oci", Kit: "image", Version: "v2"},
	)
	configuration.Sandbox.Kits.Sources["unused"] = Source{Type: SourceGit, Repo: "https://unused.example/org/repo.git", Ref: "main"}

	linked, err := Link(configuration)
	if err != nil {
		t.Fatalf("Link() error = %v", err)
	}
	if len(linked.Selections) != len(configuration.Sandbox.Kits.Use) {
		t.Fatalf("selection count = %d, want %d", len(linked.Selections), len(configuration.Sandbox.Kits.Use))
	}
	for index, selection := range linked.Selections {
		if selection.Index != index || selection.Selection != configuration.Sandbox.Kits.Use[index] {
			t.Fatalf("selection %d lost declaration order: %#v", index, selection)
		}
	}
	wantSources := []string{"docker.io/", "github.com/example/kits", "ghcr.io/example/image"}
	if !reflect.DeepEqual(linked.Trust.AllowedSources, wantSources) {
		t.Fatalf("AllowedSources = %#v, want %#v", linked.Trust.AllowedSources, wantSources)
	}
	if !linked.Trust.AllowLocalKits {
		t.Fatal("AllowLocalKits = false, want true")
	}
	if got := linked.Selections[0].RemoteReference; got != "git+https://github.com/example/kits.git#ref=v1&dir=tooling" {
		t.Fatalf("Git RemoteReference = %q", got)
	}
	if got := linked.Selections[1].RemoteReference; got != "ghcr.io/example/image:v1" {
		t.Fatalf("OCI RemoteReference = %q", got)
	}
}

func TestNormalizeGitBuildsDockerExecutionReferences(t *testing.T) {
	for _, test := range []struct {
		name       string
		repository string
		want       string
	}{
		{
			name:       "HTTPS URL",
			repository: "https://github.com/docker/sbx-kits-contrib.git",
			want:       "git+https://github.com/docker/sbx-kits-contrib.git#ref=v0.14.0&dir=mise",
		},
		{
			name:       "SSH URL",
			repository: "ssh://git@github.com/docker/sbx-kits-contrib.git",
			want:       "git+ssh://git@github.com/docker/sbx-kits-contrib.git#ref=v0.14.0&dir=mise",
		},
		{
			name:       "SCP style",
			repository: "git@github.com:docker/sbx-kits-contrib.git",
			want:       "git+ssh://git@github.com/docker/sbx-kits-contrib.git#ref=v0.14.0&dir=mise",
		},
		{
			name:       "host path",
			repository: "github.com/docker/sbx-kits-contrib.git",
			want:       "git+https://github.com/docker/sbx-kits-contrib.git#ref=v0.14.0&dir=mise",
		},
		{
			name:       "already prefixed",
			repository: "git+https://github.com/docker/sbx-kits-contrib.git",
			want:       "git+https://github.com/docker/sbx-kits-contrib.git#ref=v0.14.0&dir=mise",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, trust, err := normalizeGit(test.repository, "v0.14.0", "mise")
			if err != nil {
				t.Fatalf("normalizeGit() error = %v", err)
			}
			if got != test.want {
				t.Fatalf("normalizeGit() = %q, want %q", got, test.want)
			}
			if trust != "github.com/docker/sbx-kits-contrib" {
				t.Fatalf("trust prefix = %q", trust)
			}
		})
	}
}

func TestLinkRemoteOnlyDoesNotEnableLocalKits(t *testing.T) {
	configuration := testConfiguration()
	configuration.Sandbox.Kits.Use = configuration.Sandbox.Kits.Use[:2]
	linked, err := Link(configuration)
	if err != nil {
		t.Fatalf("Link() error = %v", err)
	}
	if linked.Trust.AllowLocalKits {
		t.Fatal("AllowLocalKits = true, want false")
	}
}

func testConfiguration() Configuration {
	return Configuration{
		Version: 1,
		Sandbox: Sandbox{Name: "demo", Agent: "codex", Kits: Kits{
			Sources: map[string]Source{
				"git":   {Type: SourceGit, Repo: "https://github.com/example/kits.git", Ref: "v1"},
				"oci":   {Type: SourceOCI, Base: "ghcr.io/example"},
				"local": {Type: SourceLocal, Root: "./kits"},
			},
			Use: []Selection{
				{Source: "git", Kit: "tooling"},
				{Source: "oci", Kit: "image", Version: "v1"},
				{Source: "local", Kit: "local-kit"},
			},
		}},
	}
}

func cloneConfiguration(configuration Configuration) Configuration {
	clone := configuration
	clone.Sandbox.Kits.Use = append([]Selection(nil), configuration.Sandbox.Kits.Use...)
	clone.Sandbox.Kits.Sources = make(map[string]Source, len(configuration.Sandbox.Kits.Sources))
	for name, source := range configuration.Sandbox.Kits.Sources {
		clone.Sandbox.Kits.Sources[name] = source
	}
	return clone
}

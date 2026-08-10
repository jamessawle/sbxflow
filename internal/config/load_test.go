package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestLoadRepositoryExamples(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate test source")
	}
	repositoryRoot := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))

	paths, err := filepath.Glob(filepath.Join(repositoryRoot, "examples", "*", "sbxflow.yaml"))
	if err != nil {
		t.Fatalf("find examples: %v", err)
	}
	if len(paths) == 0 {
		t.Fatal("no repository examples found")
	}
	for _, path := range paths {
		path := path
		t.Run(filepath.Base(filepath.Dir(path)), func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read example: %v", err)
			}
			configuration, err := Load(data)
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}
			if configuration.Version != 1 || len(configuration.Sandbox.Kits.Use) == 0 {
				t.Fatalf("unexpected decoded configuration: %#v", configuration)
			}
		})
	}
}

func TestLoadSupportedAgents(t *testing.T) {
	valid := `version: 1
sandbox:
  name: demo
  agent: codex
  kits:
    sources:
      community:
        type: git
        repo: https://github.com/example/kits.git
        ref: v1
    use:
      - source: community
        kit: tooling
`

	for _, agent := range []string{
		"claude", "codex", "copilot", "cursor", "docker-agent",
		"droid", "gemini", "kiro", "opencode", "shell",
	} {
		t.Run(agent, func(t *testing.T) {
			document := strings.Replace(valid, "agent: codex", "agent: "+agent, 1)
			configuration, err := Load([]byte(document))
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}
			if configuration.Sandbox.Agent != agent {
				t.Fatalf("agent = %q, want %q", configuration.Sandbox.Agent, agent)
			}
		})
	}
}

func TestLoadRepositoryDeclaration(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate test source")
	}
	path := filepath.Join(filepath.Dir(file), "..", "..", Filename)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read repository declaration: %v", err)
	}
	configuration, err := Load(data)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if configuration.Sandbox.Name != "sbxflow" || configuration.Sandbox.Agent != "codex" {
		t.Fatalf("unexpected repository declaration: %#v", configuration.Sandbox)
	}
}

func TestLoadRejectsInvalidDocuments(t *testing.T) {
	valid := `version: 1
sandbox:
  name: demo
  agent: codex
  kits:
    sources:
      community:
        type: git
        repo: https://github.com/example/kits.git
        ref: v1
    use:
      - source: community
        kit: tooling
`

	tests := map[string]struct {
		document string
		want     string
	}{
		"empty":              {document: "", want: "empty"},
		"malformed YAML":     {document: "version: [", want: "parse YAML"},
		"duplicate key":      {document: strings.Replace(valid, "version: 1", "version: 1\nversion: 1", 1), want: "already defined"},
		"multiple documents": {document: valid + "---\n" + valid, want: "exactly one"},
		"unsupported version": {document: strings.Replace(valid, "version: 1", "version: 2", 1),
			want: "version"},
		"unsupported agent":  {document: strings.Replace(valid, "agent: codex", "agent: typo", 1), want: "/sandbox/agent"},
		"unknown root field": {document: valid + "unknown: true\n", want: "additional properties"},
		"wrong source shape": {document: strings.Replace(valid, "        ref: v1", "        base: ghcr.io/example", 1), want: "sources"},
		"unknown source field": {document: strings.Replace(valid, "        ref: v1", "        ref: v1\n        mystery: value", 1),
			want: "mystery"},
		"duplicate selection": {document: valid + "      - source: community\n        kit: tooling\n", want: "items at 0 and 1 are equal"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := Load([]byte(test.document))
			if err == nil {
				t.Fatal("Load() error = nil, want error")
			}
			if !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Load() error = %q, want substring %q", err, test.want)
			}
		})
	}
}

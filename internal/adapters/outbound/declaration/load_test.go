package declaration

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	declarationport "github.com/jamessawle/sbxflow/internal/ports/declaration"
)

func TestLoadRepositoryExamples(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate test source")
	}
	repositoryRoot := filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..", ".."))

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

func TestLoadPreservesWorkspaceModePresence(t *testing.T) {
	base := `version: 1
sandbox:
  name: demo
  agent: codex
  kits:
    sources:
      community: {type: git, repo: https://example.test/kits.git, ref: v1}
    use:
      - {source: community, kit: tooling}
`
	omitted, err := Load([]byte(base))
	if err != nil || omitted.Sandbox.Workspace != nil {
		t.Fatalf("omitted workspace = %#v, %v", omitted.Sandbox.Workspace, err)
	}
	for _, mode := range []declarationport.WorkspaceMode{declarationport.WorkspaceModeDirect, declarationport.WorkspaceModeClone} {
		document := strings.Replace(base, "  kits:", "  workspace:\n    mode: "+string(mode)+"\n  kits:", 1)
		configuration, loadErr := Load([]byte(document))
		if loadErr != nil || configuration.Sandbox.Workspace == nil || configuration.Sandbox.Workspace.Mode != mode {
			t.Fatalf("mode %q workspace = %#v, %v", mode, configuration.Sandbox.Workspace, loadErr)
		}
	}
}

func TestLoadRepositoryDeclaration(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate test source")
	}
	path := filepath.Join(filepath.Dir(file), "..", "..", "..", "..", Filename)
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

func TestLoadPreservesOrderedNetworkAllowedHosts(t *testing.T) {
	document := `version: 1
sandbox:
  name: demo
  agent: codex
  network:
    allowedHosts:
      - api.example.com
      - packages.example.com:443
      - "*.cdn.example.com"
      - 10.0.0.1
      - "[fd00::1]:8443"
      - "edge.example.com:1"
      - "origin.example.com:65535"
      - "**"
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
	configuration, err := Load([]byte(document))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	want := []string{"api.example.com", "packages.example.com:443", "*.cdn.example.com", "10.0.0.1", "[fd00::1]:8443", "edge.example.com:1", "origin.example.com:65535", "**"}
	if !reflect.DeepEqual(configuration.Sandbox.Network.AllowedHosts, want) {
		t.Fatalf("allowed hosts = %#v, want %#v", configuration.Sandbox.Network.AllowedHosts, want)
	}
}

func TestLoadPreservesOrderedInitializationCommands(t *testing.T) {
	document := `version: 1
sandbox:
  name: demo
  agent: codex
  hooks:
    initialize:
      - command: [bash, -c, "printf '%s' literal"]
      - command: [npm, ci]
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
	configuration, err := Load([]byte(document))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	want := []declarationport.Command{{Command: []string{"bash", "-c", "printf '%s' literal"}}, {Command: []string{"npm", "ci"}}}
	if !reflect.DeepEqual(configuration.Sandbox.Hooks.Initialize, want) {
		t.Fatalf("initialize = %#v, want %#v", configuration.Sandbox.Hooks.Initialize, want)
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
		"unsupported agent":          {document: strings.Replace(valid, "agent: codex", "agent: typo", 1), want: "/sandbox/agent"},
		"unknown root field":         {document: valid + "unknown: true\n", want: "additional properties"},
		"empty workspace":            {document: strings.Replace(valid, "  kits:", "  workspace: {}\n  kits:", 1), want: "minProperties"},
		"unknown workspace field":    {document: strings.Replace(valid, "  kits:", "  workspace:\n    unknown: true\n  kits:", 1), want: "unknown"},
		"unsupported workspace mode": {document: strings.Replace(valid, "  kits:", "  workspace:\n    mode: shared\n  kits:", 1), want: "/sandbox/workspace/mode"},
		"wrong source shape":         {document: strings.Replace(valid, "        ref: v1", "        base: ghcr.io/example", 1), want: "sources"},
		"unknown source field": {document: strings.Replace(valid, "        ref: v1", "        ref: v1\n        mystery: value", 1),
			want: "mystery"},
		"duplicate selection":    {document: valid + "      - source: community\n        kit: tooling\n", want: "items at 0 and 1 are equal"},
		"empty network host":     {document: strings.Replace(valid, "  kits:", "  network:\n    allowedHosts: ['']\n  kits:", 1), want: "allowedHosts"},
		"duplicate network host": {document: strings.Replace(valid, "  kits:", "  network:\n    allowedHosts: [example.com, example.com]\n  kits:", 1), want: "items at 0 and 1 are equal"},
		"unknown network field":  {document: strings.Replace(valid, "  kits:", "  network:\n    unknown: true\n  kits:", 1), want: "unknown"},
		"empty command vector":   {document: strings.Replace(valid, "  kits:", "  hooks:\n    initialize:\n      - command: []\n  kits:", 1), want: "command"},
		"missing command vector": {document: strings.Replace(valid, "  kits:", "  hooks:\n    initialize:\n      - {}\n  kits:", 1), want: "command"},
		"empty command argument": {document: strings.Replace(valid, "  kits:", "  hooks:\n    initialize:\n      - command: [npm, '']\n  kits:", 1), want: "command"},
		"unknown hooks field":    {document: strings.Replace(valid, "  kits:", "  hooks:\n    unknown: true\n  kits:", 1), want: "unknown"},
		"unknown command field":  {document: strings.Replace(valid, "  kits:", "  hooks:\n    initialize:\n      - command: [npm, ci]\n        unknown: true\n  kits:", 1), want: "unknown"},
		// Docker Sandboxes accepts a URL resource but matches requests by host and
		// port, so such a rule would never take effect.
		"network host URL":  {document: strings.Replace(valid, "  kits:", "  network:\n    allowedHosts: ['https://example.com']\n  kits:", 1), want: "optional :port suffix"},
		"network host path": {document: strings.Replace(valid, "  kits:", "  network:\n    allowedHosts: ['example.com/v1']\n  kits:", 1), want: "optional :port suffix"},
		// Docker Sandboxes matches by host and port, so a port outside 1-65535 or
		// a malformed literal is a rule that could never match a request.
		"network port zero":         {document: strings.Replace(valid, "  kits:", "  network:\n    allowedHosts: ['example.com:0']\n  kits:", 1), want: "optional :port suffix"},
		"network port above range":  {document: strings.Replace(valid, "  kits:", "  network:\n    allowedHosts: ['example.com:65536']\n  kits:", 1), want: "optional :port suffix"},
		"network port five nines":   {document: strings.Replace(valid, "  kits:", "  network:\n    allowedHosts: ['example.com:99999']\n  kits:", 1), want: "optional :port suffix"},
		"network malformed literal": {document: strings.Replace(valid, "  kits:", "  network:\n    allowedHosts: ['[::::]']\n  kits:", 1), want: "optional :port suffix"},
		"network bracketed IPv4":    {document: strings.Replace(valid, "  kits:", "  network:\n    allowedHosts: ['[192.168.1.1]']\n  kits:", 1), want: "optional :port suffix"},
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

func TestLoadLifecycleTarget(t *testing.T) {
	target, err := LoadLifecycleTarget([]byte(`version: 1
sandbox:
  name: exact-project
  network:
    allowedHosts: [first.example, second.example]
`))
	if err != nil || target.Name != "exact-project" || !reflect.DeepEqual(target.AllowedHosts, []string{"first.example", "second.example"}) {
		t.Fatalf("LoadLifecycleTarget() = %#v, %v", target, err)
	}
}

func TestLoadLifecycleTargetRejectsUnsafeIdentity(t *testing.T) {
	valid := `version: 1
sandbox:
  name: demo
`
	tests := map[string]struct {
		document string
		want     string
	}{
		"empty":                 {document: "", want: "empty"},
		"malformed YAML":        {document: "version: [", want: "parse YAML"},
		"duplicate key":         {document: "version: 1\nversion: 1\nsandbox:\n  name: demo\n", want: "already defined"},
		"multiple documents":    {document: valid + "---\n" + valid, want: "exactly one"},
		"root is not object":    {document: "- version\n- sandbox\n", want: "must be an object"},
		"missing version":       {document: "sandbox:\n  name: demo\n", want: "missing version"},
		"non-integer version":   {document: "version: one\nsandbox:\n  name: demo\n", want: "must be an integer"},
		"unsupported version":   {document: "version: 2\nsandbox:\n  name: demo\n", want: "unsupported"},
		"missing sandbox":       {document: "version: 1\n", want: "missing sandbox"},
		"sandbox is not object": {document: "version: 1\nsandbox: demo\n", want: "must be an object"},
		"missing name":          {document: "version: 1\nsandbox: {}\n", want: "missing sandbox.name"},
		"name is not string":    {document: "version: 1\nsandbox:\n  name: 7\n", want: "must be a string"},
		"empty name":            {document: "version: 1\nsandbox:\n  name: ''\n", want: "must not be empty"},
		"network is not object": {document: valid + "  network: hosts\n", want: "sandbox.network must be an object"},
		"hosts are not array":   {document: valid + "  network:\n    allowedHosts: first.example\n", want: "must be an array"},
		"empty host":            {document: valid + "  network:\n    allowedHosts: ['']\n", want: "must be a non-empty string"},
		"host is not string":    {document: valid + "  network:\n    allowedHosts: [7]\n", want: "must be a non-empty string"},
		"host URL":              {document: valid + "  network:\n    allowedHosts: ['https://first.example']\n", want: "optional :port suffix"},
		"port zero":             {document: valid + "  network:\n    allowedHosts: ['first.example:0']\n", want: "optional :port suffix"},
		"port above range":      {document: valid + "  network:\n    allowedHosts: ['first.example:65536']\n", want: "optional :port suffix"},
		"malformed literal":     {document: valid + "  network:\n    allowedHosts: ['[::::]']\n", want: "optional :port suffix"},
		"bracketed IPv4":        {document: valid + "  network:\n    allowedHosts: ['[192.168.1.1]']\n", want: "optional :port suffix"},
		"duplicate host":        {document: valid + "  network:\n    allowedHosts: [first.example, first.example]\n", want: "duplicates"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := LoadLifecycleTarget([]byte(test.document))
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("LoadLifecycleTarget() error = %v, want substring %q", err, test.want)
			}
		})
	}
}

func TestLoadLifecycleTargetIgnoresNonIdentityConfiguration(t *testing.T) {
	document := `version: 1
sandbox:
  name: demo
  agent: unsupported
  kits:
    sources:
      local:
        type: local
        root: https://unavailable.example/kits
    use:
      - source: missing
        kit: ../unsafe
unknown: true
`
	if _, err := Load([]byte(document)); err == nil {
		t.Fatal("Load() error = nil, want complete validation failure")
	}
	target, err := LoadLifecycleTarget([]byte(document))
	if err != nil || target.Name != "demo" {
		t.Fatalf("LoadLifecycleTarget() = %#v, %v", target, err)
	}
}

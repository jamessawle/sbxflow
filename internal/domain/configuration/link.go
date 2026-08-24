package configuration

import (
	"fmt"
	"net/url"
	"path"
	"strings"
)

// LinkedConfiguration retains declaration order and the source resolved for
// every kit selection.
type LinkedConfiguration struct {
	Configuration Configuration
	Selections    []LinkedSelection
	Trust         Trust
}

// LinkedSelection pairs an ordered selection with its named source.
type LinkedSelection struct {
	Index           int
	Selection       Selection
	Source          Source
	RemoteReference string
	TrustPrefix     string
}

// Trust is the least-privilege Docker kit-source configuration required by a
// linked declaration.
type Trust struct {
	AllowedSources []string
	AllowLocalKits bool
}

// Link resolves dynamic source references, validates source-specific selection
// rules, and derives offline remote references and trust settings.
func Link(configuration Configuration) (LinkedConfiguration, error) {
	configuration.Sandbox.Hooks.Initialize = cloneCommands(configuration.Sandbox.Hooks.Initialize)
	linked := LinkedConfiguration{
		Configuration: configuration,
		Selections:    make([]LinkedSelection, 0, len(configuration.Sandbox.Kits.Use)),
		Trust: Trust{
			AllowedSources: []string{"docker.io/"},
		},
	}
	seenPrefixes := map[string]struct{}{"docker.io/": {}}

	for index, selection := range configuration.Sandbox.Kits.Use {
		item, err := linkSelection(configuration.Sandbox.Kits.Sources, index, selection)
		if err != nil {
			return LinkedConfiguration{}, err
		}
		updateTrust(&linked.Trust, seenPrefixes, item)
		linked.Selections = append(linked.Selections, item)
	}

	return linked, nil
}

func cloneCommands(commands []Command) []Command {
	cloned := make([]Command, len(commands))
	for index, command := range commands {
		cloned[index].Command = append([]string(nil), command.Command...)
	}
	return cloned
}

func linkSelection(sources map[string]Source, index int, selection Selection) (LinkedSelection, error) {
	source, ok := sources[selection.Source]
	if !ok {
		return LinkedSelection{}, fmt.Errorf("sandbox.kits.use[%d].source references unknown source %q", index, selection.Source)
	}
	item := LinkedSelection{Index: index, Selection: selection, Source: source}
	var err error
	switch source.Type {
	case SourceOCI:
		item.RemoteReference, item.TrustPrefix, err = linkOCISelection(index, selection, source)
	case SourceGit:
		item.RemoteReference, item.TrustPrefix, err = linkGitSelection(index, selection, source)
	case SourceLocal:
		err = validateLocalSelection(index, selection)
	default:
		err = fmt.Errorf("sandbox.kits.sources.%s.type %q is unsupported", selection.Source, source.Type)
	}
	return item, err
}

func linkOCISelection(index int, selection Selection, source Source) (string, string, error) {
	if selection.Version == "" {
		return "", "", fmt.Errorf("sandbox.kits.use[%d].version is required for OCI source %q", index, selection.Source)
	}
	reference, prefix, err := normalizeOCI(source.Base, selection.Kit, selection.Version)
	if err != nil {
		return "", "", fmt.Errorf("sandbox.kits.use[%d]: %w", index, err)
	}
	return reference, prefix, nil
}

func linkGitSelection(index int, selection Selection, source Source) (string, string, error) {
	if selection.Version != "" {
		return "", "", fmt.Errorf("sandbox.kits.use[%d].version is not valid for Git source %q", index, selection.Source)
	}
	reference, prefix, err := normalizeGit(source.Repo, source.Ref, selection.Kit)
	if err != nil {
		return "", "", fmt.Errorf("sandbox.kits.use[%d]: %w", index, err)
	}
	return reference, prefix, nil
}

func validateLocalSelection(index int, selection Selection) error {
	if selection.Version != "" {
		return fmt.Errorf("sandbox.kits.use[%d].version is not valid for local source %q", index, selection.Source)
	}
	return nil
}

func updateTrust(trust *Trust, seenPrefixes map[string]struct{}, item LinkedSelection) {
	if item.Source.Type == SourceLocal {
		trust.AllowLocalKits = true
	}
	if item.TrustPrefix == "" {
		return
	}
	if _, exists := seenPrefixes[item.TrustPrefix]; exists {
		return
	}
	seenPrefixes[item.TrustPrefix] = struct{}{}
	trust.AllowedSources = append(trust.AllowedSources, item.TrustPrefix)
}

func normalizeGit(repository, ref, kit string) (string, string, error) {
	host, repositoryPath, err := splitGitRepository(repository)
	if err != nil {
		return "", "", fmt.Errorf("source Git repository %q is invalid: %w", repository, err)
	}
	executionRepository, err := gitExecutionRepository(repository)
	if err != nil {
		return "", "", fmt.Errorf("source Git repository %q is invalid: %w", repository, err)
	}
	prefix := host + "/" + repositoryPath
	reference := executionRepository + "#ref=" + url.QueryEscape(ref)
	if kit != "" {
		reference += "&dir=" + url.QueryEscape(path.Clean(kit))
	}
	return reference, prefix, nil
}

func gitExecutionRepository(repository string) (string, error) {
	repository = strings.TrimSpace(repository)
	if strings.Contains(repository, "://") {
		parsed, err := url.Parse(repository)
		if err != nil || parsed.Host == "" {
			return "", fmt.Errorf("expected a repository URL with a host")
		}
		if parsed.RawQuery != "" || parsed.Fragment != "" {
			return "", fmt.Errorf("repository URL must not contain a query or fragment")
		}
		if strings.HasPrefix(strings.ToLower(parsed.Scheme), "git+") {
			return repository, nil
		}
		return "git+" + repository, nil
	}

	if colon := strings.Index(repository, ":"); colon > 0 {
		return "git+ssh://" + repository[:colon] + "/" + strings.TrimPrefix(repository[colon+1:], "/"), nil
	}
	return "git+https://" + strings.TrimPrefix(repository, "//"), nil
}

func splitGitRepository(repository string) (string, string, error) {
	repository = strings.TrimSpace(repository)
	if strings.Contains(repository, "://") {
		parsed, err := url.Parse(repository)
		if err != nil || parsed.Host == "" {
			return "", "", fmt.Errorf("expected a repository URL with a host")
		}
		repositoryPath, err := cleanRemotePath(parsed.Path)
		return parsed.Host, repositoryPath, err
	}

	if at := strings.LastIndex(repository, "@"); at >= 0 {
		repository = repository[at+1:]
	}
	if colon := strings.Index(repository, ":"); colon > 0 {
		repositoryPath, err := cleanRemotePath(repository[colon+1:])
		return repository[:colon], repositoryPath, err
	}
	parts := strings.SplitN(strings.Trim(repository, "/"), "/", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("expected host and repository path")
	}
	repositoryPath, err := cleanRemotePath(parts[1])
	return parts[0], repositoryPath, err
}

func cleanRemotePath(value string) (string, error) {
	cleaned := strings.TrimSuffix(strings.Trim(path.Clean("/"+value), "/"), ".git")
	if cleaned == "" || cleaned == "." {
		return "", fmt.Errorf("expected a repository path")
	}
	return cleaned, nil
}

func normalizeOCI(base, kit, version string) (string, string, error) {
	base = strings.TrimSpace(strings.TrimPrefix(base, "oci://"))
	base = strings.Trim(base, "/")
	if strings.ContainsAny(base, "?#@") || !strings.Contains(base, "/") {
		return "", "", fmt.Errorf("OCI source base %q must contain a registry and namespace", base)
	}
	prefix := base + "/" + strings.Trim(path.Clean("/"+kit), "/")
	return prefix + ":" + version, prefix, nil
}

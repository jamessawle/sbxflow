package config

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
	linked := LinkedConfiguration{
		Configuration: configuration,
		Selections:    make([]LinkedSelection, 0, len(configuration.Sandbox.Kits.Use)),
		Trust: Trust{
			AllowedSources: []string{"docker.io/"},
		},
	}
	seenPrefixes := map[string]struct{}{"docker.io/": {}}

	for index, selection := range configuration.Sandbox.Kits.Use {
		source, ok := configuration.Sandbox.Kits.Sources[selection.Source]
		if !ok {
			return LinkedConfiguration{}, fmt.Errorf("sandbox.kits.use[%d].source references unknown source %q", index, selection.Source)
		}

		item := LinkedSelection{Index: index, Selection: selection, Source: source}
		switch source.Type {
		case SourceOCI:
			if selection.Version == "" {
				return LinkedConfiguration{}, fmt.Errorf("sandbox.kits.use[%d].version is required for OCI source %q", index, selection.Source)
			}
			var err error
			item.RemoteReference, item.TrustPrefix, err = normalizeOCI(source.Base, selection.Kit, selection.Version)
			if err != nil {
				return LinkedConfiguration{}, fmt.Errorf("sandbox.kits.use[%d]: %w", index, err)
			}
		case SourceGit:
			if selection.Version != "" {
				return LinkedConfiguration{}, fmt.Errorf("sandbox.kits.use[%d].version is not valid for Git source %q", index, selection.Source)
			}
			var err error
			item.RemoteReference, item.TrustPrefix, err = normalizeGit(source.Repo, source.Ref, selection.Kit)
			if err != nil {
				return LinkedConfiguration{}, fmt.Errorf("sandbox.kits.use[%d]: %w", index, err)
			}
		case SourceLocal:
			if selection.Version != "" {
				return LinkedConfiguration{}, fmt.Errorf("sandbox.kits.use[%d].version is not valid for local source %q", index, selection.Source)
			}
			linked.Trust.AllowLocalKits = true
		default:
			return LinkedConfiguration{}, fmt.Errorf("sandbox.kits.sources.%s.type %q is unsupported", selection.Source, source.Type)
		}

		if item.TrustPrefix != "" {
			if _, exists := seenPrefixes[item.TrustPrefix]; !exists {
				seenPrefixes[item.TrustPrefix] = struct{}{}
				linked.Trust.AllowedSources = append(linked.Trust.AllowedSources, item.TrustPrefix)
			}
		}
		linked.Selections = append(linked.Selections, item)
	}

	return linked, nil
}

func normalizeGit(repository, ref, kit string) (string, string, error) {
	host, repositoryPath, err := splitGitRepository(repository)
	if err != nil {
		return "", "", fmt.Errorf("source Git repository %q is invalid: %w", repository, err)
	}
	prefix := host + "/" + repositoryPath
	reference := prefix + "@" + ref
	if kit != "" {
		reference += ":" + path.Clean(kit)
	}
	return reference, prefix, nil
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

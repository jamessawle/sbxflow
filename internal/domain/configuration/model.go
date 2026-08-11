// Package configuration defines declaration linking, semantic rules, and
// derived trust policy.
package configuration

import "github.com/jamessawle/sbxflow/internal/ports/declaration"

type SourceType = declaration.SourceType

const (
	SourceGit   = declaration.SourceGit
	SourceOCI   = declaration.SourceOCI
	SourceLocal = declaration.SourceLocal
)

type Configuration = declaration.Configuration
type LifecycleTarget = declaration.LifecycleTarget
type LocalKit = declaration.LocalKit
type LocalKitRequest = declaration.LocalKitRequest
type Sandbox = declaration.Sandbox
type Network = declaration.Network
type Kits = declaration.Kits
type Source = declaration.Source
type Selection = declaration.Selection
type TargetResolver = declaration.TargetResolver

// LocalKitRequests returns the filesystem inputs needed for linked local
// selections in declaration order.
func LocalKitRequests(linked LinkedConfiguration) []LocalKitRequest {
	requests := make([]LocalKitRequest, 0)
	for _, selection := range linked.Selections {
		if selection.Source.Type != SourceLocal {
			continue
		}
		requests = append(requests, LocalKitRequest{
			Index:  selection.Index,
			Source: selection.Selection.Source,
			Root:   selection.Source.Root,
			Kit:    selection.Selection.Kit,
		})
	}
	return requests
}

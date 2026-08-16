// Package schema exposes the published sbxflow configuration schema.
package schema

// Import embed for the //go:embed directive below without adding its API to
// this package's namespace.
import _ "embed"

// Configuration contains the Draft 2020-12 schema for sbxflow.yaml.
//
//go:embed sbxflow.schema.json
var Configuration []byte

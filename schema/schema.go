// Package schema exposes the published sbxflow configuration schema.
package schema

import _ "embed"

// Configuration contains the Draft 2020-12 schema for sbxflow.yaml.
//
//go:embed sbxflow.schema.json
var Configuration []byte

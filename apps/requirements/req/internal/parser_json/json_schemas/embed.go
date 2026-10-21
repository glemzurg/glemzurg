package json_schemas

import "embed"

// Schemas contains all embedded JSON schema files.
//
//go:embed *.json
var Schemas embed.FS

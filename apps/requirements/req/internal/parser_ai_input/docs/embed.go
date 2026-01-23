package docs

import "embed"

// Docs contains all embedded documentation files.
//
//go:embed *.md
var Docs embed.FS

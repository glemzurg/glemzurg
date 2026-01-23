package errors

import "embed"

// ErrorDocs contains all embedded error documentation markdown files.
//
//go:embed *.md
var ErrorDocs embed.FS

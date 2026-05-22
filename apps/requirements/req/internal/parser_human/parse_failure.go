package parser_human

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"

// ParseFailure records a single .class file that failed to parse.
//
// The class is still represented in the model as an empty placeholder, so the
// rest of the model renders normally and the class still lists (unpopulated)
// on its domain page. This record carries the error message so the class's own
// generated page can show it instead of the empty placeholder content.
type ParseFailure struct {
	ClassKey identity.Key // Key of the class that failed to parse.
	Name     string       // Display name of the class.
	Path     string       // Source file path, relative to the model root.
	Err      string       // The parse error message.
}

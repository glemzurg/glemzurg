// Package annot reads PostgreSQL annotation files that sit beside a requirements JSON tree.
// The requirements parser leaves these files untouched. This package is the parser that understands them.
package annot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// Suffix is the filename ending of a PostgreSQL annotation.
// The requirements JSON tree uses the same ending for siblings it does not read.
const Suffix = ".annot.json"

// identPattern matches an unquoted PostgreSQL identifier.
var identPattern = regexp.MustCompile(`^[a-z_][a-z0-9_]*$`)

// onDeleteActions are the foreign-key delete actions PostgreSQL accepts.
var onDeleteActions = map[string]struct{}{
	"CASCADE":     {},
	"RESTRICT":    {},
	"SET NULL":    {},
	"SET DEFAULT": {},
	"NO ACTION":   {},
}

// Note is the PostgreSQL compilation note for one model file.
// RelPath is the model file it describes, relative to the tree root, with slash separators.
type Note struct {
	RelPath    string               `json:"-"`
	Schema     string               `json:"schema,omitempty"`
	Include    *bool                `json:"include,omitempty"`
	OnDelete   string               `json:"on_delete,omitempty"`
	Attributes map[string]Attribute `json:"attributes,omitempty"`
}

// Attribute is the PostgreSQL note for one class attribute.
// An empty object means the attribute is included with compiler defaults.
type Attribute struct {
	Type    string `json:"type,omitempty"`
	Default string `json:"default,omitempty"`
	Include *bool  `json:"include,omitempty"`
}

func parseNote(relPath string, content []byte) (Note, error) {
	dec := json.NewDecoder(bytes.NewReader(content))
	dec.DisallowUnknownFields()
	var note Note
	if err := dec.Decode(&note); err != nil {
		return Note{}, fmt.Errorf("%s: %w", relPath, err)
	}
	if dec.More() {
		return Note{}, fmt.Errorf("%s: extra JSON value", relPath)
	}
	note.RelPath = relPath
	if err := note.validate(); err != nil {
		return Note{}, fmt.Errorf("%s: %w", relPath, err)
	}
	return note, nil
}

func (n Note) validate() error {
	if err := validateSchema(n); err != nil {
		return err
	}
	if err := validateOnDelete(n); err != nil {
		return err
	}
	return validateAttributes(n)
}

func validateSchema(n Note) error {
	if n.Schema == "" {
		return nil
	}
	if n.RelPath != "model.json" {
		return fmt.Errorf("schema is only valid on model.json")
	}
	if !identPattern.MatchString(n.Schema) {
		return fmt.Errorf("schema %q is not a PostgreSQL identifier", n.Schema)
	}
	return nil
}

func validateOnDelete(n Note) error {
	if n.OnDelete == "" {
		return nil
	}
	if !strings.HasSuffix(n.RelPath, ".assoc.json") {
		return fmt.Errorf("on_delete is only valid on an association file")
	}
	if _, ok := onDeleteActions[n.OnDelete]; !ok {
		return fmt.Errorf("on_delete %q is not a PostgreSQL delete action", n.OnDelete)
	}
	return nil
}

func validateAttributes(n Note) error {
	if len(n.Attributes) == 0 {
		return nil
	}
	if !strings.HasSuffix(n.RelPath, "/class.json") {
		return fmt.Errorf("attributes are only valid on class.json")
	}
	for key, attr := range n.Attributes {
		if !identPattern.MatchString(key) {
			return fmt.Errorf("attribute key %q is not an identifier", key)
		}
		if err := attr.validate(); err != nil {
			return fmt.Errorf("attribute %s: %w", key, err)
		}
	}
	return nil
}

func (a Attribute) validate() error {
	if a.Type != strings.TrimSpace(a.Type) || strings.ContainsAny(a.Type, "\n\r") {
		return fmt.Errorf("type %q is not a PostgreSQL type", a.Type)
	}
	if a.Default != "" && strings.TrimSpace(a.Default) == "" {
		return fmt.Errorf("default is empty")
	}
	return nil
}

package parser_ai

import (
	"fmt"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_ai/docs"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_ai/errors"
)

// InternalErrorPrefix is the prefix used in panic messages when error documentation
// cannot be loaded. This indicates an internal failure that cannot be resolved by
// altering input data.
const InternalErrorPrefix = "INTERNAL ERROR: "

// formatDocs is the cached content of JSON_AI_MODEL_FORMAT.md
var formatDocs string

func init() {
	content, err := docs.Docs.ReadFile("JSON_AI_MODEL_FORMAT.md")
	if err != nil {
		panic("failed to read JSON_AI_MODEL_FORMAT.md: " + err.Error())
	}
	formatDocs = string(content)
}

// ParseError represents a validation error during AI input parsing.
// It includes a unique error number, detailed advice, and optional attachments.
type ParseError struct {
	Code        int    // Unique error number
	Message     string // Human-readable error message
	ErrorFile   string // Name of the markdown file in errors/ with detailed error info
	ErrorDetail string // Content of the error markdown file
	Schema      string // The JSON schema content (if applicable)
	Docs        string // The JSON_AI_MODEL_FORMAT.md content (if applicable)
	File        string // The JSON file being parsed where the error occurred
	Field       string // JSON path to the field that caused the error (optional), e.g., "name", "attributes.myAttr.name"
}

// Error implements the error interface.
// Returns a comprehensive error block with all available information.
func (e *ParseError) Error() string {
	var b strings.Builder

	// Error number and message
	fmt.Fprintf(&b, "=== ERROR E%d ===\n", e.Code)
	fmt.Fprintf(&b, "Message: %s\n", e.Message)

	// File
	fmt.Fprintf(&b, "File: %s\n", e.File)

	// Field (if set)
	if e.Field != "" {
		fmt.Fprintf(&b, "Field: %s\n", e.Field)
	}

	// Error detail (from markdown file)
	b.WriteString("\n--- Error Detail ---\n")
	b.WriteString(e.ErrorDetail)
	if !strings.HasSuffix(e.ErrorDetail, "\n") {
		b.WriteString("\n")
	}

	// Schema (if set)
	if e.Schema != "" {
		b.WriteString("\n--- Schema ---\n")
		b.WriteString(e.Schema)
		if !strings.HasSuffix(e.Schema, "\n") {
			b.WriteString("\n")
		}
	}

	// Docs
	b.WriteString("\n--- Format Documentation ---\n")
	b.WriteString(e.Docs)
	if !strings.HasSuffix(e.Docs, "\n") {
		b.WriteString("\n")
	}

	return b.String()
}

// NewParseError creates a new ParseError with the given code, message, and file.
// The file parameter is the JSON file being parsed where the error occurred.
// It automatically loads the error documentation file for the given error code
// and attaches the format documentation (JSON_AI_MODEL_FORMAT.md) to all errors.
// Panics if no error documentation exists for the code - this indicates an internal
// error that cannot be resolved by altering input data.
func NewParseError(code int, message string, file string) *ParseError {
	content, filename, err := errors.LoadErrorDoc(code)
	if err != nil {
		panic(fmt.Sprintf("%sno error documentation for error code %d. This is an internal failure that no alteration of input will resolve.", InternalErrorPrefix, code))
	}
	return &ParseError{
		Code:        code,
		Message:     message,
		ErrorFile:   filename,
		ErrorDetail: content,
		Docs:        formatDocs,
		File:        file,
	}
}

// WithSchema returns a copy of the error with the schema content attached.
func (e *ParseError) WithSchema(schemaContent string) *ParseError {
	return &ParseError{
		Code:        e.Code,
		Message:     e.Message,
		ErrorFile:   e.ErrorFile,
		ErrorDetail: e.ErrorDetail,
		Schema:      schemaContent,
		Docs:        e.Docs,
		File:        e.File,
		Field:       e.Field,
	}
}

// WithField returns a copy of the error with the field path set.
// The field supports JSON path notation for nested fields:
//   - Top-level: "name"
//   - Nested object: "attributes.myAttr.name"
//   - Array index: "indexes.0" or "items[0]"
func (e *ParseError) WithField(field string) *ParseError {
	return &ParseError{
		Code:        e.Code,
		Message:     e.Message,
		ErrorFile:   e.ErrorFile,
		ErrorDetail: e.ErrorDetail,
		Schema:      e.Schema,
		Docs:        e.Docs,
		File:        e.File,
		Field:       field,
	}
}

package parser_ai

import (
	"fmt"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_ai/errors"
)

// InternalErrorPrefix is the prefix used in panic messages when error documentation
// cannot be loaded. This indicates an internal failure that cannot be resolved by
// altering input data.
const InternalErrorPrefix = "INTERNAL ERROR: "

// ParseError represents a validation error during AI input parsing.
// It includes a unique error number and concise remediation guidance.
// Full error documentation is available via the VerboseError() method
// or the req_check --explain flag.
type ParseError struct {
	Code        int    // Unique error number
	Message     string // Human-readable error message
	ErrorFile   string // Name of the markdown file in errors/ with detailed error info
	ErrorDetail string // Content of the error markdown file (used by VerboseError and --explain)
	Schema      string // The JSON schema content (used by VerboseError and --explain)
	File        string // The JSON file being parsed where the error occurred
	Field       string // JSON path to the field that caused the error (optional), e.g., "name", "attributes.myAttr.name"
	Hint        string // Concise remediation hint (1-3 lines)
}

// Error implements the error interface.
// Returns a concise error block optimized for AI agent consumption.
func (e *ParseError) Error() string {
	var b strings.Builder

	fmt.Fprintf(&b, "E%d: %s", e.Code, e.Message)
	fmt.Fprintf(&b, "\n  file: %s", e.File)
	if e.Field != "" {
		fmt.Fprintf(&b, "\n  field: %s", e.Field)
	}
	if e.Hint != "" {
		fmt.Fprintf(&b, "\n  hint: %s", e.Hint)
	}

	return b.String()
}

// VerboseError returns a comprehensive error block with full documentation.
// Used by req_check --explain to show detailed remediation guidance.
func (e *ParseError) VerboseError() string {
	var b strings.Builder

	// Start with the concise error.
	b.WriteString(e.Error())
	b.WriteString("\n")

	// Error detail (from markdown file).
	if e.ErrorDetail != "" {
		b.WriteString("\n--- Error Detail ---\n")
		b.WriteString(e.ErrorDetail)
		if !strings.HasSuffix(e.ErrorDetail, "\n") {
			b.WriteString("\n")
		}
	}

	// Schema (if set).
	if e.Schema != "" {
		b.WriteString("\n--- Schema ---\n")
		b.WriteString(e.Schema)
		if !strings.HasSuffix(e.Schema, "\n") {
			b.WriteString("\n")
		}
	}

	return b.String()
}

// NewParseError creates a new ParseError with the given code, message, and file.
// The file parameter is the JSON file being parsed where the error occurred.
// It automatically loads the error documentation file for the given error code.
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
		File:        e.File,
		Field:       e.Field,
		Hint:        e.Hint,
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
		File:        e.File,
		Field:       field,
		Hint:        e.Hint,
	}
}

// WithHint returns a copy of the error with a concise remediation hint.
// The hint should be 1-3 lines of actionable guidance.
func (e *ParseError) WithHint(hint string) *ParseError {
	return &ParseError{
		Code:        e.Code,
		Message:     e.Message,
		ErrorFile:   e.ErrorFile,
		ErrorDetail: e.ErrorDetail,
		Schema:      e.Schema,
		File:        e.File,
		Field:       e.Field,
		Hint:        hint,
	}
}

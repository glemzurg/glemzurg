package parser_ai_input

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_ai_input/docs"
)

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
	Code      int    // Unique error number
	Message   string // Human-readable error message
	ErrorFile string // Name of the markdown file in errors/ with detailed error info
	Schema    string // The JSON schema content (if applicable)
	Docs      string // The JSON_AI_MODEL_FORMAT.md content (if applicable)
	File      string // File where the error occurred (optional)
	Field     string // Field name that caused the error (optional)
}

// Error implements the error interface.
func (e *ParseError) Error() string {
	if e.File != "" && e.Field != "" {
		return fmt.Sprintf("[E%d] %s (file: %s, field: %s)", e.Code, e.Message, e.File, e.Field)
	}
	if e.File != "" {
		return fmt.Sprintf("[E%d] %s (file: %s)", e.Code, e.Message, e.File)
	}
	if e.Field != "" {
		return fmt.Sprintf("[E%d] %s (field: %s)", e.Code, e.Message, e.Field)
	}
	return fmt.Sprintf("[E%d] %s", e.Code, e.Message)
}

// NewParseError creates a new ParseError with the given code, message, and error file.
func NewParseError(code int, message, errorFile string) *ParseError {
	return &ParseError{
		Code:      code,
		Message:   message,
		ErrorFile: errorFile,
	}
}

// WithSchema returns a copy of the error with the schema content attached.
func (e *ParseError) WithSchema(schemaContent string) *ParseError {
	return &ParseError{
		Code:      e.Code,
		Message:   e.Message,
		ErrorFile: e.ErrorFile,
		Schema:    schemaContent,
		Docs:      e.Docs,
		File:      e.File,
		Field:     e.Field,
	}
}

// WithDocs returns a copy of the error with the JSON_AI_MODEL_FORMAT.md content attached.
func (e *ParseError) WithDocs() *ParseError {
	return &ParseError{
		Code:      e.Code,
		Message:   e.Message,
		ErrorFile: e.ErrorFile,
		Schema:    e.Schema,
		Docs:      formatDocs,
		File:      e.File,
		Field:     e.Field,
	}
}

// WithFile returns a copy of the error with the file field set.
func (e *ParseError) WithFile(file string) *ParseError {
	return &ParseError{
		Code:      e.Code,
		Message:   e.Message,
		ErrorFile: e.ErrorFile,
		Schema:    e.Schema,
		Docs:      e.Docs,
		File:      file,
		Field:     e.Field,
	}
}

// WithField returns a copy of the error with the field field set.
func (e *ParseError) WithField(field string) *ParseError {
	return &ParseError{
		Code:      e.Code,
		Message:   e.Message,
		ErrorFile: e.ErrorFile,
		Schema:    e.Schema,
		Docs:      e.Docs,
		File:      e.File,
		Field:     field,
	}
}

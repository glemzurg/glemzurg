package parser_ai_input

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_ai_input/docs"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_ai_input/errors"
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
	File        string // File where the error occurred (optional)
	Field       string // Field name that caused the error (optional)
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

// NewParseError creates a new ParseError with the given code and message.
// It automatically loads the error documentation file for the given error code
// and attaches the format documentation (JSON_AI_MODEL_FORMAT.md) to all errors.
// Panics if no error documentation exists for the code - this indicates an internal
// error that cannot be resolved by altering input data.
func NewParseError(code int, message string) *ParseError {
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

// WithFile returns a copy of the error with the file field set.
func (e *ParseError) WithFile(file string) *ParseError {
	return &ParseError{
		Code:        e.Code,
		Message:     e.Message,
		ErrorFile:   e.ErrorFile,
		ErrorDetail: e.ErrorDetail,
		Schema:      e.Schema,
		Docs:        e.Docs,
		File:        file,
		Field:       e.Field,
	}
}

// WithField returns a copy of the error with the field field set.
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

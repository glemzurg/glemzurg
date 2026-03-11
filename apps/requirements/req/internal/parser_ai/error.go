package parser_ai

import (
	"fmt"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_ai/errors"
)

// InternalErrorPrefix is the prefix used in panic messages when error documentation
// cannot be loaded. This indicates an internal failure that cannot be resolved by
// altering input data.
const InternalErrorPrefix = "INTERNAL ERROR: "

// CoreValidationDetail is an exported mirror of coreerr.ValidationError.
// It carries all structured data from core model validation so that the
// calling AI receives pre-parsed JSON without needing to parse strings.
type CoreValidationDetail struct {
	Code    string                `json:"code"`            // Core error code, e.g. "CLASS_KEY_TYPE_INVALID".
	Message string                `json:"message"`         // Human-readable description of what went wrong.
	Path    []coreerr.PathSegment `json:"path,omitempty"`  // Location in the model tree.
	Field   string                `json:"field,omitempty"` // The specific field that failed validation.
	Got     string                `json:"got,omitempty"`   // The invalid value that was provided.
	Want    string                `json:"want,omitempty"`  // What valid values look like.
}

// ParseError represents a validation error during AI input parsing.
// It includes a unique error number and concise remediation guidance.
// Full error documentation is available via the VerboseError() method
// or the req_check --explain flag.
type ParseError struct {
	Code        int                   // Unique error number
	Message     string                // Human-readable error message
	ErrorFile   string                // Name of the markdown file in errors/ with detailed error info
	ErrorDetail string                // Content of the error markdown file (used by VerboseError and --explain)
	File        string                // The JSON file being parsed where the error occurred
	Field       string                // JSON path to the field that caused the error (optional), e.g., "name", "attributes.myAttr.name"
	Hint        string                // Concise remediation hint (1-3 lines)
	Context     *CoreValidationDetail // Structured core validation detail (nil for non-core errors).
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
	if e.Context != nil {
		if len(e.Context.Path) > 0 {
			fmt.Fprintf(&b, "\n  context: %s", coreerr.FormatPath(e.Context.Path))
		}
		if e.Context.Got != "" {
			fmt.Fprintf(&b, "\n  got: %s", e.Context.Got)
		}
		if e.Context.Want != "" {
			fmt.Fprintf(&b, "\n  want: %s", e.Context.Want)
		}
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

// WithField returns a copy of the error with the field path set.
// The field supports JSON path notation for nested fields:
//   - Top-level: "name"
//   - Nested object: "attributes.myAttr.name"
//   - Array index: "indexes.0" or "items[0]"
func (e *ParseError) WithField(field string) *ParseError {
	c := e.copy()
	c.Field = field
	return c
}

// WithHint returns a copy of the error with a concise remediation hint.
// The hint should be 1-3 lines of actionable guidance.
func (e *ParseError) WithHint(hint string) *ParseError {
	c := e.copy()
	c.Hint = hint
	return c
}

// WithCoreValidation returns a copy of the error with the full core validation detail populated.
func (e *ParseError) WithCoreValidation(ve *coreerr.ValidationError) *ParseError {
	c := e.copy()
	c.Context = &CoreValidationDetail{
		Code:    string(ve.Code()),
		Message: ve.Message(),
		Path:    ve.Path(),
		Field:   ve.Field(),
		Got:     ve.Got(),
		Want:    ve.Want(),
	}
	return c
}

// copy returns a shallow copy of the ParseError.
func (e *ParseError) copy() *ParseError {
	cp := *e
	return &cp
}

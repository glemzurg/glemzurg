package coreerr

import (
	"fmt"
	"strings"
)

// Code is a unique identifier for a validation failure.
// Codes are string constants like "ACTION_NAME_REQUIRED".
type Code string

// PathSegment is one level in the model tree.
type PathSegment struct {
	Entity string // "model", "domain", "subdomain", "class", "action", etc.
	Key    string // The identity key string or index.
}

// ValidationError is the structured error type for all core validation failures.
type ValidationError struct {
	code    Code          // Unique stable identifier, e.g., "CLASS_KEY_TYPE_INVALID".
	message string        // Human-readable description of what went wrong.
	path    []PathSegment // Location in the model tree where the error occurred.
	field   string        // The specific field that failed validation.
	got     string        // The invalid value that was provided (stringified).
	want    string        // What valid values look like (e.g., "one of: person, system").
}

// New creates a ValidationError with code, message, and field.
func New(code Code, message, field string) *ValidationError {
	return &ValidationError{
		code:    code,
		message: message,
		field:   field,
	}
}

// NewWithValues creates a ValidationError with code, message, field, got, and want.
func NewWithValues(code Code, message, field, got, want string) *ValidationError {
	return &ValidationError{
		code:    code,
		message: message,
		field:   field,
		got:     got,
		want:    want,
	}
}

// NewWithPath creates a ValidationError with code, message, path, and field.
func NewWithPath(code Code, message string, path []PathSegment, field string) *ValidationError {
	return &ValidationError{
		code:    code,
		message: message,
		path:    path,
		field:   field,
	}
}

// Code returns the error code.
func (e *ValidationError) Code() Code { return e.code }

// Message returns the human-readable description.
func (e *ValidationError) Message() string { return e.message }

// Path returns the location in the model tree.
func (e *ValidationError) Path() []PathSegment { return e.path }

// Field returns the specific field that failed validation.
func (e *ValidationError) Field() string { return e.field }

// Got returns the invalid value that was provided.
func (e *ValidationError) Got() string { return e.got }

// Want returns what valid values look like.
func (e *ValidationError) Want() string { return e.want }

// Error implements the error interface.
func (e *ValidationError) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "[%s] %s", e.code, e.message)
	if e.field != "" {
		fmt.Fprintf(&b, " (field: %s", e.field)
		if e.got != "" {
			fmt.Fprintf(&b, ", got: %q", e.got)
		}
		if e.want != "" {
			fmt.Fprintf(&b, ", want: %s", e.want)
		}
		b.WriteString(")")
	}
	if len(e.path) > 0 {
		fmt.Fprintf(&b, " at %s", FormatPath(e.path))
	}
	return b.String()
}

// Is allows errors.Is matching by Code alone.
func (e *ValidationError) Is(target error) bool {
	if t, ok := target.(*ValidationError); ok {
		return t.code == e.code
	}
	return false
}

// FormatPath returns a dotted representation of a path.
// Example: "model.domains[domain1].subdomains[default].classes[order]".
func FormatPath(path []PathSegment) string {
	var b strings.Builder
	for i, seg := range path {
		if i > 0 {
			b.WriteString(".")
		}
		b.WriteString(seg.Entity)
		if seg.Key != "" {
			fmt.Fprintf(&b, "[%s]", seg.Key)
		}
	}
	return b.String()
}

// ValidationContext carries the current position in the model tree during validation.
type ValidationContext struct {
	path []PathSegment
}

// NewContext creates a root validation context with the given entity and key.
func NewContext(entity, key string) *ValidationContext {
	return &ValidationContext{
		path: []PathSegment{{Entity: entity, Key: key}},
	}
}

// Child returns a new ValidationContext with an additional path segment.
func (vc *ValidationContext) Child(entity, key string) *ValidationContext {
	newPath := make([]PathSegment, len(vc.path)+1)
	copy(newPath, vc.path)
	newPath[len(vc.path)] = PathSegment{Entity: entity, Key: key}
	return &ValidationContext{path: newPath}
}

// Err creates a new ValidationError at the current context path.
func (vc *ValidationContext) Err(code Code, field, got, want, message string) *ValidationError {
	return &ValidationError{
		code:    code,
		message: message,
		path:    vc.path,
		field:   field,
		got:     got,
		want:    want,
	}
}

// ContextPath returns the current path segments.
func (vc *ValidationContext) ContextPath() []PathSegment {
	return vc.path
}

// EnsureContext returns the given context if non-nil, otherwise creates
// a new root context with the given entity and key. This allows Validate
// methods to be called with or without a context.
func EnsureContext(ctx *ValidationContext, entity, key string) *ValidationContext {
	if ctx != nil {
		return ctx
	}
	return NewContext(entity, key)
}

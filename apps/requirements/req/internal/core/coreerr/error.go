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

// New creates a ValidationError with context, code, message, and field.
// The context provides the location in the model tree. It must not be nil.
// Panics if ctx is nil, code is empty, or message is empty — these are programming errors.
func New(ctx *ValidationContext, code Code, message, field string) *ValidationError {
	requireContext(ctx)
	requireCodeAndMessage(code, message)
	return &ValidationError{
		code:    code,
		message: message,
		path:    ctx.path,
		field:   field,
	}
}

// NewWithValues creates a ValidationError with context, code, message, field, got, and want.
// The context provides the location in the model tree. It must not be nil.
// Panics if ctx is nil, code is empty, or message is empty — these are programming errors.
func NewWithValues(ctx *ValidationContext, code Code, message, field, got, want string) *ValidationError {
	requireContext(ctx)
	requireCodeAndMessage(code, message)
	return &ValidationError{
		code:    code,
		message: message,
		path:    ctx.path,
		field:   field,
		got:     got,
		want:    want,
	}
}

// requireContext panics if ctx is nil.
// An error without a context is always a programming bug.
func requireContext(ctx *ValidationContext) {
	if ctx == nil {
		panic("coreerr: context must not be nil")
	}
}

// requireCodeAndMessage panics if code or message is empty.
// An error without a code or message is always a programming bug.
func requireCodeAndMessage(code Code, message string) {
	if code == "" {
		panic("coreerr: code must not be empty")
	}
	if message == "" {
		panic("coreerr: message must not be empty")
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

// ContextPath returns the current path segments.
func (vc *ValidationContext) ContextPath() []PathSegment {
	return vc.path
}

// ValidateNameChars checks that name contains only A-Za-z0-9, space, hyphen, and underscore.
// Returns the first invalid character found, or empty string if valid.
func ValidateNameChars(name string) string {
	for _, r := range name {
		if (r < 'A' || r > 'Z') && (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != ' ' && r != '-' && r != '_' {
			return string(r)
		}
	}
	return ""
}

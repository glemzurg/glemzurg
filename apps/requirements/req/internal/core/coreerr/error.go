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
	Code    Code          // Unique stable identifier, e.g., "CLASS_KEY_TYPE_INVALID".
	Message string        // Human-readable description of what went wrong.
	Path    []PathSegment // Location in the model tree where the error occurred.
	Field   string        // The specific field that failed validation.
	Got     string        // The invalid value that was provided (stringified).
	Want    string        // What valid values look like (e.g., "one of: person, system").
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "[%s] %s", e.Code, e.Message)
	if e.Field != "" {
		fmt.Fprintf(&b, " (field: %s", e.Field)
		if e.Got != "" {
			fmt.Fprintf(&b, ", got: %q", e.Got)
		}
		if e.Want != "" {
			fmt.Fprintf(&b, ", want: %s", e.Want)
		}
		b.WriteString(")")
	}
	if len(e.Path) > 0 {
		fmt.Fprintf(&b, " at %s", FormatPath(e.Path))
	}
	return b.String()
}

// Is allows errors.Is matching by Code alone.
func (e *ValidationError) Is(target error) bool {
	if t, ok := target.(*ValidationError); ok {
		return t.Code == e.Code
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
		Code:    code,
		Message: message,
		Path:    vc.path,
		Field:   field,
		Got:     got,
		Want:    want,
	}
}

// Path returns the current path segments.
func (vc *ValidationContext) Path() []PathSegment {
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

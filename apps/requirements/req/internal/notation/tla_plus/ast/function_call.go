package ast

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
)

// FunctionCall represents a function call with optional scope path.
// Patterns supported:
//   - Built-in module: _Module!FunctionName(args...) e.g., _Seq!Len(seq)
//   - Global function: _FunctionName(args...) - leading underscore, no scope path
//   - Class-scoped action (full): Domain!Subdomain!Class!ActionName(args...)
//   - Class-scoped action (from domain): Subdomain!Class!ActionName(args...)
//   - Class-scoped action (from subdomain): Class!ActionName(args...)
//   - Class-scoped action (current class): ActionName(args...) - no underscore, no scope
//
// Leading underscore distinguishes global/built-in from class-scoped:
//   - _Name or _Module!Name = global/built-in
//   - Name or Scope!Name = class-scoped action
type FunctionCall struct {
	// ScopePath contains the scope segments before the function name.
	// For "_Seq!Len", ScopePath = ["_Seq"]
	// For "Domain!Subdomain!Class!Action", ScopePath = ["Domain", "Subdomain", "Class"]
	// For "Action", ScopePath = [] (empty)
	ScopePath []*Identifier

	// Name is the function/action name (required)
	Name *Identifier `validate:"required"`

	// Args are the function arguments (can be empty)
	Args []Expression
}

func (f *FunctionCall) expressionNode() {}

// FullName returns the complete scoped name as a string (e.g., "Domain!Subdomain!Class!Action").
func (f *FunctionCall) FullName() string {
	parts := make([]string, 0, len(f.ScopePath)+1)
	for _, seg := range f.ScopePath {
		parts = append(parts, seg.Value)
	}
	parts = append(parts, f.Name.Value)
	return strings.Join(parts, "!")
}

// IsGlobalOrBuiltin returns true if this is a global function or built-in module call.
// Global/built-in calls are distinguished by a leading underscore:
//   - _FunctionName() - global function
//   - _Module!FunctionName() - built-in module function
func (f *FunctionCall) IsGlobalOrBuiltin() bool {
	if len(f.ScopePath) > 0 {
		// Check first scope segment for underscore (e.g., _Seq!Len)
		return strings.HasPrefix(f.ScopePath[0].Value, "_")
	}
	// Check function name for underscore (e.g., _GlobalFunc)
	return strings.HasPrefix(f.Name.Value, "_")
}

// IsSystemEvent reports whether this call is a reserved system event constructor.
// Canonical TLA uses guillemets («new»); ASCII authoring uses a leading underscore (_new).
func (f *FunctionCall) IsSystemEvent() bool {
	return len(f.ScopePath) == 0 && model_state.IsSystemEventTLAName(f.Name.Value)
}

func (f *FunctionCall) String() string {
	var out bytes.Buffer
	for _, seg := range f.ScopePath {
		out.WriteString(seg.String())
		out.WriteString("!")
	}
	out.WriteString(f.Name.String())
	out.WriteString("(")
	for i, arg := range f.Args {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(arg.String())
	}
	out.WriteString(")")
	return out.String()
}

var systemEventASCII = map[string]string{
	model_state.EventTLANameNew:     model_state.EventNameNew,
	model_state.EventTLANameDestroy: model_state.EventNameDestroy,
}

func (f *FunctionCall) ASCII() string {
	var out bytes.Buffer
	for _, seg := range f.ScopePath {
		out.WriteString(seg.ASCII())
		out.WriteString("!")
	}
	name := f.Name.ASCII()
	if ascii, ok := systemEventASCII[name]; ok {
		name = ascii
	}
	out.WriteString(name)
	out.WriteString("(")
	for i, arg := range f.Args {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(arg.ASCII())
	}
	out.WriteString(")")
	return out.String()
}

func (f *FunctionCall) Validate() error {
	if err := _validate.Struct(f); err != nil {
		return err
	}
	for i, seg := range f.ScopePath {
		if seg == nil {
			return fmt.Errorf("ScopePath[%d]: is nil", i)
		}
		if err := seg.Validate(); err != nil {
			return fmt.Errorf("ScopePath[%d]: %w", i, err)
		}
	}
	if err := f.Name.Validate(); err != nil {
		return fmt.Errorf("name: %w", err)
	}
	for i, arg := range f.Args {
		if arg == nil {
			return fmt.Errorf("args[%d]: is nil", i)
		}
		if err := arg.Validate(); err != nil {
			return fmt.Errorf("args[%d]: %w", i, err)
		}
	}
	return nil
}

// Package types defines the TLA+ type system.
//
// This package represents types in the TLA+ language, separate from Go's type system.
// Types are inferred and checked by the typechecker package before evaluation.
//
// The type system follows Hindley-Milner with:
// - Concrete types: Boolean, Number, String, Set, Tuple, Record, Bag, Function
// - Type variables: For polymorphic inference
// - Type schemes: For let-polymorphism (∀a. Type)
package types

import (
	"bytes"
	"fmt"
	"sort"
)

// Type represents a TLA+ type.
type Type interface {
	// String returns the type in a human-readable format.
	String() string

	// Equals checks structural equality with another type.
	Equals(Type) bool

	// FreeTypeVars returns the set of free type variable IDs in this type.
	FreeTypeVars() map[int]struct{}

	// substitute replaces type variables according to the substitution.
	substitute(subst Substitution) Type
}

// Substitution maps type variable IDs to their resolved types.
type Substitution map[int]Type

// Apply applies the substitution to a type.
func (s Substitution) Apply(t Type) Type {
	if t == nil {
		return nil
	}
	return t.substitute(s)
}

// Compose combines two substitutions: s1 then s2.
// Returns a new substitution that applies s1 first, then s2.
func (s Substitution) Compose(other Substitution) Substitution {
	result := make(Substitution)

	// Apply other to all types in s
	for id, t := range s {
		result[id] = other.Apply(t)
	}

	// Add bindings from other that aren't in s
	for id, t := range other {
		if _, exists := result[id]; !exists {
			result[id] = t
		}
	}

	return result
}

// ----------------------------------------------------------------------------
// Concrete Types
// ----------------------------------------------------------------------------

// Boolean represents the TLA+ BOOLEAN type.
type Boolean struct{}

func (Boolean) String() string                     { return "Boolean" }
func (Boolean) Equals(t Type) bool                 { _, ok := t.(Boolean); return ok }
func (Boolean) FreeTypeVars() map[int]struct{}     { return nil }
func (b Boolean) substitute(_ Substitution) Type   { return b }

// Number represents TLA+ numeric types (Natural, Integer, Real).
// TLA+ treats all numbers uniformly for most operations.
type Number struct{}

func (Number) String() string                     { return "Number" }
func (Number) Equals(t Type) bool                 { _, ok := t.(Number); return ok }
func (Number) FreeTypeVars() map[int]struct{}     { return nil }
func (n Number) substitute(_ Substitution) Type   { return n }

// String represents the TLA+ STRING type.
type String struct{}

func (String) String() string                     { return "String" }
func (String) Equals(t Type) bool                 { _, ok := t.(String); return ok }
func (String) FreeTypeVars() map[int]struct{}     { return nil }
func (s String) substitute(_ Substitution) Type   { return s }

// Set represents a TLA+ set type with a specific element type.
type Set struct {
	Element Type
}

func (s Set) String() string {
	return fmt.Sprintf("Set[%s]", s.Element)
}

func (s Set) Equals(t Type) bool {
	other, ok := t.(Set)
	if !ok {
		return false
	}
	return s.Element.Equals(other.Element)
}

func (s Set) FreeTypeVars() map[int]struct{} {
	return s.Element.FreeTypeVars()
}

func (s Set) substitute(subst Substitution) Type {
	return Set{Element: subst.Apply(s.Element)}
}

// Tuple represents a TLA+ tuple/sequence type.
// In TLA+, sequences are homogeneous, so we use a single element type.
type Tuple struct {
	Element Type
}

func (t Tuple) String() string {
	return fmt.Sprintf("Tuple[%s]", t.Element)
}

func (t Tuple) Equals(other Type) bool {
	o, ok := other.(Tuple)
	if !ok {
		return false
	}
	return t.Element.Equals(o.Element)
}

func (t Tuple) FreeTypeVars() map[int]struct{} {
	return t.Element.FreeTypeVars()
}

func (t Tuple) substitute(subst Substitution) Type {
	return Tuple{Element: subst.Apply(t.Element)}
}

// Record represents a TLA+ record type with named fields.
type Record struct {
	Fields map[string]Type
}

func (r Record) String() string {
	if len(r.Fields) == 0 {
		return "Record{}"
	}

	// Sort field names for consistent output
	names := make([]string, 0, len(r.Fields))
	for name := range r.Fields {
		names = append(names, name)
	}
	sort.Strings(names)

	var buf bytes.Buffer
	buf.WriteString("Record{")
	for i, name := range names {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(name)
		buf.WriteString(": ")
		buf.WriteString(r.Fields[name].String())
	}
	buf.WriteString("}")
	return buf.String()
}

func (r Record) Equals(t Type) bool {
	other, ok := t.(Record)
	if !ok {
		return false
	}
	if len(r.Fields) != len(other.Fields) {
		return false
	}
	for name, ft := range r.Fields {
		ot, exists := other.Fields[name]
		if !exists || !ft.Equals(ot) {
			return false
		}
	}
	return true
}

func (r Record) FreeTypeVars() map[int]struct{} {
	result := make(map[int]struct{})
	for _, ft := range r.Fields {
		for id := range ft.FreeTypeVars() {
			result[id] = struct{}{}
		}
	}
	return result
}

func (r Record) substitute(subst Substitution) Type {
	newFields := make(map[string]Type, len(r.Fields))
	for name, ft := range r.Fields {
		newFields[name] = subst.Apply(ft)
	}
	return Record{Fields: newFields}
}

// Bag represents a TLA+ bag (multiset) type.
type Bag struct {
	Element Type
}

func (b Bag) String() string {
	return fmt.Sprintf("Bag[%s]", b.Element)
}

func (b Bag) Equals(t Type) bool {
	other, ok := t.(Bag)
	if !ok {
		return false
	}
	return b.Element.Equals(other.Element)
}

func (b Bag) FreeTypeVars() map[int]struct{} {
	return b.Element.FreeTypeVars()
}

func (b Bag) substitute(subst Substitution) Type {
	return Bag{Element: subst.Apply(b.Element)}
}

// Function represents a TLA+ function type (domain → range).
type Function struct {
	Params []Type // Parameter types
	Return Type   // Return type
}

func (f Function) String() string {
	if len(f.Params) == 0 {
		return fmt.Sprintf("() → %s", f.Return)
	}
	if len(f.Params) == 1 {
		return fmt.Sprintf("%s → %s", f.Params[0], f.Return)
	}

	var buf bytes.Buffer
	buf.WriteString("(")
	for i, p := range f.Params {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(p.String())
	}
	buf.WriteString(") → ")
	buf.WriteString(f.Return.String())
	return buf.String()
}

func (f Function) Equals(t Type) bool {
	other, ok := t.(Function)
	if !ok {
		return false
	}
	if len(f.Params) != len(other.Params) {
		return false
	}
	for i, p := range f.Params {
		if !p.Equals(other.Params[i]) {
			return false
		}
	}
	return f.Return.Equals(other.Return)
}

func (f Function) FreeTypeVars() map[int]struct{} {
	result := make(map[int]struct{})
	for _, p := range f.Params {
		for id := range p.FreeTypeVars() {
			result[id] = struct{}{}
		}
	}
	for id := range f.Return.FreeTypeVars() {
		result[id] = struct{}{}
	}
	return result
}

func (f Function) substitute(subst Substitution) Type {
	newParams := make([]Type, len(f.Params))
	for i, p := range f.Params {
		newParams[i] = subst.Apply(p)
	}
	return Function{Params: newParams, Return: subst.Apply(f.Return)}
}

// ----------------------------------------------------------------------------
// Type Variables (for polymorphism)
// ----------------------------------------------------------------------------

// TypeVar represents a type variable used during type inference.
// Type variables are unified with concrete types during inference.
type TypeVar struct {
	ID   int    // Unique identifier
	Name string // Display name (e.g., "a", "b")
}

func (tv TypeVar) String() string {
	if tv.Name != "" {
		return tv.Name
	}
	return fmt.Sprintf("t%d", tv.ID)
}

func (tv TypeVar) Equals(t Type) bool {
	other, ok := t.(TypeVar)
	if !ok {
		return false
	}
	return tv.ID == other.ID
}

func (tv TypeVar) FreeTypeVars() map[int]struct{} {
	return map[int]struct{}{tv.ID: {}}
}

func (tv TypeVar) substitute(subst Substitution) Type {
	if resolved, ok := subst[tv.ID]; ok {
		// Continue substituting in case the resolved type contains more variables
		return resolved.substitute(subst)
	}
	return tv
}

// ----------------------------------------------------------------------------
// Type Schemes (for let-polymorphism)
// ----------------------------------------------------------------------------

// Scheme represents a polymorphic type scheme: ∀a b c. Type
// The TypeVars are bound (quantified) in the Type.
type Scheme struct {
	TypeVars []int // IDs of bound type variables
	Type     Type  // The polymorphic type body
}

func (s Scheme) String() string {
	if len(s.TypeVars) == 0 {
		return s.Type.String()
	}

	var buf bytes.Buffer
	buf.WriteString("∀")
	for i := range s.TypeVars {
		if i > 0 {
			buf.WriteString(" ")
		}
		// Use letters for display
		buf.WriteByte('a' + byte(i%26))
	}
	buf.WriteString(". ")
	buf.WriteString(s.Type.String())
	return buf.String()
}

// FreeTypeVars returns type variables free in the scheme (not bound by ∀).
func (s Scheme) FreeTypeVars() map[int]struct{} {
	free := s.Type.FreeTypeVars()
	for _, bound := range s.TypeVars {
		delete(free, bound)
	}
	return free
}

// ----------------------------------------------------------------------------
// Any type (top type for flexibility)
// ----------------------------------------------------------------------------

// Any represents a type that accepts any value.
// Used for operators that work on any type without constraints.
type Any struct{}

func (Any) String() string                   { return "Any" }
func (Any) Equals(t Type) bool               { _, ok := t.(Any); return ok }
func (Any) FreeTypeVars() map[int]struct{}   { return nil }
func (a Any) substitute(_ Substitution) Type { return a }

// ----------------------------------------------------------------------------
// Constructor helpers
// ----------------------------------------------------------------------------

// NewTypeVar creates a fresh type variable with a unique ID.
var nextTypeVarID = 0

func NewTypeVar(name string) TypeVar {
	id := nextTypeVarID
	nextTypeVarID++
	return TypeVar{ID: id, Name: name}
}

// ResetTypeVarCounter resets the type variable ID counter (for testing).
func ResetTypeVarCounter() {
	nextTypeVarID = 0
}

// Monotype creates a Scheme with no quantified variables.
func Monotype(t Type) Scheme {
	return Scheme{TypeVars: nil, Type: t}
}

// ForAll creates a polymorphic Scheme.
// varNames maps display names to IDs for the type.
func ForAll(varIDs []int, t Type) Scheme {
	return Scheme{TypeVars: varIDs, Type: t}
}

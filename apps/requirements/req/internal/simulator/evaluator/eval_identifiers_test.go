package evaluator

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/stretchr/testify/suite"
)

func TestIdentifiersSuite(t *testing.T) {
	suite.Run(t, new(IdentifiersSuite))
}

type IdentifiersSuite struct {
	suite.Suite
}

// === Basic Identifier ===

func (s *IdentifiersSuite) TestIdentifier_Simple() {
	node := &ast.Identifier{Value: "x"}
	bindings := NewBindings()
	bindings.Set("x", object.NewNatural(42), NamespaceGlobal)

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("42", num.Inspect())
}

func (s *IdentifiersSuite) TestIdentifier_NotFound() {
	node := &ast.Identifier{Value: "undefined_var"}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.True(result.IsError())
	s.Contains(result.Error.Message, "identifier not found")
}

func (s *IdentifiersSuite) TestIdentifier_NestedScope() {
	node := &ast.Identifier{Value: "x"}

	// Create parent with x = 10
	outer := NewBindings()
	outer.Set("x", object.NewNatural(10), NamespaceGlobal)

	// Create inner scope
	inner := NewEnclosedBindings(outer)

	result := Eval(node, inner)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("10", num.Inspect())
}

func (s *IdentifiersSuite) TestIdentifier_ShadowedInInnerScope() {
	node := &ast.Identifier{Value: "x"}

	// Create parent with x = 10
	outer := NewBindings()
	outer.Set("x", object.NewNatural(10), NamespaceGlobal)

	// Create inner scope with x = 20 (shadows outer)
	inner := NewEnclosedBindings(outer)
	inner.Set("x", object.NewNatural(20), NamespaceLocal)

	result := Eval(node, inner)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("20", num.Inspect())
}

// === Self Identifier ===

func (s *IdentifiersSuite) TestIdentifier_Self() {
	node := &ast.Identifier{Value: "self"}

	// Create a record and set it as self
	selfRecord := object.NewRecordFromFields(map[string]object.Object{
		"id":   object.NewNatural(123),
		"name": object.NewString("test"),
	})

	bindings := NewBindings()
	innerBindings := bindings.WithSelf(selfRecord)

	result := Eval(node, innerBindings)

	s.False(result.IsError())
	record := result.Value.(*object.Record)
	s.Equal(selfRecord.Inspect(), record.Inspect())
}

func (s *IdentifiersSuite) TestIdentifier_SelfNotDefined() {
	node := &ast.Identifier{Value: "self"}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.True(result.IsError())
	s.Contains(result.Error.Message, "self is not defined")
}

// === Field Identifier ===

func (s *IdentifiersSuite) TestFieldIdentifier_Simple() {
	// person.name
	node := &ast.FieldIdentifier{
		Identifier: &ast.Identifier{Value: "person"},
		Member:     "name",
	}

	personRecord := object.NewRecordFromFields(map[string]object.Object{
		"name": object.NewString("Alice"),
		"age":  object.NewNatural(30),
	})

	bindings := NewBindings()
	bindings.Set("person", personRecord, NamespaceGlobal)

	result := Eval(node, bindings)

	s.False(result.IsError())
	str := result.Value.(*object.String)
	s.Equal("Alice", str.Value())
}

func (s *IdentifiersSuite) TestFieldIdentifier_FieldNotFound() {
	// person.nonexistent
	node := &ast.FieldIdentifier{
		Identifier: &ast.Identifier{Value: "person"},
		Member:     "nonexistent",
	}

	personRecord := object.NewRecordFromFields(map[string]object.Object{
		"name": object.NewString("Alice"),
	})

	bindings := NewBindings()
	bindings.Set("person", personRecord, NamespaceGlobal)

	result := Eval(node, bindings)

	s.True(result.IsError())
	s.Contains(result.Error.Message, "field not found")
}

func (s *IdentifiersSuite) TestFieldIdentifier_NotARecord() {
	// x.member where x is a number
	node := &ast.FieldIdentifier{
		Identifier: &ast.Identifier{Value: "x"},
		Member:     "value",
	}

	bindings := NewBindings()
	bindings.Set("x", object.NewNatural(42), NamespaceGlobal)

	result := Eval(node, bindings)

	s.True(result.IsError())
	s.Contains(result.Error.Message, "requires Record")
}

func (s *IdentifiersSuite) TestFieldIdentifier_ExclamationMember() {
	// !.member - access field from existing value in EXCEPT context
	node := &ast.FieldIdentifier{
		Identifier: nil, // nil means use existing value
		Member:     "name",
	}

	existingRecord := object.NewRecordFromFields(map[string]object.Object{
		"name": object.NewString("Bob"),
		"age":  object.NewNatural(25),
	})

	bindings := NewBindings()
	bindings.SetExistingValue(existingRecord)

	result := Eval(node, bindings)

	s.False(result.IsError())
	str := result.Value.(*object.String)
	s.Equal("Bob", str.Value())
}

func (s *IdentifiersSuite) TestFieldIdentifier_ExclamationOutsideExcept() {
	// !.member outside of EXCEPT context
	node := &ast.FieldIdentifier{
		Identifier: nil,
		Member:     "name",
	}

	bindings := NewBindings()

	result := Eval(node, bindings)

	s.True(result.IsError())
	s.Contains(result.Error.Message, "EXCEPT context")
}

// === Existing Value (@) ===

func (s *IdentifiersSuite) TestExistingValue_Simple() {
	node := &ast.ExistingValue{}

	existingValue := object.NewNatural(42)
	bindings := NewBindings()
	bindings.SetExistingValue(existingValue)

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("42", num.Inspect())
}

func (s *IdentifiersSuite) TestExistingValue_OutsideExceptContext() {
	node := &ast.ExistingValue{}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.True(result.IsError())
	s.Contains(result.Error.Message, "EXCEPT context")
}

func (s *IdentifiersSuite) TestExistingValue_NestedScope() {
	node := &ast.ExistingValue{}

	existingValue := object.NewString("test")
	outer := NewBindings()
	outer.SetExistingValue(existingValue)

	// Inner scope should inherit existing value
	inner := NewEnclosedBindings(outer)

	result := Eval(node, inner)

	s.False(result.IsError())
	str := result.Value.(*object.String)
	s.Equal("test", str.Value())
}

// === Bindings Tests ===

func (s *IdentifiersSuite) TestBindings_PrimedVariable() {
	bindings := NewBindings()
	bindings.Set("x", object.NewNatural(10), NamespaceGlobal)

	// Prime the variable
	bindings.SetPrimed("x", object.NewNatural(20))

	// Check that it's primed
	s.True(bindings.IsPrimed("x"))

	// GetValue returns the current (unprimed) value
	val, found := bindings.GetValue("x")
	s.True(found)
	s.Equal("10", val.Inspect())

	// GetPrimedValue returns the primed (next-state) value
	primedVal, found := bindings.GetPrimedValue("x")
	s.True(found)
	s.Equal("20", primedVal.Inspect())

	// Check primed bindings map
	primed := bindings.GetPrimedBindings()
	s.Len(primed, 1)
	s.Equal("20", primed["x"].Inspect())
}

func (s *IdentifiersSuite) TestBindings_GetByNamespace() {
	bindings := NewBindings()
	bindings.Set("globalVar", object.NewNatural(1), NamespaceGlobal)
	bindings.Set("localVar", object.NewNatural(2), NamespaceLocal)
	bindings.Set("returnVar", object.NewNatural(3), NamespaceReturn)

	globals := bindings.GetByNamespace(NamespaceGlobal)
	s.Len(globals, 1)
	s.Equal("1", globals["globalVar"].Inspect())

	locals := bindings.GetByNamespace(NamespaceLocal)
	s.Len(locals, 1)
	s.Equal("2", locals["localVar"].Inspect())
}

func (s *IdentifiersSuite) TestBindings_Clone() {
	original := NewBindings()
	original.Set("x", object.NewNatural(10), NamespaceGlobal)

	cloned := original.Clone()

	// Modify the clone
	cloned.Set("x", object.NewNatural(20), NamespaceGlobal)

	// Original should be unchanged
	val, _ := original.GetValue("x")
	s.Equal("10", val.Inspect())

	// Clone should have new value
	val2, _ := cloned.GetValue("x")
	s.Equal("20", val2.Inspect())
}

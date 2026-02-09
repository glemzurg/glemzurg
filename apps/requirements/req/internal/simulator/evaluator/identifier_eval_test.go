package evaluator

import (
	"testing"

	"github.com/glemzurg/go-tlaplus/internal/simulator/object"
	"github.com/glemzurg/go-tlaplus/internal/simulator/parser"
	"github.com/stretchr/testify/suite"
)

// IdentifierEvalSuite tests end-to-end: parse -> evaluate for identifiers
func TestIdentifierEvalSuite(t *testing.T) {
	suite.Run(t, new(IdentifierEvalSuite))
}

type IdentifierEvalSuite struct {
	suite.Suite
}

// =============================================================================
// Simple Identifier Evaluation
// =============================================================================

func (s *IdentifierEvalSuite) TestIdentifier_EvalSimple() {
	// Parse "x" and evaluate it with x = 42
	expr, err := parser.ParseExpression("x")
	s.NoError(err)

	bindings := NewBindings()
	bindings.Set("x", object.NewNatural(42), NamespaceGlobal)

	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	num := result.Value.(*object.Number)
	s.Equal("42", num.Inspect())
}

func (s *IdentifierEvalSuite) TestIdentifier_EvalNotFound() {
	// Parse "undefined" and evaluate it without binding
	expr, err := parser.ParseExpression("undefined")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.True(result.IsError())
	s.Contains(result.Error.Message, "identifier not found")
}

func (s *IdentifierEvalSuite) TestIdentifier_EvalSelf() {
	// Parse "self" and evaluate with self bound
	expr, err := parser.ParseExpression("self")
	s.NoError(err)

	selfRecord := object.NewRecordFromFields(map[string]object.Object{
		"id":   object.NewNatural(123),
		"name": object.NewString("test"),
	})

	bindings := NewBindings()
	innerBindings := bindings.WithSelf(selfRecord)

	result := Eval(expr, innerBindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	record := result.Value.(*object.Record)
	s.Equal(selfRecord.Inspect(), record.Inspect())
}

// =============================================================================
// Existing Value (@) Evaluation
// =============================================================================

func (s *IdentifierEvalSuite) TestExistingValue_Eval() {
	// Parse "@" and evaluate with existing value set
	expr, err := parser.ParseExpression("@")
	s.NoError(err)

	existingValue := object.NewNatural(99)
	bindings := NewBindings()
	bindings.SetExistingValue(existingValue)

	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	num := result.Value.(*object.Number)
	s.Equal("99", num.Inspect())
}

func (s *IdentifierEvalSuite) TestExistingValue_OutsideExceptContext() {
	// Parse "@" and evaluate without existing value set
	expr, err := parser.ParseExpression("@")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.True(result.IsError())
	s.Contains(result.Error.Message, "EXCEPT context")
}

// =============================================================================
// Field Access Evaluation
// =============================================================================

func (s *IdentifierEvalSuite) TestFieldAccess_Eval() {
	// Parse "person.name" and evaluate
	expr, err := parser.ParseExpression("person.name")
	s.NoError(err)

	personRecord := object.NewRecordFromFields(map[string]object.Object{
		"name": object.NewString("Alice"),
		"age":  object.NewNatural(30),
	})

	bindings := NewBindings()
	bindings.Set("person", personRecord, NamespaceGlobal)

	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	str := result.Value.(*object.String)
	s.Equal("Alice", str.Value())
}

func (s *IdentifierEvalSuite) TestFieldAccess_FieldNotFound() {
	// Parse "person.unknown" and evaluate
	expr, err := parser.ParseExpression("person.unknown")
	s.NoError(err)

	personRecord := object.NewRecordFromFields(map[string]object.Object{
		"name": object.NewString("Alice"),
	})

	bindings := NewBindings()
	bindings.Set("person", personRecord, NamespaceGlobal)

	result := Eval(expr, bindings)

	s.True(result.IsError())
	s.Contains(result.Error.Message, "field not found")
}

func (s *IdentifierEvalSuite) TestFieldAccess_NotARecord() {
	// Parse "x.value" where x is a number
	expr, err := parser.ParseExpression("x.value")
	s.NoError(err)

	bindings := NewBindings()
	bindings.Set("x", object.NewNatural(42), NamespaceGlobal)

	result := Eval(expr, bindings)

	s.True(result.IsError())
	s.Contains(result.Error.Message, "requires Record")
}

func (s *IdentifierEvalSuite) TestFieldAccess_ExistingValue() {
	// Parse "@.status" and evaluate with existing record
	expr, err := parser.ParseExpression("@.status")
	s.NoError(err)

	existingRecord := object.NewRecordFromFields(map[string]object.Object{
		"status": object.NewString("active"),
		"count":  object.NewNatural(5),
	})

	bindings := NewBindings()
	bindings.SetExistingValue(existingRecord)

	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	str := result.Value.(*object.String)
	s.Equal("active", str.Value())
}

// =============================================================================
// Identifiers in Complex Expressions
// =============================================================================

func (s *IdentifierEvalSuite) TestIdentifier_InArithmeticExpr() {
	// Parse "x + y * 2" and evaluate
	expr, err := parser.ParseExpression("x + y * 2")
	s.NoError(err)

	bindings := NewBindings()
	bindings.Set("x", object.NewNatural(10), NamespaceGlobal)
	bindings.Set("y", object.NewNatural(5), NamespaceGlobal)

	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	num := result.Value.(*object.Number)
	// x + y * 2 = 10 + 5 * 2 = 10 + 10 = 20
	s.Equal("20", num.Inspect())
}

func (s *IdentifierEvalSuite) TestIdentifier_InComparisonExpr() {
	// Parse "age > 18" and evaluate
	expr, err := parser.ParseExpression("age > 18")
	s.NoError(err)

	bindings := NewBindings()
	bindings.Set("age", object.NewNatural(25), NamespaceGlobal)

	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

func (s *IdentifierEvalSuite) TestIdentifier_InLogicExpr() {
	// Parse "a /\ b" and evaluate
	expr, err := parser.ParseExpression("a /\\ b")
	s.NoError(err)

	bindings := NewBindings()
	bindings.Set("a", object.NewBoolean(true), NamespaceGlobal)
	bindings.Set("b", object.NewBoolean(false), NamespaceGlobal)

	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	b := result.Value.(*object.Boolean)
	s.False(b.Value()) // TRUE /\ FALSE = FALSE
}

func (s *IdentifierEvalSuite) TestFieldAccess_InComparisonExpr() {
	// Parse "person.age > 18" and evaluate
	expr, err := parser.ParseExpression("person.age > 18")
	s.NoError(err)

	personRecord := object.NewRecordFromFields(map[string]object.Object{
		"name": object.NewString("Alice"),
		"age":  object.NewNatural(25),
	})

	bindings := NewBindings()
	bindings.Set("person", personRecord, NamespaceGlobal)

	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

func (s *IdentifierEvalSuite) TestExistingValue_InArithmeticExpr() {
	// Parse "@ + 1" and evaluate
	expr, err := parser.ParseExpression("@ + 1")
	s.NoError(err)

	bindings := NewBindings()
	bindings.SetExistingValue(object.NewNatural(41))

	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	num := result.Value.(*object.Number)
	s.Equal("42", num.Inspect())
}

func (s *IdentifierEvalSuite) TestFieldAccess_ExistingValueInExpr() {
	// Parse "@.count + 1" and evaluate
	expr, err := parser.ParseExpression("@.count + 1")
	s.NoError(err)

	existingRecord := object.NewRecordFromFields(map[string]object.Object{
		"count": object.NewNatural(9),
	})

	bindings := NewBindings()
	bindings.SetExistingValue(existingRecord)

	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	num := result.Value.(*object.Number)
	s.Equal("10", num.Inspect())
}

// =============================================================================
// Equality and Comparison with Identifiers
// =============================================================================

func (s *IdentifierEvalSuite) TestIdentifier_Equality() {
	// Parse "x = y" and evaluate
	expr, err := parser.ParseExpression("x = y")
	s.NoError(err)

	bindings := NewBindings()
	bindings.Set("x", object.NewNatural(42), NamespaceGlobal)
	bindings.Set("y", object.NewNatural(42), NamespaceGlobal)

	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

func (s *IdentifierEvalSuite) TestIdentifier_NotEqual() {
	// Parse "x /= y" and evaluate (or x ≠ y)
	expr, err := parser.ParseExpression("x /= y")
	s.NoError(err)

	bindings := NewBindings()
	bindings.Set("x", object.NewNatural(42), NamespaceGlobal)
	bindings.Set("y", object.NewNatural(43), NamespaceGlobal)

	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

// =============================================================================
// Negation with Identifiers
// =============================================================================

func (s *IdentifierEvalSuite) TestIdentifier_Negation() {
	// Parse "-x" and evaluate
	expr, err := parser.ParseExpression("-x")
	s.NoError(err)

	bindings := NewBindings()
	bindings.Set("x", object.NewNatural(42), NamespaceGlobal)

	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	num := result.Value.(*object.Number)
	s.Equal("-42", num.Inspect())
}

func (s *IdentifierEvalSuite) TestIdentifier_LogicNegation() {
	// Parse "~flag" and evaluate
	expr, err := parser.ParseExpression("~flag")
	s.NoError(err)

	bindings := NewBindings()
	bindings.Set("flag", object.NewBoolean(true), NamespaceGlobal)

	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	b := result.Value.(*object.Boolean)
	s.False(b.Value())
}

// =============================================================================
// Chained Field Access Evaluation
// =============================================================================

func (s *IdentifierEvalSuite) TestFieldAccess_Chained() {
	// Parse "person.address.city" and evaluate
	expr, err := parser.ParseExpression("person.address.city")
	s.NoError(err)

	// Create nested records
	addressRecord := object.NewRecordFromFields(map[string]object.Object{
		"city":   object.NewString("New York"),
		"street": object.NewString("123 Main St"),
	})
	personRecord := object.NewRecordFromFields(map[string]object.Object{
		"name":    object.NewString("Alice"),
		"address": addressRecord,
	})

	bindings := NewBindings()
	bindings.Set("person", personRecord, NamespaceGlobal)

	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	str := result.Value.(*object.String)
	s.Equal("New York", str.Value())
}

func (s *IdentifierEvalSuite) TestFieldAccess_ChainedExistingValue() {
	// Parse "@.a.b" and evaluate
	expr, err := parser.ParseExpression("@.a.b")
	s.NoError(err)

	// Create nested records
	innerRecord := object.NewRecordFromFields(map[string]object.Object{
		"b": object.NewNatural(42),
	})
	outerRecord := object.NewRecordFromFields(map[string]object.Object{
		"a": innerRecord,
	})

	bindings := NewBindings()
	bindings.SetExistingValue(outerRecord)

	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	num := result.Value.(*object.Number)
	s.Equal("42", num.Inspect())
}

// =============================================================================
// Primed Expression Evaluation
// =============================================================================

func (s *IdentifierEvalSuite) TestPrimed_ReadCurrentValue() {
	// Parse "x'" and evaluate - should read current value if not primed
	expr, err := parser.ParseExpression("x'")
	s.NoError(err)

	bindings := NewBindings()
	bindings.Set("x", object.NewNatural(42), NamespaceGlobal)

	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	num := result.Value.(*object.Number)
	s.Equal("42", num.Inspect())
}

func (s *IdentifierEvalSuite) TestPrimed_ReadPrimedValue() {
	// Parse "x'" and evaluate - should read primed value if set
	expr, err := parser.ParseExpression("x'")
	s.NoError(err)

	bindings := NewBindings()
	bindings.Set("x", object.NewNatural(42), NamespaceGlobal)
	bindings.SetPrimed("x", object.NewNatural(100))

	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	num := result.Value.(*object.Number)
	s.Equal("100", num.Inspect())
}

func (s *IdentifierEvalSuite) TestPrimed_InExpression() {
	// Parse "x' + 1" and evaluate
	expr, err := parser.ParseExpression("x' + 1")
	s.NoError(err)

	bindings := NewBindings()
	bindings.Set("x", object.NewNatural(41), NamespaceGlobal)

	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	num := result.Value.(*object.Number)
	s.Equal("42", num.Inspect())
}

func (s *IdentifierEvalSuite) TestPrimed_UndefinedVariable() {
	// Parse "undefined'" and evaluate without binding
	expr, err := parser.ParseExpression("undefined'")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.True(result.IsError())
	s.Contains(result.Error.Message, "identifier not found")
}

func (s *IdentifierEvalSuite) TestPrimed_FieldAccess() {
	// Parse "record.field'" and evaluate
	expr, err := parser.ParseExpression("record.field'")
	s.NoError(err)

	recordObj := object.NewRecordFromFields(map[string]object.Object{
		"field": object.NewNatural(42),
	})

	bindings := NewBindings()
	bindings.Set("record", recordObj, NamespaceGlobal)

	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	num := result.Value.(*object.Number)
	s.Equal("42", num.Inspect())
}

func (s *IdentifierEvalSuite) TestPrimed_FieldAccessWithPrimedRecord() {
	// Test that record.field' reads from the primed record, not the current one.
	// This is the key semantic: record.field' should read from record' (next state).
	expr, err := parser.ParseExpression("record.field'")
	s.NoError(err)

	// Current state: record.field = 10
	currentRecord := object.NewRecordFromFields(map[string]object.Object{
		"field": object.NewNatural(10),
	})

	// Next state: record'.field = 20
	primedRecord := object.NewRecordFromFields(map[string]object.Object{
		"field": object.NewNatural(20),
	})

	bindings := NewBindings()
	bindings.Set("record", currentRecord, NamespaceGlobal)
	bindings.SetPrimed("record", primedRecord)

	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	num := result.Value.(*object.Number)
	// Should get 20 from the primed record, not 10 from current
	s.Equal("20", num.Inspect())
}

func (s *IdentifierEvalSuite) TestPrimed_CompareCurrentAndNextState() {
	// Test the use case: record.field' > record.field
	// This tests that an integer field increased in value.
	expr, err := parser.ParseExpression("record.value' > record.value")
	s.NoError(err)

	// Current state: record.value = 10
	currentRecord := object.NewRecordFromFields(map[string]object.Object{
		"value": object.NewNatural(10),
	})

	// Next state: record'.value = 15 (increased)
	primedRecord := object.NewRecordFromFields(map[string]object.Object{
		"value": object.NewNatural(15),
	})

	bindings := NewBindings()
	bindings.Set("record", currentRecord, NamespaceGlobal)
	bindings.SetPrimed("record", primedRecord)

	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	b := result.Value.(*object.Boolean)
	// 15 > 10 should be TRUE
	s.True(b.Value())
}

func (s *IdentifierEvalSuite) TestPrimed_CompareCurrentAndNextState_NotIncreased() {
	// Test the opposite: when the value didn't increase
	expr, err := parser.ParseExpression("record.value' > record.value")
	s.NoError(err)

	// Current state: record.value = 10
	currentRecord := object.NewRecordFromFields(map[string]object.Object{
		"value": object.NewNatural(10),
	})

	// Next state: record'.value = 5 (decreased)
	primedRecord := object.NewRecordFromFields(map[string]object.Object{
		"value": object.NewNatural(5),
	})

	bindings := NewBindings()
	bindings.Set("record", currentRecord, NamespaceGlobal)
	bindings.SetPrimed("record", primedRecord)

	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	b := result.Value.(*object.Boolean)
	// 5 > 10 should be FALSE
	s.False(b.Value())
}

func (s *IdentifierEvalSuite) TestPrimed_ChainedFieldAccessWithPrimedRecord() {
	// Test record.a.b' with primed record
	expr, err := parser.ParseExpression("record.inner.value'")
	s.NoError(err)

	// Current state
	currentInner := object.NewRecordFromFields(map[string]object.Object{
		"value": object.NewNatural(100),
	})
	currentRecord := object.NewRecordFromFields(map[string]object.Object{
		"inner": currentInner,
	})

	// Next state
	primedInner := object.NewRecordFromFields(map[string]object.Object{
		"value": object.NewNatural(200),
	})
	primedRecord := object.NewRecordFromFields(map[string]object.Object{
		"inner": primedInner,
	})

	bindings := NewBindings()
	bindings.Set("record", currentRecord, NamespaceGlobal)
	bindings.SetPrimed("record", primedRecord)

	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	num := result.Value.(*object.Number)
	// Should get 200 from the primed record's inner.value
	s.Equal("200", num.Inspect())
}

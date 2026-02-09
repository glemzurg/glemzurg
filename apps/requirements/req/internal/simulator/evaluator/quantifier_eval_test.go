package evaluator

import (
	"testing"

	"github.com/glemzurg/go-tlaplus/internal/simulator/object"
	"github.com/glemzurg/go-tlaplus/internal/simulator/parser"
	"github.com/stretchr/testify/suite"
)

func TestQuantifierEvalSuite(t *testing.T) {
	suite.Run(t, new(QuantifierEvalSuite))
}

type QuantifierEvalSuite struct {
	suite.Suite
}

// =============================================================================
// Universal Quantifier Evaluation (ForAll)
// =============================================================================

func (s *QuantifierEvalSuite) TestForAll_AllTrue() {
	// All elements in {1, 2, 3} are greater than 0
	expr, err := parser.ParseExpression("∀ x ∈ {1, 2, 3} : x > 0")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	b, ok := result.Value.(*object.Boolean)
	s.True(ok, "expected *object.Boolean, got %T", result.Value)
	s.True(b.Value())
}

func (s *QuantifierEvalSuite) TestForAll_SomeFalse() {
	// Not all elements in {1, 2, 3} are greater than 2
	expr, err := parser.ParseExpression("∀ x ∈ {1, 2, 3} : x > 2")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	b, ok := result.Value.(*object.Boolean)
	s.True(ok, "expected *object.Boolean, got %T", result.Value)
	s.False(b.Value())
}

func (s *QuantifierEvalSuite) TestForAll_EmptySet() {
	// Vacuously true: all elements in {} satisfy any predicate
	expr, err := parser.ParseExpression("∀ x ∈ {} : x > 1000")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	b, ok := result.Value.(*object.Boolean)
	s.True(ok, "expected *object.Boolean, got %T", result.Value)
	s.True(b.Value()) // Vacuously true for empty set
}

func (s *QuantifierEvalSuite) TestForAll_WithRange() {
	// All elements in 1..5 are less than 10
	expr, err := parser.ParseExpression("∀ n ∈ 1..5 : n < 10")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	b, ok := result.Value.(*object.Boolean)
	s.True(ok, "expected *object.Boolean, got %T", result.Value)
	s.True(b.Value())
}

func (s *QuantifierEvalSuite) TestForAll_ASCII() {
	// Using ASCII syntax
	expr, err := parser.ParseExpression("\\A x \\in {1, 2, 3} : x > 0")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	b, ok := result.Value.(*object.Boolean)
	s.True(ok, "expected *object.Boolean, got %T", result.Value)
	s.True(b.Value())
}

// =============================================================================
// Existential Quantifier Evaluation (Exists)
// =============================================================================

func (s *QuantifierEvalSuite) TestExists_SomeTrue() {
	// At least one element in {1, 2, 3} equals 2
	expr, err := parser.ParseExpression("∃ x ∈ {1, 2, 3} : x = 2")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	b, ok := result.Value.(*object.Boolean)
	s.True(ok, "expected *object.Boolean, got %T", result.Value)
	s.True(b.Value())
}

func (s *QuantifierEvalSuite) TestExists_NoneTrue() {
	// No element in {1, 2, 3} equals 5
	expr, err := parser.ParseExpression("∃ x ∈ {1, 2, 3} : x = 5")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	b, ok := result.Value.(*object.Boolean)
	s.True(ok, "expected *object.Boolean, got %T", result.Value)
	s.False(b.Value())
}

func (s *QuantifierEvalSuite) TestExists_EmptySet() {
	// No element exists in empty set
	expr, err := parser.ParseExpression("∃ x ∈ {} : x = 1")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	b, ok := result.Value.(*object.Boolean)
	s.True(ok, "expected *object.Boolean, got %T", result.Value)
	s.False(b.Value()) // False for empty set
}

func (s *QuantifierEvalSuite) TestExists_WithRange() {
	// There exists n in 1..10 where n = 5
	expr, err := parser.ParseExpression("∃ n ∈ 1..10 : n = 5")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	b, ok := result.Value.(*object.Boolean)
	s.True(ok, "expected *object.Boolean, got %T", result.Value)
	s.True(b.Value())
}

func (s *QuantifierEvalSuite) TestExists_ASCII() {
	// Using ASCII syntax
	expr, err := parser.ParseExpression("\\E x \\in {1, 2, 3} : x = 2")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	b, ok := result.Value.(*object.Boolean)
	s.True(ok, "expected *object.Boolean, got %T", result.Value)
	s.True(b.Value())
}

// =============================================================================
// Quantifiers with External Variables
// =============================================================================

func (s *QuantifierEvalSuite) TestForAll_WithExternalVariable() {
	// ∀ x ∈ S : x > threshold
	expr, err := parser.ParseExpression("∀ x ∈ S : x > threshold")
	s.NoError(err)

	bindings := NewBindings()
	bindings.Set("S", object.NewSetFromElements([]object.Object{
		object.NewInteger(5),
		object.NewInteger(10),
		object.NewInteger(15),
	}), NamespaceGlobal)
	bindings.Set("threshold", object.NewInteger(4), NamespaceGlobal)

	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	b, ok := result.Value.(*object.Boolean)
	s.True(ok, "expected *object.Boolean, got %T", result.Value)
	s.True(b.Value())
}

func (s *QuantifierEvalSuite) TestExists_WithExternalVariable() {
	// ∃ x ∈ S : x = target
	expr, err := parser.ParseExpression("∃ x ∈ S : x = target")
	s.NoError(err)

	bindings := NewBindings()
	bindings.Set("S", object.NewSetFromElements([]object.Object{
		object.NewInteger(1),
		object.NewInteger(2),
		object.NewInteger(3),
	}), NamespaceGlobal)
	bindings.Set("target", object.NewInteger(2), NamespaceGlobal)

	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	b, ok := result.Value.(*object.Boolean)
	s.True(ok, "expected *object.Boolean, got %T", result.Value)
	s.True(b.Value())
}

// =============================================================================
// Nested Quantifier Evaluation
// =============================================================================

func (s *QuantifierEvalSuite) TestNested_ForAllExists_True() {
	// For all x in {1, 2}, there exists y in {1, 2, 3} such that x < y
	// This is true because: for x=1, y=2 or y=3 work; for x=2, y=3 works
	expr, err := parser.ParseExpression("∀ x ∈ {1, 2} : ∃ y ∈ {1, 2, 3} : x < y")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	b, ok := result.Value.(*object.Boolean)
	s.True(ok, "expected *object.Boolean, got %T", result.Value)
	s.True(b.Value())
}

func (s *QuantifierEvalSuite) TestNested_ForAllExists_False() {
	// For all x in {1, 2, 3}, there exists y in {1, 2} such that x < y
	// This is false because: for x=3, there's no y in {1, 2} where 3 < y
	expr, err := parser.ParseExpression("∀ x ∈ {1, 2, 3} : ∃ y ∈ {1, 2} : x < y")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	b, ok := result.Value.(*object.Boolean)
	s.True(ok, "expected *object.Boolean, got %T", result.Value)
	s.False(b.Value())
}

func (s *QuantifierEvalSuite) TestNested_ExistsForAll_True() {
	// There exists x in {3} such that for all y in {1, 2}, x > y
	// This is true because x=3 > y for all y in {1, 2}
	expr, err := parser.ParseExpression("∃ x ∈ {3} : ∀ y ∈ {1, 2} : x > y")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	b, ok := result.Value.(*object.Boolean)
	s.True(ok, "expected *object.Boolean, got %T", result.Value)
	s.True(b.Value())
}

func (s *QuantifierEvalSuite) TestNested_ExistsForAll_False() {
	// There exists x in {1, 2} such that for all y in {1, 2, 3}, x > y
	// This is false because neither x=1 nor x=2 is greater than all of {1, 2, 3}
	expr, err := parser.ParseExpression("∃ x ∈ {1, 2} : ∀ y ∈ {1, 2, 3} : x > y")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	b, ok := result.Value.(*object.Boolean)
	s.True(ok, "expected *object.Boolean, got %T", result.Value)
	s.False(b.Value())
}

// =============================================================================
// Complex Predicate Evaluation
// =============================================================================

func (s *QuantifierEvalSuite) TestForAll_ComplexPredicate() {
	// All elements in 1..5 are >= 1 AND <= 5
	expr, err := parser.ParseExpression("∀ x ∈ 1..5 : x >= 1 ∧ x <= 5")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	b, ok := result.Value.(*object.Boolean)
	s.True(ok, "expected *object.Boolean, got %T", result.Value)
	s.True(b.Value())
}

func (s *QuantifierEvalSuite) TestForAll_ImpliesPredicate() {
	// For all x: if x > 5 then x > 0 (trivially true for 1..5)
	expr, err := parser.ParseExpression("∀ x ∈ 1..5 : x > 5 => x > 0")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	b, ok := result.Value.(*object.Boolean)
	s.True(ok, "expected *object.Boolean, got %T", result.Value)
	s.True(b.Value()) // All implications are vacuously true since antecedent is always false
}

func (s *QuantifierEvalSuite) TestExists_OrPredicate() {
	// There exists x in {1, 5, 10} where x = 1 OR x = 10
	expr, err := parser.ParseExpression("∃ x ∈ {1, 5, 10} : x = 1 ∨ x = 10")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	b, ok := result.Value.(*object.Boolean)
	s.True(ok, "expected *object.Boolean, got %T", result.Value)
	s.True(b.Value())
}

// =============================================================================
// Quantifiers with Arithmetic
// =============================================================================

func (s *QuantifierEvalSuite) TestForAll_ArithmeticPredicate() {
	// All x in {2, 4, 6} satisfy x % 2 = 0 (all even)
	expr, err := parser.ParseExpression("∀ x ∈ {2, 4, 6} : x % 2 = 0")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	b, ok := result.Value.(*object.Boolean)
	s.True(ok, "expected *object.Boolean, got %T", result.Value)
	s.True(b.Value())
}

func (s *QuantifierEvalSuite) TestExists_ArithmeticPredicate() {
	// There exists x in 1..10 where x * x = 25 (x = 5)
	expr, err := parser.ParseExpression("∃ x ∈ 1..10 : x * x = 25")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	b, ok := result.Value.(*object.Boolean)
	s.True(ok, "expected *object.Boolean, got %T", result.Value)
	s.True(b.Value())
}

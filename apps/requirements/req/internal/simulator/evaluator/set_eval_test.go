package evaluator

import (
	"testing"

	"github.com/glemzurg/go-tlaplus/internal/simulator/object"
	"github.com/glemzurg/go-tlaplus/internal/simulator/parser"
	"github.com/stretchr/testify/suite"
)

func TestSetEvalSuite(t *testing.T) {
	suite.Run(t, new(SetEvalSuite))
}

type SetEvalSuite struct {
	suite.Suite
}

// =============================================================================
// Set Literal Evaluation
// =============================================================================

func (s *SetEvalSuite) TestSetLiteral_Empty() {
	expr, err := parser.ParseExpression("{}")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	set, ok := result.Value.(*object.Set)
	s.True(ok, "expected *object.Set, got %T", result.Value)
	s.Equal(0, set.Size())
}

func (s *SetEvalSuite) TestSetLiteral_Integers() {
	expr, err := parser.ParseExpression("{1, 2, 3}")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	set, ok := result.Value.(*object.Set)
	s.True(ok, "expected *object.Set, got %T", result.Value)
	s.Equal(3, set.Size())
}

func (s *SetEvalSuite) TestSetLiteral_WithVariables() {
	expr, err := parser.ParseExpression("{x, y, z}")
	s.NoError(err)

	bindings := NewBindings()
	bindings.Set("x", object.NewInteger(10), NamespaceGlobal)
	bindings.Set("y", object.NewInteger(20), NamespaceGlobal)
	bindings.Set("z", object.NewInteger(30), NamespaceGlobal)

	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	set, ok := result.Value.(*object.Set)
	s.True(ok, "expected *object.Set, got %T", result.Value)
	s.Equal(3, set.Size())
}

func (s *SetEvalSuite) TestSetLiteral_Duplicates() {
	// {1, 1, 2} should result in {1, 2} (sets don't have duplicates)
	expr, err := parser.ParseExpression("{1, 1, 2}")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	set, ok := result.Value.(*object.Set)
	s.True(ok, "expected *object.Set, got %T", result.Value)
	s.Equal(2, set.Size())
}

// =============================================================================
// Set Range Evaluation
// =============================================================================

func (s *SetEvalSuite) TestSetRange_Simple() {
	expr, err := parser.ParseExpression("1..5")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	set, ok := result.Value.(*object.Set)
	s.True(ok, "expected *object.Set, got %T", result.Value)
	s.Equal(5, set.Size()) // {1, 2, 3, 4, 5}
}

func (s *SetEvalSuite) TestSetRange_Single() {
	expr, err := parser.ParseExpression("5..5")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	set, ok := result.Value.(*object.Set)
	s.True(ok, "expected *object.Set, got %T", result.Value)
	s.Equal(1, set.Size()) // {5}
}

func (s *SetEvalSuite) TestSetRange_Empty() {
	// 5..1 should be empty (start > end)
	expr, err := parser.ParseExpression("5..1")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	set, ok := result.Value.(*object.Set)
	s.True(ok, "expected *object.Set, got %T", result.Value)
	s.Equal(0, set.Size())
}

func (s *SetEvalSuite) TestSetRange_WithVariables() {
	expr, err := parser.ParseExpression("x..y")
	s.NoError(err)

	bindings := NewBindings()
	bindings.Set("x", object.NewInteger(3), NamespaceGlobal)
	bindings.Set("y", object.NewInteger(7), NamespaceGlobal)

	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	set, ok := result.Value.(*object.Set)
	s.True(ok, "expected *object.Set, got %T", result.Value)
	s.Equal(5, set.Size()) // {3, 4, 5, 6, 7}
}

// =============================================================================
// Set Membership Evaluation
// =============================================================================

func (s *SetEvalSuite) TestSetMembership_In_True() {
	expr, err := parser.ParseExpression("2 ∈ {1, 2, 3}")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	b, ok := result.Value.(*object.Boolean)
	s.True(ok, "expected *object.Boolean, got %T", result.Value)
	s.True(b.Value())
}

func (s *SetEvalSuite) TestSetMembership_In_False() {
	expr, err := parser.ParseExpression("5 ∈ {1, 2, 3}")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	b, ok := result.Value.(*object.Boolean)
	s.True(ok, "expected *object.Boolean, got %T", result.Value)
	s.False(b.Value())
}

func (s *SetEvalSuite) TestSetMembership_NotIn_True() {
	expr, err := parser.ParseExpression("5 ∉ {1, 2, 3}")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	b, ok := result.Value.(*object.Boolean)
	s.True(ok, "expected *object.Boolean, got %T", result.Value)
	s.True(b.Value())
}

func (s *SetEvalSuite) TestSetMembership_NotIn_False() {
	expr, err := parser.ParseExpression("2 ∉ {1, 2, 3}")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	b, ok := result.Value.(*object.Boolean)
	s.True(ok, "expected *object.Boolean, got %T", result.Value)
	s.False(b.Value())
}

func (s *SetEvalSuite) TestSetMembership_InRange() {
	expr, err := parser.ParseExpression("5 ∈ 1..10")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	b, ok := result.Value.(*object.Boolean)
	s.True(ok, "expected *object.Boolean, got %T", result.Value)
	s.True(b.Value())
}

// =============================================================================
// Set Operations Evaluation
// =============================================================================

func (s *SetEvalSuite) TestSetUnion() {
	expr, err := parser.ParseExpression("{1, 2} ∪ {2, 3}")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	set, ok := result.Value.(*object.Set)
	s.True(ok, "expected *object.Set, got %T", result.Value)
	s.Equal(3, set.Size()) // {1, 2, 3}
}

func (s *SetEvalSuite) TestSetIntersection() {
	expr, err := parser.ParseExpression("{1, 2, 3} ∩ {2, 3, 4}")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	set, ok := result.Value.(*object.Set)
	s.True(ok, "expected *object.Set, got %T", result.Value)
	s.Equal(2, set.Size()) // {2, 3}
}

func (s *SetEvalSuite) TestSetDifference() {
	expr, err := parser.ParseExpression("{1, 2, 3} \\ {2}")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	set, ok := result.Value.(*object.Set)
	s.True(ok, "expected *object.Set, got %T", result.Value)
	s.Equal(2, set.Size()) // {1, 3}
}

func (s *SetEvalSuite) TestSetOperations_Chained() {
	// {1, 2} ∪ {3} ∪ {4} should be {1, 2, 3, 4}
	expr, err := parser.ParseExpression("{1, 2} ∪ {3} ∪ {4}")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	set, ok := result.Value.(*object.Set)
	s.True(ok, "expected *object.Set, got %T", result.Value)
	s.Equal(4, set.Size())
}

// =============================================================================
// Set Comparisons Evaluation
// =============================================================================

func (s *SetEvalSuite) TestSetComparison_SubsetEq_True() {
	expr, err := parser.ParseExpression("{1, 2} ⊆ {1, 2, 3}")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	b, ok := result.Value.(*object.Boolean)
	s.True(ok, "expected *object.Boolean, got %T", result.Value)
	s.True(b.Value())
}

func (s *SetEvalSuite) TestSetComparison_SubsetEq_Equal() {
	// Equal sets should satisfy ⊆
	expr, err := parser.ParseExpression("{1, 2} ⊆ {1, 2}")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	b, ok := result.Value.(*object.Boolean)
	s.True(ok, "expected *object.Boolean, got %T", result.Value)
	s.True(b.Value())
}

func (s *SetEvalSuite) TestSetComparison_SubsetEq_False() {
	expr, err := parser.ParseExpression("{1, 2, 4} ⊆ {1, 2, 3}")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	b, ok := result.Value.(*object.Boolean)
	s.True(ok, "expected *object.Boolean, got %T", result.Value)
	s.False(b.Value())
}

func (s *SetEvalSuite) TestSetComparison_Subset_True() {
	// Proper subset - must not be equal
	expr, err := parser.ParseExpression("{1, 2} ⊂ {1, 2, 3}")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	b, ok := result.Value.(*object.Boolean)
	s.True(ok, "expected *object.Boolean, got %T", result.Value)
	s.True(b.Value())
}

func (s *SetEvalSuite) TestSetComparison_Subset_EqualSets() {
	// Equal sets should NOT satisfy proper subset ⊂
	expr, err := parser.ParseExpression("{1, 2} ⊂ {1, 2}")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	b, ok := result.Value.(*object.Boolean)
	s.True(ok, "expected *object.Boolean, got %T", result.Value)
	s.False(b.Value())
}

func (s *SetEvalSuite) TestSetComparison_SupersetEq_True() {
	expr, err := parser.ParseExpression("{1, 2, 3} ⊇ {1, 2}")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	b, ok := result.Value.(*object.Boolean)
	s.True(ok, "expected *object.Boolean, got %T", result.Value)
	s.True(b.Value())
}

func (s *SetEvalSuite) TestSetComparison_Superset_True() {
	expr, err := parser.ParseExpression("{1, 2, 3} ⊃ {1, 2}")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	b, ok := result.Value.(*object.Boolean)
	s.True(ok, "expected *object.Boolean, got %T", result.Value)
	s.True(b.Value())
}

// =============================================================================
// Combined Tests
// =============================================================================

func (s *SetEvalSuite) TestCombined_RangeUnion() {
	// 1..3 ∪ 5..7 = {1, 2, 3, 5, 6, 7}
	expr, err := parser.ParseExpression("1..3 ∪ 5..7")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	set, ok := result.Value.(*object.Set)
	s.True(ok, "expected *object.Set, got %T", result.Value)
	s.Equal(6, set.Size())
}

func (s *SetEvalSuite) TestCombined_MembershipInUnion() {
	// 5 ∈ {1, 2} ∪ {3, 4, 5}
	expr, err := parser.ParseExpression("5 ∈ {1, 2} ∪ {3, 4, 5}")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	b, ok := result.Value.(*object.Boolean)
	s.True(ok, "expected *object.Boolean, got %T", result.Value)
	s.True(b.Value())
}

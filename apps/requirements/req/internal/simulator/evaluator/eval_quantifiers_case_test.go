package evaluator

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/stretchr/testify/suite"
)

func TestQuantifiersCaseSuite(t *testing.T) {
	suite.Run(t, new(QuantifiersCaseSuite))
}

type QuantifiersCaseSuite struct {
	suite.Suite
}

// === Universal Quantifier (∀) ===

func (s *QuantifiersCaseSuite) TestForAll_AllTrue() {
	// ∀x ∈ {1, 2, 3} : TRUE = TRUE
	node := &ast.LogicBoundQuantifier{
		Quantifier: "∀",
		Membership: &ast.LogicMembership{
			Left:     &ast.Identifier{Value: "x"},
			Operator: "∈",
			Right:    &ast.SetLiteralInt{Values: []int{1, 2, 3}},
		},
		// Predicate: Using a literal TRUE as a simple test
		// (Identifier doesn't implement Real, so we can't use x > 0)
		Predicate: &ast.BooleanLiteral{Value: true},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

func (s *QuantifiersCaseSuite) TestForAll_SomeFalse() {
	// ∀x ∈ {1, 2, 3} : FALSE = FALSE
	node := &ast.LogicBoundQuantifier{
		Quantifier: "∀",
		Membership: &ast.LogicMembership{
			Left:     &ast.Identifier{Value: "x"},
			Operator: "∈",
			Right:    &ast.SetLiteralInt{Values: []int{1, 2, 3}},
		},
		Predicate: &ast.BooleanLiteral{Value: false},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.False(b.Value())
}

func (s *QuantifiersCaseSuite) TestForAll_EmptySet() {
	// ∀x ∈ {} : FALSE = TRUE (vacuously true)
	node := &ast.LogicBoundQuantifier{
		Quantifier: "∀",
		Membership: &ast.LogicMembership{
			Left:     &ast.Identifier{Value: "x"},
			Operator: "∈",
			Right:    &ast.SetLiteralInt{Values: []int{}},
		},
		Predicate: &ast.BooleanLiteral{Value: false},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value()) // Vacuously true
}

// === Existential Quantifier (∃) ===

func (s *QuantifiersCaseSuite) TestExists_SomeTrue() {
	// ∃x ∈ {1, 2, 3} : TRUE = TRUE
	node := &ast.LogicBoundQuantifier{
		Quantifier: "∃",
		Membership: &ast.LogicMembership{
			Left:     &ast.Identifier{Value: "x"},
			Operator: "∈",
			Right:    &ast.SetLiteralInt{Values: []int{1, 2, 3}},
		},
		Predicate: &ast.BooleanLiteral{Value: true},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

func (s *QuantifiersCaseSuite) TestExists_NoneTrue() {
	// ∃x ∈ {1, 2, 3} : FALSE = FALSE
	node := &ast.LogicBoundQuantifier{
		Quantifier: "∃",
		Membership: &ast.LogicMembership{
			Left:     &ast.Identifier{Value: "x"},
			Operator: "∈",
			Right:    &ast.SetLiteralInt{Values: []int{1, 2, 3}},
		},
		Predicate: &ast.BooleanLiteral{Value: false},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.False(b.Value())
}

func (s *QuantifiersCaseSuite) TestExists_EmptySet() {
	// ∃x ∈ {} : TRUE = FALSE (nothing to exist)
	node := &ast.LogicBoundQuantifier{
		Quantifier: "∃",
		Membership: &ast.LogicMembership{
			Left:     &ast.Identifier{Value: "x"},
			Operator: "∈",
			Right:    &ast.SetLiteralInt{Values: []int{}},
		},
		Predicate: &ast.BooleanLiteral{Value: true},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.False(b.Value())
}

// === CASE Expressions ===

func (s *QuantifiersCaseSuite) TestCase_FirstBranchMatches() {
	// CASE TRUE → 1 □ FALSE → 2 = 1
	node := &ast.ExpressionCase{
		Branches: []*ast.CaseBranch{
			{
				Condition: &ast.BooleanLiteral{Value: true},
				Result:    ast.NewIntLiteral(1),
			},
			{
				Condition: &ast.BooleanLiteral{Value: false},
				Result:    ast.NewIntLiteral(2),
			},
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("1", num.Inspect())
}

func (s *QuantifiersCaseSuite) TestCase_SecondBranchMatches() {
	// CASE FALSE → 1 □ TRUE → 2 = 2
	node := &ast.ExpressionCase{
		Branches: []*ast.CaseBranch{
			{
				Condition: &ast.BooleanLiteral{Value: false},
				Result:    ast.NewIntLiteral(1),
			},
			{
				Condition: &ast.BooleanLiteral{Value: true},
				Result:    ast.NewIntLiteral(2),
			},
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("2", num.Inspect())
}

func (s *QuantifiersCaseSuite) TestCase_OtherBranch() {
	// CASE FALSE → 1 □ FALSE → 2 □ OTHER → 3 = 3
	node := &ast.ExpressionCase{
		Branches: []*ast.CaseBranch{
			{
				Condition: &ast.BooleanLiteral{Value: false},
				Result:    ast.NewIntLiteral(1),
			},
			{
				Condition: &ast.BooleanLiteral{Value: false},
				Result:    ast.NewIntLiteral(2),
			},
		},
		Other: ast.NewIntLiteral(3),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("3", num.Inspect())
}

func (s *QuantifiersCaseSuite) TestCase_NoMatchNoOther_Error() {
	// CASE FALSE → 1 □ FALSE → 2 = error
	node := &ast.ExpressionCase{
		Branches: []*ast.CaseBranch{
			{
				Condition: &ast.BooleanLiteral{Value: false},
				Result:    ast.NewIntLiteral(1),
			},
			{
				Condition: &ast.BooleanLiteral{Value: false},
				Result:    ast.NewIntLiteral(2),
			},
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.True(result.IsError())
	s.Contains(result.Error.Message, "no branch matched")
}

func (s *QuantifiersCaseSuite) TestCase_WithStringResults() {
	// CASE TRUE → "yes" □ FALSE → "no" = "yes"
	node := &ast.ExpressionCase{
		Branches: []*ast.CaseBranch{
			{
				Condition: &ast.BooleanLiteral{Value: true},
				Result:    &ast.StringLiteral{Value: "yes"},
			},
			{
				Condition: &ast.BooleanLiteral{Value: false},
				Result:    &ast.StringLiteral{Value: "no"},
			},
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	str := result.Value.(*object.String)
	s.Equal("yes", str.Value())
}

func (s *QuantifiersCaseSuite) TestCase_WithRecordResults() {
	// CASE TRUE → [val ↦ 1] □ OTHER → [val ↦ 0]
	node := &ast.ExpressionCase{
		Branches: []*ast.CaseBranch{
			{
				Condition: &ast.BooleanLiteral{Value: true},
				Result: &ast.RecordInstance{
					Bindings: []*ast.FieldBinding{
						{Field: &ast.Identifier{Value: "val"}, Expression: ast.NewIntLiteral(1)},
					},
				},
			},
		},
		Other: &ast.RecordInstance{
			Bindings: []*ast.FieldBinding{
				{Field: &ast.Identifier{Value: "val"}, Expression: ast.NewIntLiteral(0)},
			},
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	record := result.Value.(*object.Record)
	s.Equal("1", record.Get("val").Inspect())
}

// === Bag Conversions (using direct builtin calls) ===

func (s *QuantifiersCaseSuite) TestSetToBag_Simple() {
	// _Bags!SetToBag({1, 2, 3}) = bag with each element count 1
	// Test using direct builtin call since SetLiteralInt doesn't implement Expression
	set := object.NewSet()
	set.Add(object.NewNatural(1))
	set.Add(object.NewNatural(2))
	set.Add(object.NewNatural(3))

	result := builtinSetToBag([]object.Object{set})

	s.False(result.IsError())
	bag := result.Value.(*object.Bag)
	s.Equal(3, len(bag.Elements()))
}

func (s *QuantifiersCaseSuite) TestSetToBag_Empty() {
	// _Bags!SetToBag({}) = empty bag
	set := object.NewSet()

	result := builtinSetToBag([]object.Object{set})

	s.False(result.IsError())
	bag := result.Value.(*object.Bag)
	s.Equal(0, len(bag.Elements()))
}

func (s *QuantifiersCaseSuite) TestBagToSet_Simple() {
	// Create a bag with elements, then convert to set
	bag := object.NewBag()
	bag.Add(object.NewNatural(1), 1)
	bag.Add(object.NewNatural(2), 1)

	result := builtinBagToSet([]object.Object{bag})

	s.False(result.IsError())
	set := result.Value.(*object.Set)
	s.Equal(2, set.Size())
}

// === Nested Quantifiers with Comparisons ===

func (s *QuantifiersCaseSuite) TestForAll_WithComparison() {
	// ∀x ∈ {5, 10, 15} : 5 < 20 = TRUE (using literal comparison since x isn't Real)
	node := &ast.LogicBoundQuantifier{
		Quantifier: "∀",
		Membership: &ast.LogicMembership{
			Left:     &ast.Identifier{Value: "x"},
			Operator: "∈",
			Right:    &ast.SetLiteralInt{Values: []int{5, 10, 15}},
		},
		Predicate: &ast.LogicRealComparison{
			Left:     ast.NewIntLiteral(5),
			Operator: "<",
			Right:    ast.NewIntLiteral(20),
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

// === Combined CASE with nested IF-THEN-ELSE ===

func (s *QuantifiersCaseSuite) TestCase_WithNestedIfElse() {
	// CASE TRUE → (IF TRUE THEN 10 ELSE 20) □ OTHER → 0
	node := &ast.ExpressionCase{
		Branches: []*ast.CaseBranch{
			{
				Condition: &ast.BooleanLiteral{Value: true},
				Result: &ast.ExpressionIfElse{
					Condition: &ast.BooleanLiteral{Value: true},
					Then:      ast.NewIntLiteral(10),
					Else:      ast.NewIntLiteral(20),
				},
			},
		},
		Other: ast.NewIntLiteral(0),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("10", num.Inspect())
}

// === Quantifier with Nested Logic ===

func (s *QuantifiersCaseSuite) TestForAll_WithLogicInfix() {
	// ∀x ∈ {1, 2} : TRUE ∧ TRUE = TRUE
	node := &ast.LogicBoundQuantifier{
		Quantifier: "∀",
		Membership: &ast.LogicMembership{
			Left:     &ast.Identifier{Value: "x"},
			Operator: "∈",
			Right:    &ast.SetLiteralInt{Values: []int{1, 2}},
		},
		Predicate: &ast.LogicInfixExpression{
			Left:     &ast.BooleanLiteral{Value: true},
			Operator: "∧",
			Right:    &ast.BooleanLiteral{Value: true},
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

func (s *QuantifiersCaseSuite) TestExists_WithLogicOr() {
	// ∃x ∈ {1} : FALSE ∨ TRUE = TRUE
	node := &ast.LogicBoundQuantifier{
		Quantifier: "∃",
		Membership: &ast.LogicMembership{
			Left:     &ast.Identifier{Value: "x"},
			Operator: "∈",
			Right:    &ast.SetLiteralInt{Values: []int{1}},
		},
		Predicate: &ast.LogicInfixExpression{
			Left:     &ast.BooleanLiteral{Value: false},
			Operator: "∨",
			Right:    &ast.BooleanLiteral{Value: true},
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

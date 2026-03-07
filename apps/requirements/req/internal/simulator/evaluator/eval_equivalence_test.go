package evaluator

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/stretchr/testify/suite"
)

func TestEquivalenceSuite(t *testing.T) {
	suite.Run(t, new(EquivalenceSuite))
}

type EquivalenceSuite struct {
	suite.Suite
}

// =============================================================================
// Basic Equivalence (≡) Truth Table
// =============================================================================

func (s *EquivalenceSuite) TestEquivalence_TrueTrue() {
	// TRUE ≡ TRUE = TRUE
	node := &ast.BinaryLogic{
		Left:     &ast.BooleanLiteral{Value: true},
		Operator: "≡",
		Right:    &ast.BooleanLiteral{Value: true},
	}
	bindings := NewBindings()

	result := EvalAST(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

func (s *EquivalenceSuite) TestEquivalence_TrueFalse() {
	// TRUE ≡ FALSE = FALSE
	node := &ast.BinaryLogic{
		Left:     &ast.BooleanLiteral{Value: true},
		Operator: "≡",
		Right:    &ast.BooleanLiteral{Value: false},
	}
	bindings := NewBindings()

	result := EvalAST(node, bindings)

	s.False(result.IsError())
	s.False(result.Value.(*object.Boolean).Value())
}

func (s *EquivalenceSuite) TestEquivalence_FalseTrue() {
	// FALSE ≡ TRUE = FALSE
	node := &ast.BinaryLogic{
		Left:     &ast.BooleanLiteral{Value: false},
		Operator: "≡",
		Right:    &ast.BooleanLiteral{Value: true},
	}
	bindings := NewBindings()

	result := EvalAST(node, bindings)

	s.False(result.IsError())
	s.False(result.Value.(*object.Boolean).Value())
}

func (s *EquivalenceSuite) TestEquivalence_FalseFalse() {
	// FALSE ≡ FALSE = TRUE
	node := &ast.BinaryLogic{
		Left:     &ast.BooleanLiteral{Value: false},
		Operator: "≡",
		Right:    &ast.BooleanLiteral{Value: false},
	}
	bindings := NewBindings()

	result := EvalAST(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

// =============================================================================
// Equivalence with Nested Expressions
// =============================================================================

func (s *EquivalenceSuite) TestEquivalence_WithAndExpressions() {
	// (TRUE ∧ FALSE) ≡ FALSE = FALSE ≡ FALSE = TRUE
	node := &ast.BinaryLogic{
		Left: &ast.BinaryLogic{
			Left:     &ast.BooleanLiteral{Value: true},
			Operator: "∧",
			Right:    &ast.BooleanLiteral{Value: false},
		},
		Operator: "≡",
		Right:    &ast.BooleanLiteral{Value: false},
	}
	bindings := NewBindings()

	result := EvalAST(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

func (s *EquivalenceSuite) TestEquivalence_WithOrExpressions() {
	// (TRUE ∨ FALSE) ≡ TRUE = TRUE ≡ TRUE = TRUE
	node := &ast.BinaryLogic{
		Left: &ast.BinaryLogic{
			Left:     &ast.BooleanLiteral{Value: true},
			Operator: "∨",
			Right:    &ast.BooleanLiteral{Value: false},
		},
		Operator: "≡",
		Right:    &ast.BooleanLiteral{Value: true},
	}
	bindings := NewBindings()

	result := EvalAST(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

func (s *EquivalenceSuite) TestEquivalence_WithNotExpressions() {
	// ¬TRUE ≡ FALSE = FALSE ≡ FALSE = TRUE
	node := &ast.BinaryLogic{
		Left: &ast.UnaryLogic{
			Operator: "¬",
			Right:    &ast.BooleanLiteral{Value: true},
		},
		Operator: "≡",
		Right:    &ast.BooleanLiteral{Value: false},
	}
	bindings := NewBindings()

	result := EvalAST(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

func (s *EquivalenceSuite) TestEquivalence_DoubleNegation() {
	// ¬¬TRUE ≡ TRUE = TRUE ≡ TRUE = TRUE
	node := &ast.BinaryLogic{
		Left: &ast.UnaryLogic{
			Operator: "¬",
			Right: &ast.UnaryLogic{
				Operator: "¬",
				Right:    &ast.BooleanLiteral{Value: true},
			},
		},
		Operator: "≡",
		Right:    &ast.BooleanLiteral{Value: true},
	}
	bindings := NewBindings()

	result := EvalAST(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

// =============================================================================
// Equivalence with Comparisons
// =============================================================================

func (s *EquivalenceSuite) TestEquivalence_WithComparisons() {
	// (3 < 5) ≡ (5 > 3) = TRUE ≡ TRUE = TRUE
	node := &ast.BinaryLogic{
		Left: &ast.BinaryComparison{
			Left:     ast.NewIntLiteral(3),
			Operator: "<",
			Right:    ast.NewIntLiteral(5),
		},
		Operator: "≡",
		Right: &ast.BinaryComparison{
			Left:     ast.NewIntLiteral(5),
			Operator: ">",
			Right:    ast.NewIntLiteral(3),
		},
	}
	bindings := NewBindings()

	result := EvalAST(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

func (s *EquivalenceSuite) TestEquivalence_ComparisonMismatch() {
	// (3 < 5) ≡ (3 > 5) = TRUE ≡ FALSE = FALSE
	node := &ast.BinaryLogic{
		Left: &ast.BinaryComparison{
			Left:     ast.NewIntLiteral(3),
			Operator: "<",
			Right:    ast.NewIntLiteral(5),
		},
		Operator: "≡",
		Right: &ast.BinaryComparison{
			Left:     ast.NewIntLiteral(3),
			Operator: ">",
			Right:    ast.NewIntLiteral(5),
		},
	}
	bindings := NewBindings()

	result := EvalAST(node, bindings)

	s.False(result.IsError())
	s.False(result.Value.(*object.Boolean).Value())
}

func (s *EquivalenceSuite) TestEquivalence_LessOrEqualVsGreaterOrEqual() {
	// (5 ≤ 5) ≡ (5 ≥ 5) = TRUE ≡ TRUE = TRUE
	node := &ast.BinaryLogic{
		Left: &ast.BinaryComparison{
			Left:     ast.NewIntLiteral(5),
			Operator: "≤",
			Right:    ast.NewIntLiteral(5),
		},
		Operator: "≡",
		Right: &ast.BinaryComparison{
			Left:     ast.NewIntLiteral(5),
			Operator: "≥",
			Right:    ast.NewIntLiteral(5),
		},
	}
	bindings := NewBindings()

	result := EvalAST(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

// =============================================================================
// Equivalence with Equality
// =============================================================================

func (s *EquivalenceSuite) TestEquivalence_WithEquality() {
	// (5 = 5) ≡ TRUE = TRUE ≡ TRUE = TRUE
	node := &ast.BinaryLogic{
		Left: &ast.BinaryEquality{
			Left:     ast.NewIntLiteral(5),
			Operator: "=",
			Right:    ast.NewIntLiteral(5),
		},
		Operator: "≡",
		Right:    &ast.BooleanLiteral{Value: true},
	}
	bindings := NewBindings()

	result := EvalAST(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

func (s *EquivalenceSuite) TestEquivalence_EqualityVsNotEquality() {
	// (5 = 5) ≡ (5 ≠ 6) = TRUE ≡ TRUE = TRUE
	node := &ast.BinaryLogic{
		Left: &ast.BinaryEquality{
			Left:     ast.NewIntLiteral(5),
			Operator: "=",
			Right:    ast.NewIntLiteral(5),
		},
		Operator: "≡",
		Right: &ast.BinaryEquality{
			Left:     ast.NewIntLiteral(5),
			Operator: "≠",
			Right:    ast.NewIntLiteral(6),
		},
	}
	bindings := NewBindings()

	result := EvalAST(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

// =============================================================================
// Equivalence Chaining (left-associative)
// =============================================================================

func (s *EquivalenceSuite) TestEquivalence_Chained() {
	// TRUE ≡ TRUE ≡ TRUE = (TRUE ≡ TRUE) ≡ TRUE = TRUE ≡ TRUE = TRUE
	node := &ast.BinaryLogic{
		Left: &ast.BinaryLogic{
			Left:     &ast.BooleanLiteral{Value: true},
			Operator: "≡",
			Right:    &ast.BooleanLiteral{Value: true},
		},
		Operator: "≡",
		Right:    &ast.BooleanLiteral{Value: true},
	}
	bindings := NewBindings()

	result := EvalAST(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

func (s *EquivalenceSuite) TestEquivalence_ChainedFalse() {
	// TRUE ≡ FALSE ≡ FALSE = (TRUE ≡ FALSE) ≡ FALSE = FALSE ≡ FALSE = TRUE
	node := &ast.BinaryLogic{
		Left: &ast.BinaryLogic{
			Left:     &ast.BooleanLiteral{Value: true},
			Operator: "≡",
			Right:    &ast.BooleanLiteral{Value: false},
		},
		Operator: "≡",
		Right:    &ast.BooleanLiteral{Value: false},
	}
	bindings := NewBindings()

	result := EvalAST(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

// =============================================================================
// Equivalence with Implication
// =============================================================================

func (s *EquivalenceSuite) TestEquivalence_WithImplication() {
	// (TRUE ⇒ FALSE) ≡ FALSE = FALSE ≡ FALSE = TRUE
	node := &ast.BinaryLogic{
		Left: &ast.BinaryLogic{
			Left:     &ast.BooleanLiteral{Value: true},
			Operator: "⇒",
			Right:    &ast.BooleanLiteral{Value: false},
		},
		Operator: "≡",
		Right:    &ast.BooleanLiteral{Value: false},
	}
	bindings := NewBindings()

	result := EvalAST(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

func (s *EquivalenceSuite) TestEquivalence_ImplicationTautology() {
	// (FALSE ⇒ TRUE) ≡ TRUE = TRUE ≡ TRUE = TRUE
	// This tests the "ex falso quodlibet" property
	node := &ast.BinaryLogic{
		Left: &ast.BinaryLogic{
			Left:     &ast.BooleanLiteral{Value: false},
			Operator: "⇒",
			Right:    &ast.BooleanLiteral{Value: true},
		},
		Operator: "≡",
		Right:    &ast.BooleanLiteral{Value: true},
	}
	bindings := NewBindings()

	result := EvalAST(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

func (s *EquivalenceSuite) TestEquivalence_VacuousImplication() {
	// (FALSE ⇒ FALSE) ≡ TRUE = TRUE ≡ TRUE = TRUE
	// Vacuously true implication
	node := &ast.BinaryLogic{
		Left: &ast.BinaryLogic{
			Left:     &ast.BooleanLiteral{Value: false},
			Operator: "⇒",
			Right:    &ast.BooleanLiteral{Value: false},
		},
		Operator: "≡",
		Right:    &ast.BooleanLiteral{Value: true},
	}
	bindings := NewBindings()

	result := EvalAST(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

// =============================================================================
// Complex Expressions
// =============================================================================

func (s *EquivalenceSuite) TestEquivalence_DeMorganLaw1() {
	// ¬(TRUE ∧ FALSE) ≡ (¬TRUE ∨ ¬FALSE) = TRUE ≡ TRUE = TRUE
	// De Morgan's Law: ¬(A ∧ B) ≡ (¬A ∨ ¬B)
	left := &ast.UnaryLogic{
		Operator: "¬",
		Right: &ast.BinaryLogic{
			Left:     &ast.BooleanLiteral{Value: true},
			Operator: "∧",
			Right:    &ast.BooleanLiteral{Value: false},
		},
	}
	right := &ast.BinaryLogic{
		Left: &ast.UnaryLogic{
			Operator: "¬",
			Right:    &ast.BooleanLiteral{Value: true},
		},
		Operator: "∨",
		Right: &ast.UnaryLogic{
			Operator: "¬",
			Right:    &ast.BooleanLiteral{Value: false},
		},
	}
	node := &ast.BinaryLogic{
		Left:     left,
		Operator: "≡",
		Right:    right,
	}
	bindings := NewBindings()

	result := EvalAST(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

func (s *EquivalenceSuite) TestEquivalence_DeMorganLaw2() {
	// ¬(TRUE ∨ FALSE) ≡ (¬TRUE ∧ ¬FALSE) = FALSE ≡ FALSE = TRUE
	// De Morgan's Law: ¬(A ∨ B) ≡ (¬A ∧ ¬B)
	left := &ast.UnaryLogic{
		Operator: "¬",
		Right: &ast.BinaryLogic{
			Left:     &ast.BooleanLiteral{Value: true},
			Operator: "∨",
			Right:    &ast.BooleanLiteral{Value: false},
		},
	}
	right := &ast.BinaryLogic{
		Left: &ast.UnaryLogic{
			Operator: "¬",
			Right:    &ast.BooleanLiteral{Value: true},
		},
		Operator: "∧",
		Right: &ast.UnaryLogic{
			Operator: "¬",
			Right:    &ast.BooleanLiteral{Value: false},
		},
	}
	node := &ast.BinaryLogic{
		Left:     left,
		Operator: "≡",
		Right:    right,
	}
	bindings := NewBindings()

	result := EvalAST(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

// =============================================================================
// ASCII Operator Variant (<=>)
// =============================================================================

func (s *EquivalenceSuite) TestEquivalence_AsciiOperator() {
	// TRUE <=> TRUE = TRUE (using ASCII operator)
	node := &ast.BinaryLogic{
		Left:     &ast.BooleanLiteral{Value: true},
		Operator: "<=>",
		Right:    &ast.BooleanLiteral{Value: true},
	}
	bindings := NewBindings()

	result := EvalAST(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

func (s *EquivalenceSuite) TestEquivalence_AsciiOperator_False() {
	// TRUE <=> FALSE = FALSE (using ASCII operator)
	node := &ast.BinaryLogic{
		Left:     &ast.BooleanLiteral{Value: true},
		Operator: "<=>",
		Right:    &ast.BooleanLiteral{Value: false},
	}
	bindings := NewBindings()

	result := EvalAST(node, bindings)

	s.False(result.IsError())
	s.False(result.Value.(*object.Boolean).Value())
}

package evaluator

import (
	"testing"

	"github.com/glemzurg/go-tlaplus/internal/simulator/ast"
	"github.com/glemzurg/go-tlaplus/internal/simulator/object"
	"github.com/stretchr/testify/suite"
)

func TestComparisonSuite(t *testing.T) {
	suite.Run(t, new(ComparisonSuite))
}

type ComparisonSuite struct {
	suite.Suite
}

// =============================================================================
// Less Than (<)
// =============================================================================

func (s *ComparisonSuite) TestLessThan_True() {
	node := &ast.LogicRealComparison{
		Left:     ast.NewIntLiteral(3),
		Operator: "<",
		Right:    ast.NewIntLiteral(5),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

func (s *ComparisonSuite) TestLessThan_False_Greater() {
	node := &ast.LogicRealComparison{
		Left:     ast.NewIntLiteral(5),
		Operator: "<",
		Right:    ast.NewIntLiteral(3),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.False(result.Value.(*object.Boolean).Value())
}

func (s *ComparisonSuite) TestLessThan_False_Equal() {
	node := &ast.LogicRealComparison{
		Left:     ast.NewIntLiteral(5),
		Operator: "<",
		Right:    ast.NewIntLiteral(5),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.False(result.Value.(*object.Boolean).Value())
}

func (s *ComparisonSuite) TestLessThan_NegativeNumbers() {
	// -5 < -3 = TRUE
	node := &ast.LogicRealComparison{
		Left:     ast.NewIntLiteral(-5),
		Operator: "<",
		Right:    ast.NewIntLiteral(-3),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

func (s *ComparisonSuite) TestLessThan_Zero() {
	// -1 < 0 = TRUE
	node := &ast.LogicRealComparison{
		Left:     ast.NewIntLiteral(-1),
		Operator: "<",
		Right:    ast.NewIntLiteral(0),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

func (s *ComparisonSuite) TestLessThan_Rationals() {
	// 1/3 < 1/2 = TRUE
	node := &ast.LogicRealComparison{
		Left:     ast.NewFractionExpr(ast.NewIntLiteral(1), ast.NewIntLiteral(3)),
		Operator: "<",
		Right:    ast.NewFractionExpr(ast.NewIntLiteral(1), ast.NewIntLiteral(2)),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

func (s *ComparisonSuite) TestLessThan_MixedTypes() {
	// 1/2 < 1 = TRUE (rational vs natural)
	node := &ast.LogicRealComparison{
		Left:     ast.NewFractionExpr(ast.NewIntLiteral(1), ast.NewIntLiteral(2)),
		Operator: "<",
		Right:    ast.NewIntLiteral(1),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

// =============================================================================
// Greater Than (>)
// =============================================================================

func (s *ComparisonSuite) TestGreaterThan_True() {
	node := &ast.LogicRealComparison{
		Left:     ast.NewIntLiteral(5),
		Operator: ">",
		Right:    ast.NewIntLiteral(3),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

func (s *ComparisonSuite) TestGreaterThan_False_Less() {
	node := &ast.LogicRealComparison{
		Left:     ast.NewIntLiteral(3),
		Operator: ">",
		Right:    ast.NewIntLiteral(5),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.False(result.Value.(*object.Boolean).Value())
}

func (s *ComparisonSuite) TestGreaterThan_False_Equal() {
	node := &ast.LogicRealComparison{
		Left:     ast.NewIntLiteral(5),
		Operator: ">",
		Right:    ast.NewIntLiteral(5),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.False(result.Value.(*object.Boolean).Value())
}

func (s *ComparisonSuite) TestGreaterThan_NegativeNumbers() {
	// -3 > -5 = TRUE
	node := &ast.LogicRealComparison{
		Left:     ast.NewIntLiteral(-3),
		Operator: ">",
		Right:    ast.NewIntLiteral(-5),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

func (s *ComparisonSuite) TestGreaterThan_Zero() {
	// 0 > -1 = TRUE
	node := &ast.LogicRealComparison{
		Left:     ast.NewIntLiteral(0),
		Operator: ">",
		Right:    ast.NewIntLiteral(-1),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

func (s *ComparisonSuite) TestGreaterThan_Rationals() {
	// 1/2 > 1/3 = TRUE
	node := &ast.LogicRealComparison{
		Left:     ast.NewFractionExpr(ast.NewIntLiteral(1), ast.NewIntLiteral(2)),
		Operator: ">",
		Right:    ast.NewFractionExpr(ast.NewIntLiteral(1), ast.NewIntLiteral(3)),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

// =============================================================================
// Less Than or Equal (≤)
// =============================================================================

func (s *ComparisonSuite) TestLessOrEqual_Less() {
	node := &ast.LogicRealComparison{
		Left:     ast.NewIntLiteral(3),
		Operator: "≤",
		Right:    ast.NewIntLiteral(5),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

func (s *ComparisonSuite) TestLessOrEqual_Equal() {
	node := &ast.LogicRealComparison{
		Left:     ast.NewIntLiteral(5),
		Operator: "≤",
		Right:    ast.NewIntLiteral(5),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

func (s *ComparisonSuite) TestLessOrEqual_Greater() {
	node := &ast.LogicRealComparison{
		Left:     ast.NewIntLiteral(5),
		Operator: "≤",
		Right:    ast.NewIntLiteral(3),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.False(result.Value.(*object.Boolean).Value())
}

func (s *ComparisonSuite) TestLessOrEqual_NegativeNumbers() {
	// -5 ≤ -5 = TRUE
	node := &ast.LogicRealComparison{
		Left:     ast.NewIntLiteral(-5),
		Operator: "≤",
		Right:    ast.NewIntLiteral(-5),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

func (s *ComparisonSuite) TestLessOrEqual_Rationals() {
	// 1/2 ≤ 2/4 = TRUE (equal)
	node := &ast.LogicRealComparison{
		Left:     ast.NewFractionExpr(ast.NewIntLiteral(1), ast.NewIntLiteral(2)),
		Operator: "≤",
		Right:    ast.NewFractionExpr(ast.NewIntLiteral(2), ast.NewIntLiteral(4)),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

func (s *ComparisonSuite) TestLessOrEqual_RationalLessThanInteger() {
	// 3/2 ≤ 2 = TRUE (1.5 ≤ 2)
	node := &ast.LogicRealComparison{
		Left:     ast.NewFractionExpr(ast.NewIntLiteral(3), ast.NewIntLiteral(2)),
		Operator: "≤",
		Right:    ast.NewIntLiteral(2),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

// =============================================================================
// Greater Than or Equal (≥)
// =============================================================================

func (s *ComparisonSuite) TestGreaterOrEqual_Greater() {
	node := &ast.LogicRealComparison{
		Left:     ast.NewIntLiteral(5),
		Operator: "≥",
		Right:    ast.NewIntLiteral(3),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

func (s *ComparisonSuite) TestGreaterOrEqual_Equal() {
	node := &ast.LogicRealComparison{
		Left:     ast.NewIntLiteral(5),
		Operator: "≥",
		Right:    ast.NewIntLiteral(5),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

func (s *ComparisonSuite) TestGreaterOrEqual_Less() {
	node := &ast.LogicRealComparison{
		Left:     ast.NewIntLiteral(3),
		Operator: "≥",
		Right:    ast.NewIntLiteral(5),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.False(result.Value.(*object.Boolean).Value())
}

func (s *ComparisonSuite) TestGreaterOrEqual_Zero() {
	// 0 ≥ 0 = TRUE
	node := &ast.LogicRealComparison{
		Left:     ast.NewIntLiteral(0),
		Operator: "≥",
		Right:    ast.NewIntLiteral(0),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

func (s *ComparisonSuite) TestGreaterOrEqual_Rationals() {
	// 2/3 ≥ 1/2 = TRUE
	node := &ast.LogicRealComparison{
		Left:     ast.NewFractionExpr(ast.NewIntLiteral(2), ast.NewIntLiteral(3)),
		Operator: "≥",
		Right:    ast.NewFractionExpr(ast.NewIntLiteral(1), ast.NewIntLiteral(2)),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

func (s *ComparisonSuite) TestGreaterOrEqual_IntegerEqualToRational() {
	// 2 ≥ 4/2 = TRUE (2 ≥ 2)
	node := &ast.LogicRealComparison{
		Left:     ast.NewIntLiteral(2),
		Operator: "≥",
		Right:    ast.NewFractionExpr(ast.NewIntLiteral(4), ast.NewIntLiteral(2)),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

// =============================================================================
// Comparisons with Arithmetic Expressions
// =============================================================================

func (s *ComparisonSuite) TestComparison_WithArithmetic() {
	// (2 + 3) > 4 = 5 > 4 = TRUE
	node := &ast.LogicRealComparison{
		Left: &ast.RealInfixExpression{
			Left:     ast.NewIntLiteral(2),
			Operator: "+",
			Right:    ast.NewIntLiteral(3),
		},
		Operator: ">",
		Right:    ast.NewIntLiteral(4),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

func (s *ComparisonSuite) TestComparison_BothSidesArithmetic() {
	// (2 * 3) < (4 + 5) = 6 < 9 = TRUE
	node := &ast.LogicRealComparison{
		Left: &ast.RealInfixExpression{
			Left:     ast.NewIntLiteral(2),
			Operator: "*",
			Right:    ast.NewIntLiteral(3),
		},
		Operator: "<",
		Right: &ast.RealInfixExpression{
			Left:     ast.NewIntLiteral(4),
			Operator: "+",
			Right:    ast.NewIntLiteral(5),
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

func (s *ComparisonSuite) TestComparison_WithPower() {
	// 2^3 ≥ 8 = 8 ≥ 8 = TRUE
	node := &ast.LogicRealComparison{
		Left: &ast.RealInfixExpression{
			Left:     ast.NewIntLiteral(2),
			Operator: "^",
			Right:    ast.NewIntLiteral(3),
		},
		Operator: "≥",
		Right:    ast.NewIntLiteral(8),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

// =============================================================================
// Comparisons in Logic Expressions
// =============================================================================

func (s *ComparisonSuite) TestComparison_InAndExpression() {
	// (3 < 5) ∧ (5 > 3) = TRUE ∧ TRUE = TRUE
	node := &ast.LogicInfixExpression{
		Left: &ast.LogicRealComparison{
			Left:     ast.NewIntLiteral(3),
			Operator: "<",
			Right:    ast.NewIntLiteral(5),
		},
		Operator: "∧",
		Right: &ast.LogicRealComparison{
			Left:     ast.NewIntLiteral(5),
			Operator: ">",
			Right:    ast.NewIntLiteral(3),
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

func (s *ComparisonSuite) TestComparison_InOrExpression() {
	// (3 > 5) ∨ (5 > 3) = FALSE ∨ TRUE = TRUE
	node := &ast.LogicInfixExpression{
		Left: &ast.LogicRealComparison{
			Left:     ast.NewIntLiteral(3),
			Operator: ">",
			Right:    ast.NewIntLiteral(5),
		},
		Operator: "∨",
		Right: &ast.LogicRealComparison{
			Left:     ast.NewIntLiteral(5),
			Operator: ">",
			Right:    ast.NewIntLiteral(3),
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

func (s *ComparisonSuite) TestComparison_InImplication() {
	// (x < y) ⇒ (x ≤ y): for x=3, y=5: TRUE ⇒ TRUE = TRUE
	node := &ast.LogicInfixExpression{
		Left: &ast.LogicRealComparison{
			Left:     ast.NewIntLiteral(3),
			Operator: "<",
			Right:    ast.NewIntLiteral(5),
		},
		Operator: "⇒",
		Right: &ast.LogicRealComparison{
			Left:     ast.NewIntLiteral(3),
			Operator: "≤",
			Right:    ast.NewIntLiteral(5),
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

func (s *ComparisonSuite) TestComparison_NegatedComparison() {
	// ¬(3 > 5) = ¬FALSE = TRUE
	node := &ast.LogicPrefixExpression{
		Operator: "¬",
		Right: &ast.LogicRealComparison{
			Left:     ast.NewIntLiteral(3),
			Operator: ">",
			Right:    ast.NewIntLiteral(5),
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

// =============================================================================
// Edge Cases
// =============================================================================

func (s *ComparisonSuite) TestComparison_LargeNumbers() {
	// 1000000 < 1000001 = TRUE
	node := &ast.LogicRealComparison{
		Left:     ast.NewIntLiteral(1000000),
		Operator: "<",
		Right:    ast.NewIntLiteral(1000001),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

func (s *ComparisonSuite) TestComparison_VerySmallRationals() {
	// 1/1000 < 1/100 = TRUE
	node := &ast.LogicRealComparison{
		Left:     ast.NewFractionExpr(ast.NewIntLiteral(1), ast.NewIntLiteral(1000)),
		Operator: "<",
		Right:    ast.NewFractionExpr(ast.NewIntLiteral(1), ast.NewIntLiteral(100)),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

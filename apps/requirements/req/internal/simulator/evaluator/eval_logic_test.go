package evaluator

import (
	"testing"

	"github.com/glemzurg/go-tlaplus/internal/simulator/ast"
	"github.com/glemzurg/go-tlaplus/internal/simulator/object"
	"github.com/stretchr/testify/suite"
)

func TestLogicSuite(t *testing.T) {
	suite.Run(t, new(LogicSuite))
}

type LogicSuite struct {
	suite.Suite
}

// === Logic AND (∧) ===

func (s *LogicSuite) TestAnd_TrueTrue() {
	node := &ast.LogicInfixExpression{
		Left:     &ast.BooleanLiteral{Value: true},
		Operator: "∧",
		Right:    &ast.BooleanLiteral{Value: true},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

func (s *LogicSuite) TestAnd_TrueFalse() {
	node := &ast.LogicInfixExpression{
		Left:     &ast.BooleanLiteral{Value: true},
		Operator: "∧",
		Right:    &ast.BooleanLiteral{Value: false},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.False(b.Value())
}

func (s *LogicSuite) TestAnd_FalseTrue() {
	node := &ast.LogicInfixExpression{
		Left:     &ast.BooleanLiteral{Value: false},
		Operator: "∧",
		Right:    &ast.BooleanLiteral{Value: true},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.False(b.Value())
}

func (s *LogicSuite) TestAnd_FalseFalse() {
	node := &ast.LogicInfixExpression{
		Left:     &ast.BooleanLiteral{Value: false},
		Operator: "∧",
		Right:    &ast.BooleanLiteral{Value: false},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.False(b.Value())
}

// === Logic OR (∨) ===

func (s *LogicSuite) TestOr_TrueTrue() {
	node := &ast.LogicInfixExpression{
		Left:     &ast.BooleanLiteral{Value: true},
		Operator: "∨",
		Right:    &ast.BooleanLiteral{Value: true},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

func (s *LogicSuite) TestOr_TrueFalse() {
	node := &ast.LogicInfixExpression{
		Left:     &ast.BooleanLiteral{Value: true},
		Operator: "∨",
		Right:    &ast.BooleanLiteral{Value: false},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

func (s *LogicSuite) TestOr_FalseFalse() {
	node := &ast.LogicInfixExpression{
		Left:     &ast.BooleanLiteral{Value: false},
		Operator: "∨",
		Right:    &ast.BooleanLiteral{Value: false},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.False(b.Value())
}

// === Logic NOT (¬) ===

func (s *LogicSuite) TestNot_True() {
	node := &ast.LogicPrefixExpression{
		Operator: "¬",
		Right:    &ast.BooleanLiteral{Value: true},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.False(b.Value())
}

func (s *LogicSuite) TestNot_False() {
	node := &ast.LogicPrefixExpression{
		Operator: "¬",
		Right:    &ast.BooleanLiteral{Value: false},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

// === Implication (⇒) ===

func (s *LogicSuite) TestImplication_TrueTrue() {
	// TRUE ⇒ TRUE = TRUE
	node := &ast.LogicInfixExpression{
		Left:     &ast.BooleanLiteral{Value: true},
		Operator: "⇒",
		Right:    &ast.BooleanLiteral{Value: true},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

func (s *LogicSuite) TestImplication_TrueFalse() {
	// TRUE ⇒ FALSE = FALSE
	node := &ast.LogicInfixExpression{
		Left:     &ast.BooleanLiteral{Value: true},
		Operator: "⇒",
		Right:    &ast.BooleanLiteral{Value: false},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.False(b.Value())
}

func (s *LogicSuite) TestImplication_FalseTrue() {
	// FALSE ⇒ TRUE = TRUE
	node := &ast.LogicInfixExpression{
		Left:     &ast.BooleanLiteral{Value: false},
		Operator: "⇒",
		Right:    &ast.BooleanLiteral{Value: true},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

func (s *LogicSuite) TestImplication_FalseFalse() {
	// FALSE ⇒ FALSE = TRUE (vacuously true)
	node := &ast.LogicInfixExpression{
		Left:     &ast.BooleanLiteral{Value: false},
		Operator: "⇒",
		Right:    &ast.BooleanLiteral{Value: false},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

// === Equivalence (≡) ===

func (s *LogicSuite) TestEquivalence_TrueTrue() {
	// TRUE ≡ TRUE = TRUE
	node := &ast.LogicInfixExpression{
		Left:     &ast.BooleanLiteral{Value: true},
		Operator: "≡",
		Right:    &ast.BooleanLiteral{Value: true},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

func (s *LogicSuite) TestEquivalence_TrueFalse() {
	// TRUE ≡ FALSE = FALSE
	node := &ast.LogicInfixExpression{
		Left:     &ast.BooleanLiteral{Value: true},
		Operator: "≡",
		Right:    &ast.BooleanLiteral{Value: false},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.False(b.Value())
}

func (s *LogicSuite) TestEquivalence_FalseFalse() {
	// FALSE ≡ FALSE = TRUE
	node := &ast.LogicInfixExpression{
		Left:     &ast.BooleanLiteral{Value: false},
		Operator: "≡",
		Right:    &ast.BooleanLiteral{Value: false},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

// === Numeric Comparisons ===

func (s *LogicSuite) TestComparison_LessThan() {
	// 3 < 5 = TRUE
	node := &ast.LogicRealComparison{
		Left:     ast.NewIntLiteral(3),
		Operator: "<",
		Right:    ast.NewIntLiteral(5),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

func (s *LogicSuite) TestComparison_LessThan_False() {
	// 5 < 3 = FALSE
	node := &ast.LogicRealComparison{
		Left:     ast.NewIntLiteral(5),
		Operator: "<",
		Right:    ast.NewIntLiteral(3),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.False(b.Value())
}

func (s *LogicSuite) TestComparison_GreaterThan() {
	// 5 > 3 = TRUE
	node := &ast.LogicRealComparison{
		Left:     ast.NewIntLiteral(5),
		Operator: ">",
		Right:    ast.NewIntLiteral(3),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

func (s *LogicSuite) TestComparison_LessThanOrEqual() {
	// 3 ≤ 3 = TRUE
	node := &ast.LogicRealComparison{
		Left:     ast.NewIntLiteral(3),
		Operator: "≤",
		Right:    ast.NewIntLiteral(3),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

func (s *LogicSuite) TestComparison_GreaterThanOrEqual() {
	// 3 ≥ 3 = TRUE
	node := &ast.LogicRealComparison{
		Left:     ast.NewIntLiteral(3),
		Operator: "≥",
		Right:    ast.NewIntLiteral(3),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

func (s *LogicSuite) TestComparison_Real() {
	// 1/3 < 1/2 = TRUE
	node := &ast.LogicRealComparison{
		Left:     ast.NewFractionExpr(ast.NewIntLiteral(1), ast.NewIntLiteral(3)),
		Operator: "<",
		Right:    ast.NewFractionExpr(ast.NewIntLiteral(1), ast.NewIntLiteral(2)),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

// === Nested Logic ===

func (s *LogicSuite) TestNested_AndOr() {
	// (TRUE ∧ FALSE) ∨ TRUE = FALSE ∨ TRUE = TRUE
	inner := &ast.LogicInfixExpression{
		Left:     &ast.BooleanLiteral{Value: true},
		Operator: "∧",
		Right:    &ast.BooleanLiteral{Value: false},
	}
	outer := &ast.LogicInfixExpression{
		Left:     inner,
		Operator: "∨",
		Right:    &ast.BooleanLiteral{Value: true},
	}
	bindings := NewBindings()

	result := Eval(outer, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

func (s *LogicSuite) TestNested_NotAnd() {
	// ¬(TRUE ∧ FALSE) = ¬FALSE = TRUE
	inner := &ast.LogicInfixExpression{
		Left:     &ast.BooleanLiteral{Value: true},
		Operator: "∧",
		Right:    &ast.BooleanLiteral{Value: false},
	}
	outer := &ast.LogicPrefixExpression{
		Operator: "¬",
		Right:    inner,
	}
	bindings := NewBindings()

	result := Eval(outer, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

// === Comparison as Logic ===

func (s *LogicSuite) TestComparison_InLogic() {
	// (3 < 5) ∧ (5 > 3) = TRUE ∧ TRUE = TRUE
	left := &ast.LogicRealComparison{
		Left:     ast.NewIntLiteral(3),
		Operator: "<",
		Right:    ast.NewIntLiteral(5),
	}
	right := &ast.LogicRealComparison{
		Left:     ast.NewIntLiteral(5),
		Operator: ">",
		Right:    ast.NewIntLiteral(3),
	}
	node := &ast.LogicInfixExpression{
		Left:     left,
		Operator: "∧",
		Right:    right,
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

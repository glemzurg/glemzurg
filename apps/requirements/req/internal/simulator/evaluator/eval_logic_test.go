package evaluator

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
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
	node := &ast.BinaryLogic{
		Left:     &ast.BooleanLiteral{Value: true},
		Operator: "∧",
		Right:    &ast.BooleanLiteral{Value: true},
	}
	bindings := NewBindings()

	result := EvalAST(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

func (s *LogicSuite) TestAnd_TrueFalse() {
	node := &ast.BinaryLogic{
		Left:     &ast.BooleanLiteral{Value: true},
		Operator: "∧",
		Right:    &ast.BooleanLiteral{Value: false},
	}
	bindings := NewBindings()

	result := EvalAST(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.False(b.Value())
}

func (s *LogicSuite) TestAnd_FalseTrue() {
	node := &ast.BinaryLogic{
		Left:     &ast.BooleanLiteral{Value: false},
		Operator: "∧",
		Right:    &ast.BooleanLiteral{Value: true},
	}
	bindings := NewBindings()

	result := EvalAST(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.False(b.Value())
}

func (s *LogicSuite) TestAnd_FalseFalse() {
	node := &ast.BinaryLogic{
		Left:     &ast.BooleanLiteral{Value: false},
		Operator: "∧",
		Right:    &ast.BooleanLiteral{Value: false},
	}
	bindings := NewBindings()

	result := EvalAST(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.False(b.Value())
}

// === Logic OR (∨) ===

func (s *LogicSuite) TestOr_TrueTrue() {
	node := &ast.BinaryLogic{
		Left:     &ast.BooleanLiteral{Value: true},
		Operator: "∨",
		Right:    &ast.BooleanLiteral{Value: true},
	}
	bindings := NewBindings()

	result := EvalAST(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

func (s *LogicSuite) TestOr_TrueFalse() {
	node := &ast.BinaryLogic{
		Left:     &ast.BooleanLiteral{Value: true},
		Operator: "∨",
		Right:    &ast.BooleanLiteral{Value: false},
	}
	bindings := NewBindings()

	result := EvalAST(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

func (s *LogicSuite) TestOr_FalseFalse() {
	node := &ast.BinaryLogic{
		Left:     &ast.BooleanLiteral{Value: false},
		Operator: "∨",
		Right:    &ast.BooleanLiteral{Value: false},
	}
	bindings := NewBindings()

	result := EvalAST(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.False(b.Value())
}

// === Logic NOT (¬) ===

func (s *LogicSuite) TestNot_True() {
	node := &ast.UnaryLogic{
		Operator: "¬",
		Right:    &ast.BooleanLiteral{Value: true},
	}
	bindings := NewBindings()

	result := EvalAST(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.False(b.Value())
}

func (s *LogicSuite) TestNot_False() {
	node := &ast.UnaryLogic{
		Operator: "¬",
		Right:    &ast.BooleanLiteral{Value: false},
	}
	bindings := NewBindings()

	result := EvalAST(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

// === Implication (⇒) ===

func (s *LogicSuite) TestImplication_TrueTrue() {
	// TRUE ⇒ TRUE = TRUE
	node := &ast.BinaryLogic{
		Left:     &ast.BooleanLiteral{Value: true},
		Operator: "⇒",
		Right:    &ast.BooleanLiteral{Value: true},
	}
	bindings := NewBindings()

	result := EvalAST(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

func (s *LogicSuite) TestImplication_TrueFalse() {
	// TRUE ⇒ FALSE = FALSE
	node := &ast.BinaryLogic{
		Left:     &ast.BooleanLiteral{Value: true},
		Operator: "⇒",
		Right:    &ast.BooleanLiteral{Value: false},
	}
	bindings := NewBindings()

	result := EvalAST(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.False(b.Value())
}

func (s *LogicSuite) TestImplication_FalseTrue() {
	// FALSE ⇒ TRUE = TRUE
	node := &ast.BinaryLogic{
		Left:     &ast.BooleanLiteral{Value: false},
		Operator: "⇒",
		Right:    &ast.BooleanLiteral{Value: true},
	}
	bindings := NewBindings()

	result := EvalAST(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

func (s *LogicSuite) TestImplication_FalseFalse() {
	// FALSE ⇒ FALSE = TRUE (vacuously true)
	node := &ast.BinaryLogic{
		Left:     &ast.BooleanLiteral{Value: false},
		Operator: "⇒",
		Right:    &ast.BooleanLiteral{Value: false},
	}
	bindings := NewBindings()

	result := EvalAST(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

// === Equivalence (≡) ===

func (s *LogicSuite) TestEquivalence_TrueTrue() {
	// TRUE ≡ TRUE = TRUE
	node := &ast.BinaryLogic{
		Left:     &ast.BooleanLiteral{Value: true},
		Operator: "≡",
		Right:    &ast.BooleanLiteral{Value: true},
	}
	bindings := NewBindings()

	result := EvalAST(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

func (s *LogicSuite) TestEquivalence_TrueFalse() {
	// TRUE ≡ FALSE = FALSE
	node := &ast.BinaryLogic{
		Left:     &ast.BooleanLiteral{Value: true},
		Operator: "≡",
		Right:    &ast.BooleanLiteral{Value: false},
	}
	bindings := NewBindings()

	result := EvalAST(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.False(b.Value())
}

func (s *LogicSuite) TestEquivalence_FalseFalse() {
	// FALSE ≡ FALSE = TRUE
	node := &ast.BinaryLogic{
		Left:     &ast.BooleanLiteral{Value: false},
		Operator: "≡",
		Right:    &ast.BooleanLiteral{Value: false},
	}
	bindings := NewBindings()

	result := EvalAST(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

// === Numeric Comparisons ===

func (s *LogicSuite) TestComparison_LessThan() {
	// 3 < 5 = TRUE
	node := &ast.BinaryComparison{
		Left:     ast.NewIntLiteral(3),
		Operator: "<",
		Right:    ast.NewIntLiteral(5),
	}
	bindings := NewBindings()

	result := EvalAST(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

func (s *LogicSuite) TestComparison_LessThan_False() {
	// 5 < 3 = FALSE
	node := &ast.BinaryComparison{
		Left:     ast.NewIntLiteral(5),
		Operator: "<",
		Right:    ast.NewIntLiteral(3),
	}
	bindings := NewBindings()

	result := EvalAST(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.False(b.Value())
}

func (s *LogicSuite) TestComparison_GreaterThan() {
	// 5 > 3 = TRUE
	node := &ast.BinaryComparison{
		Left:     ast.NewIntLiteral(5),
		Operator: ">",
		Right:    ast.NewIntLiteral(3),
	}
	bindings := NewBindings()

	result := EvalAST(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

func (s *LogicSuite) TestComparison_LessThanOrEqual() {
	// 3 ≤ 3 = TRUE
	node := &ast.BinaryComparison{
		Left:     ast.NewIntLiteral(3),
		Operator: "≤",
		Right:    ast.NewIntLiteral(3),
	}
	bindings := NewBindings()

	result := EvalAST(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

func (s *LogicSuite) TestComparison_GreaterThanOrEqual() {
	// 3 ≥ 3 = TRUE
	node := &ast.BinaryComparison{
		Left:     ast.NewIntLiteral(3),
		Operator: "≥",
		Right:    ast.NewIntLiteral(3),
	}
	bindings := NewBindings()

	result := EvalAST(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

func (s *LogicSuite) TestComparison_Real() {
	// 1/3 < 1/2 = TRUE
	node := &ast.BinaryComparison{
		Left:     ast.NewFraction(ast.NewIntLiteral(1), ast.NewIntLiteral(3)),
		Operator: "<",
		Right:    ast.NewFraction(ast.NewIntLiteral(1), ast.NewIntLiteral(2)),
	}
	bindings := NewBindings()

	result := EvalAST(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

// === Nested Logic ===

func (s *LogicSuite) TestNested_AndOr() {
	// (TRUE ∧ FALSE) ∨ TRUE = FALSE ∨ TRUE = TRUE
	inner := &ast.BinaryLogic{
		Left:     &ast.BooleanLiteral{Value: true},
		Operator: "∧",
		Right:    &ast.BooleanLiteral{Value: false},
	}
	outer := &ast.BinaryLogic{
		Left:     inner,
		Operator: "∨",
		Right:    &ast.BooleanLiteral{Value: true},
	}
	bindings := NewBindings()

	result := EvalAST(outer, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

func (s *LogicSuite) TestNested_NotAnd() {
	// ¬(TRUE ∧ FALSE) = ¬FALSE = TRUE
	inner := &ast.BinaryLogic{
		Left:     &ast.BooleanLiteral{Value: true},
		Operator: "∧",
		Right:    &ast.BooleanLiteral{Value: false},
	}
	outer := &ast.UnaryLogic{
		Operator: "¬",
		Right:    inner,
	}
	bindings := NewBindings()

	result := EvalAST(outer, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

// === Comparison as Logic ===

func (s *LogicSuite) TestComparison_InLogic() {
	// (3 < 5) ∧ (5 > 3) = TRUE ∧ TRUE = TRUE
	left := &ast.BinaryComparison{
		Left:     ast.NewIntLiteral(3),
		Operator: "<",
		Right:    ast.NewIntLiteral(5),
	}
	right := &ast.BinaryComparison{
		Left:     ast.NewIntLiteral(5),
		Operator: ">",
		Right:    ast.NewIntLiteral(3),
	}
	node := &ast.BinaryLogic{
		Left:     left,
		Operator: "∧",
		Right:    right,
	}
	bindings := NewBindings()

	result := EvalAST(node, bindings)

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

package evaluator

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/stretchr/testify/suite"
)

func TestLogicEqualitySuite(t *testing.T) {
	suite.Run(t, new(LogicEqualitySuite))
}

type LogicEqualitySuite struct {
	suite.Suite
}

// =============================================================================
// Number Equality
// =============================================================================

func (s *LogicEqualitySuite) TestNumberEqual_True() {
	node := &ast.LogicEquality{
		Operator: "=",
		Left:     ast.NewIntLiteral(42),
		Right:    ast.NewIntLiteral(42),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

func (s *LogicEqualitySuite) TestNumberEqual_False() {
	node := &ast.LogicEquality{
		Operator: "=",
		Left:     ast.NewIntLiteral(42),
		Right:    ast.NewIntLiteral(43),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.False(result.Value.(*object.Boolean).Value())
}

func (s *LogicEqualitySuite) TestNumberNotEqual_True() {
	node := &ast.LogicEquality{
		Operator: "≠",
		Left:     ast.NewIntLiteral(42),
		Right:    ast.NewIntLiteral(43),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

func (s *LogicEqualitySuite) TestNumberNotEqual_False() {
	node := &ast.LogicEquality{
		Operator: "≠",
		Left:     ast.NewIntLiteral(42),
		Right:    ast.NewIntLiteral(42),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.False(result.Value.(*object.Boolean).Value())
}

func (s *LogicEqualitySuite) TestNumberEqual_DifferentKindsSameValue() {
	// Natural 5 should equal Rational 10/2
	node := &ast.LogicEquality{
		Operator: "=",
		Left:     ast.NewIntLiteral(5),
		Right:    ast.NewFractionExpr(ast.NewIntLiteral(10), ast.NewIntLiteral(2)),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

// =============================================================================
// String Equality
// =============================================================================

func (s *LogicEqualitySuite) TestStringEqual_True() {
	node := &ast.LogicEquality{
		Operator: "=",
		Left:     &ast.StringLiteral{Value: "hello"},
		Right:    &ast.StringLiteral{Value: "hello"},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

func (s *LogicEqualitySuite) TestStringEqual_False() {
	node := &ast.LogicEquality{
		Operator: "=",
		Left:     &ast.StringLiteral{Value: "hello"},
		Right:    &ast.StringLiteral{Value: "world"},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.False(result.Value.(*object.Boolean).Value())
}

func (s *LogicEqualitySuite) TestStringNotEqual_True() {
	node := &ast.LogicEquality{
		Operator: "≠",
		Left:     &ast.StringLiteral{Value: "hello"},
		Right:    &ast.StringLiteral{Value: "world"},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

// =============================================================================
// Boolean Equality
// =============================================================================

func (s *LogicEqualitySuite) TestBooleanEqual_True() {
	node := &ast.LogicEquality{
		Operator: "=",
		Left:     &ast.BooleanLiteral{Value: true},
		Right:    &ast.BooleanLiteral{Value: true},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

func (s *LogicEqualitySuite) TestBooleanEqual_False() {
	node := &ast.LogicEquality{
		Operator: "=",
		Left:     &ast.BooleanLiteral{Value: true},
		Right:    &ast.BooleanLiteral{Value: false},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.False(result.Value.(*object.Boolean).Value())
}

func (s *LogicEqualitySuite) TestBooleanNotEqual_True() {
	node := &ast.LogicEquality{
		Operator: "≠",
		Left:     &ast.BooleanLiteral{Value: true},
		Right:    &ast.BooleanLiteral{Value: false},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

// =============================================================================
// Set Equality
// =============================================================================

func (s *LogicEqualitySuite) TestSetEqual_True() {
	// {1, 2, 3} = {1, 2, 3}
	node := &ast.LogicEquality{
		Operator: "=",
		Left:     &ast.SetLiteralInt{Values: []int{1, 2, 3}},
		Right:    &ast.SetLiteralInt{Values: []int{1, 2, 3}},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

func (s *LogicEqualitySuite) TestSetEqual_DifferentOrder() {
	// {1, 2, 3} = {3, 2, 1} - sets are unordered
	node := &ast.LogicEquality{
		Operator: "=",
		Left:     &ast.SetLiteralInt{Values: []int{1, 2, 3}},
		Right:    &ast.SetLiteralInt{Values: []int{3, 2, 1}},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

func (s *LogicEqualitySuite) TestSetEqual_False() {
	// {1, 2, 3} = {1, 2, 4}
	node := &ast.LogicEquality{
		Operator: "=",
		Left:     &ast.SetLiteralInt{Values: []int{1, 2, 3}},
		Right:    &ast.SetLiteralInt{Values: []int{1, 2, 4}},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.False(result.Value.(*object.Boolean).Value())
}

func (s *LogicEqualitySuite) TestSetNotEqual_True() {
	// {1, 2} ≠ {1, 2, 3}
	node := &ast.LogicEquality{
		Operator: "≠",
		Left:     &ast.SetLiteralInt{Values: []int{1, 2}},
		Right:    &ast.SetLiteralInt{Values: []int{1, 2, 3}},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

// =============================================================================
// Tuple Equality
// =============================================================================

func (s *LogicEqualitySuite) TestTupleEqual_True() {
	// <<1, 2, 3>> = <<1, 2, 3>>
	node := &ast.LogicEquality{
		Operator: "=",
		Left: &ast.TupleLiteral{Elements: []ast.Expression{
			ast.NewIntLiteral(1),
			ast.NewIntLiteral(2),
			ast.NewIntLiteral(3),
		}},
		Right: &ast.TupleLiteral{Elements: []ast.Expression{
			ast.NewIntLiteral(1),
			ast.NewIntLiteral(2),
			ast.NewIntLiteral(3),
		}},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

func (s *LogicEqualitySuite) TestTupleEqual_DifferentOrder() {
	// <<1, 2>> = <<2, 1>> - tuples are ordered
	node := &ast.LogicEquality{
		Operator: "=",
		Left: &ast.TupleLiteral{Elements: []ast.Expression{
			ast.NewIntLiteral(1),
			ast.NewIntLiteral(2),
		}},
		Right: &ast.TupleLiteral{Elements: []ast.Expression{
			ast.NewIntLiteral(2),
			ast.NewIntLiteral(1),
		}},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.False(result.Value.(*object.Boolean).Value())
}

func (s *LogicEqualitySuite) TestTupleEqual_DifferentLength() {
	// <<1, 2>> = <<1, 2, 3>>
	node := &ast.LogicEquality{
		Operator: "=",
		Left: &ast.TupleLiteral{Elements: []ast.Expression{
			ast.NewIntLiteral(1),
			ast.NewIntLiteral(2),
		}},
		Right: &ast.TupleLiteral{Elements: []ast.Expression{
			ast.NewIntLiteral(1),
			ast.NewIntLiteral(2),
			ast.NewIntLiteral(3),
		}},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.False(result.Value.(*object.Boolean).Value())
}

// =============================================================================
// Record Equality
// =============================================================================

func (s *LogicEqualitySuite) TestRecordEqual_True() {
	// [x |-> 1, y |-> 2] = [x |-> 1, y |-> 2]
	node := &ast.LogicEquality{
		Operator: "=",
		Left: &ast.RecordInstance{
			Bindings: []*ast.FieldBinding{
				{Field: &ast.Identifier{Value: "x"}, Expression: ast.NewIntLiteral(1)},
				{Field: &ast.Identifier{Value: "y"}, Expression: ast.NewIntLiteral(2)},
			},
		},
		Right: &ast.RecordInstance{
			Bindings: []*ast.FieldBinding{
				{Field: &ast.Identifier{Value: "x"}, Expression: ast.NewIntLiteral(1)},
				{Field: &ast.Identifier{Value: "y"}, Expression: ast.NewIntLiteral(2)},
			},
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

func (s *LogicEqualitySuite) TestRecordEqual_DifferentFieldOrder() {
	// [x |-> 1, y |-> 2] = [y |-> 2, x |-> 1] - fields can be in any order
	node := &ast.LogicEquality{
		Operator: "=",
		Left: &ast.RecordInstance{
			Bindings: []*ast.FieldBinding{
				{Field: &ast.Identifier{Value: "x"}, Expression: ast.NewIntLiteral(1)},
				{Field: &ast.Identifier{Value: "y"}, Expression: ast.NewIntLiteral(2)},
			},
		},
		Right: &ast.RecordInstance{
			Bindings: []*ast.FieldBinding{
				{Field: &ast.Identifier{Value: "y"}, Expression: ast.NewIntLiteral(2)},
				{Field: &ast.Identifier{Value: "x"}, Expression: ast.NewIntLiteral(1)},
			},
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

func (s *LogicEqualitySuite) TestRecordEqual_DifferentValue() {
	// [x |-> 1] = [x |-> 2]
	node := &ast.LogicEquality{
		Operator: "=",
		Left: &ast.RecordInstance{
			Bindings: []*ast.FieldBinding{
				{Field: &ast.Identifier{Value: "x"}, Expression: ast.NewIntLiteral(1)},
			},
		},
		Right: &ast.RecordInstance{
			Bindings: []*ast.FieldBinding{
				{Field: &ast.Identifier{Value: "x"}, Expression: ast.NewIntLiteral(2)},
			},
		},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.False(result.Value.(*object.Boolean).Value())
}

// =============================================================================
// Cross-Type Equality (always false for different types)
// =============================================================================

func (s *LogicEqualitySuite) TestCrossTypeEqual_NumberVsString() {
	// 42 = "42" - different types, always false
	node := &ast.LogicEquality{
		Operator: "=",
		Left:     ast.NewIntLiteral(42),
		Right:    &ast.StringLiteral{Value: "42"},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.False(result.Value.(*object.Boolean).Value())
}

func (s *LogicEqualitySuite) TestCrossTypeEqual_BooleanVsNumber() {
	// TRUE = 1 - different types, always false
	node := &ast.LogicEquality{
		Operator: "=",
		Left:     &ast.BooleanLiteral{Value: true},
		Right:    ast.NewIntLiteral(1),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.False(result.Value.(*object.Boolean).Value())
}

func (s *LogicEqualitySuite) TestCrossTypeNotEqual_NumberVsBoolean() {
	// 1 ≠ TRUE - different types, always true
	node := &ast.LogicEquality{
		Operator: "≠",
		Left:     ast.NewIntLiteral(1),
		Right:    &ast.BooleanLiteral{Value: true},
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

// =============================================================================
// ASCII Operator Variants
// =============================================================================

func (s *LogicEqualitySuite) TestNotEqual_SlashEquals() {
	// 1 /= 2 using ASCII operator
	node := &ast.LogicEquality{
		Operator: "/=",
		Left:     ast.NewIntLiteral(1),
		Right:    ast.NewIntLiteral(2),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

func (s *LogicEqualitySuite) TestNotEqual_Hash() {
	// 1 # 2 using hash operator
	node := &ast.LogicEquality{
		Operator: "#",
		Left:     ast.NewIntLiteral(1),
		Right:    ast.NewIntLiteral(2),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	s.True(result.Value.(*object.Boolean).Value())
}

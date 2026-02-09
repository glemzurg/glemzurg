package evaluator

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/stretchr/testify/suite"
)

func TestArithmeticSuite(t *testing.T) {
	suite.Run(t, new(ArithmeticSuite))
}

type ArithmeticSuite struct {
	suite.Suite
}

// === Addition ===

func (s *ArithmeticSuite) TestAddition_Natural() {
	// 2 + 3 = 5
	node := &ast.RealInfixExpression{
		Left:     ast.NewIntLiteral(2),
		Operator: "+",
		Right:    ast.NewIntLiteral(3),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("5", num.Inspect())
}

func (s *ArithmeticSuite) TestAddition_Integer() {
	// -2 + 5 = 3
	node := &ast.RealInfixExpression{
		Left:     ast.NewIntLiteral(-2),
		Operator: "+",
		Right:    ast.NewIntLiteral(5),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("3", num.Inspect())
}

func (s *ArithmeticSuite) TestAddition_Real() {
	// 1/2 + 1/4 = 3/4
	node := &ast.RealInfixExpression{
		Left:     ast.NewFractionExpr(ast.NewIntLiteral(1), ast.NewIntLiteral(2)),
		Operator: "+",
		Right:    ast.NewFractionExpr(ast.NewIntLiteral(1), ast.NewIntLiteral(4)),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("3/4", num.Inspect())
}

func (s *ArithmeticSuite) TestAddition_MixedTypes() {
	// 1/2 + 1 = 3/2
	node := &ast.RealInfixExpression{
		Left:     ast.NewFractionExpr(ast.NewIntLiteral(1), ast.NewIntLiteral(2)),
		Operator: "+",
		Right:    ast.NewIntLiteral(1),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("3/2", num.Inspect())
}

// === Subtraction ===

func (s *ArithmeticSuite) TestSubtraction_Natural() {
	// 5 - 3 = 2
	node := &ast.RealInfixExpression{
		Left:     ast.NewIntLiteral(5),
		Operator: "-",
		Right:    ast.NewIntLiteral(3),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("2", num.Inspect())
}

func (s *ArithmeticSuite) TestSubtraction_NegativeResult() {
	// 3 - 5 = -2
	node := &ast.RealInfixExpression{
		Left:     ast.NewIntLiteral(3),
		Operator: "-",
		Right:    ast.NewIntLiteral(5),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("-2", num.Inspect())
	s.Equal(object.KindInteger, num.Kind())
}

// === Multiplication ===

func (s *ArithmeticSuite) TestMultiplication_Natural() {
	// 3 * 4 = 12
	node := &ast.RealInfixExpression{
		Left:     ast.NewIntLiteral(3),
		Operator: "*",
		Right:    ast.NewIntLiteral(4),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("12", num.Inspect())
}

func (s *ArithmeticSuite) TestMultiplication_ByZero() {
	// 5 * 0 = 0
	node := &ast.RealInfixExpression{
		Left:     ast.NewIntLiteral(5),
		Operator: "*",
		Right:    ast.NewIntLiteral(0),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("0", num.Inspect())
}

func (s *ArithmeticSuite) TestMultiplication_Real() {
	// 1/2 * 2/3 = 1/3
	node := &ast.RealInfixExpression{
		Left:     ast.NewFractionExpr(ast.NewIntLiteral(1), ast.NewIntLiteral(2)),
		Operator: "*",
		Right:    ast.NewFractionExpr(ast.NewIntLiteral(2), ast.NewIntLiteral(3)),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("1/3", num.Inspect())
}

// === Division ===

func (s *ArithmeticSuite) TestDivision_Natural() {
	// 10 ÷ 2 = 5
	node := &ast.RealInfixExpression{
		Left:     ast.NewIntLiteral(10),
		Operator: "÷",
		Right:    ast.NewIntLiteral(2),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("5", num.Inspect())
}

func (s *ArithmeticSuite) TestDivision_WithSlash() {
	// 10 / 2 = 5
	node := &ast.RealInfixExpression{
		Left:     ast.NewIntLiteral(10),
		Operator: "/",
		Right:    ast.NewIntLiteral(2),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("5", num.Inspect())
}

func (s *ArithmeticSuite) TestDivision_FractionalResult() {
	// 1 ÷ 3 = 1/3
	node := &ast.RealInfixExpression{
		Left:     ast.NewIntLiteral(1),
		Operator: "÷",
		Right:    ast.NewIntLiteral(3),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("1/3", num.Inspect())
}

func (s *ArithmeticSuite) TestDivision_ByZero() {
	// 5 ÷ 0 = error
	node := &ast.RealInfixExpression{
		Left:     ast.NewIntLiteral(5),
		Operator: "÷",
		Right:    ast.NewIntLiteral(0),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.True(result.IsError())
	s.Contains(result.Error.Message, "division by zero")
}

// === Real Contamination in Division ===

func (s *ArithmeticSuite) TestDivision_RealLeftOperand() {
	// sqrt(2) ÷ 2 = Real (contaminated by Real operand)
	// First create sqrt(2) via power
	sqrt2 := &ast.RealInfixExpression{
		Left:     ast.NewIntLiteral(2),
		Operator: "^",
		Right:    ast.NewFractionExpr(ast.NewIntLiteral(1), ast.NewIntLiteral(2)),
	}
	node := &ast.RealInfixExpression{
		Left:     sqrt2,
		Operator: "÷",
		Right:    ast.NewIntLiteral(2),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal(object.KindReal, num.Kind())
	// sqrt(2)/2 ≈ 0.7071
	s.InDelta(0.7071067811865476, num.Float64(), 0.0000001)
}

func (s *ArithmeticSuite) TestDivision_RealRightOperand() {
	// 2 ÷ sqrt(2) = Real (contaminated by Real operand)
	sqrt2 := &ast.RealInfixExpression{
		Left:     ast.NewIntLiteral(2),
		Operator: "^",
		Right:    ast.NewFractionExpr(ast.NewIntLiteral(1), ast.NewIntLiteral(2)),
	}
	node := &ast.RealInfixExpression{
		Left:     ast.NewIntLiteral(2),
		Operator: "÷",
		Right:    sqrt2,
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal(object.KindReal, num.Kind())
	// 2/sqrt(2) = sqrt(2) ≈ 1.4142
	s.InDelta(1.4142135623730951, num.Float64(), 0.0000001)
}

func (s *ArithmeticSuite) TestDivision_SlashWithRealLeftOperand() {
	// sqrt(2) / 2 = Real (using / operator)
	sqrt2 := &ast.RealInfixExpression{
		Left:     ast.NewIntLiteral(2),
		Operator: "^",
		Right:    ast.NewFractionExpr(ast.NewIntLiteral(1), ast.NewIntLiteral(2)),
	}
	node := &ast.RealInfixExpression{
		Left:     sqrt2,
		Operator: "/",
		Right:    ast.NewIntLiteral(2),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal(object.KindReal, num.Kind())
}

func (s *ArithmeticSuite) TestDivision_SlashWithRealRightOperand() {
	// 2 / sqrt(2) = Real (using / operator)
	sqrt2 := &ast.RealInfixExpression{
		Left:     ast.NewIntLiteral(2),
		Operator: "^",
		Right:    ast.NewFractionExpr(ast.NewIntLiteral(1), ast.NewIntLiteral(2)),
	}
	node := &ast.RealInfixExpression{
		Left:     ast.NewIntLiteral(2),
		Operator: "/",
		Right:    sqrt2,
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal(object.KindReal, num.Kind())
	// 2/sqrt(2) = sqrt(2) ≈ 1.4142
	s.InDelta(1.4142135623730951, num.Float64(), 0.0000001)
}

func (s *ArithmeticSuite) TestDivision_SameRealDividedBySelf() {
	// sqrt(2) / sqrt(2) = Real (still contaminated, even though mathematically = 1)
	sqrt2a := &ast.RealInfixExpression{
		Left:     ast.NewIntLiteral(2),
		Operator: "^",
		Right:    ast.NewFractionExpr(ast.NewIntLiteral(1), ast.NewIntLiteral(2)),
	}
	sqrt2b := &ast.RealInfixExpression{
		Left:     ast.NewIntLiteral(2),
		Operator: "^",
		Right:    ast.NewFractionExpr(ast.NewIntLiteral(1), ast.NewIntLiteral(2)),
	}
	node := &ast.RealInfixExpression{
		Left:     sqrt2a,
		Operator: "/",
		Right:    sqrt2b,
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal(object.KindReal, num.Kind())
	// sqrt(2)/sqrt(2) = 1
	s.InDelta(1.0, num.Float64(), 0.0000001)
}

func (s *ArithmeticSuite) TestDivision_BothRealOperands() {
	// sqrt(2) ÷ sqrt(3) = Real
	sqrt2 := &ast.RealInfixExpression{
		Left:     ast.NewIntLiteral(2),
		Operator: "^",
		Right:    ast.NewFractionExpr(ast.NewIntLiteral(1), ast.NewIntLiteral(2)),
	}
	sqrt3 := &ast.RealInfixExpression{
		Left:     ast.NewIntLiteral(3),
		Operator: "^",
		Right:    ast.NewFractionExpr(ast.NewIntLiteral(1), ast.NewIntLiteral(2)),
	}
	node := &ast.RealInfixExpression{
		Left:     sqrt2,
		Operator: "÷",
		Right:    sqrt3,
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal(object.KindReal, num.Kind())
	// sqrt(2)/sqrt(3) ≈ 0.8165
	s.InDelta(0.816496580927726, num.Float64(), 0.0000001)
}

// === Modulo ===

func (s *ArithmeticSuite) TestModulo_Simple() {
	// 10 % 3 = 1
	node := &ast.RealInfixExpression{
		Left:     ast.NewIntLiteral(10),
		Operator: "%",
		Right:    ast.NewIntLiteral(3),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("1", num.Inspect())
}

func (s *ArithmeticSuite) TestModulo_EvenDivision() {
	// 9 % 3 = 0
	node := &ast.RealInfixExpression{
		Left:     ast.NewIntLiteral(9),
		Operator: "%",
		Right:    ast.NewIntLiteral(3),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("0", num.Inspect())
}

// === Nested Expressions ===

func (s *ArithmeticSuite) TestNested_Expression() {
	// (2 + 3) * 4 = 20
	inner := &ast.RealInfixExpression{
		Left:     ast.NewIntLiteral(2),
		Operator: "+",
		Right:    ast.NewIntLiteral(3),
	}
	outer := &ast.RealInfixExpression{
		Left:     inner,
		Operator: "*",
		Right:    ast.NewIntLiteral(4),
	}
	bindings := NewBindings()

	result := Eval(outer, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("20", num.Inspect())
}

func (s *ArithmeticSuite) TestNested_DeepExpression() {
	// ((1 + 2) * 3) - 4 = 5
	innermost := &ast.RealInfixExpression{
		Left:     ast.NewIntLiteral(1),
		Operator: "+",
		Right:    ast.NewIntLiteral(2),
	}
	middle := &ast.RealInfixExpression{
		Left:     innermost,
		Operator: "*",
		Right:    ast.NewIntLiteral(3),
	}
	outer := &ast.RealInfixExpression{
		Left:     middle,
		Operator: "-",
		Right:    ast.NewIntLiteral(4),
	}
	bindings := NewBindings()

	result := Eval(outer, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("5", num.Inspect())
}

// Note: RealInfixExpression.Left/Right must be ast.Real types,
// so Identifiers (which implement ast.Expression, not ast.Real) cannot
// be used directly. This is a type-safe design choice in the AST.

// === Power ===

func (s *ArithmeticSuite) TestPower_Simple() {
	// 2 ^ 3 = 8
	node := &ast.RealInfixExpression{
		Left:     ast.NewIntLiteral(2),
		Operator: "^",
		Right:    ast.NewIntLiteral(3),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("8", num.Inspect())
}

func (s *ArithmeticSuite) TestPower_Zero() {
	// 5 ^ 0 = 1
	node := &ast.RealInfixExpression{
		Left:     ast.NewIntLiteral(5),
		Operator: "^",
		Right:    ast.NewIntLiteral(0),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("1", num.Inspect())
}

func (s *ArithmeticSuite) TestPower_One() {
	// 7 ^ 1 = 7
	node := &ast.RealInfixExpression{
		Left:     ast.NewIntLiteral(7),
		Operator: "^",
		Right:    ast.NewIntLiteral(1),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("7", num.Inspect())
}

func (s *ArithmeticSuite) TestPower_FractionalBase() {
	// (1/2) ^ 3 = 1/8
	node := &ast.RealInfixExpression{
		Left:     ast.NewFractionExpr(ast.NewIntLiteral(1), ast.NewIntLiteral(2)),
		Operator: "^",
		Right:    ast.NewIntLiteral(3),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("1/8", num.Inspect())
}

func (s *ArithmeticSuite) TestPower_LargeExponent() {
	// 2 ^ 10 = 1024
	node := &ast.RealInfixExpression{
		Left:     ast.NewIntLiteral(2),
		Operator: "^",
		Right:    ast.NewIntLiteral(10),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("1024", num.Inspect())
}

func (s *ArithmeticSuite) TestPower_ZeroBase() {
	// 0 ^ 5 = 0
	node := &ast.RealInfixExpression{
		Left:     ast.NewIntLiteral(0),
		Operator: "^",
		Right:    ast.NewIntLiteral(5),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("0", num.Inspect())
}

func (s *ArithmeticSuite) TestPower_ZeroToZero_Error() {
	// 0 ^ 0 = error (undefined)
	node := &ast.RealInfixExpression{
		Left:     ast.NewIntLiteral(0),
		Operator: "^",
		Right:    ast.NewIntLiteral(0),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.True(result.IsError())
	s.Contains(result.Error.Message, "0^0 is undefined")
}

func (s *ArithmeticSuite) TestPower_NegativeExponent_Error() {
	// 2 ^ -1 = error (negative exponent not supported)
	node := &ast.RealInfixExpression{
		Left:     ast.NewIntLiteral(2),
		Operator: "^",
		Right:    ast.NewIntLiteral(-1),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.True(result.IsError())
	s.Contains(result.Error.Message, "non-negative")
}

func (s *ArithmeticSuite) TestPower_SquareRoot() {
	// 4 ^ (1/2) = 2 (square root)
	node := &ast.RealInfixExpression{
		Left:     ast.NewIntLiteral(4),
		Operator: "^",
		Right:    ast.NewFractionExpr(ast.NewIntLiteral(1), ast.NewIntLiteral(2)),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("2", num.Inspect())
}

func (s *ArithmeticSuite) TestPower_CubeRoot() {
	// 27 ^ (1/3) = 3 (cube root)
	node := &ast.RealInfixExpression{
		Left:     ast.NewIntLiteral(27),
		Operator: "^",
		Right:    ast.NewFractionExpr(ast.NewIntLiteral(1), ast.NewIntLiteral(3)),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("3", num.Inspect())
}

func (s *ArithmeticSuite) TestPower_FractionalExponentWithNumerator() {
	// 8 ^ (2/3) = (8^2)^(1/3) = 64^(1/3) = 4
	node := &ast.RealInfixExpression{
		Left:     ast.NewIntLiteral(8),
		Operator: "^",
		Right:    ast.NewFractionExpr(ast.NewIntLiteral(2), ast.NewIntLiteral(3)),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("4", num.Inspect())
}

func (s *ArithmeticSuite) TestPower_FractionalBaseWithRoot() {
	// (1/4) ^ (1/2) = 1/2 (square root of 1/4)
	node := &ast.RealInfixExpression{
		Left:     ast.NewFractionExpr(ast.NewIntLiteral(1), ast.NewIntLiteral(4)),
		Operator: "^",
		Right:    ast.NewFractionExpr(ast.NewIntLiteral(1), ast.NewIntLiteral(2)),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("1/2", num.Inspect())
}

func (s *ArithmeticSuite) TestPower_IrrationalResult_Real() {
	// 2 ^ (1/2) = sqrt(2) ≈ 1.414... (Real number)
	node := &ast.RealInfixExpression{
		Left:     ast.NewIntLiteral(2),
		Operator: "^",
		Right:    ast.NewFractionExpr(ast.NewIntLiteral(1), ast.NewIntLiteral(2)),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal(object.KindReal, num.Kind())
	// sqrt(2) ≈ 1.4142135623730951
	s.InDelta(1.4142135623730951, num.Float64(), 0.0000001)
}

// === Real Contamination in Power ===

func (s *ArithmeticSuite) TestPower_RealBase() {
	// sqrt(2) ^ 2 = Real (contaminated by Real base, even though result is 2)
	sqrt2 := &ast.RealInfixExpression{
		Left:     ast.NewIntLiteral(2),
		Operator: "^",
		Right:    ast.NewFractionExpr(ast.NewIntLiteral(1), ast.NewIntLiteral(2)),
	}
	node := &ast.RealInfixExpression{
		Left:     sqrt2,
		Operator: "^",
		Right:    ast.NewIntLiteral(2),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal(object.KindReal, num.Kind())
	// sqrt(2)^2 = 2
	s.InDelta(2.0, num.Float64(), 0.0000001)
}

func (s *ArithmeticSuite) TestPower_RealExponent() {
	// 2 ^ sqrt(2) = Real (contaminated by Real exponent)
	sqrt2 := &ast.RealInfixExpression{
		Left:     ast.NewIntLiteral(2),
		Operator: "^",
		Right:    ast.NewFractionExpr(ast.NewIntLiteral(1), ast.NewIntLiteral(2)),
	}
	node := &ast.RealInfixExpression{
		Left:     ast.NewIntLiteral(2),
		Operator: "^",
		Right:    sqrt2,
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal(object.KindReal, num.Kind())
	// 2^sqrt(2) ≈ 2.6651
	s.InDelta(2.6651441426902252, num.Float64(), 0.0000001)
}

func (s *ArithmeticSuite) TestPower_BothReal() {
	// sqrt(2) ^ sqrt(3) = Real
	sqrt2 := &ast.RealInfixExpression{
		Left:     ast.NewIntLiteral(2),
		Operator: "^",
		Right:    ast.NewFractionExpr(ast.NewIntLiteral(1), ast.NewIntLiteral(2)),
	}
	sqrt3 := &ast.RealInfixExpression{
		Left:     ast.NewIntLiteral(3),
		Operator: "^",
		Right:    ast.NewFractionExpr(ast.NewIntLiteral(1), ast.NewIntLiteral(2)),
	}
	node := &ast.RealInfixExpression{
		Left:     sqrt2,
		Operator: "^",
		Right:    sqrt3,
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal(object.KindReal, num.Kind())
	// sqrt(2)^sqrt(3) ≈ 1.8226
	s.InDelta(1.8226346549662427, num.Float64(), 0.0000001)
}

// === Modulo Error Cases ===

func (s *ArithmeticSuite) TestModulo_FractionalLeft_Error() {
	// (1/2) % 3 = error (modulo requires integer operands)
	node := &ast.RealInfixExpression{
		Left:     ast.NewFractionExpr(ast.NewIntLiteral(1), ast.NewIntLiteral(2)),
		Operator: "%",
		Right:    ast.NewIntLiteral(3),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.True(result.IsError())
	s.Contains(result.Error.Message, "integer operands")
}

func (s *ArithmeticSuite) TestModulo_FractionalRight_Error() {
	// 10 % (1/2) = error (modulo requires integer operands)
	node := &ast.RealInfixExpression{
		Left:     ast.NewIntLiteral(10),
		Operator: "%",
		Right:    ast.NewFractionExpr(ast.NewIntLiteral(1), ast.NewIntLiteral(2)),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.True(result.IsError())
	s.Contains(result.Error.Message, "integer operands")
}

func (s *ArithmeticSuite) TestModulo_BothFractional_Error() {
	// (3/4) % (1/2) = error (modulo requires integer operands)
	node := &ast.RealInfixExpression{
		Left:     ast.NewFractionExpr(ast.NewIntLiteral(3), ast.NewIntLiteral(4)),
		Operator: "%",
		Right:    ast.NewFractionExpr(ast.NewIntLiteral(1), ast.NewIntLiteral(2)),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.True(result.IsError())
	s.Contains(result.Error.Message, "integer operands")
}

func (s *ArithmeticSuite) TestModulo_ByZero_Error() {
	// 10 % 0 = error (division by zero)
	node := &ast.RealInfixExpression{
		Left:     ast.NewIntLiteral(10),
		Operator: "%",
		Right:    ast.NewIntLiteral(0),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.True(result.IsError())
	s.Contains(result.Error.Message, "zero")
}

// === Error Cases ===

func (s *ArithmeticSuite) TestUnknownOperator() {
	// 5 & 3 (invalid operator)
	node := &ast.RealInfixExpression{
		Left:     ast.NewIntLiteral(5),
		Operator: "&", // not a valid arithmetic operator
		Right:    ast.NewIntLiteral(3),
	}
	bindings := NewBindings()

	result := Eval(node, bindings)

	s.True(result.IsError())
	s.Contains(result.Error.Message, "unknown")
}

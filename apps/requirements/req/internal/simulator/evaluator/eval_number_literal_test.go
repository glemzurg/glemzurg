package evaluator

import (
	"testing"

	"github.com/glemzurg/go-tlaplus/internal/simulator/ast"
	"github.com/glemzurg/go-tlaplus/internal/simulator/object"
	"github.com/stretchr/testify/suite"
)

func TestNumberLiteralSuite(t *testing.T) {
	suite.Run(t, new(NumberLiteralSuite))
}

type NumberLiteralSuite struct {
	suite.Suite
}

// === NumberLiteral Evaluation ===

func (s *NumberLiteralSuite) TestNumberLiteral_DecimalInteger() {
	tests := []struct {
		testName    string
		integerPart string
		expected    string
		kind        object.NumberKind
	}{
		{"zero", "0", "0", object.KindNatural},
		{"single digit", "7", "7", object.KindNatural},
		{"multiple digits", "42", "42", object.KindNatural},
		{"large number", "1000000", "1000000", object.KindNatural},
		{"leading zeros", "007", "7", object.KindNatural},
	}
	for _, tt := range tests {
		s.Run(tt.testName, func() {
			node := &ast.NumberLiteral{
				Base:        ast.BaseDecimal,
				IntegerPart: tt.integerPart,
			}
			result := Eval(node, NewBindings())

			s.False(result.IsError())
			num := result.Value.(*object.Number)
			s.Equal(tt.kind, num.Kind())
			s.Equal(tt.expected, num.Inspect())
		})
	}
}

func (s *NumberLiteralSuite) TestNumberLiteral_DecimalWithFraction() {
	tests := []struct {
		testName       string
		integerPart    string
		fractionalPart string
		expected       string
		kind           object.NumberKind
	}{
		{"simple decimal", "3", "14", "157/50", object.KindRational},
		{"half", "0", "5", "1/2", object.KindRational},
		{"quarter", "0", "25", "1/4", object.KindRational},
		{"whole after reduction", "1", "0", "1", object.KindNatural},
		{"two point five", "2", "5", "5/2", object.KindRational},
		{"trailing zeros preserved", "1", "500", "3/2", object.KindRational},
	}
	for _, tt := range tests {
		s.Run(tt.testName, func() {
			node := &ast.NumberLiteral{
				Base:            ast.BaseDecimal,
				IntegerPart:     tt.integerPart,
				HasDecimalPoint: true,
				FractionalPart:  tt.fractionalPart,
			}
			result := Eval(node, NewBindings())

			s.False(result.IsError())
			num := result.Value.(*object.Number)
			s.Equal(tt.kind, num.Kind())
			s.Equal(tt.expected, num.Inspect())
		})
	}
}

func (s *NumberLiteralSuite) TestNumberLiteral_BinaryBase() {
	tests := []struct {
		testName    string
		prefix      string
		integerPart string
		expected    string
	}{
		{"binary zero", "\\b", "0", "0"},
		{"binary one", "\\b", "1", "1"},
		{"binary 1010", "\\b", "1010", "10"},
		{"binary 11111111", "\\B", "11111111", "255"},
		{"binary with leading zeros", "\\b", "00001010", "10"},
	}
	for _, tt := range tests {
		s.Run(tt.testName, func() {
			node := &ast.NumberLiteral{
				Base:        ast.BaseBinary,
				BasePrefix:  tt.prefix,
				IntegerPart: tt.integerPart,
			}
			result := Eval(node, NewBindings())

			s.False(result.IsError())
			num := result.Value.(*object.Number)
			s.Equal(tt.expected, num.Inspect())
		})
	}
}

func (s *NumberLiteralSuite) TestNumberLiteral_OctalBase() {
	tests := []struct {
		testName    string
		prefix      string
		integerPart string
		expected    string
	}{
		{"octal zero", "\\o", "0", "0"},
		{"octal 7", "\\o", "7", "7"},
		{"octal 10", "\\o", "10", "8"},
		{"octal 17", "\\O", "17", "15"},
		{"octal 777", "\\o", "777", "511"},
	}
	for _, tt := range tests {
		s.Run(tt.testName, func() {
			node := &ast.NumberLiteral{
				Base:        ast.BaseOctal,
				BasePrefix:  tt.prefix,
				IntegerPart: tt.integerPart,
			}
			result := Eval(node, NewBindings())

			s.False(result.IsError())
			num := result.Value.(*object.Number)
			s.Equal(tt.expected, num.Inspect())
		})
	}
}

func (s *NumberLiteralSuite) TestNumberLiteral_HexBase() {
	tests := []struct {
		testName    string
		prefix      string
		integerPart string
		expected    string
	}{
		{"hex zero", "\\h", "0", "0"},
		{"hex F", "\\h", "F", "15"},
		{"hex f lowercase", "\\h", "f", "15"},
		{"hex FF", "\\H", "FF", "255"},
		{"hex mixed case", "\\h", "aB", "171"},
		{"hex DEADBEEF", "\\h", "DEADBEEF", "3735928559"},
	}
	for _, tt := range tests {
		s.Run(tt.testName, func() {
			node := &ast.NumberLiteral{
				Base:        ast.BaseHex,
				BasePrefix:  tt.prefix,
				IntegerPart: tt.integerPart,
			}
			result := Eval(node, NewBindings())

			s.False(result.IsError())
			num := result.Value.(*object.Number)
			s.Equal(tt.expected, num.Inspect())
		})
	}
}

func (s *NumberLiteralSuite) TestNumberLiteral_InvalidDigits() {
	tests := []struct {
		testName    string
		base        ast.NumberBase
		integerPart string
	}{
		{"invalid decimal", ast.BaseDecimal, "12a3"},
		{"invalid binary", ast.BaseBinary, "1012"},
		{"invalid octal", ast.BaseOctal, "189"},
		{"invalid hex", ast.BaseHex, "GHIJ"},
	}
	for _, tt := range tests {
		s.Run(tt.testName, func() {
			node := &ast.NumberLiteral{
				Base:        tt.base,
				IntegerPart: tt.integerPart,
			}
			result := Eval(node, NewBindings())

			s.True(result.IsError())
			s.Contains(result.Error.Message, "invalid")
		})
	}
}

// === NumericPrefixExpression (Negation) ===

func (s *NumberLiteralSuite) TestNumericPrefix_Negation() {
	tests := []struct {
		testName string
		value    int
		expected string
		kind     object.NumberKind
	}{
		{"negate positive", 42, "-42", object.KindInteger},
		{"negate zero", 0, "0", object.KindNatural},
		{"negate large", 1000000, "-1000000", object.KindInteger},
	}
	for _, tt := range tests {
		s.Run(tt.testName, func() {
			node := &ast.NumericPrefixExpression{
				Operator: "-",
				Right:    &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: string(rune('0' + tt.value%10))},
			}
			// Use proper number construction
			node.Right = &ast.NumberLiteral{
				Base:        ast.BaseDecimal,
				IntegerPart: intToString(tt.value),
			}
			result := Eval(node, NewBindings())

			s.False(result.IsError())
			num := result.Value.(*object.Number)
			s.Equal(tt.kind, num.Kind())
			s.Equal(tt.expected, num.Inspect())
		})
	}
}

func (s *NumberLiteralSuite) TestNumericPrefix_DoubleNegation() {
	// --5 should equal 5
	node := &ast.NumericPrefixExpression{
		Operator: "-",
		Right: &ast.NumericPrefixExpression{
			Operator: "-",
			Right:    &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "5"},
		},
	}
	result := Eval(node, NewBindings())

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal(object.KindNatural, num.Kind())
	s.Equal("5", num.Inspect())
}

func (s *NumberLiteralSuite) TestNumericPrefix_NegateDecimal() {
	// -(3.14)
	node := &ast.NumericPrefixExpression{
		Operator: "-",
		Right: &ast.NumberLiteral{
			Base:            ast.BaseDecimal,
			IntegerPart:     "3",
			HasDecimalPoint: true,
			FractionalPart:  "14",
		},
	}
	result := Eval(node, NewBindings())

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal(object.KindRational, num.Kind())
	s.Equal("-157/50", num.Inspect())
}

func (s *NumberLiteralSuite) TestNumericPrefix_NegateNonNumeric() {
	node := &ast.NumericPrefixExpression{
		Operator: "-",
		Right:    &ast.StringLiteral{Value: "hello"},
	}
	result := Eval(node, NewBindings())

	s.True(result.IsError())
	s.Contains(result.Error.Message, "cannot negate non-numeric")
}

func (s *NumberLiteralSuite) TestNumericPrefix_UnknownOperator() {
	node := &ast.NumericPrefixExpression{
		Operator: "+",
		Right:    &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "5"},
	}
	result := Eval(node, NewBindings())

	s.True(result.IsError())
	s.Contains(result.Error.Message, "unknown numeric prefix operator")
}

// === FractionExpr ===

func (s *NumberLiteralSuite) TestFractionExpr_Simple() {
	tests := []struct {
		testName    string
		numerator   int
		denominator int
		expected    string
		kind        object.NumberKind
	}{
		{"one half", 1, 2, "1/2", object.KindRational},
		{"three quarters", 3, 4, "3/4", object.KindRational},
		{"reduces to whole", 4, 2, "2", object.KindNatural},
		{"reduces to fraction", 6, 4, "3/2", object.KindRational},
		{"large fraction", 22, 7, "22/7", object.KindRational},
	}
	for _, tt := range tests {
		s.Run(tt.testName, func() {
			node := ast.NewFractionExpr(
				&ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: intToString(tt.numerator)},
				&ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: intToString(tt.denominator)},
			)
			result := Eval(node, NewBindings())

			s.False(result.IsError())
			num := result.Value.(*object.Number)
			s.Equal(tt.kind, num.Kind())
			s.Equal(tt.expected, num.Inspect())
		})
	}
}

func (s *NumberLiteralSuite) TestFractionExpr_NegativeNumerator() {
	// -3/4
	node := ast.NewFractionExpr(
		ast.NewNegation(&ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "3"}),
		&ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "4"},
	)
	result := Eval(node, NewBindings())

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal(object.KindRational, num.Kind())
	s.Equal("-3/4", num.Inspect())
}

func (s *NumberLiteralSuite) TestFractionExpr_NegativeDenominator() {
	// 3/-4
	node := ast.NewFractionExpr(
		&ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "3"},
		ast.NewNegation(&ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "4"}),
	)
	result := Eval(node, NewBindings())

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal(object.KindRational, num.Kind())
	s.Equal("-3/4", num.Inspect())
}

func (s *NumberLiteralSuite) TestFractionExpr_BothNegative() {
	// -3/-4 = 3/4
	node := ast.NewFractionExpr(
		ast.NewNegation(&ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "3"}),
		ast.NewNegation(&ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "4"}),
	)
	result := Eval(node, NewBindings())

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal(object.KindRational, num.Kind())
	s.Equal("3/4", num.Inspect())
}

func (s *NumberLiteralSuite) TestFractionExpr_DivisionByZero() {
	node := ast.NewFractionExpr(
		&ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "1"},
		&ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "0"},
	)
	result := Eval(node, NewBindings())

	s.True(result.IsError())
	s.Contains(result.Error.Message, "division by zero")
}

func (s *NumberLiteralSuite) TestFractionExpr_NonNumericNumerator() {
	node := ast.NewFractionExpr(
		&ast.StringLiteral{Value: "hello"},
		&ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "2"},
	)
	result := Eval(node, NewBindings())

	s.True(result.IsError())
	s.Contains(result.Error.Message, "numerator must be numeric")
}

func (s *NumberLiteralSuite) TestFractionExpr_NonNumericDenominator() {
	node := ast.NewFractionExpr(
		&ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "2"},
		&ast.StringLiteral{Value: "hello"},
	)
	result := Eval(node, NewBindings())

	s.True(result.IsError())
	s.Contains(result.Error.Message, "denominator must be numeric")
}

func (s *NumberLiteralSuite) TestFractionExpr_NestedFractions() {
	// (1/2)/3 = 1/6
	node := ast.NewFractionExpr(
		ast.NewFractionExpr(
			&ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "1"},
			&ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "2"},
		),
		&ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "3"},
	)
	result := Eval(node, NewBindings())

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("1/6", num.Inspect())
}

func (s *NumberLiteralSuite) TestFractionExpr_DecimalOperands() {
	// 1.5/0.5 = 3
	node := ast.NewFractionExpr(
		&ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "1", HasDecimalPoint: true, FractionalPart: "5"},
		&ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "0", HasDecimalPoint: true, FractionalPart: "5"},
	)
	result := Eval(node, NewBindings())

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal(object.KindNatural, num.Kind())
	s.Equal("3", num.Inspect())
}

// === ParenExpr ===

func (s *NumberLiteralSuite) TestParenExpr_Number() {
	node := &ast.ParenExpr{
		Inner: &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "42"},
	}
	result := Eval(node, NewBindings())

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("42", num.Inspect())
}

func (s *NumberLiteralSuite) TestParenExpr_String() {
	node := &ast.ParenExpr{
		Inner: &ast.StringLiteral{Value: "hello"},
	}
	result := Eval(node, NewBindings())

	s.False(result.IsError())
	str := result.Value.(*object.String)
	s.Equal("hello", str.Value())
}

func (s *NumberLiteralSuite) TestParenExpr_Boolean() {
	node := &ast.ParenExpr{
		Inner: &ast.BooleanLiteral{Value: true},
	}
	result := Eval(node, NewBindings())

	s.False(result.IsError())
	b := result.Value.(*object.Boolean)
	s.True(b.Value())
}

func (s *NumberLiteralSuite) TestParenExpr_NestedParens() {
	// ((42))
	node := &ast.ParenExpr{
		Inner: &ast.ParenExpr{
			Inner: &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "42"},
		},
	}
	result := Eval(node, NewBindings())

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("42", num.Inspect())
}

func (s *NumberLiteralSuite) TestParenExpr_WithNegation() {
	// -(42)
	node := &ast.NumericPrefixExpression{
		Operator: "-",
		Right: &ast.ParenExpr{
			Inner: &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "42"},
		},
	}
	result := Eval(node, NewBindings())

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("-42", num.Inspect())
}

func (s *NumberLiteralSuite) TestParenExpr_WithFraction() {
	// (3)/(4)
	node := ast.NewFractionExpr(
		&ast.ParenExpr{Inner: &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "3"}},
		&ast.ParenExpr{Inner: &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "4"}},
	)
	result := Eval(node, NewBindings())

	s.False(result.IsError())
	num := result.Value.(*object.Number)
	s.Equal("3/4", num.Inspect())
}

// Helper function to convert int to string
func intToString(n int) string {
	if n < 0 {
		return "-" + intToString(-n)
	}
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}

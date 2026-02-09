package parser

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/ast"
	"github.com/stretchr/testify/suite"
)

type NumberTestSuite struct {
	suite.Suite
}

func TestNumberSuite(t *testing.T) {
	suite.Run(t, new(NumberTestSuite))
}

// =============================================================================
// Decimal Integers
// =============================================================================

func (s *NumberTestSuite) TestParseZero() {
	expr, err := ParseExpression("0")
	s.NoError(err)
	s.Equal(&ast.NumberLiteral{
		Base:            ast.BaseDecimal,
		BasePrefix:      "",
		IntegerPart:     "0",
		HasDecimalPoint: false,
		FractionalPart:  "",
	}, expr)
}

func (s *NumberTestSuite) TestParsePositive() {
	expr, err := ParseExpression("42")
	s.NoError(err)
	s.Equal(&ast.NumberLiteral{
		Base:            ast.BaseDecimal,
		BasePrefix:      "",
		IntegerPart:     "42",
		HasDecimalPoint: false,
		FractionalPart:  "",
	}, expr)
}

func (s *NumberTestSuite) TestParseWithLeadingZeros() {
	expr, err := ParseExpression("007")
	s.NoError(err)
	s.Equal(&ast.NumberLiteral{
		Base:            ast.BaseDecimal,
		BasePrefix:      "",
		IntegerPart:     "007",
		HasDecimalPoint: false,
		FractionalPart:  "",
	}, expr)
}

// =============================================================================
// Decimals with Fractional Part
// =============================================================================

func (s *NumberTestSuite) TestParseDecimal() {
	expr, err := ParseExpression("3.14")
	s.NoError(err)
	s.Equal(&ast.NumberLiteral{
		Base:            ast.BaseDecimal,
		BasePrefix:      "",
		IntegerPart:     "3",
		HasDecimalPoint: true,
		FractionalPart:  "14",
	}, expr)
}

func (s *NumberTestSuite) TestParseDecimalNoWholePart() {
	expr, err := ParseExpression(".5")
	s.NoError(err)
	s.Equal(&ast.NumberLiteral{
		Base:            ast.BaseDecimal,
		BasePrefix:      "",
		IntegerPart:     "",
		HasDecimalPoint: true,
		FractionalPart:  "5",
	}, expr)
}

func (s *NumberTestSuite) TestParseDecimalZeroWholePart() {
	expr, err := ParseExpression("0.123")
	s.NoError(err)
	s.Equal(&ast.NumberLiteral{
		Base:            ast.BaseDecimal,
		BasePrefix:      "",
		IntegerPart:     "0",
		HasDecimalPoint: true,
		FractionalPart:  "123",
	}, expr)
}

func (s *NumberTestSuite) TestParseDecimalWithTrailingZeros() {
	expr, err := ParseExpression("3.140")
	s.NoError(err)
	s.Equal(&ast.NumberLiteral{
		Base:            ast.BaseDecimal,
		BasePrefix:      "",
		IntegerPart:     "3",
		HasDecimalPoint: true,
		FractionalPart:  "140",
	}, expr)
}

// =============================================================================
// Binary
// =============================================================================

func (s *NumberTestSuite) TestParseBinaryLowerCase() {
	expr, err := ParseExpression("\\b1010")
	s.NoError(err)
	s.Equal(&ast.NumberLiteral{
		Base:            ast.BaseBinary,
		BasePrefix:      "\\b",
		IntegerPart:     "1010",
		HasDecimalPoint: false,
		FractionalPart:  "",
	}, expr)
}

func (s *NumberTestSuite) TestParseBinaryUpperCase() {
	expr, err := ParseExpression("\\B0011")
	s.NoError(err)
	s.Equal(&ast.NumberLiteral{
		Base:            ast.BaseBinary,
		BasePrefix:      "\\B",
		IntegerPart:     "0011",
		HasDecimalPoint: false,
		FractionalPart:  "",
	}, expr)
}

// =============================================================================
// Octal
// =============================================================================

func (s *NumberTestSuite) TestParseOctalLowerCase() {
	expr, err := ParseExpression("\\o17")
	s.NoError(err)
	s.Equal(&ast.NumberLiteral{
		Base:            ast.BaseOctal,
		BasePrefix:      "\\o",
		IntegerPart:     "17",
		HasDecimalPoint: false,
		FractionalPart:  "",
	}, expr)
}

func (s *NumberTestSuite) TestParseOctalUpperCase() {
	expr, err := ParseExpression("\\O777")
	s.NoError(err)
	s.Equal(&ast.NumberLiteral{
		Base:            ast.BaseOctal,
		BasePrefix:      "\\O",
		IntegerPart:     "777",
		HasDecimalPoint: false,
		FractionalPart:  "",
	}, expr)
}

// =============================================================================
// Hexadecimal
// =============================================================================

func (s *NumberTestSuite) TestParseHexLowerCase() {
	expr, err := ParseExpression("\\hff")
	s.NoError(err)
	s.Equal(&ast.NumberLiteral{
		Base:            ast.BaseHex,
		BasePrefix:      "\\h",
		IntegerPart:     "ff",
		HasDecimalPoint: false,
		FractionalPart:  "",
	}, expr)
}

func (s *NumberTestSuite) TestParseHexUpperCase() {
	expr, err := ParseExpression("\\H0ABC")
	s.NoError(err)
	s.Equal(&ast.NumberLiteral{
		Base:            ast.BaseHex,
		BasePrefix:      "\\H",
		IntegerPart:     "0ABC",
		HasDecimalPoint: false,
		FractionalPart:  "",
	}, expr)
}

func (s *NumberTestSuite) TestParseHexMixedCase() {
	expr, err := ParseExpression("\\hDeAdBeEf")
	s.NoError(err)
	s.Equal(&ast.NumberLiteral{
		Base:            ast.BaseHex,
		BasePrefix:      "\\h",
		IntegerPart:     "DeAdBeEf",
		HasDecimalPoint: false,
		FractionalPart:  "",
	}, expr)
}

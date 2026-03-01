package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestNumberLiteralSuite(t *testing.T) {
	suite.Run(t, new(NumberLiteralSuite))
}

type NumberLiteralSuite struct {
	suite.Suite
}

func (suite *NumberLiteralSuite) TestString() {
	tests := []struct {
		testName string
		n        *NumberLiteral
		expected string
	}{
		{
			testName: "decimal integer",
			n:        &NumberLiteral{Base: BaseDecimal, IntegerPart: "42"},
			expected: "42",
		},
		{
			testName: "decimal with leading zeros",
			n:        &NumberLiteral{Base: BaseDecimal, IntegerPart: "007"},
			expected: "007",
		},
		{
			testName: "decimal with fractional part",
			n:        &NumberLiteral{Base: BaseDecimal, IntegerPart: "3", HasDecimalPoint: true, FractionalPart: "14"},
			expected: "3.14",
		},
		{
			testName: "decimal no whole part",
			n:        &NumberLiteral{Base: BaseDecimal, IntegerPart: "", HasDecimalPoint: true, FractionalPart: "5"},
			expected: ".5",
		},
		{
			testName: "binary lowercase",
			n:        &NumberLiteral{Base: BaseBinary, BasePrefix: "\\b", IntegerPart: "1010"},
			expected: "\\b1010",
		},
		{
			testName: "binary uppercase",
			n:        &NumberLiteral{Base: BaseBinary, BasePrefix: "\\B", IntegerPart: "0011"},
			expected: "\\B0011",
		},
		{
			testName: "octal lowercase",
			n:        &NumberLiteral{Base: BaseOctal, BasePrefix: "\\o", IntegerPart: "17"},
			expected: "\\o17",
		},
		{
			testName: "octal uppercase",
			n:        &NumberLiteral{Base: BaseOctal, BasePrefix: "\\O", IntegerPart: "777"},
			expected: "\\O777",
		},
		{
			testName: "hex lowercase",
			n:        &NumberLiteral{Base: BaseHex, BasePrefix: "\\h", IntegerPart: "ff"},
			expected: "\\hff",
		},
		{
			testName: "hex uppercase",
			n:        &NumberLiteral{Base: BaseHex, BasePrefix: "\\H", IntegerPart: "DEADBEEF"},
			expected: "\\HDEADBEEF",
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.n.String())
		})
	}
}

func (suite *NumberLiteralSuite) TestAscii() {
	// Ascii should be same as String for numbers
	n := &NumberLiteral{Base: BaseDecimal, IntegerPart: "42"}
	assert.Equal(suite.T(), n.String(), n.Ascii())
}

func (suite *NumberLiteralSuite) TestValidate() {
	tests := []struct {
		testName string
		n        *NumberLiteral
		errstr   string
	}{
		// OK.
		{
			testName: "valid decimal integer",
			n:        &NumberLiteral{Base: BaseDecimal, IntegerPart: "42"},
		},
		{
			testName: "valid decimal with fraction",
			n:        &NumberLiteral{Base: BaseDecimal, IntegerPart: "3", HasDecimalPoint: true, FractionalPart: "14"},
		},
		{
			testName: "valid decimal no whole part",
			n:        &NumberLiteral{Base: BaseDecimal, IntegerPart: "", HasDecimalPoint: true, FractionalPart: "5"},
		},
		{
			testName: "valid binary",
			n:        &NumberLiteral{Base: BaseBinary, BasePrefix: "\\b", IntegerPart: "1010"},
		},
		{
			testName: "valid octal",
			n:        &NumberLiteral{Base: BaseOctal, BasePrefix: "\\o", IntegerPart: "17"},
		},
		{
			testName: "valid hex",
			n:        &NumberLiteral{Base: BaseHex, BasePrefix: "\\h", IntegerPart: "ff"},
		},
		// Errors.
		{
			testName: "error empty number",
			n:        &NumberLiteral{Base: BaseDecimal, IntegerPart: ""},
			errstr:   "must have integer part or fractional part",
		},
		{
			testName: "error decimal point without fractional",
			n:        &NumberLiteral{Base: BaseDecimal, IntegerPart: "3", HasDecimalPoint: true, FractionalPart: ""},
			errstr:   "decimal point requires fractional part",
		},
		{
			testName: "error binary with decimal point",
			n:        &NumberLiteral{Base: BaseBinary, BasePrefix: "\\b", IntegerPart: "10", HasDecimalPoint: true, FractionalPart: "01"},
			errstr:   "non-decimal bases cannot have decimal points",
		},
		{
			testName: "error invalid binary digit",
			n:        &NumberLiteral{Base: BaseBinary, BasePrefix: "\\b", IntegerPart: "102"},
			errstr:   "invalid digit",
		},
		{
			testName: "error invalid octal digit",
			n:        &NumberLiteral{Base: BaseOctal, BasePrefix: "\\o", IntegerPart: "89"},
			errstr:   "invalid digit",
		},
		{
			testName: "error invalid hex digit",
			n:        &NumberLiteral{Base: BaseHex, BasePrefix: "\\h", IntegerPart: "xyz"},
			errstr:   "invalid digit",
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.n.Validate()
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

func (suite *NumberLiteralSuite) TestIsInteger() {
	tests := []struct {
		testName string
		n        *NumberLiteral
		expected bool
	}{
		{
			testName: "integer",
			n:        &NumberLiteral{Base: BaseDecimal, IntegerPart: "42"},
			expected: true,
		},
		{
			testName: "decimal",
			n:        &NumberLiteral{Base: BaseDecimal, IntegerPart: "3", HasDecimalPoint: true, FractionalPart: "14"},
			expected: false,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.n.IsInteger())
		})
	}
}

func (suite *NumberLiteralSuite) TestIsDecimal() {
	tests := []struct {
		testName string
		n        *NumberLiteral
		expected bool
	}{
		{
			testName: "integer",
			n:        &NumberLiteral{Base: BaseDecimal, IntegerPart: "42"},
			expected: false,
		},
		{
			testName: "decimal",
			n:        &NumberLiteral{Base: BaseDecimal, IntegerPart: "3", HasDecimalPoint: true, FractionalPart: "14"},
			expected: true,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.n.IsDecimal())
		})
	}
}

func (suite *NumberLiteralSuite) TestConstructors() {
	tests := []struct {
		testName string
		n        *NumberLiteral
		expected string
	}{
		{
			testName: "NewNumberLiteral",
			n:        NewNumberLiteral("42"),
			expected: "42",
		},
		{
			testName: "NewDecimalNumberLiteral",
			n:        NewDecimalNumberLiteral("3", "14"),
			expected: "3.14",
		},
		{
			testName: "NewBinaryNumberLiteral",
			n:        NewBinaryNumberLiteral("\\b", "1010"),
			expected: "\\b1010",
		},
		{
			testName: "NewOctalNumberLiteral",
			n:        NewOctalNumberLiteral("\\o", "17"),
			expected: "\\o17",
		},
		{
			testName: "NewHexNumberLiteral",
			n:        NewHexNumberLiteral("\\h", "ff"),
			expected: "\\hff",
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.n.String())
		})
	}
}

func (suite *NumberLiteralSuite) TestNewIntLiteral() {
	// Positive
	expr := NewIntLiteral(42)
	n, ok := expr.(*NumberLiteral)
	assert.True(suite.T(), ok, "expected *NumberLiteral for positive")
	assert.Equal(suite.T(), "42", n.String())

	// Zero
	expr = NewIntLiteral(0)
	n, ok = expr.(*NumberLiteral)
	assert.True(suite.T(), ok, "expected *NumberLiteral for zero")
	assert.Equal(suite.T(), "0", n.String())

	// Negative - should be NumericPrefixExpression
	expr = NewIntLiteral(-5)
	neg, ok := expr.(*NumericPrefixExpression)
	assert.True(suite.T(), ok, "expected *NumericPrefixExpression for negative")
	assert.Equal(suite.T(), "-", neg.Operator)
	inner, ok := neg.Right.(*NumberLiteral)
	assert.True(suite.T(), ok, "expected inner *NumberLiteral")
	assert.Equal(suite.T(), "5", inner.IntegerPart)
}

func (suite *NumberLiteralSuite) TestExpressionNode() {
	n := &NumberLiteral{Base: BaseDecimal, IntegerPart: "42"}
	n.expressionNode()
}

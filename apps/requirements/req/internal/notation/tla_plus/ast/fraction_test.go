package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestFractionExprSuite(t *testing.T) {
	suite.Run(t, new(FractionExprSuite))
}

type FractionExprSuite struct {
	suite.Suite
}

func (suite *FractionExprSuite) TestString() {
	tests := []struct {
		testName    string
		numerator   Expression
		denominator Expression
		expected    string
	}{
		{
			testName:    "simple integer fraction",
			numerator:   &NumberLiteral{Base: BaseDecimal, IntegerPart: "1"},
			denominator: &NumberLiteral{Base: BaseDecimal, IntegerPart: "2"},
			expected:    "1/2",
		},
		{
			testName:    "larger integers",
			numerator:   &NumberLiteral{Base: BaseDecimal, IntegerPart: "22"},
			denominator: &NumberLiteral{Base: BaseDecimal, IntegerPart: "7"},
			expected:    "22/7",
		},
		{
			testName:    "decimal numerator",
			numerator:   &NumberLiteral{Base: BaseDecimal, IntegerPart: "3", HasDecimalPoint: true, FractionalPart: "14"},
			denominator: &NumberLiteral{Base: BaseDecimal, IntegerPart: "2"},
			expected:    "3.14/2",
		},
		{
			testName:    "decimal denominator",
			numerator:   &NumberLiteral{Base: BaseDecimal, IntegerPart: "1"},
			denominator: &NumberLiteral{Base: BaseDecimal, IntegerPart: "0", HasDecimalPoint: true, FractionalPart: "5"},
			expected:    "1/0.5",
		},
		{
			testName: "negated numerator",
			numerator: &NumericPrefixExpression{
				Operator: "-",
				Right:    &NumberLiteral{Base: BaseDecimal, IntegerPart: "3"},
			},
			denominator: &NumberLiteral{Base: BaseDecimal, IntegerPart: "4"},
			expected:    "-3/4",
		},
		{
			testName:  "negated denominator",
			numerator: &NumberLiteral{Base: BaseDecimal, IntegerPart: "3"},
			denominator: &NumericPrefixExpression{
				Operator: "-",
				Right:    &NumberLiteral{Base: BaseDecimal, IntegerPart: "4"},
			},
			expected: "3/-4",
		},
		{
			testName: "nested fraction in numerator",
			numerator: &FractionExpr{
				Numerator:   &NumberLiteral{Base: BaseDecimal, IntegerPart: "1"},
				Denominator: &NumberLiteral{Base: BaseDecimal, IntegerPart: "2"},
			},
			denominator: &NumberLiteral{Base: BaseDecimal, IntegerPart: "3"},
			expected:    "1/2/3",
		},
		{
			testName:  "parenthesized numerator",
			numerator: &ParenExpr{Inner: &NumberLiteral{Base: BaseDecimal, IntegerPart: "5"}},
			denominator: &NumberLiteral{Base: BaseDecimal, IntegerPart: "10"},
			expected:    "(5)/10",
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			f := &FractionExpr{
				Numerator:   tt.numerator,
				Denominator: tt.denominator,
			}
			assert.Equal(t, tt.expected, f.String())
		})
	}
}

func (suite *FractionExprSuite) TestAscii() {
	// Ascii should be same as String for FractionExpr
	f := &FractionExpr{
		Numerator:   &NumberLiteral{Base: BaseDecimal, IntegerPart: "1"},
		Denominator: &NumberLiteral{Base: BaseDecimal, IntegerPart: "2"},
	}
	assert.Equal(suite.T(), f.String(), f.Ascii())
}

func (suite *FractionExprSuite) TestValidate() {
	tests := []struct {
		testName string
		f        *FractionExpr
		errstr   string
	}{
		// OK.
		{
			testName: "valid simple fraction",
			f: &FractionExpr{
				Numerator:   &NumberLiteral{Base: BaseDecimal, IntegerPart: "1"},
				Denominator: &NumberLiteral{Base: BaseDecimal, IntegerPart: "2"},
			},
		},
		{
			testName: "valid with negation",
			f: &FractionExpr{
				Numerator: &NumericPrefixExpression{
					Operator: "-",
					Right:    &NumberLiteral{Base: BaseDecimal, IntegerPart: "3"},
				},
				Denominator: &NumberLiteral{Base: BaseDecimal, IntegerPart: "4"},
			},
		},
		{
			testName: "valid nested fraction",
			f: &FractionExpr{
				Numerator: &FractionExpr{
					Numerator:   &NumberLiteral{Base: BaseDecimal, IntegerPart: "1"},
					Denominator: &NumberLiteral{Base: BaseDecimal, IntegerPart: "2"},
				},
				Denominator: &NumberLiteral{Base: BaseDecimal, IntegerPart: "3"},
			},
		},

		// Errors.
		{
			testName: "error nil numerator",
			f: &FractionExpr{
				Numerator:   nil,
				Denominator: &NumberLiteral{Base: BaseDecimal, IntegerPart: "2"},
			},
			errstr: "Numerator",
		},
		{
			testName: "error nil denominator",
			f: &FractionExpr{
				Numerator:   &NumberLiteral{Base: BaseDecimal, IntegerPart: "1"},
				Denominator: nil,
			},
			errstr: "Denominator",
		},
		{
			testName: "error invalid numerator",
			f: &FractionExpr{
				Numerator:   &NumberLiteral{Base: BaseDecimal, IntegerPart: "", FractionalPart: ""},
				Denominator: &NumberLiteral{Base: BaseDecimal, IntegerPart: "2"},
			},
			errstr: "Numerator",
		},
		{
			testName: "error invalid denominator",
			f: &FractionExpr{
				Numerator:   &NumberLiteral{Base: BaseDecimal, IntegerPart: "1"},
				Denominator: &NumberLiteral{Base: BaseDecimal, IntegerPart: "", FractionalPart: ""},
			},
			errstr: "Denominator",
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.f.Validate()
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

func (suite *FractionExprSuite) TestNewFractionExpr() {
	numerator := &NumberLiteral{Base: BaseDecimal, IntegerPart: "3"}
	denominator := &NumberLiteral{Base: BaseDecimal, IntegerPart: "4"}
	f := NewFractionExpr(numerator, denominator)
	assert.Equal(suite.T(), numerator, f.Numerator)
	assert.Equal(suite.T(), denominator, f.Denominator)
}

func (suite *FractionExprSuite) TestExpressionNode() {
	f := &FractionExpr{
		Numerator:   &NumberLiteral{Base: BaseDecimal, IntegerPart: "1"},
		Denominator: &NumberLiteral{Base: BaseDecimal, IntegerPart: "2"},
	}
	f.expressionNode()
}

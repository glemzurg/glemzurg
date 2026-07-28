package ast

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestExpressionTupleIndexSuite(t *testing.T) {
	suite.Run(t, new(ExpressionTupleIndexSuite))
}

type ExpressionTupleIndexSuite struct {
	suite.Suite
}

func (suite *ExpressionTupleIndexSuite) TestString() {
	tests := []struct {
		testName string
		tuple    Expression
		index    Expression
		expected string
	}{
		{
			testName: `literal tuple with literal index`,
			tuple: &TupleLiteral{
				Elements: []Expression{
					newIntLiteral(1),
					newIntLiteral(2),
					newIntLiteral(3),
				},
			},
			index:    newIntLiteral(1),
			expected: `⟨1, 2, 3⟩[1]`,
		},
		{
			testName: `literal tuple with index 0`,
			tuple: &TupleLiteral{
				Elements: []Expression{
					&StringLiteral{Value: `a`},
					&StringLiteral{Value: `b`},
				},
			},
			index:    newIntLiteral(0),
			expected: `⟨"a", "b"⟩[0]`,
		},
		{
			testName: `appended tuple with index`,
			tuple: &BuiltinCall{
				Name: "_Seq!Append",
				Args: []Expression{
					&TupleLiteral{
						Elements: []Expression{
							newIntLiteral(1),
						},
					},
					newIntLiteral(2),
				},
			},
			index:    newIntLiteral(2),
			expected: `_Seq!Append(⟨1⟩, 2)[2]`,
		},
		{
			testName: `tuple tail with index`,
			tuple: &BuiltinCall{
				Name: "_Seq!Tail",
				Args: []Expression{
					&TupleLiteral{
						Elements: []Expression{
							newIntLiteral(1),
							newIntLiteral(2),
						},
					},
				},
			},
			index:    newIntLiteral(1),
			expected: `_Seq!Tail(⟨1, 2⟩)[1]`,
		},
	}
	for _, tt := range tests {
		_ = suite.Run(tt.testName, func() {
			expr := &ExpressionTupleIndex{
				Tuple: tt.tuple,
				Index: tt.index,
			}
			suite.Equal(tt.expected, expr.String())
		})
	}
}

func (suite *ExpressionTupleIndexSuite) TestASCII() {
	tests := []struct {
		testName string
		tuple    Expression
		index    Expression
		expected string
	}{
		{
			testName: `literal tuple with literal index`,
			tuple: &TupleLiteral{
				Elements: []Expression{
					newIntLiteral(1),
					newIntLiteral(2),
					newIntLiteral(3),
				},
			},
			index:    newIntLiteral(1),
			expected: `<<1, 2, 3>>[1]`,
		},
		{
			testName: `appended tuple with index`,
			tuple: &BuiltinCall{
				Name: "_Seq!Append",
				Args: []Expression{
					&TupleLiteral{
						Elements: []Expression{
							newIntLiteral(1),
						},
					},
					newIntLiteral(2),
				},
			},
			index:    newIntLiteral(2),
			expected: `_Seq!Append(<<1>>, 2)[2]`,
		},
	}
	for _, tt := range tests {
		_ = suite.Run(tt.testName, func() {
			expr := &ExpressionTupleIndex{
				Tuple: tt.tuple,
				Index: tt.index,
			}
			suite.Equal(tt.expected, expr.ASCII())
		})
	}
}

func (suite *ExpressionTupleIndexSuite) TestValidate() {
	tests := []struct {
		testName string
		e        *ExpressionTupleIndex
		errstr   string
	}{
		// OK.
		{
			testName: `valid index`,
			e: &ExpressionTupleIndex{
				Tuple: &TupleLiteral{
					Elements: []Expression{
						newIntLiteral(1),
					},
				},
				Index: newIntLiteral(0),
			},
		},

		// Errors.
		{
			testName: `error missing tuple`,
			e: &ExpressionTupleIndex{
				Index: newIntLiteral(0),
			},
			errstr: `Tuple`,
		},
		{
			testName: `error missing index`,
			e: &ExpressionTupleIndex{
				Tuple: &TupleLiteral{
					Elements: []Expression{
						newIntLiteral(1),
					},
				},
			},
			errstr: `Index`,
		},
		{
			testName: `error invalid tuple`,
			e: &ExpressionTupleIndex{
				Tuple: &TupleLiteral{
					Elements: []Expression{
						&Identifier{Value: ``},
					},
				},
				Index: newIntLiteral(0),
			},
			errstr: `Value`,
		},
	}
	for _, tt := range tests {
		_ = suite.Run(tt.testName, func() {
			err := tt.e.Validate()
			if tt.errstr == `` {
				suite.Require().NoError(err)
			} else {
				suite.Require().ErrorContains(err, tt.errstr)
			}
		})
	}
}

func (suite *ExpressionTupleIndexSuite) TestExpressionNode() {
	e := &ExpressionTupleIndex{
		Tuple: &TupleLiteral{
			Elements: []Expression{
				newIntLiteral(1),
			},
		},
		Index: newIntLiteral(0),
	}
	e.expressionNode()
}

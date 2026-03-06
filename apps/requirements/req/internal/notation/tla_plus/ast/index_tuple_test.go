package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
					NewIntLiteral(1),
					NewIntLiteral(2),
					NewIntLiteral(3),
				},
			},
			index:    NewIntLiteral(1),
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
			index:    NewIntLiteral(0),
			expected: `⟨"a", "b"⟩[0]`,
		},
		{
			testName: `appended tuple with index`,
			tuple: &BuiltinCall{
				Name: "_Seq!Append",
				Args: []Expression{
					&TupleLiteral{
						Elements: []Expression{
							NewIntLiteral(1),
						},
					},
					NewIntLiteral(2),
				},
			},
			index:    NewIntLiteral(2),
			expected: `_Seq!Append(⟨1⟩, 2)[2]`,
		},
		{
			testName: `tuple tail with index`,
			tuple: &BuiltinCall{
				Name: "_Seq!Tail",
				Args: []Expression{
					&TupleLiteral{
						Elements: []Expression{
							NewIntLiteral(1),
							NewIntLiteral(2),
						},
					},
				},
			},
			index:    NewIntLiteral(1),
			expected: `_Seq!Tail(⟨1, 2⟩)[1]`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			expr := &ExpressionTupleIndex{
				Tuple: tt.tuple,
				Index: tt.index,
			}
			assert.Equal(t, tt.expected, expr.String())
		})
	}
}

func (suite *ExpressionTupleIndexSuite) TestAscii() {
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
					NewIntLiteral(1),
					NewIntLiteral(2),
					NewIntLiteral(3),
				},
			},
			index:    NewIntLiteral(1),
			expected: `<<1, 2, 3>>[1]`,
		},
		{
			testName: `appended tuple with index`,
			tuple: &BuiltinCall{
				Name: "_Seq!Append",
				Args: []Expression{
					&TupleLiteral{
						Elements: []Expression{
							NewIntLiteral(1),
						},
					},
					NewIntLiteral(2),
				},
			},
			index:    NewIntLiteral(2),
			expected: `_Seq!Append(<<1>>, 2)[2]`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			expr := &ExpressionTupleIndex{
				Tuple: tt.tuple,
				Index: tt.index,
			}
			assert.Equal(t, tt.expected, expr.Ascii())
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
						NewIntLiteral(1),
					},
				},
				Index: NewIntLiteral(0),
			},
		},

		// Errors.
		{
			testName: `error missing tuple`,
			e: &ExpressionTupleIndex{
				Index: NewIntLiteral(0),
			},
			errstr: `Tuple`,
		},
		{
			testName: `error missing index`,
			e: &ExpressionTupleIndex{
				Tuple: &TupleLiteral{
					Elements: []Expression{
						NewIntLiteral(1),
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
				Index: NewIntLiteral(0),
			},
			errstr: `Value`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.e.Validate()
			if tt.errstr == `` {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

func (suite *ExpressionTupleIndexSuite) TestExpressionNode() {
	e := &ExpressionTupleIndex{
		Tuple: &TupleLiteral{
			Elements: []Expression{
				NewIntLiteral(1),
			},
		},
		Index: NewIntLiteral(0),
	}
	e.expressionNode()
}

package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestTupleInfixExpressionSuite(t *testing.T) {
	suite.Run(t, new(TupleInfixExpressionSuite))
}

type TupleInfixExpressionSuite struct {
	suite.Suite
}

func (suite *TupleInfixExpressionSuite) TestString() {
	tests := []struct {
		testName string
		operator string
		operands []Expression
		expected string
	}{
		{
			testName: `two tuples`,
			operator: TupleOperatorConcat,
			operands: []Expression{
				&TupleLiteral{
					Elements: []Expression{
						NewIntLiteral(1),
						NewIntLiteral(2),
					},
				},
				&TupleLiteral{
					Elements: []Expression{
						NewIntLiteral(3),
						NewIntLiteral(4),
					},
				},
			},
			expected: `⟨1, 2⟩ ∘ ⟨3, 4⟩`,
		},
		{
			testName: `three tuples`,
			operator: TupleOperatorConcat,
			operands: []Expression{
				&TupleLiteral{
					Elements: []Expression{
						NewIntLiteral(1),
					},
				},
				&TupleLiteral{
					Elements: []Expression{
						NewIntLiteral(2),
					},
				},
				&TupleLiteral{
					Elements: []Expression{
						NewIntLiteral(3),
					},
				},
			},
			expected: `⟨1⟩ ∘ ⟨2⟩ ∘ ⟨3⟩`,
		},
		{
			testName: `with empty tuple`,
			operator: TupleOperatorConcat,
			operands: []Expression{
				&TupleLiteral{
					Elements: []Expression{
						NewIntLiteral(1),
					},
				},
				&TupleLiteral{
					Elements: []Expression{},
				},
			},
			expected: `⟨1⟩ ∘ ⟨⟩`,
		},
		{
			testName: `with tuple tail`,
			operator: TupleOperatorConcat,
			operands: []Expression{
				&TupleLiteral{
					Elements: []Expression{
						NewIntLiteral(1),
					},
				},
				&BuiltinCall{
					Name: "_Seq!Tail",
					Args: []Expression{
						&TupleLiteral{
							Elements: []Expression{
								NewIntLiteral(2),
								NewIntLiteral(3),
							},
						},
					},
				},
			},
			expected: `⟨1⟩ ∘ _Seq!Tail(⟨2, 3⟩)`,
		},
		{
			testName: `with tuple append`,
			operator: TupleOperatorConcat,
			operands: []Expression{
				&BuiltinCall{
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
				&TupleLiteral{
					Elements: []Expression{
						NewIntLiteral(3),
					},
				},
			},
			expected: `_Seq!Append(⟨1⟩, 2) ∘ ⟨3⟩`,
		},
		{
			testName: `nested concat`,
			operator: TupleOperatorConcat,
			operands: []Expression{
				&TupleInfixExpression{
					Operator: TupleOperatorConcat,
					Operands: []Expression{
						&TupleLiteral{
							Elements: []Expression{
								NewIntLiteral(1),
							},
						},
						&TupleLiteral{
							Elements: []Expression{
								NewIntLiteral(2),
							},
						},
					},
				},
				&TupleLiteral{
					Elements: []Expression{
						NewIntLiteral(3),
					},
				},
			},
			expected: `⟨1⟩ ∘ ⟨2⟩ ∘ ⟨3⟩`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			expr := &TupleInfixExpression{
				Operator: tt.operator,
				Operands: tt.operands,
			}
			assert.Equal(t, tt.expected, expr.String())
		})
	}
}

func (suite *TupleInfixExpressionSuite) TestAscii() {
	tests := []struct {
		testName string
		operator string
		operands []Expression
		expected string
	}{
		{
			testName: `two tuples`,
			operator: TupleOperatorConcat,
			operands: []Expression{
				&TupleLiteral{
					Elements: []Expression{
						NewIntLiteral(1),
						NewIntLiteral(2),
					},
				},
				&TupleLiteral{
					Elements: []Expression{
						NewIntLiteral(3),
						NewIntLiteral(4),
					},
				},
			},
			expected: `<<1, 2>> \o <<3, 4>>`,
		},
		{
			testName: `three tuples`,
			operator: TupleOperatorConcat,
			operands: []Expression{
				&TupleLiteral{
					Elements: []Expression{
						NewIntLiteral(1),
					},
				},
				&TupleLiteral{
					Elements: []Expression{
						NewIntLiteral(2),
					},
				},
				&TupleLiteral{
					Elements: []Expression{
						NewIntLiteral(3),
					},
				},
			},
			expected: `<<1>> \o <<2>> \o <<3>>`,
		},
		{
			testName: `with tuple tail`,
			operator: TupleOperatorConcat,
			operands: []Expression{
				&TupleLiteral{
					Elements: []Expression{
						NewIntLiteral(1),
					},
				},
				&BuiltinCall{
					Name: "_Seq!Tail",
					Args: []Expression{
						&TupleLiteral{
							Elements: []Expression{
								NewIntLiteral(2),
								NewIntLiteral(3),
							},
						},
					},
				},
			},
			expected: `<<1>> \o _Seq!Tail(<<2, 3>>)`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			expr := &TupleInfixExpression{
				Operator: tt.operator,
				Operands: tt.operands,
			}
			assert.Equal(t, tt.expected, expr.Ascii())
		})
	}
}

func (suite *TupleInfixExpressionSuite) TestValidate() {
	tests := []struct {
		testName string
		t        *TupleInfixExpression
		errstr   string
	}{
		// OK.
		{
			testName: `valid two operands`,
			t: &TupleInfixExpression{
				Operator: TupleOperatorConcat,
				Operands: []Expression{
					&TupleLiteral{
						Elements: []Expression{
							NewIntLiteral(1),
						},
					},
					&TupleLiteral{
						Elements: []Expression{
							NewIntLiteral(2),
						},
					},
				},
			},
		},
		{
			testName: `valid three operands`,
			t: &TupleInfixExpression{
				Operator: TupleOperatorConcat,
				Operands: []Expression{
					&TupleLiteral{
						Elements: []Expression{
							NewIntLiteral(1),
						},
					},
					&TupleLiteral{
						Elements: []Expression{
							NewIntLiteral(2),
						},
					},
					&TupleLiteral{
						Elements: []Expression{
							NewIntLiteral(3),
						},
					},
				},
			},
		},
		{
			testName: `valid nested`,
			t: &TupleInfixExpression{
				Operator: TupleOperatorConcat,
				Operands: []Expression{
					&TupleInfixExpression{
						Operator: TupleOperatorConcat,
						Operands: []Expression{
							&TupleLiteral{
								Elements: []Expression{
									NewIntLiteral(1),
								},
							},
							&TupleLiteral{
								Elements: []Expression{
									NewIntLiteral(2),
								},
							},
						},
					},
					&TupleLiteral{
						Elements: []Expression{
							NewIntLiteral(3),
						},
					},
				},
			},
		},

		// Errors.
		{
			testName: `error missing operator`,
			t: &TupleInfixExpression{
				Operands: []Expression{
					&TupleLiteral{
						Elements: []Expression{
							NewIntLiteral(1),
						},
					},
					&TupleLiteral{
						Elements: []Expression{
							NewIntLiteral(2),
						},
					},
				},
			},
			errstr: `Operator`,
		},
		{
			testName: `error invalid operator`,
			t: &TupleInfixExpression{
				Operator: "+",
				Operands: []Expression{
					&TupleLiteral{
						Elements: []Expression{
							NewIntLiteral(1),
						},
					},
					&TupleLiteral{
						Elements: []Expression{
							NewIntLiteral(2),
						},
					},
				},
			},
			errstr: `Operator`,
		},
		{
			testName: `error missing operands`,
			t: &TupleInfixExpression{
				Operator: TupleOperatorConcat,
			},
			errstr: `Operands`,
		},
		{
			testName: `error single operand`,
			t: &TupleInfixExpression{
				Operator: TupleOperatorConcat,
				Operands: []Expression{
					&TupleLiteral{
						Elements: []Expression{
							NewIntLiteral(1),
						},
					},
				},
			},
			errstr: `Operands`,
		},
		{
			testName: `error nil operand`,
			t: &TupleInfixExpression{
				Operator: TupleOperatorConcat,
				Operands: []Expression{
					&TupleLiteral{
						Elements: []Expression{
							NewIntLiteral(1),
						},
					},
					nil,
				},
			},
			errstr: `Operands[1]`,
		},
		{
			testName: `error invalid operand element`,
			t: &TupleInfixExpression{
				Operator: TupleOperatorConcat,
				Operands: []Expression{
					&TupleLiteral{
						Elements: []Expression{
							&Identifier{Value: ``},
						},
					},
					&TupleLiteral{
						Elements: []Expression{
							NewIntLiteral(2),
						},
					},
				},
			},
			errstr: `Value`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.t.Validate()
			if tt.errstr == `` {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

func (suite *TupleInfixExpressionSuite) TestExpressionNode() {
	t := &TupleInfixExpression{
		Operator: TupleOperatorConcat,
		Operands: []Expression{
			&TupleLiteral{
				Elements: []Expression{
					NewIntLiteral(1),
				},
			},
			&TupleLiteral{
				Elements: []Expression{
					NewIntLiteral(2),
				},
			},
		},
	}
	t.expressionNode()
}


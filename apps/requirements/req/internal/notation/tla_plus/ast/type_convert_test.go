package ast

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_expression_type"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestTypeConvertSuite(t *testing.T) {
	suite.Run(t, new(TypeConvertSuite))
}

type TypeConvertSuite struct {
	suite.Suite
}

func (suite *TypeConvertSuite) TestConvertToExpressionType() {
	tests := []struct {
		testName string
		expr     Expression
		expected model_expression_type.ExpressionType
		errstr   string
	}{
		// --- Scalar types via Identifier ---
		{
			testName: "BOOLEAN identifier",
			expr:     &Identifier{Value: "BOOLEAN"},
			expected: &model_expression_type.BooleanType{},
		},
		{
			testName: "Nat identifier",
			expr:     &Identifier{Value: "Nat"},
			expected: &model_expression_type.IntegerType{},
		},
		{
			testName: "Int identifier",
			expr:     &Identifier{Value: "Int"},
			expected: &model_expression_type.IntegerType{},
		},
		{
			testName: "Real identifier",
			expr:     &Identifier{Value: "Real"},
			expected: &model_expression_type.RationalType{},
		},
		{
			testName: "STRING identifier",
			expr:     &Identifier{Value: "STRING"},
			expected: &model_expression_type.StringType{},
		},
		{
			testName: "error unknown identifier",
			expr:     &Identifier{Value: "Unknown"},
			errstr:   "unknown type identifier: Unknown",
		},

		// --- Scalar types via SetConstant ---
		{
			testName: "BOOLEAN set constant",
			expr:     &SetConstant{Value: SetConstantBoolean},
			expected: &model_expression_type.BooleanType{},
		},
		{
			testName: "Nat set constant",
			expr:     &SetConstant{Value: SetConstantNat},
			expected: &model_expression_type.IntegerType{},
		},
		{
			testName: "Int set constant",
			expr:     &SetConstant{Value: SetConstantInt},
			expected: &model_expression_type.IntegerType{},
		},
		{
			testName: "Real set constant",
			expr:     &SetConstant{Value: SetConstantReal},
			expected: &model_expression_type.RationalType{},
		},
		{
			testName: "error unknown set constant",
			expr:     &SetConstant{Value: "INVALID"},
			errstr:   "unknown set constant for type: INVALID",
		},

		// --- Enum types ---
		{
			testName: "enum via SetLiteralEnum",
			expr:     &SetLiteralEnum{Values: []string{"active", "inactive", "pending"}},
			expected: &model_expression_type.EnumType{Values: []string{"active", "inactive", "pending"}},
		},
		{
			testName: "enum via SetLiteral with string literals",
			expr: &SetLiteral{Elements: []Expression{
				&StringLiteral{Value: "red"},
				&StringLiteral{Value: "green"},
				&StringLiteral{Value: "blue"},
			}},
			expected: &model_expression_type.EnumType{Values: []string{"red", "green", "blue"}},
		},
		{
			testName: "error empty SetLiteralEnum",
			expr:     &SetLiteralEnum{Values: []string{}},
			errstr:   "enum type must have at least one value",
		},
		{
			testName: "error empty SetLiteral",
			expr:     &SetLiteral{Elements: []Expression{}},
			errstr:   "set literal as type must have at least one element",
		},
		{
			testName: "error SetLiteral with non-string element",
			expr: &SetLiteral{Elements: []Expression{
				NewNumberLiteral("42"),
			}},
			errstr: "set literal element 0 is not a string literal",
		},

		// --- Sequence types ---
		{
			testName: "Seq(Int)",
			expr: &FunctionCall{
				ScopePath: []*Identifier{{Value: "_Seq"}},
				Name:      &Identifier{Value: "Seq"},
				Args:      []Expression{&Identifier{Value: "Int"}},
			},
			expected: &model_expression_type.SequenceType{
				ElementType: &model_expression_type.IntegerType{},
				Unique:      false,
			},
		},
		{
			testName: "SeqUnique(STRING)",
			expr: &FunctionCall{
				ScopePath: []*Identifier{{Value: "_Seq"}},
				Name:      &Identifier{Value: "SeqUnique"},
				Args:      []Expression{&Identifier{Value: "STRING"}},
			},
			expected: &model_expression_type.SequenceType{
				ElementType: &model_expression_type.StringType{},
				Unique:      true,
			},
		},
		{
			testName: "error unknown _Seq function",
			expr: &FunctionCall{
				ScopePath: []*Identifier{{Value: "_Seq"}},
				Name:      &Identifier{Value: "BadFunc"},
				Args:      []Expression{&Identifier{Value: "Int"}},
			},
			errstr: "unknown _Seq function for type: BadFunc",
		},

		// --- Set type ---
		{
			testName: "_Set(Int)",
			expr: &FunctionCall{
				ScopePath: []*Identifier{{Value: "_Set"}},
				Name:      &Identifier{Value: "_Set"},
				Args:      []Expression{&Identifier{Value: "Int"}},
			},
			expected: &model_expression_type.SetType{
				ElementType: &model_expression_type.IntegerType{},
			},
		},
		{
			testName: "error unknown _Set function",
			expr: &FunctionCall{
				ScopePath: []*Identifier{{Value: "_Set"}},
				Name:      &Identifier{Value: "BadFunc"},
				Args:      []Expression{&Identifier{Value: "Int"}},
			},
			errstr: "unknown _Set function for type: BadFunc",
		},

		// --- Bag type ---
		{
			testName: "_Bag(STRING)",
			expr: &FunctionCall{
				ScopePath: []*Identifier{{Value: "_Bags"}},
				Name:      &Identifier{Value: "_Bag"},
				Args:      []Expression{&Identifier{Value: "STRING"}},
			},
			expected: &model_expression_type.BagType{
				ElementType: &model_expression_type.StringType{},
			},
		},
		{
			testName: "error unknown _Bags function",
			expr: &FunctionCall{
				ScopePath: []*Identifier{{Value: "_Bags"}},
				Name:      &Identifier{Value: "BadFunc"},
				Args:      []Expression{&Identifier{Value: "Int"}},
			},
			errstr: "unknown _Bags function for type: BadFunc",
		},

		// --- Function call errors ---
		{
			testName: "error unknown module",
			expr: &FunctionCall{
				ScopePath: []*Identifier{{Value: "_Unknown"}},
				Name:      &Identifier{Value: "Func"},
				Args:      []Expression{&Identifier{Value: "Int"}},
			},
			errstr: "unknown module for type expression: _Unknown",
		},
		{
			testName: "error no scope path",
			expr: &FunctionCall{
				ScopePath: []*Identifier{},
				Name:      &Identifier{Value: "Func"},
				Args:      []Expression{&Identifier{Value: "Int"}},
			},
			errstr: "not a valid type expression",
		},
		{
			testName: "error wrong number of args",
			expr: &FunctionCall{
				ScopePath: []*Identifier{{Value: "_Seq"}},
				Name:      &Identifier{Value: "Seq"},
				Args:      []Expression{&Identifier{Value: "Int"}, &Identifier{Value: "STRING"}},
			},
			errstr: "requires exactly 1 argument, got 2",
		},
		{
			testName: "error invalid arg type",
			expr: &FunctionCall{
				ScopePath: []*Identifier{{Value: "_Seq"}},
				Name:      &Identifier{Value: "Seq"},
				Args:      []Expression{&Identifier{Value: "Unknown"}},
			},
			errstr: "type constructor _Seq!Seq argument",
		},

		// --- Record type ---
		{
			testName: "record type [name: STRING, age: Int]",
			expr: &RecordTypeExpr{
				Fields: []*RecordTypeField{
					{Name: &Identifier{Value: "name"}, Type: &Identifier{Value: "STRING"}},
					{Name: &Identifier{Value: "age"}, Type: &Identifier{Value: "Int"}},
				},
			},
			expected: &model_expression_type.RecordType{
				Fields: []model_expression_type.RecordFieldType{
					{Name: "name", Type: &model_expression_type.StringType{}},
					{Name: "age", Type: &model_expression_type.IntegerType{}},
				},
			},
		},
		{
			testName: "error record field with invalid type",
			expr: &RecordTypeExpr{
				Fields: []*RecordTypeField{
					{Name: &Identifier{Value: "x"}, Type: &Identifier{Value: "Unknown"}},
				},
			},
			errstr: "record field x",
		},

		// --- Cartesian product / Tuple type ---
		{
			testName: "Int \\X STRING",
			expr: &CartesianProduct{
				Operands: []Expression{
					&Identifier{Value: "Int"},
					&Identifier{Value: "STRING"},
				},
			},
			expected: &model_expression_type.TupleType{
				ElementTypes: []model_expression_type.ExpressionType{
					&model_expression_type.IntegerType{},
					&model_expression_type.StringType{},
				},
			},
		},
		{
			testName: "three-way cartesian product",
			expr: &CartesianProduct{
				Operands: []Expression{
					&Identifier{Value: "Int"},
					&Identifier{Value: "STRING"},
					&Identifier{Value: "BOOLEAN"},
				},
			},
			expected: &model_expression_type.TupleType{
				ElementTypes: []model_expression_type.ExpressionType{
					&model_expression_type.IntegerType{},
					&model_expression_type.StringType{},
					&model_expression_type.BooleanType{},
				},
			},
		},
		{
			testName: "error cartesian product with invalid operand",
			expr: &CartesianProduct{
				Operands: []Expression{
					&Identifier{Value: "Int"},
					&Identifier{Value: "Unknown"},
				},
			},
			errstr: "cartesian product operand 1",
		},

		// --- Nested types ---
		{
			testName: "Seq of records",
			expr: &FunctionCall{
				ScopePath: []*Identifier{{Value: "_Seq"}},
				Name:      &Identifier{Value: "Seq"},
				Args: []Expression{
					&RecordTypeExpr{
						Fields: []*RecordTypeField{
							{Name: &Identifier{Value: "id"}, Type: &Identifier{Value: "Int"}},
							{Name: &Identifier{Value: "label"}, Type: &Identifier{Value: "STRING"}},
						},
					},
				},
			},
			expected: &model_expression_type.SequenceType{
				ElementType: &model_expression_type.RecordType{
					Fields: []model_expression_type.RecordFieldType{
						{Name: "id", Type: &model_expression_type.IntegerType{}},
						{Name: "label", Type: &model_expression_type.StringType{}},
					},
				},
				Unique: false,
			},
		},
		{
			testName: "Set of tuples",
			expr: &FunctionCall{
				ScopePath: []*Identifier{{Value: "_Set"}},
				Name:      &Identifier{Value: "_Set"},
				Args: []Expression{
					&CartesianProduct{
						Operands: []Expression{
							&Identifier{Value: "Int"},
							&Identifier{Value: "STRING"},
						},
					},
				},
			},
			expected: &model_expression_type.SetType{
				ElementType: &model_expression_type.TupleType{
					ElementTypes: []model_expression_type.ExpressionType{
						&model_expression_type.IntegerType{},
						&model_expression_type.StringType{},
					},
				},
			},
		},

		// --- Invalid AST node types ---
		{
			testName: "error nil expression",
			expr:     nil,
			errstr:   "type expression is nil",
		},
		{
			testName: "error number literal is not a type",
			expr:     NewNumberLiteral("42"),
			errstr:   "not a valid type expression: 42",
		},
		{
			testName: "error boolean literal is not a type",
			expr:     &BooleanLiteral{Value: true},
			errstr:   "not a valid type expression",
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.testName, func(t *testing.T) {
			result, err := ConvertToExpressionType(tt.expr)
			if tt.errstr == "" {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

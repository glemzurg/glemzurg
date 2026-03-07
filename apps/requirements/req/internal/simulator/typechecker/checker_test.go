package typechecker

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestCheckerSuite(t *testing.T) {
	suite.Run(t, new(CheckerSuite))
}

type CheckerSuite struct {
	suite.Suite
	tc *TypeChecker
}

func (s *CheckerSuite) SetupTest() {
	types.ResetTypeVarCounter()
	s.tc = NewTypeChecker()
}

// === Literals ===

func (s *CheckerSuite) TestCheck_BooleanLiteral() {
	node := &ast.BooleanLiteral{Value: true}
	typed, err := s.tc.Check(node)

	s.NoError(err)
	s.True( typed.Type.Equals(types.Boolean{}))
}

func (s *CheckerSuite) TestCheck_NaturalLiteral() {
	node := ast.NewIntLiteral(42)
	typed, err := s.tc.Check(node)

	s.NoError(err)
	s.True( typed.Type.Equals(types.Number{}))
}

func (s *CheckerSuite) TestCheck_IntegerLiteral() {
	node := ast.NewIntLiteral(-5)
	typed, err := s.tc.Check(node)

	s.NoError(err)
	s.True( typed.Type.Equals(types.Number{}))
}

func (s *CheckerSuite) TestCheck_RealLiteral() {
	node := ast.NewFraction(ast.NewIntLiteral(7), ast.NewIntLiteral(2))
	typed, err := s.tc.Check(node)

	s.NoError(err)
	s.True( typed.Type.Equals(types.Number{}))
}

func (s *CheckerSuite) TestCheck_StringLiteral() {
	node := &ast.StringLiteral{Value: "hello"}
	typed, err := s.tc.Check(node)

	s.NoError(err)
	s.True( typed.Type.Equals(types.String{}))
}

// === Identifiers ===

func (s *CheckerSuite) TestCheck_Identifier_Bound() {
	s.tc.env.BindMono("x", types.Number{})
	node := &ast.Identifier{Value: "x"}
	typed, err := s.tc.Check(node)

	s.NoError(err)
	s.True( typed.Type.Equals(types.Number{}))
}

func (s *CheckerSuite) TestCheck_Identifier_Unbound() {
	node := &ast.Identifier{Value: "undefined"}
	_, err := s.tc.Check(node)

	s.Error(err)
	s.Contains( err.Error(), "unbound variable")
}

// === Arithmetic ===

func (s *CheckerSuite) TestCheck_RealInfix_Valid() {
	node := &ast.BinaryArithmetic{
		Left:     ast.NewIntLiteral(1),
		Operator: "+",
		Right:    ast.NewIntLiteral(2),
	}
	typed, err := s.tc.Check(node)

	s.NoError(err)
	s.True( typed.Type.Equals(types.Number{}))
	s.Len(typed.Children, 2)
}

// === Logic ===

func (s *CheckerSuite) TestCheck_LogicInfix_Valid() {
	node := &ast.BinaryLogic{
		Left:     &ast.BooleanLiteral{Value: true},
		Operator: "∧",
		Right:    &ast.BooleanLiteral{Value: false},
	}
	typed, err := s.tc.Check(node)

	s.NoError(err)
	s.True( typed.Type.Equals(types.Boolean{}))
}

func (s *CheckerSuite) TestCheck_LogicPrefix_Valid() {
	node := &ast.UnaryLogic{
		Operator: "¬",
		Right:    &ast.BooleanLiteral{Value: true},
	}
	typed, err := s.tc.Check(node)

	s.NoError(err)
	s.True( typed.Type.Equals(types.Boolean{}))
}

func (s *CheckerSuite) TestCheck_RealComparison_Valid() {
	node := &ast.BinaryComparison{
		Left:     ast.NewIntLiteral(1),
		Operator: "<",
		Right:    ast.NewIntLiteral(2),
	}
	typed, err := s.tc.Check(node)

	s.NoError(err)
	s.True( typed.Type.Equals(types.Boolean{}))
}

// === Sets ===

func (s *CheckerSuite) TestCheck_SetLiteralInt() {
	node := &ast.SetLiteralInt{Values: []int{1, 2, 3}}
	typed, err := s.tc.Check(node)

	s.NoError(err)
	setType, ok := typed.Type.(types.Set)
	s.True(ok)
	s.True(setType.Element.Equals(types.Number{}))
}

func (s *CheckerSuite) TestCheck_SetLiteralEnum_Strings() {
	node := &ast.SetLiteralEnum{Values: []string{"a", "b", "c"}}
	typed, err := s.tc.Check(node)

	s.NoError(err)
	setType, ok := typed.Type.(types.Set)
	s.True(ok)
	s.True(setType.Element.Equals(types.String{}))
}

func (s *CheckerSuite) TestCheck_SetRange() {
	node := &ast.SetRange{Start: 1, End: 10}
	typed, err := s.tc.Check(node)

	s.NoError(err)
	setType, ok := typed.Type.(types.Set)
	s.True(ok)
	s.True(setType.Element.Equals(types.Number{}))
}

func (s *CheckerSuite) TestCheck_SetInfix_Union() {
	node := &ast.BinarySetOperation{
		Left:     &ast.SetLiteralInt{Values: []int{1, 2}},
		Operator: "∪",
		Right:    &ast.SetLiteralInt{Values: []int{3, 4}},
	}
	typed, err := s.tc.Check(node)

	s.NoError(err)
	setType, ok := typed.Type.(types.Set)
	s.True(ok)
	s.True(setType.Element.Equals(types.Number{}))
}

func (s *CheckerSuite) TestCheck_SetConstant_Boolean() {
	node := &ast.SetConstant{Value: "BOOLEAN"}
	typed, err := s.tc.Check(node)

	s.NoError(err)
	setType, ok := typed.Type.(types.Set)
	s.True(ok)
	s.True(setType.Element.Equals(types.Boolean{}))
}

// === Membership ===

func (s *CheckerSuite) TestCheck_Membership_Valid() {
	node := &ast.Membership{
		Left:     ast.NewIntLiteral(1),
		Operator: "∈",
		Right:    &ast.SetLiteralInt{Values: []int{1, 2, 3}},
	}
	typed, err := s.tc.Check(node)

	s.NoError(err)
	s.True( typed.Type.Equals(types.Boolean{}))
}

// === Quantifiers ===

func (s *CheckerSuite) TestCheck_ForAll() {
	node := &ast.Quantifier{
		Quantifier: "∀",
		Membership: &ast.Membership{
			Left:     &ast.Identifier{Value: "x"},
			Operator: "∈",
			Right:    &ast.SetLiteralInt{Values: []int{1, 2, 3}},
		},
		Predicate: &ast.BooleanLiteral{Value: true},
	}
	typed, err := s.tc.Check(node)

	s.NoError(err)
	s.True( typed.Type.Equals(types.Boolean{}))
}

func (s *CheckerSuite) TestCheck_Exists() {
	node := &ast.Quantifier{
		Quantifier: "∃",
		Membership: &ast.Membership{
			Left:     &ast.Identifier{Value: "x"},
			Operator: "∈",
			Right:    &ast.SetLiteralInt{Values: []int{1, 2, 3}},
		},
		Predicate: &ast.BooleanLiteral{Value: true},
	}
	typed, err := s.tc.Check(node)

	s.NoError(err)
	s.True( typed.Type.Equals(types.Boolean{}))
}

// === Tuples ===

func (s *CheckerSuite) TestCheck_TupleLiteral() {
	node := &ast.TupleLiteral{
		Elements: []ast.Expression{
			ast.NewIntLiteral(1),
			ast.NewIntLiteral(2),
			ast.NewIntLiteral(3),
		},
	}
	typed, err := s.tc.Check(node)

	s.NoError(err)
	tupleType, ok := typed.Type.(types.Tuple)
	s.True(ok)
	s.True(tupleType.Element.Equals(types.Number{}))
}

func (s *CheckerSuite) TestCheck_TupleIndex() {
	node := &ast.TupleIndex{
		Tuple: &ast.TupleLiteral{
			Elements: []ast.Expression{
				ast.NewIntLiteral(1),
				ast.NewIntLiteral(2),
			},
		},
		Index: ast.NewIntLiteral(1),
	}
	typed, err := s.tc.Check(node)

	s.NoError(err)
	s.True( typed.Type.Equals(types.Number{}))
}

// === Records ===

func (s *CheckerSuite) TestCheck_RecordInstance() {
	node := &ast.RecordInstance{
		Bindings: []*ast.FieldBinding{
			{Field: &ast.Identifier{Value: "x"}, Expression: ast.NewIntLiteral(1)},
			{Field: &ast.Identifier{Value: "y"}, Expression: ast.NewIntLiteral(2)},
		},
	}
	typed, err := s.tc.Check(node)

	s.NoError(err)
	recType, ok := typed.Type.(types.Record)
	s.True(ok)
	s.True(recType.Fields["x"].Equals(types.Number{}))
	s.True(recType.Fields["y"].Equals(types.Number{}))
}

func (s *CheckerSuite) TestCheck_FieldAccess() {
	// First bind a record variable
	s.tc.env.BindMono("rec", types.Record{
		Fields: map[string]types.Type{"x": types.Number{}, "y": types.Boolean{}},
	})

	node := &ast.FieldAccess{
		Identifier: &ast.Identifier{Value: "rec"},
		Member:     "x",
	}
	typed, err := s.tc.Check(node)

	s.NoError(err)
	s.True( typed.Type.Equals(types.Number{}))
}

func (s *CheckerSuite) TestCheck_FieldAccess_UndefinedField() {
	s.tc.env.BindMono("rec", types.Record{
		Fields: map[string]types.Type{"x": types.Number{}},
	})

	node := &ast.FieldAccess{
		Identifier: &ast.Identifier{Value: "rec"},
		Member:     "z",
	}
	_, err := s.tc.Check(node)

	s.Error(err)
	s.Contains( err.Error(), "does not have field")
}

// === Control Flow ===

func (s *CheckerSuite) TestCheck_IfElse() {
	node := &ast.IfThenElse{
		Condition: &ast.BooleanLiteral{Value: true},
		Then:      ast.NewIntLiteral(1),
		Else:      ast.NewIntLiteral(2),
	}
	typed, err := s.tc.Check(node)

	s.NoError(err)
	s.True( typed.Type.Equals(types.Number{}))
}

func (s *CheckerSuite) TestCheck_IfElse_BranchMismatch() {
	node := &ast.IfThenElse{
		Condition: &ast.BooleanLiteral{Value: true},
		Then:      ast.NewIntLiteral(1),
		Else:      &ast.StringLiteral{Value: "no"},
	}
	_, err := s.tc.Check(node)

	s.Error(err)
	s.Contains( err.Error(), "same type")
}

func (s *CheckerSuite) TestCheck_Case() {
	node := &ast.CaseExpr{
		Branches: []*ast.CaseBranch{
			{
				Condition: &ast.BooleanLiteral{Value: true},
				Result:    ast.NewIntLiteral(1),
			},
			{
				Condition: &ast.BooleanLiteral{Value: false},
				Result:    ast.NewIntLiteral(2),
			},
		},
	}
	typed, err := s.tc.Check(node)

	s.NoError(err)
	s.True( typed.Type.Equals(types.Number{}))
}

func (s *CheckerSuite) TestCheck_Case_WithOther() {
	node := &ast.CaseExpr{
		Branches: []*ast.CaseBranch{
			{
				Condition: &ast.BooleanLiteral{Value: false},
				Result:    ast.NewIntLiteral(1),
			},
		},
		Other: ast.NewIntLiteral(0),
	}
	typed, err := s.tc.Check(node)

	s.NoError(err)
	s.True( typed.Type.Equals(types.Number{}))
}

// === Builtin Calls ===

func (s *CheckerSuite) TestCheck_BuiltinCall_SeqHead() {
	node := &ast.BuiltinCall{
		Name: "_Seq!Head",
		Args: []ast.Expression{
			&ast.TupleLiteral{
				Elements: []ast.Expression{
					ast.NewIntLiteral(1),
					ast.NewIntLiteral(2),
				},
			},
		},
	}
	typed, err := s.tc.Check(node)

	s.NoError(err)
	// Head of Tuple[Number] should return Number
	s.True(typed.Type.Equals(types.Number{}))
}

func (s *CheckerSuite) TestCheck_BuiltinCall_SeqTail() {
	node := &ast.BuiltinCall{
		Name: "_Seq!Tail",
		Args: []ast.Expression{
			&ast.TupleLiteral{
				Elements: []ast.Expression{
					ast.NewIntLiteral(1),
					ast.NewIntLiteral(2),
				},
			},
		},
	}
	typed, err := s.tc.Check(node)

	s.NoError(err)
	tupleType, ok := typed.Type.(types.Tuple)
	s.True(ok)
	s.True(tupleType.Element.Equals(types.Number{}))
}

func (s *CheckerSuite) TestCheck_BuiltinCall_SeqAppend() {
	node := &ast.BuiltinCall{
		Name: "_Seq!Append",
		Args: []ast.Expression{
			&ast.TupleLiteral{
				Elements: []ast.Expression{
					ast.NewIntLiteral(1),
				},
			},
			ast.NewIntLiteral(2),
		},
	}
	typed, err := s.tc.Check(node)

	s.NoError(err)
	tupleType, ok := typed.Type.(types.Tuple)
	s.True(ok)
	s.True(tupleType.Element.Equals(types.Number{}))
}

func (s *CheckerSuite) TestCheck_BuiltinCall_SeqLen() {
	node := &ast.BuiltinCall{
		Name: "_Seq!Len",
		Args: []ast.Expression{
			&ast.TupleLiteral{
				Elements: []ast.Expression{
					&ast.StringLiteral{Value: "a"},
					&ast.StringLiteral{Value: "b"},
				},
			},
		},
	}
	typed, err := s.tc.Check(node)

	s.NoError(err)
	s.True( typed.Type.Equals(types.Number{}))
}

// Note: SetToBag and Cardinality tests need SetLiteralInt to implement Expression
// which is part of the AST simplification work. Skip for now.

func (s *CheckerSuite) TestCheck_BuiltinCall_UnknownBuiltin() {
	node := &ast.BuiltinCall{
		Name: "_Unknown!Function",
		Args: []ast.Expression{},
	}
	_, err := s.tc.Check(node)

	s.Error(err)
	s.Contains( err.Error(), "unknown builtin")
}

func (s *CheckerSuite) TestCheck_BuiltinCall_WrongArgCount() {
	node := &ast.BuiltinCall{
		Name: "_Seq!Head",
		Args: []ast.Expression{
			&ast.TupleLiteral{Elements: []ast.Expression{ast.NewIntLiteral(1)}},
			ast.NewIntLiteral(2), // Extra argument
		},
	}
	_, err := s.tc.Check(node)

	s.Error(err)
	s.Contains( err.Error(), "expects 1 arguments")
}

func (s *CheckerSuite) TestCheck_BuiltinCall_WrongArgType() {
	node := &ast.BuiltinCall{
		Name: "_Seq!Head",
		Args: []ast.Expression{
			ast.NewIntLiteral(42), // Not a tuple
		},
	}
	_, err := s.tc.Check(node)

	s.Error(err)
	s.Contains( err.Error(), "expected Tuple")
}

// === Polymorphism ===

func (s *CheckerSuite) TestCheck_Polymorphic_Instantiation() {
	// _Seq!Head should work with different element types

	// With Number tuple
	node1 := &ast.BuiltinCall{
		Name: "_Seq!Head",
		Args: []ast.Expression{
			&ast.TupleLiteral{Elements: []ast.Expression{ast.NewIntLiteral(1)}},
		},
	}
	typed1, err := s.tc.Check(node1)
	s.NoError(err)
	s.True( typed1.Type.Equals(types.Number{}))

	// Reset for fresh type variables
	types.ResetTypeVarCounter()
	s.tc = NewTypeChecker()

	// With String tuple
	node2 := &ast.BuiltinCall{
		Name: "_Seq!Head",
		Args: []ast.Expression{
			&ast.TupleLiteral{Elements: []ast.Expression{&ast.StringLiteral{Value: "hello"}}},
		},
	}
	typed2, err := s.tc.Check(node2)
	s.NoError(err)
	s.True( typed2.Type.Equals(types.String{}))
}

// === Type Inference with Variables ===
// Note: TestCheck_TypeInference_FromUsage requires Identifier to implement Real
// which is part of the AST simplification work. Skip for now.

func TestTupleLiteralChildren(t *testing.T) {
	node := &ast.TupleLiteral{
		Elements: []ast.Expression{
			ast.NewIntLiteral(1),
			ast.NewIntLiteral(2),
			ast.NewIntLiteral(3),
		},
	}

	tc := NewTypeChecker()
	typed, err := tc.Check(node)

	assert.NoError(t, err)
	assert.NotNil(t, typed)
	assert.Len(t, typed.Children, 3, "TypedNode should have 3 children")
}

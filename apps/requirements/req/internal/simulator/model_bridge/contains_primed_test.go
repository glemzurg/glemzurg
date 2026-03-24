package model_bridge

import (
	"math/big"
	"testing"

	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/stretchr/testify/suite"
)

type ContainsPrimedSuite struct {
	suite.Suite
}

func TestContainsPrimedSuite(t *testing.T) {
	suite.Run(t, new(ContainsPrimedSuite))
}

func (s *ContainsPrimedSuite) TestNilExpression() {
	s.False(ContainsAnyPrimedME(nil))
}

func (s *ContainsPrimedSuite) TestSimpleLocalVar() {
	expr := &me.LocalVar{Name: "x"}
	s.False(ContainsAnyPrimedME(expr))
}

func (s *ContainsPrimedSuite) TestSimpleNumber() {
	expr := &me.IntLiteral{Value: big.NewInt(42)}
	s.False(ContainsAnyPrimedME(expr))
}

func (s *ContainsPrimedSuite) TestNextStateVariable() {
	expr := &me.NextState{Expr: &me.LocalVar{Name: "x"}}
	s.True(ContainsAnyPrimedME(expr))
}

func (s *ContainsPrimedSuite) TestNextStateFieldAccess() {
	expr := &me.NextState{Expr: &me.FieldAccess{
		Base:  &me.SelfRef{},
		Field: "count",
	}}
	s.True(ContainsAnyPrimedME(expr))
}

func (s *ContainsPrimedSuite) TestBinaryArithNoPrime() {
	expr := &me.BinaryArith{
		Left:  &me.LocalVar{Name: "x"},
		Op:    me.ArithAdd,
		Right: &me.LocalVar{Name: "y"},
	}
	s.False(ContainsAnyPrimedME(expr))
}

func (s *ContainsPrimedSuite) TestBinaryArithWithPrime() {
	expr := &me.BinaryArith{
		Left:  &me.NextState{Expr: &me.LocalVar{Name: "x"}},
		Op:    me.ArithAdd,
		Right: &me.LocalVar{Name: "y"},
	}
	s.True(ContainsAnyPrimedME(expr))
}

func (s *ContainsPrimedSuite) TestComparisonWithPrime() {
	expr := &me.Compare{
		Left:  &me.NextState{Expr: &me.LocalVar{Name: "x"}},
		Op:    me.CompareGt,
		Right: &me.IntLiteral{Value: big.NewInt(0)},
	}
	s.True(ContainsAnyPrimedME(expr))
}

func (s *ContainsPrimedSuite) TestComparisonNoPrime() {
	expr := &me.Compare{
		Left:  &me.LocalVar{Name: "x"},
		Op:    me.CompareGt,
		Right: &me.IntLiteral{Value: big.NewInt(0)},
	}
	s.False(ContainsAnyPrimedME(expr))
}

func (s *ContainsPrimedSuite) TestNestedDeepPrime() {
	expr := &me.Compare{
		Left: &me.BinaryArith{
			Left:  &me.LocalVar{Name: "a"},
			Op:    me.ArithAdd,
			Right: &me.LocalVar{Name: "b"},
		},
		Op: me.CompareGt,
		Right: &me.BinaryArith{
			Left:  &me.NextState{Expr: &me.LocalVar{Name: "c"}},
			Op:    me.ArithAdd,
			Right: &me.LocalVar{Name: "d"},
		},
	}
	s.True(ContainsAnyPrimedME(expr))
}

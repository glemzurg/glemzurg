package convert

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/suite"

	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/ast"
)

type LowerTestSuite struct {
	suite.Suite
	ctx *LowerContext
}

func TestLowerSuite(t *testing.T) {
	suite.Run(t, new(LowerTestSuite))
}

func (s *LowerTestSuite) SetupTest() {
	domainKey, _ := identity.NewDomainKey("d")
	subKey, _ := identity.NewSubdomainKey(domainKey, "s")
	classKey, _ := identity.NewClassKey(subKey, "c")
	attrKey, _ := identity.NewAttributeKey(classKey, "balance")
	actionKey, _ := identity.NewActionKey(classKey, "Deposit")
	queryKey, _ := identity.NewQueryKey(classKey, "GetBalance")
	globalKey, _ := identity.NewGlobalFunctionKey("_Helper")
	namedSetKey, _ := identity.NewNamedSetKey("valid_statuses")

	// Cross-class action.
	subKey2, _ := identity.NewSubdomainKey(domainKey, "s2")
	classKey2, _ := identity.NewClassKey(subKey2, "c2")
	crossActionKey, _ := identity.NewActionKey(classKey2, "OtherAction")

	s.ctx = &LowerContext{
		ClassKey:        classKey,
		AttributeNames:  map[string]identity.Key{"balance": attrKey},
		ActionNames:     map[string]identity.Key{"Deposit": actionKey},
		QueryNames:      map[string]identity.Key{"GetBalance": queryKey},
		GlobalFunctions: map[string]identity.Key{"_Helper": globalKey},
		NamedSets:       map[string]identity.Key{"valid_statuses": namedSetKey},
		AllActions:      map[string]identity.Key{"s2!c2!OtherAction": crossActionKey},
		Parameters:      map[string]bool{"amount": true},
	}
}

// --- Literal tests ---

func (s *LowerTestSuite) TestLowerBooleanLiteral() {
	result, err := Lower(&ast.BooleanLiteral{Value: true}, s.ctx)
	s.Require().NoError(err)
	boolLit, ok := result.(*me.BoolLiteral)
	s.True(ok)
	s.True(boolLit.Value)
}

func (s *LowerTestSuite) TestLowerNumberLiteralDecimal() {
	result, err := Lower(ast.NewNumberLiteral("42"), s.ctx)
	s.Require().NoError(err)
	intLit, ok := result.(*me.IntLiteral)
	s.True(ok)
	s.Equal(0, intLit.Value.Cmp(big.NewInt(42)))
}

func (s *LowerTestSuite) TestLowerNumberLiteralHex() {
	result, err := Lower(ast.NewHexNumberLiteral("\\h", "FF"), s.ctx)
	s.Require().NoError(err)
	intLit, ok := result.(*me.IntLiteral)
	s.True(ok)
	s.Equal(0, intLit.Value.Cmp(big.NewInt(255)))
}

func (s *LowerTestSuite) TestLowerNumberLiteralBinary() {
	result, err := Lower(ast.NewBinaryNumberLiteral("\\b", "1010"), s.ctx)
	s.Require().NoError(err)
	intLit, ok := result.(*me.IntLiteral)
	s.True(ok)
	s.Equal(0, intLit.Value.Cmp(big.NewInt(10)))
}

func (s *LowerTestSuite) TestLowerNumberLiteralOctal() {
	result, err := Lower(ast.NewOctalNumberLiteral("\\o", "17"), s.ctx)
	s.Require().NoError(err)
	intLit, ok := result.(*me.IntLiteral)
	s.True(ok)
	s.Equal(0, intLit.Value.Cmp(big.NewInt(15)))
}

func (s *LowerTestSuite) TestLowerNumberLiteralDecimalPointError() {
	n := ast.NewDecimalNumberLiteral("3", "14")
	_, err := Lower(n, s.ctx)
	s.Require().Error(err)
	s.Contains(err.Error(), "Fraction")
}

func (s *LowerTestSuite) TestLowerStringLiteral() {
	result, err := Lower(&ast.StringLiteral{Value: "hello"}, s.ctx)
	s.Require().NoError(err)
	strLit, ok := result.(*me.StringLiteral)
	s.True(ok)
	s.Equal("hello", strLit.Value)
}

func (s *LowerTestSuite) TestLowerFractionLiteralIntegers() {
	// 3/4 with integer numerator and denominator → RationalLiteral.
	frac := ast.NewFraction(ast.NewNumberLiteral("3"), ast.NewNumberLiteral("4"))
	result, err := Lower(frac, s.ctx)
	s.Require().NoError(err)
	ratLit, ok := result.(*me.RationalLiteral)
	s.True(ok)
	expected := big.NewRat(3, 4)
	s.Equal(0, ratLit.Value.Cmp(expected))
}

func (s *LowerTestSuite) TestLowerFractionNonLiteral() {
	// balance / 4 → BinaryArith{Op: div, ...}
	frac := ast.NewFraction(&ast.Identifier{Value: "balance"}, ast.NewNumberLiteral("4"))
	result, err := Lower(frac, s.ctx)
	s.Require().NoError(err)
	arith, ok := result.(*me.BinaryArith)
	s.True(ok)
	s.Equal(me.ArithDiv, arith.Op)
}

// --- Collection tests ---

func (s *LowerTestSuite) TestLowerSetLiteral() {
	result, err := Lower(&ast.SetLiteral{Elements: []ast.Expression{
		ast.NewNumberLiteral("1"),
		ast.NewNumberLiteral("2"),
	}}, s.ctx)
	s.Require().NoError(err)
	setLit, ok := result.(*me.SetLiteral)
	s.True(ok)
	s.Len(setLit.Elements, 2)
}

func (s *LowerTestSuite) TestLowerSetLiteralEmpty() {
	result, err := Lower(&ast.SetLiteral{}, s.ctx)
	s.Require().NoError(err)
	setLit, ok := result.(*me.SetLiteral)
	s.True(ok)
	s.Empty(setLit.Elements)
}

func (s *LowerTestSuite) TestLowerSetLiteralEnum() {
	result, err := Lower(&ast.SetLiteralEnum{Values: []string{"a", "b"}}, s.ctx)
	s.Require().NoError(err)
	setLit, ok := result.(*me.SetLiteral)
	s.True(ok)
	s.Len(setLit.Elements, 2)
	s.Equal("a", setLit.Elements[0].(*me.StringLiteral).Value)
}

func (s *LowerTestSuite) TestLowerSetLiteralInt() {
	result, err := Lower(&ast.SetLiteralInt{Values: []int{1, 2, 3}}, s.ctx)
	s.Require().NoError(err)
	setLit, ok := result.(*me.SetLiteral)
	s.True(ok)
	s.Len(setLit.Elements, 3)
	s.Equal(0, setLit.Elements[0].(*me.IntLiteral).Value.Cmp(big.NewInt(1)))
}

func (s *LowerTestSuite) TestLowerSetConstant() {
	tests := []struct {
		astValue string
		kind     me.SetConstantKind
	}{
		{ast.SetConstantNat, me.SetConstantNat},
		{ast.SetConstantInt, me.SetConstantInt},
		{ast.SetConstantReal, me.SetConstantReal},
		{ast.SetConstantBoolean, me.SetConstantBoolean},
	}
	for _, tt := range tests {
		s.Run(tt.astValue, func() {
			result, err := Lower(&ast.SetConstant{Value: tt.astValue}, s.ctx)
			s.Require().NoError(err)
			sc, ok := result.(*me.SetConstant)
			s.True(ok)
			s.Equal(tt.kind, sc.Kind)
		})
	}
}

func (s *LowerTestSuite) TestLowerSetRange() {
	result, err := Lower(&ast.SetRange{Start: 1, End: 10}, s.ctx)
	s.Require().NoError(err)
	sr, ok := result.(*me.SetRange)
	s.True(ok)
	s.Equal(0, sr.Start.(*me.IntLiteral).Value.Cmp(big.NewInt(1)))
	s.Equal(0, sr.End.(*me.IntLiteral).Value.Cmp(big.NewInt(10)))
}

func (s *LowerTestSuite) TestLowerSetRangeExpr() {
	result, err := Lower(&ast.SetRangeExpr{
		Start: ast.NewNumberLiteral("1"),
		End:   ast.NewNumberLiteral("10"),
	}, s.ctx)
	s.Require().NoError(err)
	sr, ok := result.(*me.SetRange)
	s.True(ok)
	s.NotNil(sr.Start)
	s.NotNil(sr.End)
}

func (s *LowerTestSuite) TestLowerTupleLiteral() {
	result, err := Lower(&ast.TupleLiteral{Elements: []ast.Expression{
		ast.NewNumberLiteral("1"),
		&ast.BooleanLiteral{Value: true},
	}}, s.ctx)
	s.Require().NoError(err)
	tup, ok := result.(*me.TupleLiteral)
	s.True(ok)
	s.Len(tup.Elements, 2)
}

func (s *LowerTestSuite) TestLowerRecordInstance() {
	result, err := Lower(&ast.RecordInstance{Bindings: []*ast.FieldBinding{
		{Field: &ast.Identifier{Value: "x"}, Expression: ast.NewNumberLiteral("1")},
		{Field: &ast.Identifier{Value: "y"}, Expression: ast.NewNumberLiteral("2")},
	}}, s.ctx)
	s.Require().NoError(err)
	rec, ok := result.(*me.RecordLiteral)
	s.True(ok)
	s.Len(rec.Fields, 2)
	s.Equal("x", rec.Fields[0].Name)
}

// --- Reference tests ---

func (s *LowerTestSuite) TestLowerIdentifierSelf() {
	result, err := Lower(&ast.Identifier{Value: "self"}, s.ctx)
	s.Require().NoError(err)
	_, ok := result.(*me.SelfRef)
	s.True(ok)
}

func (s *LowerTestSuite) TestLowerIdentifierAttribute() {
	result, err := Lower(&ast.Identifier{Value: "balance"}, s.ctx)
	s.Require().NoError(err)
	attrRef, ok := result.(*me.AttributeRef)
	s.True(ok)
	s.Contains(attrRef.AttributeKey.String(), "balance")
}

func (s *LowerTestSuite) TestLowerIdentifierParameter() {
	result, err := Lower(&ast.Identifier{Value: "amount"}, s.ctx)
	s.Require().NoError(err)
	localVar, ok := result.(*me.LocalVar)
	s.True(ok)
	s.Equal("amount", localVar.Name)
}

func (s *LowerTestSuite) TestLowerIdentifierNamedSet() {
	result, err := Lower(&ast.Identifier{Value: "valid_statuses"}, s.ctx)
	s.Require().NoError(err)
	ref, ok := result.(*me.NamedSetRef)
	s.True(ok)
	s.Contains(ref.SetKey.String(), "valid_statuses")
}

func (s *LowerTestSuite) TestLowerIdentifierUnresolved() {
	_, err := Lower(&ast.Identifier{Value: "unknown"}, s.ctx)
	s.Require().Error(err)
	s.Contains(err.Error(), "unresolved identifier")
}

func (s *LowerTestSuite) TestLowerExistingValueInExcept() {
	ctx := *s.ctx
	ctx.exceptField = "count"
	result, err := Lower(&ast.ExistingValue{}, &ctx)
	s.Require().NoError(err)
	pf, ok := result.(*me.PriorFieldValue)
	s.True(ok)
	s.Equal("count", pf.Field)
}

func (s *LowerTestSuite) TestLowerExistingValueOutsideExcept() {
	_, err := Lower(&ast.ExistingValue{}, s.ctx)
	s.Require().Error(err)
	s.Contains(err.Error(), "EXCEPT")
}

// --- Unary operator tests ---

func (s *LowerTestSuite) TestLowerUnaryNegation() {
	result, err := Lower(&ast.UnaryNegation{Operator: "-", Right: ast.NewNumberLiteral("5")}, s.ctx)
	s.Require().NoError(err)
	neg, ok := result.(*me.Negate)
	s.True(ok)
	s.NotNil(neg.Expr)
}

func (s *LowerTestSuite) TestLowerUnaryLogic() {
	result, err := Lower(&ast.UnaryLogic{Operator: "¬", Right: &ast.BooleanLiteral{Value: true}}, s.ctx)
	s.Require().NoError(err)
	not, ok := result.(*me.Not)
	s.True(ok)
	s.NotNil(not.Expr)
}

func (s *LowerTestSuite) TestLowerPrimed() {
	result, err := Lower(&ast.Primed{Base: &ast.Identifier{Value: "balance"}}, s.ctx)
	s.Require().NoError(err)
	ns, ok := result.(*me.NextState)
	s.True(ok)
	_, isAttr := ns.Expr.(*me.AttributeRef)
	s.True(isAttr)
}

// --- Binary operator tests ---

func (s *LowerTestSuite) TestLowerBinaryArithmetic() {
	tests := []struct {
		op       string
		expected me.ArithOp
	}{
		{"+", me.ArithAdd},
		{"-", me.ArithSub},
		{"*", me.ArithMul},
		{"÷", me.ArithDiv},
		{"^", me.ArithPow},
		{"%", me.ArithMod},
	}
	for _, tt := range tests {
		s.Run(tt.op, func() {
			result, err := Lower(&ast.BinaryArithmetic{
				Left: ast.NewNumberLiteral("1"), Operator: tt.op, Right: ast.NewNumberLiteral("2"),
			}, s.ctx)
			s.Require().NoError(err)
			arith, ok := result.(*me.BinaryArith)
			s.True(ok)
			s.Equal(tt.expected, arith.Op)
		})
	}
}

func (s *LowerTestSuite) TestLowerBinaryLogic() {
	tests := []struct {
		op       string
		expected me.LogicOp
	}{
		{"∧", me.LogicAnd},
		{"∨", me.LogicOr},
		{"⇒", me.LogicImplies},
		{"≡", me.LogicEquiv},
	}
	for _, tt := range tests {
		s.Run(tt.op, func() {
			result, err := Lower(&ast.BinaryLogic{
				Operator: tt.op, Left: &ast.BooleanLiteral{Value: true}, Right: &ast.BooleanLiteral{Value: false},
			}, s.ctx)
			s.Require().NoError(err)
			logic, ok := result.(*me.BinaryLogic)
			s.True(ok)
			s.Equal(tt.expected, logic.Op)
		})
	}
}

func (s *LowerTestSuite) TestLowerBinaryEquality() {
	tests := []struct {
		op       string
		expected me.CompareOp
	}{
		{"=", me.CompareEq},
		{"≠", me.CompareNeq},
	}
	for _, tt := range tests {
		s.Run(tt.op, func() {
			result, err := Lower(&ast.BinaryEquality{
				Operator: tt.op, Left: ast.NewNumberLiteral("1"), Right: ast.NewNumberLiteral("2"),
			}, s.ctx)
			s.Require().NoError(err)
			cmp, ok := result.(*me.Compare)
			s.True(ok)
			s.Equal(tt.expected, cmp.Op)
		})
	}
}

func (s *LowerTestSuite) TestLowerBinaryComparison() {
	tests := []struct {
		op       string
		expected me.CompareOp
	}{
		{"<", me.CompareLt},
		{">", me.CompareGt},
		{"≤", me.CompareLte},
		{"≥", me.CompareGte},
	}
	for _, tt := range tests {
		s.Run(tt.op, func() {
			result, err := Lower(&ast.BinaryComparison{
				Operator: tt.op, Left: ast.NewNumberLiteral("1"), Right: ast.NewNumberLiteral("2"),
			}, s.ctx)
			s.Require().NoError(err)
			cmp, ok := result.(*me.Compare)
			s.True(ok)
			s.Equal(tt.expected, cmp.Op)
		})
	}
}

func (s *LowerTestSuite) TestLowerBinarySetOperation() {
	tests := []struct {
		op       string
		expected me.SetOpKind
	}{
		{"∪", me.SetUnion},
		{"∩", me.SetIntersect},
		{"\\", me.SetDifference},
	}
	for _, tt := range tests {
		s.Run(tt.op, func() {
			result, err := Lower(&ast.BinarySetOperation{
				Operator: tt.op, Left: &ast.SetLiteral{}, Right: &ast.SetLiteral{},
			}, s.ctx)
			s.Require().NoError(err)
			setOp, ok := result.(*me.SetOp)
			s.True(ok)
			s.Equal(tt.expected, setOp.Op)
		})
	}
}

func (s *LowerTestSuite) TestLowerBinarySetComparison() {
	tests := []struct {
		op       string
		expected me.SetCompareOp
	}{
		{"⊆", me.SetCompareSubsetEq},
		{"⊂", me.SetCompareSubset},
		{"⊇", me.SetCompareSupersetEq},
		{"⊃", me.SetCompareSuperset},
	}
	for _, tt := range tests {
		s.Run(tt.op, func() {
			result, err := Lower(&ast.BinarySetComparison{
				Operator: tt.op, Left: &ast.SetLiteral{}, Right: &ast.SetLiteral{},
			}, s.ctx)
			s.Require().NoError(err)
			sc, ok := result.(*me.SetCompare)
			s.True(ok)
			s.Equal(tt.expected, sc.Op)
		})
	}
}

func (s *LowerTestSuite) TestLowerBinarySetComparisonEquality() {
	// Set = and ≠ lower to Compare.
	result, err := Lower(&ast.BinarySetComparison{
		Operator: "=", Left: &ast.SetLiteral{}, Right: &ast.SetLiteral{},
	}, s.ctx)
	s.Require().NoError(err)
	cmp, ok := result.(*me.Compare)
	s.True(ok)
	s.Equal(me.CompareEq, cmp.Op)
}

func (s *LowerTestSuite) TestLowerMembership() {
	result, err := Lower(&ast.Membership{
		Operator: "∈", Left: ast.NewNumberLiteral("1"), Right: &ast.SetLiteral{},
	}, s.ctx)
	s.Require().NoError(err)
	mem, ok := result.(*me.Membership)
	s.True(ok)
	s.False(mem.Negated)
}

func (s *LowerTestSuite) TestLowerMembershipNegated() {
	result, err := Lower(&ast.Membership{
		Operator: "∉", Left: ast.NewNumberLiteral("1"), Right: &ast.SetLiteral{},
	}, s.ctx)
	s.Require().NoError(err)
	mem, ok := result.(*me.Membership)
	s.True(ok)
	s.True(mem.Negated)
}

func (s *LowerTestSuite) TestLowerBagOperation() {
	tests := []struct {
		op       string
		expected me.BagOpKind
	}{
		{"⊕", me.BagSum},
		{"⊖", me.BagDifference},
	}
	for _, tt := range tests {
		s.Run(tt.op, func() {
			result, err := Lower(&ast.BinaryBagOperation{
				Operator: tt.op, Left: &ast.SetLiteral{}, Right: &ast.SetLiteral{},
			}, s.ctx)
			s.Require().NoError(err)
			bo, ok := result.(*me.BagOp)
			s.True(ok)
			s.Equal(tt.expected, bo.Op)
		})
	}
}

func (s *LowerTestSuite) TestLowerBagComparison() {
	tests := []struct {
		op       string
		expected me.BagCompareOp
	}{
		{"⊏", me.BagCompareProperSubBag},
		{"⊑", me.BagCompareSubBag},
		{"⊐", me.BagCompareProperSupBag},
		{"⊒", me.BagCompareSupBag},
	}
	for _, tt := range tests {
		s.Run(tt.op, func() {
			result, err := Lower(&ast.BinaryBagComparison{
				Operator: tt.op, Left: &ast.SetLiteral{}, Right: &ast.SetLiteral{},
			}, s.ctx)
			s.Require().NoError(err)
			bc, ok := result.(*me.BagCompare)
			s.True(ok)
			s.Equal(tt.expected, bc.Op)
		})
	}
}

// --- Concatenation tests ---

func (s *LowerTestSuite) TestLowerStringConcat() {
	result, err := Lower(&ast.StringConcat{
		Operator: "∘",
		Operands: []ast.Expression{&ast.StringLiteral{Value: "a"}, &ast.StringLiteral{Value: "b"}},
	}, s.ctx)
	s.Require().NoError(err)
	sc, ok := result.(*me.StringConcat)
	s.True(ok)
	s.Len(sc.Operands, 2)
}

func (s *LowerTestSuite) TestLowerTupleConcat() {
	result, err := Lower(&ast.TupleConcat{
		Operator: "∘",
		Operands: []ast.Expression{
			&ast.TupleLiteral{Elements: []ast.Expression{ast.NewNumberLiteral("1")}},
			&ast.TupleLiteral{Elements: []ast.Expression{ast.NewNumberLiteral("2")}},
		},
	}, s.ctx)
	s.Require().NoError(err)
	tc, ok := result.(*me.TupleConcat)
	s.True(ok)
	s.Len(tc.Operands, 2)
}

// --- Indexing tests ---

func (s *LowerTestSuite) TestLowerTupleIndex() {
	result, err := Lower(&ast.TupleIndex{
		Tuple: &ast.TupleLiteral{Elements: []ast.Expression{ast.NewNumberLiteral("1")}},
		Index: ast.NewNumberLiteral("1"),
	}, s.ctx)
	s.Require().NoError(err)
	ti, ok := result.(*me.TupleIndex)
	s.True(ok)
	s.NotNil(ti.Tuple)
	s.NotNil(ti.Index)
}

func (s *LowerTestSuite) TestLowerStringIndex() {
	result, err := Lower(&ast.StringIndex{
		Str:   &ast.StringLiteral{Value: "abc"},
		Index: ast.NewNumberLiteral("1"),
	}, s.ctx)
	s.Require().NoError(err)
	si, ok := result.(*me.StringIndex)
	s.True(ok)
	s.NotNil(si.Str)
	s.NotNil(si.Index)
}

func (s *LowerTestSuite) TestLowerFieldAccess() {
	result, err := Lower(&ast.FieldAccess{
		Base:   &ast.Identifier{Value: "self"},
		Member: "name",
	}, s.ctx)
	s.Require().NoError(err)
	fa, ok := result.(*me.FieldAccess)
	s.True(ok)
	s.Equal("name", fa.Field)
	_, isSelf := fa.Base.(*me.SelfRef)
	s.True(isSelf)
}

// --- Record alteration tests ---

func (s *LowerTestSuite) TestLowerRecordAltered() {
	result, err := Lower(&ast.RecordAltered{
		Base: &ast.Identifier{Value: "self"},
		Alterations: []*ast.FieldAlteration{
			{
				Field:      &ast.FieldAccess{Member: "count"},
				Expression: ast.NewNumberLiteral("5"),
			},
		},
	}, s.ctx)
	s.Require().NoError(err)
	ru, ok := result.(*me.RecordUpdate)
	s.True(ok)
	s.Len(ru.Alterations, 1)
	s.Equal("count", ru.Alterations[0].Field)
}

func (s *LowerTestSuite) TestLowerRecordAlteredWithExistingValue() {
	// [self EXCEPT !.count = @ + 1]
	result, err := Lower(&ast.RecordAltered{
		Base: &ast.Identifier{Value: "self"},
		Alterations: []*ast.FieldAlteration{
			{
				Field: &ast.FieldAccess{Member: "count"},
				Expression: &ast.BinaryArithmetic{
					Left:     &ast.ExistingValue{},
					Operator: "+",
					Right:    ast.NewNumberLiteral("1"),
				},
			},
		},
	}, s.ctx)
	s.Require().NoError(err)
	ru, ok := result.(*me.RecordUpdate)
	s.True(ok)

	// The value should be BinaryArith with PriorFieldValue on the left.
	arith, ok := ru.Alterations[0].Value.(*me.BinaryArith)
	s.True(ok)
	pfv, ok := arith.Left.(*me.PriorFieldValue)
	s.True(ok)
	s.Equal("count", pfv.Field)
}

// --- Control flow tests ---

func (s *LowerTestSuite) TestLowerIfThenElse() {
	result, err := Lower(&ast.IfThenElse{
		Condition: &ast.BooleanLiteral{Value: true},
		Then:      ast.NewNumberLiteral("1"),
		Else:      ast.NewNumberLiteral("2"),
	}, s.ctx)
	s.Require().NoError(err)
	ite, ok := result.(*me.IfThenElse)
	s.True(ok)
	s.NotNil(ite.Condition)
	s.NotNil(ite.Then)
	s.NotNil(ite.Else)
}

func (s *LowerTestSuite) TestLowerCaseExpr() {
	result, err := Lower(&ast.CaseExpr{
		Branches: []*ast.CaseBranch{
			{Condition: &ast.BooleanLiteral{Value: true}, Result: ast.NewNumberLiteral("1")},
		},
		Other: ast.NewNumberLiteral("0"),
	}, s.ctx)
	s.Require().NoError(err)
	c, ok := result.(*me.Case)
	s.True(ok)
	s.Len(c.Branches, 1)
	s.NotNil(c.Otherwise)
}

func (s *LowerTestSuite) TestLowerCaseExprNoOtherwise() {
	result, err := Lower(&ast.CaseExpr{
		Branches: []*ast.CaseBranch{
			{Condition: &ast.BooleanLiteral{Value: true}, Result: ast.NewNumberLiteral("1")},
		},
	}, s.ctx)
	s.Require().NoError(err)
	c, ok := result.(*me.Case)
	s.True(ok)
	s.Nil(c.Otherwise)
}

// --- Quantifier tests ---

func (s *LowerTestSuite) TestLowerQuantifierForall() {
	result, err := Lower(&ast.Quantifier{
		Quantifier: "∀",
		Membership: &ast.Membership{
			Operator: "∈",
			Left:     &ast.Identifier{Value: "x"},
			Right:    &ast.SetConstant{Value: "Nat"},
		},
		Predicate: &ast.BinaryComparison{
			Operator: ">",
			Left:     &ast.Identifier{Value: "x"},
			Right:    ast.NewNumberLiteral("0"),
		},
	}, s.ctx)
	s.Require().NoError(err)
	q, ok := result.(*me.Quantifier)
	s.True(ok)
	s.Equal(me.QuantifierForall, q.Kind)
	s.Equal("x", q.Variable)
	// x should resolve as LocalVar inside the predicate.
	cmp, ok := q.Predicate.(*me.Compare)
	s.True(ok)
	lv, ok := cmp.Left.(*me.LocalVar)
	s.True(ok)
	s.Equal("x", lv.Name)
}

func (s *LowerTestSuite) TestLowerQuantifierExists() {
	result, err := Lower(&ast.Quantifier{
		Quantifier: "∃",
		Membership: &ast.Membership{
			Operator: "∈",
			Left:     &ast.Identifier{Value: "x"},
			Right:    &ast.SetConstant{Value: "Nat"},
		},
		Predicate: &ast.BooleanLiteral{Value: true},
	}, s.ctx)
	s.Require().NoError(err)
	q, ok := result.(*me.Quantifier)
	s.True(ok)
	s.Equal(me.QuantifierExists, q.Kind)
}

func (s *LowerTestSuite) TestLowerSetFilter() {
	result, err := Lower(&ast.SetFilter{
		Membership: &ast.Membership{
			Operator: "∈",
			Left:     &ast.Identifier{Value: "x"},
			Right:    &ast.SetConstant{Value: "Nat"},
		},
		Predicate: &ast.BinaryComparison{
			Operator: ">",
			Left:     &ast.Identifier{Value: "x"},
			Right:    ast.NewNumberLiteral("0"),
		},
	}, s.ctx)
	s.Require().NoError(err)
	sf, ok := result.(*me.SetFilter)
	s.True(ok)
	s.Equal("x", sf.Variable)
}

// --- Call tests ---

func (s *LowerTestSuite) TestLowerBuiltinCallAst() {
	// AST BuiltinCall node: _Seq!Len(args...)
	result, err := Lower(&ast.BuiltinCall{
		Name: "_Seq!Len",
		Args: []ast.Expression{&ast.TupleLiteral{Elements: []ast.Expression{ast.NewNumberLiteral("1")}}},
	}, s.ctx)
	s.Require().NoError(err)
	bc, ok := result.(*me.BuiltinCall)
	s.True(ok)
	s.Equal("_Seq", bc.Module)
	s.Equal("Len", bc.Function)
	s.Len(bc.Args, 1)
}

func (s *LowerTestSuite) TestLowerFunctionCallBuiltin() {
	// FunctionCall with scope path: _Seq!Head(args...)
	result, err := Lower(&ast.FunctionCall{
		ScopePath: []*ast.Identifier{{Value: "_Seq"}},
		Name:      &ast.Identifier{Value: "Head"},
		Args:      []ast.Expression{&ast.TupleLiteral{Elements: []ast.Expression{ast.NewNumberLiteral("1")}}},
	}, s.ctx)
	s.Require().NoError(err)
	bc, ok := result.(*me.BuiltinCall)
	s.True(ok)
	s.Equal("_Seq", bc.Module)
	s.Equal("Head", bc.Function)
}

func (s *LowerTestSuite) TestLowerFunctionCallGlobal() {
	// FunctionCall with no scope path, underscore: _Helper(arg)
	result, err := Lower(&ast.FunctionCall{
		Name: &ast.Identifier{Value: "_Helper"},
		Args: []ast.Expression{ast.NewNumberLiteral("1")},
	}, s.ctx)
	s.Require().NoError(err)
	gc, ok := result.(*me.GlobalCall)
	s.True(ok)
	s.Contains(gc.FunctionKey.String(), "gfunc")
}

func (s *LowerTestSuite) TestLowerFunctionCallSameClassAction() {
	result, err := Lower(&ast.FunctionCall{
		Name: &ast.Identifier{Value: "Deposit"},
		Args: []ast.Expression{ast.NewNumberLiteral("100")},
	}, s.ctx)
	s.Require().NoError(err)
	ac, ok := result.(*me.ActionCall)
	s.True(ok)
	s.Contains(ac.ActionKey.String(), "action")
}

func (s *LowerTestSuite) TestLowerFunctionCallSameClassQuery() {
	result, err := Lower(&ast.FunctionCall{
		Name: &ast.Identifier{Value: "GetBalance"},
		Args: nil,
	}, s.ctx)
	s.Require().NoError(err)
	ac, ok := result.(*me.ActionCall)
	s.True(ok)
	s.Contains(ac.ActionKey.String(), "query")
}

func (s *LowerTestSuite) TestLowerFunctionCallCrossClass() {
	result, err := Lower(&ast.FunctionCall{
		ScopePath: []*ast.Identifier{{Value: "s2"}, {Value: "c2"}},
		Name:      &ast.Identifier{Value: "OtherAction"},
		Args:      nil,
	}, s.ctx)
	s.Require().NoError(err)
	ac, ok := result.(*me.ActionCall)
	s.True(ok)
	s.Contains(ac.ActionKey.String(), "otheraction")
}

func (s *LowerTestSuite) TestLowerScopedCallModelScope() {
	result, err := Lower(&ast.ScopedCall{
		ModelScope:   true,
		FunctionName: &ast.Identifier{Value: "Helper"},
		Parameter:    &ast.RecordInstance{Bindings: []*ast.FieldBinding{{Field: &ast.Identifier{Value: "x"}, Expression: ast.NewNumberLiteral("1")}}},
	}, s.ctx)
	s.Require().NoError(err)
	gc, ok := result.(*me.GlobalCall)
	s.True(ok)
	s.Contains(gc.FunctionKey.String(), "gfunc")
}

// --- Parenthesized passthrough ---

func (s *LowerTestSuite) TestLowerParenthesized() {
	result, err := Lower(&ast.Parenthesized{Inner: ast.NewNumberLiteral("42")}, s.ctx)
	s.Require().NoError(err)
	intLit, ok := result.(*me.IntLiteral)
	s.True(ok)
	s.Equal(0, intLit.Value.Cmp(big.NewInt(42)))
}

// --- Error tests ---

func (s *LowerTestSuite) TestLowerNilExpression() {
	_, err := Lower(nil, s.ctx)
	s.Require().Error(err)
	s.Contains(err.Error(), "nil")
}

func (s *LowerTestSuite) TestLowerRecordTypeExprError() {
	_, err := Lower(&ast.RecordTypeExpr{}, s.ctx)
	s.Require().Error(err)
	s.Contains(err.Error(), "type expression")
}

// --- Validation of lowered output ---

func (s *LowerTestSuite) TestLoweredExpressionValidates() {
	// A complex expression that exercises multiple node types.
	expr := &ast.BinaryLogic{
		Operator: "∧",
		Left: &ast.BinaryComparison{
			Operator: ">",
			Left:     &ast.Identifier{Value: "balance"},
			Right:    ast.NewNumberLiteral("0"),
		},
		Right: &ast.Quantifier{
			Quantifier: "∀",
			Membership: &ast.Membership{
				Operator: "∈",
				Left:     &ast.Identifier{Value: "x"},
				Right:    &ast.SetConstant{Value: "Nat"},
			},
			Predicate: &ast.BinaryComparison{
				Operator: "≥",
				Left:     &ast.Identifier{Value: "x"},
				Right:    ast.NewNumberLiteral("0"),
			},
		},
	}
	result, err := Lower(expr, s.ctx)
	s.Require().NoError(err)
	s.NoError(result.Validate())
}

// --- Test local variable scoping ---

func (s *LowerTestSuite) TestLocalVarShadowsAttribute() {
	// A quantifier-bound variable with the same name as an attribute should resolve as LocalVar.
	result, err := Lower(&ast.Quantifier{
		Quantifier: "∀",
		Membership: &ast.Membership{
			Operator: "∈",
			Left:     &ast.Identifier{Value: "balance"},
			Right:    &ast.SetConstant{Value: "Nat"},
		},
		Predicate: &ast.BinaryComparison{
			Operator: ">",
			Left:     &ast.Identifier{Value: "balance"},
			Right:    ast.NewNumberLiteral("0"),
		},
	}, s.ctx)
	s.Require().NoError(err)
	q, ok := result.(*me.Quantifier)
	s.True(ok)
	// Inside predicate, "balance" should be LocalVar, not AttributeRef.
	cmp := q.Predicate.(*me.Compare)
	lv, ok := cmp.Left.(*me.LocalVar)
	s.True(ok)
	s.Equal("balance", lv.Name)
}

func (s *LowerTestSuite) TestLowerBigInteger() {
	// Very large integer that exceeds int64.
	n := &ast.NumberLiteral{
		Base:        ast.BaseDecimal,
		IntegerPart: "999999999999999999999999999999",
	}
	result, err := Lower(n, s.ctx)
	s.Require().NoError(err)
	intLit, ok := result.(*me.IntLiteral)
	s.True(ok)
	expected, _ := new(big.Int).SetString("999999999999999999999999999999", 10)
	s.Equal(0, intLit.Value.Cmp(expected))
}

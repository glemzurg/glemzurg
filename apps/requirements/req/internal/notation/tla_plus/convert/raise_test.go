package convert

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/parser"
	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_expression"
)

// RaiseTestSuite tests the Raise() function, primarily via round-trip:
//
//	Expression → Raise() → AST → Print() → string → Parse() → AST → Lower() → Expression → compare
type RaiseTestSuite struct {
	suite.Suite
	raiseCtx *RaiseContext
	lowerCtx *LowerContext
}

func TestRaiseSuite(t *testing.T) {
	suite.Run(t, new(RaiseTestSuite))
}

func (s *RaiseTestSuite) SetupTest() {
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

	// Lower context (for re-lowering after parse).
	s.lowerCtx = &LowerContext{
		ClassKey:        classKey,
		AttributeNames:  map[string]identity.Key{"balance": attrKey},
		ActionNames:     map[string]identity.Key{"Deposit": actionKey},
		QueryNames:      map[string]identity.Key{"GetBalance": queryKey},
		GlobalFunctions: map[string]identity.Key{"_Helper": globalKey},
		NamedSets:       map[string]identity.Key{"valid_statuses": namedSetKey},
		AllActions:      map[string]identity.Key{"s2!c2!OtherAction": crossActionKey},
		Parameters:      map[string]bool{"amount": true},
	}

	// Raise context (inverse mappings).
	s.raiseCtx = &RaiseContext{
		AttributeNames:   map[identity.Key]string{attrKey: "balance"},
		ActionNames:      map[identity.Key]string{actionKey: "Deposit"},
		QueryNames:       map[identity.Key]string{queryKey: "GetBalance"},
		GlobalFunctions:  map[identity.Key]string{globalKey: "_Helper"},
		NamedSets:        map[identity.Key]string{namedSetKey: "valid_statuses"},
		ActionScopePaths: map[identity.Key]string{crossActionKey: "s2!c2!OtherAction"},
	}
}

// roundTrip performs: Expression → Raise → AST → Print → string → Parse → AST → Lower → Expression.
// Returns the final Expression and the printed string.
func (s *RaiseTestSuite) roundTrip(expr me.Expression) (me.Expression, string) {
	// Step 1: Raise to AST.
	raised, err := Raise(expr, s.raiseCtx)
	s.Require().NoError(err, "Raise failed")

	// Step 2: Print AST to string.
	printed := ast.Print(raised)

	// Step 3: Parse string back to AST.
	parsed, err := parser.ParseExpression(printed)
	s.Require().NoError(err, "Parse failed for: %q", printed)

	// Step 4: Lower AST back to Expression.
	lowered, err := Lower(parsed, s.lowerCtx)
	s.Require().NoError(err, "Lower failed for: %q", printed)

	return lowered, printed
}

// assertRoundTrip verifies that an expression survives a full round-trip.
func (s *RaiseTestSuite) assertRoundTrip(expr me.Expression) string {
	result, printed := s.roundTrip(expr)
	s.True(reflect.DeepEqual(expr, result),
		"round-trip mismatch for %q\n  original:  %#v\n  result:    %#v", printed, expr, result)
	return printed
}

// --- Literal round-trips ---

func (s *RaiseTestSuite) TestRoundTripBoolLiteral() {
	s.assertRoundTrip(&me.BoolLiteral{Value: true})
	s.assertRoundTrip(&me.BoolLiteral{Value: false})
}

func (s *RaiseTestSuite) TestRoundTripIntLiteral() {
	s.assertRoundTrip(&me.IntLiteral{Value: big.NewInt(42)})
	s.assertRoundTrip(&me.IntLiteral{Value: big.NewInt(0)})
}

func (s *RaiseTestSuite) TestRoundTripNegativeIntLiteral() {
	// Negative int → Negate{IntLiteral{positive}} after round-trip
	// because parser produces UnaryNegation{NumberLiteral} → Lower → Negate{IntLiteral}.
	expr := &me.IntLiteral{Value: big.NewInt(-5)}
	_, printed := s.roundTrip(expr)
	s.Equal("-5", printed)

	// Round-trip produces Negate{IntLiteral{5}}, not IntLiteral{-5}.
	result, _ := s.roundTrip(expr)
	negate, ok := result.(*me.Negate)
	s.True(ok, "expected Negate, got %T", result)
	intLit, ok := negate.Expr.(*me.IntLiteral)
	s.True(ok, "expected IntLiteral inside Negate")
	s.Equal(0, intLit.Value.Cmp(big.NewInt(5)))
}

func (s *RaiseTestSuite) TestRoundTripRationalLiteral() {
	// 3/4
	rat := new(big.Rat).SetFrac64(3, 4)
	expr := &me.RationalLiteral{Value: rat}
	result, printed := s.roundTrip(expr)
	s.Equal("3 / 4", printed)

	// Round-trip produces RationalLiteral{3/4}.
	ratResult, ok := result.(*me.RationalLiteral)
	s.True(ok, "expected RationalLiteral, got %T", result)
	s.Equal(0, ratResult.Value.Cmp(rat))
}

func (s *RaiseTestSuite) TestRoundTripRationalLiteralWholeNumber() {
	// 6/2 normalizes to 3, which is an integer.
	rat := new(big.Rat).SetFrac64(6, 2)
	expr := &me.RationalLiteral{Value: rat}
	result, printed := s.roundTrip(expr)
	s.Equal("3", printed)

	// Round-trip produces IntLiteral{3}.
	intResult, ok := result.(*me.IntLiteral)
	s.True(ok, "expected IntLiteral, got %T", result)
	s.Equal(0, intResult.Value.Cmp(big.NewInt(3)))
}

func (s *RaiseTestSuite) TestRoundTripStringLiteral() {
	s.assertRoundTrip(&me.StringLiteral{Value: "hello"})
	s.assertRoundTrip(&me.StringLiteral{Value: ""})
}

func (s *RaiseTestSuite) TestRoundTripStringLiteralEscape() {
	expr := &me.StringLiteral{Value: "say \"hi\""}
	result, printed := s.roundTrip(expr)
	s.Equal(`"say \"hi\""`, printed)
	s.Equal(expr, result)
}

// --- Collection round-trips ---

func (s *RaiseTestSuite) TestRoundTripSetLiteral() {
	s.assertRoundTrip(&me.SetLiteral{
		Elements: []me.Expression{
			&me.IntLiteral{Value: big.NewInt(1)},
			&me.IntLiteral{Value: big.NewInt(2)},
			&me.IntLiteral{Value: big.NewInt(3)},
		},
	})
}

func (s *RaiseTestSuite) TestRoundTripEmptySet() {
	s.assertRoundTrip(&me.SetLiteral{Elements: []me.Expression{}})
}

func (s *RaiseTestSuite) TestRoundTripTupleLiteral() {
	s.assertRoundTrip(&me.TupleLiteral{
		Elements: []me.Expression{
			&me.IntLiteral{Value: big.NewInt(1)},
			&me.StringLiteral{Value: "x"},
		},
	})
}

func (s *RaiseTestSuite) TestRoundTripRecordLiteral() {
	s.assertRoundTrip(&me.RecordLiteral{
		Fields: []me.RecordField{
			{Name: "a", Value: &me.IntLiteral{Value: big.NewInt(1)}},
			{Name: "b", Value: &me.StringLiteral{Value: "hello"}},
		},
	})
}

func (s *RaiseTestSuite) TestRoundTripSetConstant() {
	tests := []struct {
		kind me.SetConstantKind
	}{
		{me.SetConstantNat},
		{me.SetConstantInt},
		{me.SetConstantReal},
		{me.SetConstantBoolean},
	}
	for _, tt := range tests {
		s.assertRoundTrip(&me.SetConstant{Kind: tt.kind})
	}
}

func (s *RaiseTestSuite) TestRoundTripSetRange() {
	s.assertRoundTrip(&me.SetRange{
		Start: &me.IntLiteral{Value: big.NewInt(1)},
		End:   &me.IntLiteral{Value: big.NewInt(10)},
	})
}

// --- Reference round-trips ---

func (s *RaiseTestSuite) TestRoundTripSelfRef() {
	s.assertRoundTrip(&me.SelfRef{})
}

func (s *RaiseTestSuite) TestRoundTripAttributeRef() {
	attrKey := s.lowerCtx.AttributeNames["balance"]
	s.assertRoundTrip(&me.AttributeRef{AttributeKey: attrKey})
}

func (s *RaiseTestSuite) TestRoundTripLocalVar() {
	s.assertRoundTrip(&me.LocalVar{Name: "amount"})
}

func (s *RaiseTestSuite) TestRoundTripNamedSetRef() {
	setKey := s.lowerCtx.NamedSets["valid_statuses"]
	s.assertRoundTrip(&me.NamedSetRef{SetKey: setKey})
}

// --- Unary operator round-trips ---

func (s *RaiseTestSuite) TestRoundTripNegate() {
	s.assertRoundTrip(&me.Negate{
		Expr: &me.LocalVar{Name: "amount"},
	})
}

func (s *RaiseTestSuite) TestRoundTripNot() {
	s.assertRoundTrip(&me.Not{
		Expr: &me.BoolLiteral{Value: true},
	})
}

func (s *RaiseTestSuite) TestRoundTripNextState() {
	attrKey := s.lowerCtx.AttributeNames["balance"]
	s.assertRoundTrip(&me.NextState{
		Expr: &me.AttributeRef{AttributeKey: attrKey},
	})
}

// --- Binary operator round-trips ---

func (s *RaiseTestSuite) TestRoundTripBinaryArith() {
	ops := []me.ArithOp{me.ArithAdd, me.ArithSub, me.ArithMul, me.ArithDiv, me.ArithMod, me.ArithPow}
	for _, op := range ops {
		s.assertRoundTrip(&me.BinaryArith{
			Op:    op,
			Left:  &me.IntLiteral{Value: big.NewInt(2)},
			Right: &me.IntLiteral{Value: big.NewInt(3)},
		})
	}
}

func (s *RaiseTestSuite) TestRoundTripBinaryLogic() {
	ops := []me.LogicOp{me.LogicAnd, me.LogicOr, me.LogicImplies, me.LogicEquiv}
	for _, op := range ops {
		s.assertRoundTrip(&me.BinaryLogic{
			Op:    op,
			Left:  &me.BoolLiteral{Value: true},
			Right: &me.BoolLiteral{Value: false},
		})
	}
}

func (s *RaiseTestSuite) TestRoundTripCompare() {
	ops := []me.CompareOp{me.CompareLt, me.CompareGt, me.CompareLte, me.CompareGte, me.CompareEq, me.CompareNeq}
	for _, op := range ops {
		s.assertRoundTrip(&me.Compare{
			Op:    op,
			Left:  &me.IntLiteral{Value: big.NewInt(1)},
			Right: &me.IntLiteral{Value: big.NewInt(2)},
		})
	}
}

func (s *RaiseTestSuite) TestRoundTripSetOp() {
	ops := []me.SetOpKind{me.SetUnion, me.SetIntersect, me.SetDifference}
	setA := &me.LocalVar{Name: "amount"} // Uses a known parameter name
	setB := &me.LocalVar{Name: "amount"}
	for _, op := range ops {
		s.assertRoundTrip(&me.SetOp{
			Op:    op,
			Left:  setA,
			Right: setB,
		})
	}
}

func (s *RaiseTestSuite) TestRoundTripSetCompare() {
	ops := []me.SetCompareOp{me.SetCompareSubsetEq, me.SetCompareSubset, me.SetCompareSupersetEq, me.SetCompareSuperset}
	for _, op := range ops {
		s.assertRoundTrip(&me.SetCompare{
			Op:    op,
			Left:  &me.LocalVar{Name: "amount"},
			Right: &me.LocalVar{Name: "amount"},
		})
	}
}

func (s *RaiseTestSuite) TestRoundTripBagOp() {
	ops := []me.BagOpKind{me.BagSum, me.BagDifference}
	for _, op := range ops {
		s.assertRoundTrip(&me.BagOp{
			Op:    op,
			Left:  &me.LocalVar{Name: "amount"},
			Right: &me.LocalVar{Name: "amount"},
		})
	}
}

func (s *RaiseTestSuite) TestRoundTripBagCompare() {
	ops := []me.BagCompareOp{me.BagCompareProperSubBag, me.BagCompareSubBag, me.BagCompareProperSupBag, me.BagCompareSupBag}
	for _, op := range ops {
		s.assertRoundTrip(&me.BagCompare{
			Op:    op,
			Left:  &me.LocalVar{Name: "amount"},
			Right: &me.LocalVar{Name: "amount"},
		})
	}
}

func (s *RaiseTestSuite) TestRoundTripMembership() {
	s.assertRoundTrip(&me.Membership{
		Element: &me.IntLiteral{Value: big.NewInt(1)},
		Set: &me.SetLiteral{
			Elements: []me.Expression{
				&me.IntLiteral{Value: big.NewInt(1)},
				&me.IntLiteral{Value: big.NewInt(2)},
			},
		},
		Negated: false,
	})
	s.assertRoundTrip(&me.Membership{
		Element: &me.IntLiteral{Value: big.NewInt(1)},
		Set:     &me.SetConstant{Kind: me.SetConstantNat},
		Negated: true,
	})
}

// --- Concatenation round-trips ---

func (s *RaiseTestSuite) TestRoundTripTupleConcat() {
	// Note: StringConcat and TupleConcat both produce TupleConcat after
	// round-trip because the parser always produces TupleConcat for ∘.
	s.assertRoundTrip(&me.TupleConcat{
		Operands: []me.Expression{
			&me.LocalVar{Name: "amount"},
			&me.LocalVar{Name: "amount"},
		},
	})
}

// --- Field/Index access round-trips ---

func (s *RaiseTestSuite) TestRoundTripFieldAccess() {
	s.assertRoundTrip(&me.FieldAccess{
		Base:  &me.SelfRef{},
		Field: "balance",
	})
}

func (s *RaiseTestSuite) TestRoundTripTupleIndex() {
	s.assertRoundTrip(&me.TupleIndex{
		Tuple: &me.LocalVar{Name: "amount"},
		Index: &me.IntLiteral{Value: big.NewInt(1)},
	})
}

func (s *RaiseTestSuite) TestRoundTripStringIndex() {
	// StringIndex round-trips to TupleIndex because the parser
	// always produces TupleIndex for bracket syntax.
	expr := &me.StringIndex{
		Str:   &me.LocalVar{Name: "amount"},
		Index: &me.IntLiteral{Value: big.NewInt(1)},
	}
	result, _ := s.roundTrip(expr)
	// Parser produces TupleIndex, not StringIndex.
	tupleIdx, ok := result.(*me.TupleIndex)
	s.True(ok, "expected TupleIndex, got %T", result)
	s.True(reflect.DeepEqual(expr.Str, tupleIdx.Tuple))
	s.True(reflect.DeepEqual(expr.Index, tupleIdx.Index))
}

// --- Record alteration round-trips ---

func (s *RaiseTestSuite) TestRoundTripRecordUpdate() {
	attrKey := s.lowerCtx.AttributeNames["balance"]
	s.assertRoundTrip(&me.RecordUpdate{
		Base: &me.SelfRef{},
		Alterations: []me.FieldAlteration{
			{
				Field: "balance",
				Value: &me.BinaryArith{
					Op:    me.ArithAdd,
					Left:  &me.PriorFieldValue{Field: "balance"},
					Right: &me.LocalVar{Name: "amount"},
				},
			},
		},
	})
	// Verify the attribute ref works for base.
	s.assertRoundTrip(&me.RecordUpdate{
		Base: &me.AttributeRef{AttributeKey: attrKey},
		Alterations: []me.FieldAlteration{
			{Field: "balance", Value: &me.IntLiteral{Value: big.NewInt(0)}},
		},
	})
}

// --- Control flow round-trips ---

func (s *RaiseTestSuite) TestRoundTripIfThenElse() {
	s.assertRoundTrip(&me.IfThenElse{
		Condition: &me.Compare{
			Op:    me.CompareGt,
			Left:  &me.LocalVar{Name: "amount"},
			Right: &me.IntLiteral{Value: big.NewInt(0)},
		},
		Then: &me.LocalVar{Name: "amount"},
		Else: &me.IntLiteral{Value: big.NewInt(0)},
	})
}

func (s *RaiseTestSuite) TestRoundTripCase() {
	s.assertRoundTrip(&me.Case{
		Branches: []me.CaseBranch{
			{
				Condition: &me.Compare{
					Op:    me.CompareEq,
					Left:  &me.LocalVar{Name: "amount"},
					Right: &me.IntLiteral{Value: big.NewInt(1)},
				},
				Result: &me.IntLiteral{Value: big.NewInt(10)},
			},
			{
				Condition: &me.Compare{
					Op:    me.CompareEq,
					Left:  &me.LocalVar{Name: "amount"},
					Right: &me.IntLiteral{Value: big.NewInt(2)},
				},
				Result: &me.IntLiteral{Value: big.NewInt(20)},
			},
		},
		Otherwise: &me.IntLiteral{Value: big.NewInt(0)},
	})
}

func (s *RaiseTestSuite) TestRoundTripCaseNoOtherwise() {
	s.assertRoundTrip(&me.Case{
		Branches: []me.CaseBranch{
			{
				Condition: &me.BoolLiteral{Value: true},
				Result:    &me.IntLiteral{Value: big.NewInt(1)},
			},
		},
	})
}

// --- Quantifier round-trips ---

func (s *RaiseTestSuite) TestRoundTripQuantifierForall() {
	s.assertRoundTrip(&me.Quantifier{
		Kind:     me.QuantifierForall,
		Variable: "x",
		Domain:   &me.SetConstant{Kind: me.SetConstantNat},
		Predicate: &me.Compare{
			Op:    me.CompareGte,
			Left:  &me.LocalVar{Name: "x"},
			Right: &me.IntLiteral{Value: big.NewInt(0)},
		},
	})
}

func (s *RaiseTestSuite) TestRoundTripQuantifierExists() {
	s.assertRoundTrip(&me.Quantifier{
		Kind:     me.QuantifierExists,
		Variable: "x",
		Domain: &me.SetLiteral{
			Elements: []me.Expression{
				&me.IntLiteral{Value: big.NewInt(1)},
				&me.IntLiteral{Value: big.NewInt(2)},
			},
		},
		Predicate: &me.Compare{
			Op:    me.CompareGt,
			Left:  &me.LocalVar{Name: "x"},
			Right: &me.IntLiteral{Value: big.NewInt(0)},
		},
	})
}

func (s *RaiseTestSuite) TestRaiseSetFilter() {
	// SetFilter cannot do a full round-trip because the PEG parser does not
	// support the {x ∈ S : P(x)} syntax. Instead, verify raise + print output.
	expr := &me.SetFilter{
		Variable: "x",
		Set:      &me.SetConstant{Kind: me.SetConstantInt},
		Predicate: &me.Compare{
			Op:    me.CompareGt,
			Left:  &me.LocalVar{Name: "x"},
			Right: &me.IntLiteral{Value: big.NewInt(0)},
		},
	}
	raised, err := Raise(expr, s.raiseCtx)
	s.Require().NoError(err)
	printed := ast.Print(raised)
	s.Equal("{x ∈ Int : x > 0}", printed)
}

// --- Call round-trips ---

func (s *RaiseTestSuite) TestRoundTripActionCall() {
	actionKey := s.lowerCtx.ActionNames["Deposit"]
	s.assertRoundTrip(&me.ActionCall{
		ActionKey: actionKey,
		Args: []me.Expression{
			&me.IntLiteral{Value: big.NewInt(100)},
		},
	})
}

func (s *RaiseTestSuite) TestRoundTripQueryCall() {
	queryKey := s.lowerCtx.QueryNames["GetBalance"]
	s.assertRoundTrip(&me.ActionCall{
		ActionKey: queryKey,
		Args:      []me.Expression{},
	})
}

func (s *RaiseTestSuite) TestRoundTripGlobalCall() {
	globalKey := s.lowerCtx.GlobalFunctions["_Helper"]
	s.assertRoundTrip(&me.GlobalCall{
		FunctionKey: globalKey,
		Args: []me.Expression{
			&me.IntLiteral{Value: big.NewInt(1)},
		},
	})
}

func (s *RaiseTestSuite) TestRoundTripBuiltinCall() {
	attrKey := s.lowerCtx.AttributeNames["balance"]
	s.assertRoundTrip(&me.BuiltinCall{
		Module:   "_Seq",
		Function: "Len",
		Args: []me.Expression{
			&me.AttributeRef{AttributeKey: attrKey},
		},
	})
}

func (s *RaiseTestSuite) TestRoundTripCrossClassActionCall() {
	crossKey := s.lowerCtx.AllActions["s2!c2!OtherAction"]
	s.assertRoundTrip(&me.ActionCall{
		ActionKey: crossKey,
		Args: []me.Expression{
			&me.IntLiteral{Value: big.NewInt(1)},
		},
	})
}

// --- Precedence round-trips ---

func (s *RaiseTestSuite) TestRoundTripPrecedenceAndOr() {
	// a ∧ b ∨ c — and binds tighter than or
	expr := &me.BinaryLogic{
		Op: me.LogicOr,
		Left: &me.BinaryLogic{
			Op:    me.LogicAnd,
			Left:  &me.BoolLiteral{Value: true},
			Right: &me.BoolLiteral{Value: false},
		},
		Right: &me.BoolLiteral{Value: true},
	}
	printed := s.assertRoundTrip(expr)
	s.Equal("TRUE ∧ FALSE ∨ TRUE", printed)
}

func (s *RaiseTestSuite) TestRoundTripPrecedenceOrInsideAnd() {
	// (a ∨ b) ∧ c — needs parens
	expr := &me.BinaryLogic{
		Op: me.LogicAnd,
		Left: &me.BinaryLogic{
			Op:    me.LogicOr,
			Left:  &me.BoolLiteral{Value: true},
			Right: &me.BoolLiteral{Value: false},
		},
		Right: &me.BoolLiteral{Value: true},
	}
	printed := s.assertRoundTrip(expr)
	s.Equal("(TRUE ∨ FALSE) ∧ TRUE", printed)
}

func (s *RaiseTestSuite) TestRoundTripPrecedenceImpliesRightAssoc() {
	// a ⇒ b ⇒ c — right-assoc
	expr := &me.BinaryLogic{
		Op:   me.LogicImplies,
		Left: &me.BoolLiteral{Value: true},
		Right: &me.BinaryLogic{
			Op:    me.LogicImplies,
			Left:  &me.BoolLiteral{Value: false},
			Right: &me.BoolLiteral{Value: true},
		},
	}
	printed := s.assertRoundTrip(expr)
	s.Equal("TRUE ⇒ FALSE ⇒ TRUE", printed)
}

func (s *RaiseTestSuite) TestRoundTripPrecedenceArithmetic() {
	// 1 + 2 * 3 — mul higher than add
	expr := &me.BinaryArith{
		Op:   me.ArithAdd,
		Left: &me.IntLiteral{Value: big.NewInt(1)},
		Right: &me.BinaryArith{
			Op:    me.ArithMul,
			Left:  &me.IntLiteral{Value: big.NewInt(2)},
			Right: &me.IntLiteral{Value: big.NewInt(3)},
		},
	}
	printed := s.assertRoundTrip(expr)
	s.Equal("1 + 2 * 3", printed)
}

func (s *RaiseTestSuite) TestRoundTripPrecedenceAddInsideMul() {
	// (1 + 2) * 3 — needs parens
	expr := &me.BinaryArith{
		Op: me.ArithMul,
		Left: &me.BinaryArith{
			Op:    me.ArithAdd,
			Left:  &me.IntLiteral{Value: big.NewInt(1)},
			Right: &me.IntLiteral{Value: big.NewInt(2)},
		},
		Right: &me.IntLiteral{Value: big.NewInt(3)},
	}
	printed := s.assertRoundTrip(expr)
	s.Equal("(1 + 2) * 3", printed)
}

func (s *RaiseTestSuite) TestRoundTripPrecedenceComparisonAndLogic() {
	// x > 0 ∧ x < 10 — comparisons higher than logic
	expr := &me.BinaryLogic{
		Op: me.LogicAnd,
		Left: &me.Compare{
			Op:    me.CompareGt,
			Left:  &me.LocalVar{Name: "amount"},
			Right: &me.IntLiteral{Value: big.NewInt(0)},
		},
		Right: &me.Compare{
			Op:    me.CompareLt,
			Left:  &me.LocalVar{Name: "amount"},
			Right: &me.IntLiteral{Value: big.NewInt(10)},
		},
	}
	printed := s.assertRoundTrip(expr)
	s.Equal("amount > 0 ∧ amount < 10", printed)
}

// --- Raise-only tests (not round-trip) ---

func (s *RaiseTestSuite) TestRaiseNilExpression() {
	_, err := Raise(nil, s.raiseCtx)
	s.Error(err)
}

func (s *RaiseTestSuite) TestRaisePriorFieldValue() {
	// PriorFieldValue raises to ExistingValue (@).
	raised, err := Raise(&me.PriorFieldValue{Field: "count"}, s.raiseCtx)
	s.NoError(err)
	_, ok := raised.(*ast.ExistingValue)
	s.True(ok)
}

func (s *RaiseTestSuite) TestRaiseStringConcat() {
	// StringConcat raises to TupleConcat AST (parser always produces TupleConcat).
	raised, err := Raise(&me.StringConcat{
		Operands: []me.Expression{
			&me.StringLiteral{Value: "a"},
			&me.StringLiteral{Value: "b"},
		},
	}, s.raiseCtx)
	s.NoError(err)
	tc, ok := raised.(*ast.TupleConcat)
	s.True(ok)
	s.Len(tc.Operands, 2)
}

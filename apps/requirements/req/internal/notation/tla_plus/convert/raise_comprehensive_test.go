package convert

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"

	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/parser"
)

// RaiseComprehensiveTestSuite provides exhaustive pair-wise precedence coverage,
// associativity tests, cross-category interactions, and deeply nested expressions.
type RaiseComprehensiveTestSuite struct {
	suite.Suite
	raiseCtx *RaiseContext
	lowerCtx *LowerContext
}

func TestRaiseComprehensiveSuite(t *testing.T) {
	suite.Run(t, new(RaiseComprehensiveTestSuite))
}

func (s *RaiseComprehensiveTestSuite) SetupTest() {
	domainKey, _ := identity.NewDomainKey("d")
	subKey, _ := identity.NewSubdomainKey(domainKey, "s")
	classKey, _ := identity.NewClassKey(subKey, "c")
	attrBalanceKey, _ := identity.NewAttributeKey(classKey, "balance")
	attrCountKey, _ := identity.NewAttributeKey(classKey, "count")
	attrTotalKey, _ := identity.NewAttributeKey(classKey, "total")
	attrItemsKey, _ := identity.NewAttributeKey(classKey, "items")
	attrActiveKey, _ := identity.NewAttributeKey(classKey, "active")
	actionDepositKey, _ := identity.NewActionKey(classKey, "Deposit")
	actionWithdrawKey, _ := identity.NewActionKey(classKey, "Withdraw")
	queryGetBalanceKey, _ := identity.NewQueryKey(classKey, "GetBalance")
	queryIsValidKey, _ := identity.NewQueryKey(classKey, "IsValid")
	globalHelperKey, _ := identity.NewGlobalFunctionKey("_Helper")
	globalMaxKey, _ := identity.NewGlobalFunctionKey("_Max")
	namedSetKey, _ := identity.NewNamedSetKey("valid_statuses")
	namedSetAccountsKey, _ := identity.NewNamedSetKey("all_accounts")

	// Cross-class actions.
	subKey2, _ := identity.NewSubdomainKey(domainKey, "s2")
	classKey2, _ := identity.NewClassKey(subKey2, "c2")
	crossActionKey, _ := identity.NewActionKey(classKey2, "OtherAction")

	s.lowerCtx = &LowerContext{
		ClassKey: classKey,
		AttributeNames: map[string]identity.Key{
			"balance": attrBalanceKey,
			"count":   attrCountKey,
			"total":   attrTotalKey,
			"items":   attrItemsKey,
			"active":  attrActiveKey,
		},
		ActionNames: map[string]identity.Key{
			"Deposit":  actionDepositKey,
			"Withdraw": actionWithdrawKey,
		},
		QueryNames: map[string]identity.Key{
			"GetBalance": queryGetBalanceKey,
			"IsValid":    queryIsValidKey,
		},
		GlobalFunctions: map[string]identity.Key{
			"_Helper": globalHelperKey,
			"_Max":    globalMaxKey,
		},
		NamedSets: map[string]identity.Key{
			"valid_statuses": namedSetKey,
			"all_accounts":   namedSetAccountsKey,
		},
		AllActions: map[string]identity.Key{
			"s2!c2!OtherAction": crossActionKey,
		},
		Parameters: map[string]bool{
			"amount": true,
			"rate":   true,
			"limit":  true,
			"n":      true,
			"x":      true,
			"y":      true,
			"z":      true,
			"a":      true,
			"b":      true,
			"c":      true,
			"d":      true,
			"p":      true,
			"q":      true,
			"r":      true,
			"s":      true,
			"t":      true,
		},
	}

	s.raiseCtx = &RaiseContext{
		AttributeNames: map[identity.Key]string{
			attrBalanceKey: "balance",
			attrCountKey:   "count",
			attrTotalKey:   "total",
			attrItemsKey:   "items",
			attrActiveKey:  "active",
		},
		ActionNames: map[identity.Key]string{
			actionDepositKey:  "Deposit",
			actionWithdrawKey: "Withdraw",
		},
		QueryNames: map[identity.Key]string{
			queryGetBalanceKey: "GetBalance",
			queryIsValidKey:    "IsValid",
		},
		GlobalFunctions: map[identity.Key]string{
			globalHelperKey: "_Helper",
			globalMaxKey:    "_Max",
		},
		NamedSets: map[identity.Key]string{
			namedSetKey:         "valid_statuses",
			namedSetAccountsKey: "all_accounts",
		},
		ActionScopePaths: map[identity.Key]string{
			crossActionKey: "s2!c2!OtherAction",
		},
	}
}

// roundTrip performs: Expression → Raise → AST → Print → string → Parse → AST → Lower → Expression.
func (s *RaiseComprehensiveTestSuite) roundTrip(expr me.Expression) (me.Expression, string) {
	raised, err := Raise(expr, s.raiseCtx)
	s.Require().NoError(err, "Raise failed")

	printed := ast.Print(raised)

	parsed, err := parser.ParseExpression(printed)
	s.Require().NoError(err, "Parse failed for: %q", printed)

	lowered, err := Lower(parsed, s.lowerCtx)
	s.Require().NoError(err, "Lower failed for: %q", printed)

	return lowered, printed
}

// assertRoundTrip verifies that an expression survives a full round-trip.
func (s *RaiseComprehensiveTestSuite) assertRoundTrip(expr me.Expression) string {
	result, printed := s.roundTrip(expr)
	s.True(reflect.DeepEqual(expr, result),
		"round-trip mismatch for %q\n  original:  %#v\n  result:    %#v", printed, expr, result)
	return printed
}

// Helpers for constructing model expressions.
func intLit(v int64) *me.IntLiteral     { return &me.IntLiteral{Value: big.NewInt(v)} }
func boolLit(v bool) *me.BoolLiteral    { return &me.BoolLiteral{Value: v} }
func strLit(v string) *me.StringLiteral { return &me.StringLiteral{Value: v} }
func localVar(name string) *me.LocalVar { return &me.LocalVar{Name: name} }

func add(l, r me.Expression) *me.BinaryArith {
	return &me.BinaryArith{Op: me.ArithAdd, Left: l, Right: r}
}
func sub(l, r me.Expression) *me.BinaryArith {
	return &me.BinaryArith{Op: me.ArithSub, Left: l, Right: r}
}
func mul(l, r me.Expression) *me.BinaryArith {
	return &me.BinaryArith{Op: me.ArithMul, Left: l, Right: r}
}
func div(l, r me.Expression) *me.BinaryArith {
	return &me.BinaryArith{Op: me.ArithDiv, Left: l, Right: r}
}
func mod(l, r me.Expression) *me.BinaryArith {
	return &me.BinaryArith{Op: me.ArithMod, Left: l, Right: r}
}
func pow(l, r me.Expression) *me.BinaryArith {
	return &me.BinaryArith{Op: me.ArithPow, Left: l, Right: r}
}
func and(l, r me.Expression) *me.BinaryLogic {
	return &me.BinaryLogic{Op: me.LogicAnd, Left: l, Right: r}
}
func or(l, r me.Expression) *me.BinaryLogic {
	return &me.BinaryLogic{Op: me.LogicOr, Left: l, Right: r}
}
func implies(l, r me.Expression) *me.BinaryLogic {
	return &me.BinaryLogic{Op: me.LogicImplies, Left: l, Right: r}
}
func equiv(l, r me.Expression) *me.BinaryLogic {
	return &me.BinaryLogic{Op: me.LogicEquiv, Left: l, Right: r}
}
func lt(l, r me.Expression) *me.Compare {
	return &me.Compare{Op: me.CompareLt, Left: l, Right: r}
}
func gt(l, r me.Expression) *me.Compare {
	return &me.Compare{Op: me.CompareGt, Left: l, Right: r}
}
func lte(l, r me.Expression) *me.Compare {
	return &me.Compare{Op: me.CompareLte, Left: l, Right: r}
}
func gte(l, r me.Expression) *me.Compare {
	return &me.Compare{Op: me.CompareGte, Left: l, Right: r}
}
func eq(l, r me.Expression) *me.Compare {
	return &me.Compare{Op: me.CompareEq, Left: l, Right: r}
}
func setUnion(l, r me.Expression) *me.SetOp {
	return &me.SetOp{Op: me.SetUnion, Left: l, Right: r}
}
func setIntersect(l, r me.Expression) *me.SetOp {
	return &me.SetOp{Op: me.SetIntersect, Left: l, Right: r}
}
func setDiff(l, r me.Expression) *me.SetOp {
	return &me.SetOp{Op: me.SetDifference, Left: l, Right: r}
}
func not(e me.Expression) *me.Not    { return &me.Not{Expr: e} }
func neg(e me.Expression) *me.Negate { return &me.Negate{Expr: e} }

// =============================================================================
// SECTION 1: Pair-wise Precedence Tests
//
// For each adjacent pair of precedence levels in the PEG grammar, test both
// directions: (1) higher-prec inside lower-prec (no parens), and
// (2) lower-prec inside higher-prec (needs parens).
// =============================================================================

// --- Implies (1) vs Equiv (2) ---

func (s *RaiseComprehensiveTestSuite) TestPrecPair_EquivInsideImplies() {
	// (a ≡ b) ⇒ c — equiv higher than implies, no parens needed on left
	expr := implies(equiv(boolLit(true), boolLit(false)), boolLit(true))
	printed := s.assertRoundTrip(expr)
	s.Equal("TRUE ≡ FALSE ⇒ TRUE", printed)
}

func (s *RaiseComprehensiveTestSuite) TestPrecPair_ImpliesInsideEquiv() {
	// (a ⇒ b) ≡ c — implies lower than equiv, needs parens
	expr := equiv(implies(boolLit(true), boolLit(false)), boolLit(true))
	printed := s.assertRoundTrip(expr)
	s.Equal("(TRUE ⇒ FALSE) ≡ TRUE", printed)
}

// --- Equiv (2) vs Or (3) ---

func (s *RaiseComprehensiveTestSuite) TestPrecPair_OrInsideEquiv() {
	// (a ∨ b) ≡ c — or higher, no parens on left
	expr := equiv(or(boolLit(true), boolLit(false)), boolLit(true))
	printed := s.assertRoundTrip(expr)
	s.Equal("TRUE ∨ FALSE ≡ TRUE", printed)
}

func (s *RaiseComprehensiveTestSuite) TestPrecPair_EquivInsideOr() {
	// (a ≡ b) ∨ c — equiv lower, needs parens
	expr := or(equiv(boolLit(true), boolLit(false)), boolLit(true))
	printed := s.assertRoundTrip(expr)
	s.Equal("(TRUE ≡ FALSE) ∨ TRUE", printed)
}

// --- Or (3) vs And (4) ---

func (s *RaiseComprehensiveTestSuite) TestPrecPair_AndInsideOr() {
	// (a ∧ b) ∨ c — and higher, no parens
	expr := or(and(boolLit(true), boolLit(false)), boolLit(true))
	printed := s.assertRoundTrip(expr)
	s.Equal("TRUE ∧ FALSE ∨ TRUE", printed)
}

func (s *RaiseComprehensiveTestSuite) TestPrecPair_OrInsideAnd() {
	// (a ∨ b) ∧ c — or lower, needs parens
	expr := and(or(boolLit(true), boolLit(false)), boolLit(true))
	printed := s.assertRoundTrip(expr)
	s.Equal("(TRUE ∨ FALSE) ∧ TRUE", printed)
}

// --- And (4) vs Not (5) ---

func (s *RaiseComprehensiveTestSuite) TestPrecPair_NotInsideAnd() {
	// ¬a ∧ b — not higher, no parens
	expr := and(not(boolLit(true)), boolLit(false))
	printed := s.assertRoundTrip(expr)
	s.Equal("¬TRUE ∧ FALSE", printed)
}

func (s *RaiseComprehensiveTestSuite) TestPrecPair_AndInsideNot() {
	// ¬(a ∧ b) — and lower, needs parens
	expr := not(and(boolLit(true), boolLit(false)))
	printed := s.assertRoundTrip(expr)
	s.Equal("¬(TRUE ∧ FALSE)", printed)
}

// --- Not (5) vs Membership (7) ---

func (s *RaiseComprehensiveTestSuite) TestPrecPair_MembershipInsideNot() {
	// ¬(x ∈ S) — membership higher prec (7) than not (5), no parens needed
	expr := not(&me.Membership{
		Element: localVar("x"),
		Set:     localVar("s"),
		Negated: false,
	})
	printed := s.assertRoundTrip(expr)
	s.Equal("¬x ∈ s", printed)
}

// --- Membership (7) vs SetCompare (8) ---

func (s *RaiseComprehensiveTestSuite) TestPrecPair_MembershipInsideSetCompare() {
	// Membership prec (7) < set compare prec (8) — needs parens
	expr := &me.SetCompare{
		Op: me.SetCompareSubsetEq,
		Left: &me.Membership{
			Element: localVar("x"),
			Set:     localVar("a"),
			Negated: false,
		},
		Right: localVar("b"),
	}
	printed := s.assertRoundTrip(expr)
	s.Equal("(x ∈ a) ⊆ b", printed)
}

// --- SetCompare (8) vs BagCompare (9) ---

func (s *RaiseComprehensiveTestSuite) TestPrecPair_BagCompareInsideSetCompare() {
	// bag compare is at a higher precedence — no parens needed on left
	expr := &me.SetCompare{
		Op: me.SetCompareSubsetEq,
		Left: &me.BagCompare{
			Op:    me.BagCompareSubBag,
			Left:  localVar("x"),
			Right: localVar("y"),
		},
		Right: localVar("z"),
	}
	printed := s.assertRoundTrip(expr)
	s.Equal("x ⊑ y ⊆ z", printed)
}

// --- BagCompare (9) vs Equality (10) ---

func (s *RaiseComprehensiveTestSuite) TestPrecPair_EqualityInsideBagCompare() {
	// Equality prec (10) > bag compare prec (9) — no parens needed
	expr := &me.BagCompare{
		Op:    me.BagCompareSubBag,
		Left:  eq(localVar("x"), localVar("y")),
		Right: localVar("z"),
	}
	printed := s.assertRoundTrip(expr)
	s.Equal("x = y ⊑ z", printed)
}

func (s *RaiseComprehensiveTestSuite) TestPrecPair_BagCompareInsideEquality() {
	// Bag compare prec (9) < equality prec (10) — needs parens
	expr := eq(
		&me.BagCompare{Op: me.BagCompareSubBag, Left: localVar("x"), Right: localVar("y")},
		boolLit(true),
	)
	printed := s.assertRoundTrip(expr)
	s.Equal("(x ⊑ y) = TRUE", printed)
}

// --- Equality (10) vs NumericComparison (11) ---

func (s *RaiseComprehensiveTestSuite) TestPrecPair_NumCompareInsideEquality() {
	// numeric compare higher than equality — no parens
	expr := eq(gt(localVar("x"), intLit(0)), boolLit(true))
	printed := s.assertRoundTrip(expr)
	s.Equal("x > 0 = TRUE", printed)
}

func (s *RaiseComprehensiveTestSuite) TestPrecPair_EqualityInsideNumCompare() {
	// equality lower than numeric compare — needs parens
	expr := gt(eq(localVar("x"), intLit(1)), intLit(0))
	printed := s.assertRoundTrip(expr)
	s.Equal("(x = 1) > 0", printed)
}

// --- NumericComparison (11) vs SetDifference (12) ---

func (s *RaiseComprehensiveTestSuite) TestPrecPair_SetDiffInsideNumCompare() {
	// set difference higher than numeric compare — set diff as operand of compare
	expr := gt(
		setDiff(localVar("a"), localVar("b")),
		localVar("c"),
	)
	printed := s.assertRoundTrip(expr)
	s.Equal(`a \ b > c`, printed)
}

// --- SetDifference (12) vs SetIntersection (13) ---

func (s *RaiseComprehensiveTestSuite) TestPrecPair_IntersectInsideSetDiff() {
	// intersect higher than set diff — no parens
	expr := setDiff(
		setIntersect(localVar("a"), localVar("b")),
		localVar("c"),
	)
	printed := s.assertRoundTrip(expr)
	s.Equal(`a ∩ b \ c`, printed)
}

func (s *RaiseComprehensiveTestSuite) TestPrecPair_SetDiffInsideIntersect() {
	// set diff lower — needs parens
	expr := setIntersect(
		setDiff(localVar("a"), localVar("b")),
		localVar("c"),
	)
	printed := s.assertRoundTrip(expr)
	s.Equal(`(a \ b) ∩ c`, printed)
}

// --- SetIntersection (13) vs SetUnion (14) ---

func (s *RaiseComprehensiveTestSuite) TestPrecPair_UnionInsideIntersect() {
	// Union prec (14) > intersect prec (13) — union binds tighter, no parens needed
	expr := setIntersect(
		setUnion(localVar("a"), localVar("b")),
		localVar("c"),
	)
	printed := s.assertRoundTrip(expr)
	s.Equal("a ∪ b ∩ c", printed)
}

func (s *RaiseComprehensiveTestSuite) TestPrecPair_IntersectInsideUnion() {
	// Intersect prec (13) < union prec (14) — needs parens
	expr := setUnion(
		setIntersect(localVar("a"), localVar("b")),
		localVar("c"),
	)
	printed := s.assertRoundTrip(expr)
	s.Equal("(a ∩ b) ∪ c", printed)
}

// --- SetUnion (14) vs Range (16) ---

func (s *RaiseComprehensiveTestSuite) TestPrecPair_RangeInsideUnion() {
	// range higher than union — no parens
	expr := setUnion(
		&me.SetRange{Start: intLit(1), End: intLit(5)},
		&me.SetRange{Start: intLit(6), End: intLit(10)},
	)
	printed := s.assertRoundTrip(expr)
	s.Equal("1..5 ∪ 6..10", printed)
}

// --- BagSum (17) vs Mod (18) ---

func (s *RaiseComprehensiveTestSuite) TestPrecPair_ModInsideBagSum() {
	// mod higher than bag sum — no parens
	expr := &me.BagOp{
		Op:    me.BagSum,
		Left:  mod(localVar("x"), intLit(3)),
		Right: localVar("y"),
	}
	printed := s.assertRoundTrip(expr)
	s.Equal("x % 3 ⊕ y", printed)
}

func (s *RaiseComprehensiveTestSuite) TestPrecPair_BagSumInsideMod() {
	// bag sum lower — needs parens
	expr := mod(
		&me.BagOp{Op: me.BagSum, Left: localVar("x"), Right: localVar("y")},
		intLit(3),
	)
	printed := s.assertRoundTrip(expr)
	s.Equal("(x ⊕ y) % 3", printed)
}

// --- Mod (18) vs Add (19) ---

func (s *RaiseComprehensiveTestSuite) TestPrecPair_AddInsideMod() {
	// add higher than mod — no parens
	expr := mod(
		add(localVar("x"), intLit(1)),
		intLit(3),
	)
	printed := s.assertRoundTrip(expr)
	s.Equal("x + 1 % 3", printed)
}

func (s *RaiseComprehensiveTestSuite) TestPrecPair_ModInsideAdd() {
	// mod lower — needs parens
	expr := add(
		mod(localVar("x"), intLit(3)),
		intLit(1),
	)
	printed := s.assertRoundTrip(expr)
	s.Equal("(x % 3) + 1", printed)
}

// --- BagDiff (20) vs Sub (21) ---

func (s *RaiseComprehensiveTestSuite) TestPrecPair_SubInsideBagDiff() {
	// sub higher — no parens
	expr := &me.BagOp{
		Op:    me.BagDifference,
		Left:  sub(localVar("x"), intLit(1)),
		Right: localVar("y"),
	}
	printed := s.assertRoundTrip(expr)
	s.Equal("x - 1 ⊖ y", printed)
}

func (s *RaiseComprehensiveTestSuite) TestPrecPair_BagDiffInsideSub() {
	// bag diff lower — needs parens
	expr := sub(
		&me.BagOp{Op: me.BagDifference, Left: localVar("x"), Right: localVar("y")},
		intLit(1),
	)
	printed := s.assertRoundTrip(expr)
	s.Equal("(x ⊖ y) - 1", printed)
}

// --- Sub (21) vs Negate (22) ---

func (s *RaiseComprehensiveTestSuite) TestPrecPair_NegInsideSub() {
	// negate higher than sub — 5 - (-3)
	expr := sub(intLit(5), neg(intLit(3)))
	printed := s.assertRoundTrip(expr)
	s.Equal("5 - -3", printed)
}

// --- Negate (22) vs Div (23) ---

func (s *RaiseComprehensiveTestSuite) TestPrecPair_DivInsideNeg() {
	// Div prec (23) > negate prec (22) — div binds tighter, no parens needed
	expr := neg(div(intLit(6), intLit(2)))
	printed := s.assertRoundTrip(expr)
	s.Equal("-6 ÷ 2", printed)
}

func (s *RaiseComprehensiveTestSuite) TestPrecPair_NegInsideDiv() {
	// negate lower than div — (-x) ÷ y needs parens
	expr := div(neg(localVar("x")), localVar("y"))
	printed := s.assertRoundTrip(expr)
	s.Equal("(-x) ÷ y", printed)
}

// --- Div (23) vs Mul (25) ---

func (s *RaiseComprehensiveTestSuite) TestPrecPair_MulInsideDiv() {
	// mul higher than div — no parens needed
	expr := div(mul(localVar("x"), intLit(2)), intLit(3))
	printed := s.assertRoundTrip(expr)
	s.Equal("x * 2 ÷ 3", printed)
}

func (s *RaiseComprehensiveTestSuite) TestPrecPair_DivInsideMul() {
	// div lower — needs parens
	expr := mul(div(localVar("x"), intLit(2)), intLit(3))
	printed := s.assertRoundTrip(expr)
	s.Equal("(x ÷ 2) * 3", printed)
}

// --- Mul (25) vs Pow (27) ---

func (s *RaiseComprehensiveTestSuite) TestPrecPair_PowInsideMul() {
	// pow higher than mul — no parens
	expr := mul(pow(intLit(2), intLit(3)), intLit(4))
	printed := s.assertRoundTrip(expr)
	s.Equal("2 ^ 3 * 4", printed)
}

func (s *RaiseComprehensiveTestSuite) TestPrecPair_MulInsidePow() {
	// mul lower — needs parens
	expr := pow(mul(intLit(2), intLit(3)), intLit(4))
	printed := s.assertRoundTrip(expr)
	s.Equal("(2 * 3) ^ 4", printed)
}

// --- Concat (24) vs Mul (25) ---

func (s *RaiseComprehensiveTestSuite) TestPrecPair_MulInsideConcat() {
	// mul higher than concat — no parens
	expr := &me.TupleConcat{
		Operands: []me.Expression{
			mul(localVar("a"), localVar("b")),
			localVar("c"),
		},
	}
	printed := s.assertRoundTrip(expr)
	s.Equal("a * b ∘ c", printed)
}

// --- Pow (27) vs Prime (28) ---

func (s *RaiseComprehensiveTestSuite) TestPrecPair_PrimeInsidePow() {
	// prime higher than pow — no parens
	attrKey := s.lowerCtx.AttributeNames["balance"]
	expr := pow(
		&me.NextState{Expr: &me.AttributeRef{AttributeKey: attrKey}},
		intLit(2),
	)
	printed := s.assertRoundTrip(expr)
	s.Equal("balance' ^ 2", printed)
}

// --- Prime (28) vs FieldAccess (29) ---

func (s *RaiseComprehensiveTestSuite) TestPrecPair_FieldAccessInsidePrime() {
	// field access higher than prime — record.field' = (record.field)'
	expr := &me.NextState{
		Expr: &me.FieldAccess{
			Base:  &me.SelfRef{},
			Field: "balance",
		},
	}
	printed := s.assertRoundTrip(expr)
	s.Equal("self.balance'", printed)
}

// =============================================================================
// SECTION 2: Associativity Chain Tests
// =============================================================================

// --- Left-associative chains ---

func (s *RaiseComprehensiveTestSuite) TestAssocLeftAddChain() {
	// (a + b) + c — left-assoc, no parens
	expr := add(add(localVar("x"), localVar("y")), localVar("z"))
	printed := s.assertRoundTrip(expr)
	s.Equal("x + y + z", printed)
}

func (s *RaiseComprehensiveTestSuite) TestAssocLeftAddChainRight() {
	// a + (b + c) — right child same prec left-assoc, needs parens
	expr := add(localVar("x"), add(localVar("y"), localVar("z")))
	printed := s.assertRoundTrip(expr)
	s.Equal("x + (y + z)", printed)
}

func (s *RaiseComprehensiveTestSuite) TestAssocLeftSubChain() {
	// (a - b) - c — left-assoc
	expr := sub(sub(localVar("x"), localVar("y")), localVar("z"))
	printed := s.assertRoundTrip(expr)
	s.Equal("x - y - z", printed)
}

func (s *RaiseComprehensiveTestSuite) TestAssocLeftSubChainRight() {
	// a - (b - c) — right child same prec, needs parens
	expr := sub(localVar("x"), sub(localVar("y"), localVar("z")))
	printed := s.assertRoundTrip(expr)
	s.Equal("x - (y - z)", printed)
}

func (s *RaiseComprehensiveTestSuite) TestAssocLeftMulChain() {
	// (a * b) * c — left-assoc
	expr := mul(mul(intLit(2), intLit(3)), intLit(4))
	printed := s.assertRoundTrip(expr)
	s.Equal("2 * 3 * 4", printed)
}

func (s *RaiseComprehensiveTestSuite) TestAssocLeftDivChain() {
	// (a ÷ b) ÷ c — left-assoc
	expr := div(div(intLit(12), intLit(3)), intLit(2))
	printed := s.assertRoundTrip(expr)
	s.Equal("12 ÷ 3 ÷ 2", printed)
}

func (s *RaiseComprehensiveTestSuite) TestAssocLeftModChain() {
	// (a % b) % c — left-assoc
	expr := mod(mod(intLit(17), intLit(5)), intLit(3))
	printed := s.assertRoundTrip(expr)
	s.Equal("17 % 5 % 3", printed)
}

func (s *RaiseComprehensiveTestSuite) TestAssocLeftAndChain() {
	// (a ∧ b) ∧ c — left-assoc
	expr := and(and(boolLit(true), boolLit(false)), boolLit(true))
	printed := s.assertRoundTrip(expr)
	s.Equal("TRUE ∧ FALSE ∧ TRUE", printed)
}

func (s *RaiseComprehensiveTestSuite) TestAssocLeftOrChain() {
	// (a ∨ b) ∨ c — left-assoc
	expr := or(or(boolLit(true), boolLit(false)), boolLit(true))
	printed := s.assertRoundTrip(expr)
	s.Equal("TRUE ∨ FALSE ∨ TRUE", printed)
}

func (s *RaiseComprehensiveTestSuite) TestAssocLeftEquivChain() {
	// (a ≡ b) ≡ c — left-assoc
	expr := equiv(equiv(boolLit(true), boolLit(false)), boolLit(true))
	printed := s.assertRoundTrip(expr)
	s.Equal("TRUE ≡ FALSE ≡ TRUE", printed)
}

func (s *RaiseComprehensiveTestSuite) TestAssocLeftSetUnionChain() {
	// (a ∪ b) ∪ c — left-assoc
	expr := setUnion(setUnion(localVar("a"), localVar("b")), localVar("c"))
	printed := s.assertRoundTrip(expr)
	s.Equal("a ∪ b ∪ c", printed)
}

func (s *RaiseComprehensiveTestSuite) TestAssocLeftSetIntersectChain() {
	// (a ∩ b) ∩ c — left-assoc
	expr := setIntersect(setIntersect(localVar("a"), localVar("b")), localVar("c"))
	printed := s.assertRoundTrip(expr)
	s.Equal("a ∩ b ∩ c", printed)
}

func (s *RaiseComprehensiveTestSuite) TestAssocLeftSetDiffChain() {
	// (a \ b) \ c — left-assoc
	expr := setDiff(setDiff(localVar("a"), localVar("b")), localVar("c"))
	printed := s.assertRoundTrip(expr)
	s.Equal(`a \ b \ c`, printed)
}

func (s *RaiseComprehensiveTestSuite) TestAssocLeftBagSumChain() {
	// (a ⊕ b) ⊕ c — left-assoc
	a, b, c := localVar("a"), localVar("b"), localVar("c")
	expr := &me.BagOp{Op: me.BagSum, Left: &me.BagOp{Op: me.BagSum, Left: a, Right: b}, Right: c}
	printed := s.assertRoundTrip(expr)
	s.Equal("a ⊕ b ⊕ c", printed)
}

// --- Right-associative chains ---

func (s *RaiseComprehensiveTestSuite) TestAssocRightImpliesChain() {
	// a ⇒ (b ⇒ c) — right-assoc, no parens
	expr := implies(boolLit(true), implies(boolLit(false), boolLit(true)))
	printed := s.assertRoundTrip(expr)
	s.Equal("TRUE ⇒ FALSE ⇒ TRUE", printed)
}

func (s *RaiseComprehensiveTestSuite) TestAssocRightImpliesChainLeft() {
	// (a ⇒ b) ⇒ c — left child same prec right-assoc, needs parens
	expr := implies(implies(boolLit(true), boolLit(false)), boolLit(true))
	printed := s.assertRoundTrip(expr)
	s.Equal("(TRUE ⇒ FALSE) ⇒ TRUE", printed)
}

func (s *RaiseComprehensiveTestSuite) TestAssocRightPowChain() {
	// a ^ (b ^ c) — right-assoc, no parens
	expr := pow(intLit(2), pow(intLit(3), intLit(4)))
	printed := s.assertRoundTrip(expr)
	s.Equal("2 ^ 3 ^ 4", printed)
}

func (s *RaiseComprehensiveTestSuite) TestAssocRightPowChainLeft() {
	// (a ^ b) ^ c — left child same prec right-assoc, needs parens
	expr := pow(pow(intLit(2), intLit(3)), intLit(4))
	printed := s.assertRoundTrip(expr)
	s.Equal("(2 ^ 3) ^ 4", printed)
}

// --- Non-associative operators ---

func (s *RaiseComprehensiveTestSuite) TestAssocNoneCompareInCompare() {
	// (a > b) > c — non-assoc, needs parens on both sides
	expr := gt(gt(localVar("x"), intLit(1)), intLit(0))
	printed := s.assertRoundTrip(expr)
	s.Equal("(x > 1) > 0", printed)
}

func (s *RaiseComprehensiveTestSuite) TestAssocNoneEqInEq() {
	// (a = b) = c — non-assoc, needs parens
	expr := eq(eq(localVar("x"), intLit(1)), boolLit(true))
	printed := s.assertRoundTrip(expr)
	s.Equal("(x = 1) = TRUE", printed)
}

// =============================================================================
// SECTION 3: Cross-Category Operator Interactions
// =============================================================================

func (s *RaiseComprehensiveTestSuite) TestCross_ArithmeticInsideComparison() {
	// x + 1 > y * 2 — arithmetic higher than comparison
	expr := gt(add(localVar("x"), intLit(1)), mul(localVar("y"), intLit(2)))
	printed := s.assertRoundTrip(expr)
	s.Equal("x + 1 > y * 2", printed)
}

func (s *RaiseComprehensiveTestSuite) TestCross_ComparisonInsideLogic() {
	// (x > 0) ∧ (y < 10) ∨ (z = 5)
	expr := or(
		and(gt(localVar("x"), intLit(0)), lt(localVar("y"), intLit(10))),
		eq(localVar("z"), intLit(5)),
	)
	printed := s.assertRoundTrip(expr)
	s.Equal("x > 0 ∧ y < 10 ∨ z = 5", printed)
}

func (s *RaiseComprehensiveTestSuite) TestCross_SetOpsInsideCompare() {
	// (A ∪ B) ⊆ C — set union is higher than set compare
	expr := &me.SetCompare{
		Op:    me.SetCompareSubsetEq,
		Left:  setUnion(localVar("a"), localVar("b")),
		Right: localVar("c"),
	}
	printed := s.assertRoundTrip(expr)
	s.Equal("a ∪ b ⊆ c", printed)
}

func (s *RaiseComprehensiveTestSuite) TestCross_MembershipInsideLogic() {
	// (x ∈ S) ∧ (y ∉ T) — membership higher than logic
	expr := and(
		&me.Membership{Element: localVar("x"), Set: localVar("s"), Negated: false},
		&me.Membership{Element: localVar("y"), Set: localVar("t"), Negated: true},
	)
	printed := s.assertRoundTrip(expr)
	s.Equal("x ∈ s ∧ y ∉ t", printed)
}

func (s *RaiseComprehensiveTestSuite) TestCross_SetRangeInsideMembership() {
	// x ∈ 1..10 — range higher than membership
	expr := &me.Membership{
		Element: localVar("x"),
		Set:     &me.SetRange{Start: intLit(1), End: intLit(10)},
		Negated: false,
	}
	printed := s.assertRoundTrip(expr)
	s.Equal("x ∈ 1..10", printed)
}

func (s *RaiseComprehensiveTestSuite) TestCross_ArithmeticInsideSetRange() {
	// (n-1)..(n+1) — arithmetic higher than range
	expr := &me.SetRange{
		Start: sub(localVar("n"), intLit(1)),
		End:   add(localVar("n"), intLit(1)),
	}
	printed := s.assertRoundTrip(expr)
	s.Equal("n - 1..n + 1", printed)
}

func (s *RaiseComprehensiveTestSuite) TestCross_BagOpsInsideBagCompare() {
	// (A ⊕ B) ⊑ C — bag sum higher than bag compare
	expr := &me.BagCompare{
		Op:    me.BagCompareSubBag,
		Left:  &me.BagOp{Op: me.BagSum, Left: localVar("a"), Right: localVar("b")},
		Right: localVar("c"),
	}
	printed := s.assertRoundTrip(expr)
	s.Equal("a ⊕ b ⊑ c", printed)
}

func (s *RaiseComprehensiveTestSuite) TestCross_ConcatInsideBagDiff() {
	// concat higher than bag diff — no parens
	expr := &me.BagOp{
		Op: me.BagDifference,
		Left: &me.TupleConcat{
			Operands: []me.Expression{localVar("a"), localVar("b")},
		},
		Right: localVar("c"),
	}
	printed := s.assertRoundTrip(expr)
	s.Equal("a ∘ b ⊖ c", printed)
}

func (s *RaiseComprehensiveTestSuite) TestCross_NegationOfPower() {
	// Pow prec (27) > negate prec (22) — pow binds tighter, no parens needed
	expr := neg(pow(intLit(2), intLit(3)))
	printed := s.assertRoundTrip(expr)
	s.Equal("-2 ^ 3", printed)
}

func (s *RaiseComprehensiveTestSuite) TestCross_NegationOfAdd() {
	// -(x + 1) — negation higher than add, needs parens
	expr := neg(add(localVar("x"), intLit(1)))
	printed := s.assertRoundTrip(expr)
	s.Equal("-(x + 1)", printed)
}

func (s *RaiseComprehensiveTestSuite) TestCross_NegationThenMul() {
	// (-x) * y — negation lower than mul (in PEG), needs parens for mul's left child
	expr := mul(neg(localVar("x")), localVar("y"))
	printed := s.assertRoundTrip(expr)
	s.Equal("(-x) * y", printed)
}

func (s *RaiseComprehensiveTestSuite) TestCross_NegationThenAdd() {
	// -x + y — negation higher than add, no parens needed
	expr := add(neg(localVar("x")), localVar("y"))
	printed := s.assertRoundTrip(expr)
	s.Equal("-x + y", printed)
}

func (s *RaiseComprehensiveTestSuite) TestCross_FieldAccessInArithmetic() {
	// self.balance + amount — field access highest
	attrKey := s.lowerCtx.AttributeNames["balance"]
	expr := add(
		&me.FieldAccess{Base: &me.SelfRef{}, Field: "balance"},
		localVar("amount"),
	)
	_ = attrKey
	printed := s.assertRoundTrip(expr)
	s.Equal("self.balance + amount", printed)
}

func (s *RaiseComprehensiveTestSuite) TestCross_TupleIndexInComparison() {
	// t[1] > t[2]
	expr := gt(
		&me.TupleIndex{Tuple: localVar("t"), Index: intLit(1)},
		&me.TupleIndex{Tuple: localVar("t"), Index: intLit(2)},
	)
	printed := s.assertRoundTrip(expr)
	s.Equal("t[1] > t[2]", printed)
}

func (s *RaiseComprehensiveTestSuite) TestCross_NextStateInEquality() {
	// balance' = balance + amount
	attrKey := s.lowerCtx.AttributeNames["balance"]
	expr := eq(
		&me.NextState{Expr: &me.AttributeRef{AttributeKey: attrKey}},
		add(&me.AttributeRef{AttributeKey: attrKey}, localVar("amount")),
	)
	printed := s.assertRoundTrip(expr)
	s.Equal("balance' = balance + amount", printed)
}

func (s *RaiseComprehensiveTestSuite) TestCross_LogicInsideImplies() {
	// (a ∧ b) ⇒ (c ∨ d) — and/or higher than implies
	expr := implies(
		and(localVar("a"), localVar("b")),
		or(localVar("c"), localVar("d")),
	)
	printed := s.assertRoundTrip(expr)
	s.Equal("a ∧ b ⇒ c ∨ d", printed)
}

func (s *RaiseComprehensiveTestSuite) TestCross_NotAndOrImplies() {
	// ¬a ∧ b ∨ c ⇒ d — full precedence chain
	expr := implies(
		or(
			and(not(localVar("a")), localVar("b")),
			localVar("c"),
		),
		localVar("d"),
	)
	printed := s.assertRoundTrip(expr)
	s.Equal("¬a ∧ b ∨ c ⇒ d", printed)
}

// =============================================================================
// SECTION 4: Deep Nesting and Complex Expressions
// =============================================================================

func (s *RaiseComprehensiveTestSuite) TestDeep_FiveLevelArithmetic() {
	// 1 + 2 * 3 ^ 4 ÷ 5 - 6
	// = 1 + (((2 * (3 ^ 4)) ÷ 5) - 6) based on precedence
	expr := add(
		intLit(1),
		sub(
			div(
				mul(intLit(2), pow(intLit(3), intLit(4))),
				intLit(5),
			),
			intLit(6),
		),
	)
	printed := s.assertRoundTrip(expr)
	s.Equal("1 + 2 * 3 ^ 4 ÷ 5 - 6", printed)
}

func (s *RaiseComprehensiveTestSuite) TestDeep_NestedQuantifiers() {
	// ∀ x ∈ Nat : ∃ y ∈ Nat : x + y > 0
	expr := &me.Quantifier{
		Kind:     me.QuantifierForall,
		Variable: "x",
		Domain:   &me.SetConstant{Kind: me.SetConstantNat},
		Predicate: &me.Quantifier{
			Kind:     me.QuantifierExists,
			Variable: "y",
			Domain:   &me.SetConstant{Kind: me.SetConstantNat},
			Predicate: gt(
				add(localVar("x"), localVar("y")),
				intLit(0),
			),
		},
	}
	printed := s.assertRoundTrip(expr)
	s.Equal("∀ x ∈ Nat : ∃ y ∈ Nat : x + y > 0", printed)
}

func (s *RaiseComprehensiveTestSuite) TestDeep_QuantifierWithComplexPredicate() {
	// ∀ x ∈ Nat : x > 0 ⇒ x ≥ 1
	expr := &me.Quantifier{
		Kind:     me.QuantifierForall,
		Variable: "x",
		Domain:   &me.SetConstant{Kind: me.SetConstantNat},
		Predicate: implies(
			gt(localVar("x"), intLit(0)),
			gte(localVar("x"), intLit(1)),
		),
	}
	printed := s.assertRoundTrip(expr)
	s.Equal("∀ x ∈ Nat : x > 0 ⇒ x ≥ 1", printed)
}

func (s *RaiseComprehensiveTestSuite) TestDeep_NestedIfThenElse() {
	// IF a > 0 THEN IF b > 0 THEN 1 ELSE 2 ELSE 3
	expr := &me.IfThenElse{
		Condition: gt(localVar("a"), intLit(0)),
		Then: &me.IfThenElse{
			Condition: gt(localVar("b"), intLit(0)),
			Then:      intLit(1),
			Else:      intLit(2),
		},
		Else: intLit(3),
	}
	printed := s.assertRoundTrip(expr)
	s.Equal("IF a > 0 THEN IF b > 0 THEN 1 ELSE 2 ELSE 3", printed)
}

func (s *RaiseComprehensiveTestSuite) TestDeep_CaseWithArithmetic() {
	// CASE x > 0 → x * 2 □ x < 0 → -x □ OTHER → 0
	expr := &me.Case{
		Branches: []me.CaseBranch{
			{
				Condition: gt(localVar("x"), intLit(0)),
				Result:    mul(localVar("x"), intLit(2)),
			},
			{
				Condition: lt(localVar("x"), intLit(0)),
				Result:    neg(localVar("x")),
			},
		},
		Otherwise: intLit(0),
	}
	printed := s.assertRoundTrip(expr)
	s.Equal("CASE x > 0 → x * 2 □ x < 0 → -x □ OTHER → 0", printed)
}

func (s *RaiseComprehensiveTestSuite) TestDeep_RecordUpdateWithArithmetic() {
	// [self EXCEPT !.balance = @ + amount * 2, !.count = @ + 1]
	expr := &me.RecordUpdate{
		Base: &me.SelfRef{},
		Alterations: []me.FieldAlteration{
			{
				Field: "balance",
				Value: add(
					&me.PriorFieldValue{Field: "balance"},
					mul(localVar("amount"), intLit(2)),
				),
			},
			{
				Field: "count",
				Value: add(
					&me.PriorFieldValue{Field: "count"},
					intLit(1),
				),
			},
		},
	}
	printed := s.assertRoundTrip(expr)
	s.Equal("[self EXCEPT !.balance = @ + amount * 2, !.count = @ + 1]", printed)
}

func (s *RaiseComprehensiveTestSuite) TestDeep_SetOfArithmeticExpressions() {
	// {x + 1, x * 2, x ^ 3}
	expr := &me.SetLiteral{
		Elements: []me.Expression{
			add(localVar("x"), intLit(1)),
			mul(localVar("x"), intLit(2)),
			pow(localVar("x"), intLit(3)),
		},
	}
	printed := s.assertRoundTrip(expr)
	s.Equal("{x + 1, x * 2, x ^ 3}", printed)
}

func (s *RaiseComprehensiveTestSuite) TestDeep_TupleOfComparisons() {
	// <<x > 0, y < 10, z = 5>>
	expr := &me.TupleLiteral{
		Elements: []me.Expression{
			gt(localVar("x"), intLit(0)),
			lt(localVar("y"), intLit(10)),
			eq(localVar("z"), intLit(5)),
		},
	}
	printed := s.assertRoundTrip(expr)
	s.Equal("⟨x > 0, y < 10, z = 5⟩", printed)
}

func (s *RaiseComprehensiveTestSuite) TestDeep_RecordWithExpressionValues() {
	// [x ↦ a + b, y ↦ c * d, z ↦ ¬p]
	expr := &me.RecordLiteral{
		Fields: []me.RecordField{
			{Name: "x", Value: add(localVar("a"), localVar("b"))},
			{Name: "y", Value: mul(localVar("c"), localVar("d"))},
			{Name: "z", Value: not(localVar("p"))},
		},
	}
	printed := s.assertRoundTrip(expr)
	s.Equal("[x ↦ a + b, y ↦ c * d, z ↦ ¬p]", printed)
}

func (s *RaiseComprehensiveTestSuite) TestDeep_IfWithQuantifierCondition() {
	// IF ∀ x ∈ Nat : x ≥ 0 THEN 1 ELSE 0
	expr := &me.IfThenElse{
		Condition: &me.Quantifier{
			Kind:      me.QuantifierForall,
			Variable:  "x",
			Domain:    &me.SetConstant{Kind: me.SetConstantNat},
			Predicate: gte(localVar("x"), intLit(0)),
		},
		Then: intLit(1),
		Else: intLit(0),
	}
	printed := s.assertRoundTrip(expr)
	s.Equal("IF ∀ x ∈ Nat : x ≥ 0 THEN 1 ELSE 0", printed)
}

func (s *RaiseComprehensiveTestSuite) TestDeep_SetLiteralInsideMembership() {
	// x ∈ {1, 2, 3} ∧ y ∉ {4, 5}
	expr := and(
		&me.Membership{
			Element: localVar("x"),
			Set: &me.SetLiteral{Elements: []me.Expression{
				intLit(1), intLit(2), intLit(3),
			}},
			Negated: false,
		},
		&me.Membership{
			Element: localVar("y"),
			Set: &me.SetLiteral{Elements: []me.Expression{
				intLit(4), intLit(5),
			}},
			Negated: true,
		},
	)
	printed := s.assertRoundTrip(expr)
	s.Equal("x ∈ {1, 2, 3} ∧ y ∉ {4, 5}", printed)
}

func (s *RaiseComprehensiveTestSuite) TestDeep_QuantifierOverSetRange() {
	// ∀ i ∈ 1..10 : i ≥ 1 ∧ i ≤ 10
	expr := &me.Quantifier{
		Kind:     me.QuantifierForall,
		Variable: "i",
		Domain:   &me.SetRange{Start: intLit(1), End: intLit(10)},
		Predicate: and(
			gte(localVar("i"), intLit(1)),
			lte(localVar("i"), intLit(10)),
		),
	}
	printed := s.assertRoundTrip(expr)
	s.Equal("∀ i ∈ 1..10 : i ≥ 1 ∧ i ≤ 10", printed)
}

func (s *RaiseComprehensiveTestSuite) TestDeep_NestedFunctionCalls() {
	// _Seq!Len(_Seq!Tail(items))
	attrKey := s.lowerCtx.AttributeNames["items"]
	expr := &me.BuiltinCall{
		Module:   "_Seq",
		Function: "Len",
		Args: []me.Expression{
			&me.BuiltinCall{
				Module:   "_Seq",
				Function: "Tail",
				Args:     []me.Expression{&me.AttributeRef{AttributeKey: attrKey}},
			},
		},
	}
	printed := s.assertRoundTrip(expr)
	s.Equal("_Seq!Len(_Seq!Tail(items))", printed)
}

func (s *RaiseComprehensiveTestSuite) TestDeep_FunctionCallInComparison() {
	// _Seq!Len(items) > 0 ∧ _Seq!Len(items) ≤ limit
	attrKey := s.lowerCtx.AttributeNames["items"]
	lenCall := &me.BuiltinCall{
		Module:   "_Seq",
		Function: "Len",
		Args:     []me.Expression{&me.AttributeRef{AttributeKey: attrKey}},
	}
	expr := and(
		gt(lenCall, intLit(0)),
		lte(
			&me.BuiltinCall{
				Module:   "_Seq",
				Function: "Len",
				Args:     []me.Expression{&me.AttributeRef{AttributeKey: attrKey}},
			},
			localVar("limit"),
		),
	)
	printed := s.assertRoundTrip(expr)
	s.Equal("_Seq!Len(items) > 0 ∧ _Seq!Len(items) ≤ limit", printed)
}

func (s *RaiseComprehensiveTestSuite) TestDeep_CrossClassCallInExpression() {
	// s2!c2!OtherAction(balance + 1) > 0
	crossKey := s.lowerCtx.AllActions["s2!c2!OtherAction"]
	attrKey := s.lowerCtx.AttributeNames["balance"]
	expr := gt(
		&me.ActionCall{
			ActionKey: crossKey,
			Args:      []me.Expression{add(&me.AttributeRef{AttributeKey: attrKey}, intLit(1))},
		},
		intLit(0),
	)
	printed := s.assertRoundTrip(expr)
	s.Equal("s2!c2!OtherAction(balance + 1) > 0", printed)
}

func (s *RaiseComprehensiveTestSuite) TestDeep_MembershipInQuantifierDomain() {
	// ∃ x ∈ {1, 2, 3, 4, 5} : x * x = n
	expr := &me.Quantifier{
		Kind:     me.QuantifierExists,
		Variable: "x",
		Domain: &me.SetLiteral{
			Elements: []me.Expression{intLit(1), intLit(2), intLit(3), intLit(4), intLit(5)},
		},
		Predicate: eq(mul(localVar("x"), localVar("x")), localVar("n")),
	}
	printed := s.assertRoundTrip(expr)
	s.Equal("∃ x ∈ {1, 2, 3, 4, 5} : x * x = n", printed)
}

func (s *RaiseComprehensiveTestSuite) TestDeep_ConcatThreeWay() {
	// a ∘ b ∘ c — three-way concat
	expr := &me.TupleConcat{
		Operands: []me.Expression{localVar("a"), localVar("b"), localVar("c")},
	}
	printed := s.assertRoundTrip(expr)
	s.Equal("a ∘ b ∘ c", printed)
}

func (s *RaiseComprehensiveTestSuite) TestDeep_TupleConcatWithTupleLiterals() {
	// <<1, 2>> ∘ <<3, 4>> ∘ <<5>>
	expr := &me.TupleConcat{
		Operands: []me.Expression{
			&me.TupleLiteral{Elements: []me.Expression{intLit(1), intLit(2)}},
			&me.TupleLiteral{Elements: []me.Expression{intLit(3), intLit(4)}},
			&me.TupleLiteral{Elements: []me.Expression{intLit(5)}},
		},
	}
	printed := s.assertRoundTrip(expr)
	s.Equal("⟨1, 2⟩ ∘ ⟨3, 4⟩ ∘ ⟨5⟩", printed)
}

func (s *RaiseComprehensiveTestSuite) TestDeep_ChainedFieldAccess() {
	// self.items — simple field access
	expr := &me.FieldAccess{
		Base: &me.FieldAccess{
			Base:  &me.SelfRef{},
			Field: "balance",
		},
		Field: "count",
	}
	// This raises to self.balance.count — but lowering resolves differently.
	// Since balance resolves to an attribute, this won't round-trip via assertRoundTrip.
	// Test raise + print only.
	raised, err := Raise(expr, s.raiseCtx)
	s.Require().NoError(err)
	printed := ast.Print(raised)
	s.Equal("self.balance.count", printed)
}

func (s *RaiseComprehensiveTestSuite) TestDeep_TupleIndexOnFieldAccess() {
	// self.items[1]
	expr := &me.TupleIndex{
		Tuple: &me.FieldAccess{
			Base:  &me.SelfRef{},
			Field: "items",
		},
		Index: intLit(1),
	}
	// Raise + print check
	raised, err := Raise(expr, s.raiseCtx)
	s.Require().NoError(err)
	printed := ast.Print(raised)
	s.Equal("self.items[1]", printed)
}

func (s *RaiseComprehensiveTestSuite) TestDeep_NestedSetOps() {
	// Union(14) > Intersect(13) — no parens; SetDiff(12) < Intersect(13) — needs parens
	expr := setIntersect(
		setUnion(localVar("a"), localVar("b")),
		setDiff(localVar("c"), localVar("d")),
	)
	printed := s.assertRoundTrip(expr)
	s.Equal(`a ∪ b ∩ (c \ d)`, printed)
}

func (s *RaiseComprehensiveTestSuite) TestDeep_ComplexConditionInCase() {
	// CASE conditions with logic ops — tests CASE wrapping at OrExpr level
	expr := &me.Case{
		Branches: []me.CaseBranch{
			{
				Condition: and(gt(localVar("x"), intLit(0)), lt(localVar("x"), intLit(100))),
				Result:    strLit("valid"),
			},
			{
				Condition: or(lte(localVar("x"), intLit(0)), gte(localVar("x"), intLit(100))),
				Result:    strLit("invalid"),
			},
		},
		Otherwise: strLit("unknown"),
	}
	printed := s.assertRoundTrip(expr)
	s.Equal(`CASE x > 0 ∧ x < 100 → "valid" □ x ≤ 0 ∨ x ≥ 100 → "invalid" □ OTHER → "unknown"`, printed)
}

// =============================================================================
// SECTION 5: Edge Cases and Special Combinations
// =============================================================================

func (s *RaiseComprehensiveTestSuite) TestEdge_EmptySetInMembership() {
	// x ∈ {}
	expr := &me.Membership{
		Element: localVar("x"),
		Set:     &me.SetLiteral{Elements: []me.Expression{}},
		Negated: false,
	}
	printed := s.assertRoundTrip(expr)
	s.Equal("x ∈ {}", printed)
}

func (s *RaiseComprehensiveTestSuite) TestEdge_SingleElementTuple() {
	// <<42>>
	expr := &me.TupleLiteral{Elements: []me.Expression{intLit(42)}}
	printed := s.assertRoundTrip(expr)
	s.Equal("⟨42⟩", printed)
}

func (s *RaiseComprehensiveTestSuite) TestEdge_DoubleNegation() {
	// ¬¬TRUE
	expr := not(not(boolLit(true)))
	printed := s.assertRoundTrip(expr)
	s.Equal("¬¬TRUE", printed)
}

func (s *RaiseComprehensiveTestSuite) TestEdge_NegateNegate() {
	// --x = -(-x) arithmetic double negation
	expr := neg(neg(localVar("x")))
	printed := s.assertRoundTrip(expr)
	s.Equal("--x", printed)
}

func (s *RaiseComprehensiveTestSuite) TestEdge_ZeroLiteral() {
	expr := intLit(0)
	printed := s.assertRoundTrip(expr)
	s.Equal("0", printed)
}

func (s *RaiseComprehensiveTestSuite) TestEdge_LargeIntLiteral() {
	big := new(big.Int)
	big.SetString("999999999999999999999", 10)
	expr := &me.IntLiteral{Value: big}
	printed := s.assertRoundTrip(expr)
	s.Equal("999999999999999999999", printed)
}

func (s *RaiseComprehensiveTestSuite) TestEdge_StringWithSpecialChars() {
	expr := strLit("line1\nline2\ttab")
	result, printed := s.roundTrip(expr)
	s.Equal("\"line1\\nline2\\ttab\"", printed)
	s.Equal(expr, result)
}

func (s *RaiseComprehensiveTestSuite) TestEdge_EmptyString() {
	expr := strLit("")
	printed := s.assertRoundTrip(expr)
	s.Equal(`""`, printed)
}

func (s *RaiseComprehensiveTestSuite) TestEdge_SetConstantsInArithmetic() {
	// Nat and Int as operands in comparison
	expr := &me.Membership{
		Element: localVar("x"),
		Set:     &me.SetConstant{Kind: me.SetConstantInt},
		Negated: false,
	}
	printed := s.assertRoundTrip(expr)
	s.Equal("x ∈ Int", printed)
}

func (s *RaiseComprehensiveTestSuite) TestEdge_AllSetConstants() {
	for _, kind := range []me.SetConstantKind{
		me.SetConstantNat, me.SetConstantInt, me.SetConstantReal, me.SetConstantBoolean,
	} {
		expr := &me.Membership{
			Element: localVar("x"),
			Set:     &me.SetConstant{Kind: kind},
			Negated: false,
		}
		s.assertRoundTrip(expr)
	}
}

func (s *RaiseComprehensiveTestSuite) TestEdge_AllComparisonOps() {
	ops := []struct {
		op     me.CompareOp
		symbol string
	}{
		{me.CompareLt, "<"}, {me.CompareGt, ">"}, {me.CompareLte, "≤"},
		{me.CompareGte, "≥"}, {me.CompareEq, "="}, {me.CompareNeq, "≠"},
	}
	for _, tt := range ops {
		expr := &me.Compare{Op: tt.op, Left: localVar("x"), Right: intLit(5)}
		printed := s.assertRoundTrip(expr)
		s.Contains(printed, tt.symbol)
	}
}

func (s *RaiseComprehensiveTestSuite) TestEdge_AllSetCompareOps() {
	ops := []struct {
		op     me.SetCompareOp
		symbol string
	}{
		{me.SetCompareSubsetEq, "⊆"}, {me.SetCompareSubset, "⊂"},
		{me.SetCompareSupersetEq, "⊇"}, {me.SetCompareSuperset, "⊃"},
	}
	for _, tt := range ops {
		expr := &me.SetCompare{Op: tt.op, Left: localVar("a"), Right: localVar("b")}
		printed := s.assertRoundTrip(expr)
		s.Contains(printed, tt.symbol)
	}
}

func (s *RaiseComprehensiveTestSuite) TestEdge_AllBagCompareOps() {
	ops := []struct {
		op     me.BagCompareOp
		symbol string
	}{
		{me.BagCompareProperSubBag, "⊏"}, {me.BagCompareSubBag, "⊑"},
		{me.BagCompareProperSupBag, "⊐"}, {me.BagCompareSupBag, "⊒"},
	}
	for _, tt := range ops {
		expr := &me.BagCompare{Op: tt.op, Left: localVar("a"), Right: localVar("b")}
		printed := s.assertRoundTrip(expr)
		s.Contains(printed, tt.symbol)
	}
}

func (s *RaiseComprehensiveTestSuite) TestEdge_IfThenElseInSetLiteral() {
	// {IF x > 0 THEN 1 ELSE 0, 2}
	expr := &me.SetLiteral{
		Elements: []me.Expression{
			&me.IfThenElse{
				Condition: gt(localVar("x"), intLit(0)),
				Then:      intLit(1),
				Else:      intLit(0),
			},
			intLit(2),
		},
	}
	printed := s.assertRoundTrip(expr)
	s.Equal("{IF x > 0 THEN 1 ELSE 0, 2}", printed)
}

func (s *RaiseComprehensiveTestSuite) TestEdge_CaseInsideIfThenElse() {
	// IF flag THEN CASE x > 0 → 1 □ OTHER → 0 ELSE 42
	expr := &me.IfThenElse{
		Condition: localVar("a"),
		Then: &me.Case{
			Branches: []me.CaseBranch{
				{Condition: gt(localVar("x"), intLit(0)), Result: intLit(1)},
			},
			Otherwise: intLit(0),
		},
		Else: intLit(42),
	}
	printed := s.assertRoundTrip(expr)
	s.Equal("IF a THEN CASE x > 0 → 1 □ OTHER → 0 ELSE 42", printed)
}

func (s *RaiseComprehensiveTestSuite) TestEdge_NestedRecordLiterals() {
	// [a ↦ [b ↦ [c ↦ 1]]]
	expr := &me.RecordLiteral{
		Fields: []me.RecordField{
			{
				Name: "a",
				Value: &me.RecordLiteral{
					Fields: []me.RecordField{
						{
							Name: "b",
							Value: &me.RecordLiteral{
								Fields: []me.RecordField{
									{Name: "c", Value: intLit(1)},
								},
							},
						},
					},
				},
			},
		},
	}
	printed := s.assertRoundTrip(expr)
	s.Equal("[a ↦ [b ↦ [c ↦ 1]]]", printed)
}

func (s *RaiseComprehensiveTestSuite) TestEdge_NestedTupleLiterals() {
	// <<1, <<2, <<3, 4>>>>, 5>>
	expr := &me.TupleLiteral{
		Elements: []me.Expression{
			intLit(1),
			&me.TupleLiteral{
				Elements: []me.Expression{
					intLit(2),
					&me.TupleLiteral{Elements: []me.Expression{intLit(3), intLit(4)}},
				},
			},
			intLit(5),
		},
	}
	printed := s.assertRoundTrip(expr)
	s.Equal("⟨1, ⟨2, ⟨3, 4⟩⟩, 5⟩", printed)
}

func (s *RaiseComprehensiveTestSuite) TestEdge_SetWithRecordElements() {
	// {[x ↦ 1], [x ↦ 2]}
	expr := &me.SetLiteral{
		Elements: []me.Expression{
			&me.RecordLiteral{Fields: []me.RecordField{{Name: "x", Value: intLit(1)}}},
			&me.RecordLiteral{Fields: []me.RecordField{{Name: "x", Value: intLit(2)}}},
		},
	}
	printed := s.assertRoundTrip(expr)
	s.Equal("{[x ↦ 1], [x ↦ 2]}", printed)
}

func (s *RaiseComprehensiveTestSuite) TestEdge_FunctionCallWithComplexArgs() {
	// Deposit(balance + amount * rate, x > 0)
	attrKey := s.lowerCtx.AttributeNames["balance"]
	actionKey := s.lowerCtx.ActionNames["Deposit"]
	expr := &me.ActionCall{
		ActionKey: actionKey,
		Args: []me.Expression{
			add(&me.AttributeRef{AttributeKey: attrKey}, mul(localVar("amount"), localVar("rate"))),
			gt(localVar("x"), intLit(0)),
		},
	}
	printed := s.assertRoundTrip(expr)
	s.Equal("Deposit(balance + amount * rate, x > 0)", printed)
}

func (s *RaiseComprehensiveTestSuite) TestEdge_GlobalCallNoArgs() {
	// _Helper() with no args
	globalKey := s.lowerCtx.GlobalFunctions["_Helper"]
	expr := &me.GlobalCall{
		FunctionKey: globalKey,
		Args:        []me.Expression{},
	}
	printed := s.assertRoundTrip(expr)
	s.Equal("_Helper()", printed)
}

func (s *RaiseComprehensiveTestSuite) TestEdge_MultipleGlobalCalls() {
	// _Max(_Helper(x), _Helper(y))
	globalMaxKey := s.lowerCtx.GlobalFunctions["_Max"]
	globalHelperKey := s.lowerCtx.GlobalFunctions["_Helper"]
	expr := &me.GlobalCall{
		FunctionKey: globalMaxKey,
		Args: []me.Expression{
			&me.GlobalCall{FunctionKey: globalHelperKey, Args: []me.Expression{localVar("x")}},
			&me.GlobalCall{FunctionKey: globalHelperKey, Args: []me.Expression{localVar("y")}},
		},
	}
	printed := s.assertRoundTrip(expr)
	s.Equal("_Max(_Helper(x), _Helper(y))", printed)
}

func (s *RaiseComprehensiveTestSuite) TestEdge_NamedSetInMembership() {
	// x ∈ valid_statuses
	setKey := s.lowerCtx.NamedSets["valid_statuses"]
	expr := &me.Membership{
		Element: localVar("x"),
		Set:     &me.NamedSetRef{SetKey: setKey},
		Negated: false,
	}
	printed := s.assertRoundTrip(expr)
	s.Equal("x ∈ valid_statuses", printed)
}

func (s *RaiseComprehensiveTestSuite) TestEdge_RecordUpdateMultipleAlterations() {
	// [self EXCEPT !.balance = @ - amount, !.count = @ + 1, !.total = @ + amount]
	expr := &me.RecordUpdate{
		Base: &me.SelfRef{},
		Alterations: []me.FieldAlteration{
			{Field: "balance", Value: sub(&me.PriorFieldValue{Field: "balance"}, localVar("amount"))},
			{Field: "count", Value: add(&me.PriorFieldValue{Field: "count"}, intLit(1))},
			{Field: "total", Value: add(&me.PriorFieldValue{Field: "total"}, localVar("amount"))},
		},
	}
	printed := s.assertRoundTrip(expr)
	s.Equal("[self EXCEPT !.balance = @ - amount, !.count = @ + 1, !.total = @ + amount]", printed)
}

func (s *RaiseComprehensiveTestSuite) TestEdge_CaseNoOtherwiseMultipleBranches() {
	// CASE a = 1 → 10 □ a = 2 → 20 □ a = 3 → 30
	expr := &me.Case{
		Branches: []me.CaseBranch{
			{Condition: eq(localVar("a"), intLit(1)), Result: intLit(10)},
			{Condition: eq(localVar("a"), intLit(2)), Result: intLit(20)},
			{Condition: eq(localVar("a"), intLit(3)), Result: intLit(30)},
		},
	}
	printed := s.assertRoundTrip(expr)
	s.Equal("CASE a = 1 → 10 □ a = 2 → 20 □ a = 3 → 30", printed)
}

// =============================================================================
// SECTION 6: Complex Real-World-Style Expressions
// =============================================================================

func (s *RaiseComprehensiveTestSuite) TestRealWorld_AccountDeposit() {
	// balance' = balance + amount ∧ amount > 0 ∧ balance + amount ≤ limit
	attrKey := s.lowerCtx.AttributeNames["balance"]
	expr := and(
		and(
			eq(
				&me.NextState{Expr: &me.AttributeRef{AttributeKey: attrKey}},
				add(&me.AttributeRef{AttributeKey: attrKey}, localVar("amount")),
			),
			gt(localVar("amount"), intLit(0)),
		),
		lte(
			add(&me.AttributeRef{AttributeKey: attrKey}, localVar("amount")),
			localVar("limit"),
		),
	)
	printed := s.assertRoundTrip(expr)
	s.Equal("balance' = balance + amount ∧ amount > 0 ∧ balance + amount ≤ limit", printed)
}

func (s *RaiseComprehensiveTestSuite) TestRealWorld_StateTransitionGuard() {
	// active = TRUE ∧ balance > 0 ⇒ balance' = balance - amount ∧ amount > 0
	attrActiveKey := s.lowerCtx.AttributeNames["active"]
	attrBalanceKey := s.lowerCtx.AttributeNames["balance"]
	expr := implies(
		and(
			eq(&me.AttributeRef{AttributeKey: attrActiveKey}, boolLit(true)),
			gt(&me.AttributeRef{AttributeKey: attrBalanceKey}, intLit(0)),
		),
		and(
			eq(
				&me.NextState{Expr: &me.AttributeRef{AttributeKey: attrBalanceKey}},
				sub(&me.AttributeRef{AttributeKey: attrBalanceKey}, localVar("amount")),
			),
			gt(localVar("amount"), intLit(0)),
		),
	)
	printed := s.assertRoundTrip(expr)
	s.Equal("active = TRUE ∧ balance > 0 ⇒ balance' = balance - amount ∧ amount > 0", printed)
}

func (s *RaiseComprehensiveTestSuite) TestRealWorld_Invariant() {
	// ∀ x ∈ all_accounts : x.balance ≥ 0
	// Since all_accounts resolves to named set, we use it.
	// For the predicate, x.balance is a field access on quantifier var.
	namedSetKey := s.lowerCtx.NamedSets["all_accounts"]
	expr := &me.Quantifier{
		Kind:     me.QuantifierForall,
		Variable: "x",
		Domain:   &me.NamedSetRef{SetKey: namedSetKey},
		Predicate: gte(
			&me.FieldAccess{Base: localVar("x"), Field: "balance"},
			intLit(0),
		),
	}
	// Raise + print only (field access on quantifier var doesn't lower to AttributeRef)
	raised, err := Raise(expr, s.raiseCtx)
	s.Require().NoError(err)
	printed := ast.Print(raised)
	s.Equal("∀ x ∈ all_accounts : x.balance ≥ 0", printed)
}

func (s *RaiseComprehensiveTestSuite) TestRealWorld_ComplexArithmeticChain() {
	// sub(21) > mod(18), mul(25) > sub(21), div(23) > sub(21), pow(27) > mod(18) — no outer parens
	expr := mod(
		sub(
			mul(add(localVar("a"), localVar("b")), localVar("c")),
			div(localVar("d"), localVar("x")),
		),
		pow(localVar("y"), localVar("z")),
	)
	printed := s.assertRoundTrip(expr)
	s.Equal("(a + b) * c - d ÷ x % y ^ z", printed)
}

func (s *RaiseComprehensiveTestSuite) TestRealWorld_IfWithRecordUpdate() {
	// IF amount > 0 THEN [self EXCEPT !.balance = @ + amount] ELSE self
	expr := &me.IfThenElse{
		Condition: gt(localVar("amount"), intLit(0)),
		Then: &me.RecordUpdate{
			Base: &me.SelfRef{},
			Alterations: []me.FieldAlteration{
				{Field: "balance", Value: add(&me.PriorFieldValue{Field: "balance"}, localVar("amount"))},
			},
		},
		Else: &me.SelfRef{},
	}
	printed := s.assertRoundTrip(expr)
	s.Equal("IF amount > 0 THEN [self EXCEPT !.balance = @ + amount] ELSE self", printed)
}

func (s *RaiseComprehensiveTestSuite) TestRealWorld_SetOperationChain() {
	// Precedence: SetDiff(12) < Intersect(13) < Union(14), so union binds tightest
	// setDiff(setIntersect(setUnion(a,b), c), d) — intersect(13) > setDiff(12) no parens;
	// union(14) > intersect(13) no parens
	expr := setDiff(
		setIntersect(
			setUnion(localVar("a"), localVar("b")),
			localVar("c"),
		),
		localVar("d"),
	)
	printed := s.assertRoundTrip(expr)
	s.Equal(`a ∪ b ∩ c \ d`, printed)
}

func (s *RaiseComprehensiveTestSuite) TestRealWorld_QuantifierWithIfThenElse() {
	// ∀ x ∈ Nat : IF x > 0 THEN x * x > x ELSE x = 0
	expr := &me.Quantifier{
		Kind:     me.QuantifierForall,
		Variable: "x",
		Domain:   &me.SetConstant{Kind: me.SetConstantNat},
		Predicate: &me.IfThenElse{
			Condition: gt(localVar("x"), intLit(0)),
			Then:      gt(mul(localVar("x"), localVar("x")), localVar("x")),
			Else:      eq(localVar("x"), intLit(0)),
		},
	}
	printed := s.assertRoundTrip(expr)
	s.Equal("∀ x ∈ Nat : IF x > 0 THEN x * x > x ELSE x = 0", printed)
}

func (s *RaiseComprehensiveTestSuite) TestRealWorld_ExistsWithSetLiteralComplex() {
	// ∃ x ∈ {1, 2, 3} : ∃ y ∈ {4, 5, 6} : x + y = 7
	expr := &me.Quantifier{
		Kind:     me.QuantifierExists,
		Variable: "x",
		Domain:   &me.SetLiteral{Elements: []me.Expression{intLit(1), intLit(2), intLit(3)}},
		Predicate: &me.Quantifier{
			Kind:      me.QuantifierExists,
			Variable:  "y",
			Domain:    &me.SetLiteral{Elements: []me.Expression{intLit(4), intLit(5), intLit(6)}},
			Predicate: eq(add(localVar("x"), localVar("y")), intLit(7)),
		},
	}
	printed := s.assertRoundTrip(expr)
	s.Equal("∃ x ∈ {1, 2, 3} : ∃ y ∈ {4, 5, 6} : x + y = 7", printed)
}

// =============================================================================
// SECTION 7: Raise-Only Tests (expressions that don't perfectly round-trip)
// =============================================================================

func (s *RaiseComprehensiveTestSuite) TestRaiseOnly_SetFilterWithComplexPredicate() {
	// {x ∈ Nat : x > 0 ∧ x < 100}
	expr := &me.SetFilter{
		Variable: "x",
		Set:      &me.SetConstant{Kind: me.SetConstantNat},
		Predicate: and(
			gt(localVar("x"), intLit(0)),
			lt(localVar("x"), intLit(100)),
		),
	}
	raised, err := Raise(expr, s.raiseCtx)
	s.Require().NoError(err)
	printed := ast.Print(raised)
	s.Equal("{x ∈ Nat : x > 0 ∧ x < 100}", printed)
}

func (s *RaiseComprehensiveTestSuite) TestRaiseOnly_SetFilterNestedQuantifier() {
	// {x ∈ Int : ∃ y ∈ Nat : x = y * y}
	expr := &me.SetFilter{
		Variable: "x",
		Set:      &me.SetConstant{Kind: me.SetConstantInt},
		Predicate: &me.Quantifier{
			Kind:      me.QuantifierExists,
			Variable:  "y",
			Domain:    &me.SetConstant{Kind: me.SetConstantNat},
			Predicate: eq(localVar("x"), mul(localVar("y"), localVar("y"))),
		},
	}
	raised, err := Raise(expr, s.raiseCtx)
	s.Require().NoError(err)
	printed := ast.Print(raised)
	s.Equal("{x ∈ Int : ∃ y ∈ Nat : x = y * y}", printed)
}

func (s *RaiseComprehensiveTestSuite) TestRaiseOnly_StringConcatRaisesToTupleConcat() {
	// StringConcat raises to TupleConcat in AST — round-trip gives TupleConcat
	expr := &me.StringConcat{
		Operands: []me.Expression{strLit("hello"), strLit(" "), strLit("world")},
	}
	raised, err := Raise(expr, s.raiseCtx)
	s.Require().NoError(err)
	tc, ok := raised.(*ast.TupleConcat)
	s.True(ok, "expected TupleConcat, got %T", raised)
	s.Len(tc.Operands, 3)
	printed := ast.Print(raised)
	s.Equal(`"hello" ∘ " " ∘ "world"`, printed)
}

func (s *RaiseComprehensiveTestSuite) TestRaiseOnly_NegativeRational() {
	// -3/4 as rational — Fraction prec (26) > negate prec (22), no parens needed
	rat := new(big.Rat).SetFrac64(-3, 4)
	expr := &me.RationalLiteral{Value: rat}
	raised, err := Raise(expr, s.raiseCtx)
	s.Require().NoError(err)
	printed := ast.Print(raised)
	s.Equal("-3 / 4", printed)
}

func (s *RaiseComprehensiveTestSuite) TestRaiseOnly_RationalNormalization() {
	// 6/4 normalizes to 3/2
	rat := new(big.Rat).SetFrac64(6, 4)
	expr := &me.RationalLiteral{Value: rat}
	raised, err := Raise(expr, s.raiseCtx)
	s.Require().NoError(err)
	printed := ast.Print(raised)
	s.Equal("3 / 2", printed)
}

func (s *RaiseComprehensiveTestSuite) TestRaiseOnly_StringIndexPrintsAsTupleIndex() {
	// StringIndex raises like TupleIndex
	expr := &me.StringIndex{
		Str:   localVar("s"),
		Index: intLit(1),
	}
	raised, err := Raise(expr, s.raiseCtx)
	s.Require().NoError(err)
	printed := ast.Print(raised)
	s.Equal("s[1]", printed)
}

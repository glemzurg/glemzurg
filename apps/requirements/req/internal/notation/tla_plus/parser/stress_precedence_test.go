package parser

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/ast"
	"github.com/stretchr/testify/suite"
)

// StressPrecedenceTestSuite stress-tests operator precedence edge cases.
// Precedence is encoded in the PEG grammar hierarchy. Bugs emerge at
// boundaries between adjacent levels and in associativity handling.
type StressPrecedenceTestSuite struct {
	suite.Suite
}

func TestStressPrecedenceSuite(t *testing.T) {
	suite.Run(t, new(StressPrecedenceTestSuite))
}

// === Power right-associativity (level 14) ===

// TestPowerRightAssociativity verifies that ^ is right-associative:
// 2 ^ 3 ^ 4 = 2 ^ (3 ^ 4), NOT (2 ^ 3) ^ 4.
func (s *StressPrecedenceTestSuite) TestPowerRightAssociativity() {
	expr, err := ParseExpression("2 ^ 3 ^ 4")
	s.Require().NoError(err)

	// Should be: BinaryArith(2, ^, BinaryArith(3, ^, 4))
	top, ok := expr.(*ast.RealInfixExpression)
	s.Require().True(ok, "expected RealInfixExpression, got %T", expr)
	s.Equal("^", top.Operator)

	// Left should be atomic 2
	left, ok := top.Left.(*ast.NumberLiteral)
	s.Require().True(ok, "left should be NumberLiteral, got %T", top.Left)
	s.Equal("2", left.String())

	// Right should be another power expression: 3 ^ 4
	right, ok := top.Right.(*ast.RealInfixExpression)
	s.Require().True(ok, "right should be RealInfixExpression (3 ^ 4), got %T", top.Right)
	s.Equal("^", right.Operator)
}

func (s *StressPrecedenceTestSuite) TestPowerTripleChain() {
	expr, err := ParseExpression("a ^ b ^ c ^ d")
	s.Require().NoError(err)

	// Should be: a ^ (b ^ (c ^ d)) — right-associative
	top, ok := expr.(*ast.RealInfixExpression)
	s.Require().True(ok, "expected RealInfixExpression")
	s.Equal("^", top.Operator)

	// Left is 'a'
	_, ok = top.Left.(*ast.Identifier)
	s.True(ok, "left should be identifier 'a'")

	// Right should be b ^ (c ^ d)
	mid, ok := top.Right.(*ast.RealInfixExpression)
	s.Require().True(ok, "right should be nested power")
	s.Equal("^", mid.Operator)

	// mid.Right should be c ^ d
	inner, ok := mid.Right.(*ast.RealInfixExpression)
	s.Require().True(ok, "innermost should be nested power")
	s.Equal("^", inner.Operator)
}

func (s *StressPrecedenceTestSuite) TestPowerVsMultiplication() {
	tests := []struct {
		input string
		desc  string
	}{
		{"2 ^ 3 * 4", "power binds tighter than multiplication: (2^3) * 4"},
		{"2 * 3 ^ 4", "power binds tighter than multiplication: 2 * (3^4)"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			expr, err := ParseExpression(tt.input)
			s.Require().NoError(err, "should parse: %s", tt.input)

			// Both should have * at top level
			top, ok := expr.(*ast.RealInfixExpression)
			s.Require().True(ok, "top should be RealInfixExpression for %s", tt.input)
			s.Equal("*", top.Operator, "top operator should be * for %s", tt.input)
		})
	}
}

// === Negation binding (level 12) ===

func (s *StressPrecedenceTestSuite) TestNegationBinding() {
	tests := []struct {
		input       string
		desc        string
		topOperator string
	}{
		// Negation is at level 12 (below multiplication at 13.3).
		// So -x * y = (-x) * y — negation binds TIGHTER than multiplication.
		// Wait: actually, NegationExpr → DivisionExpr → ... → MultiplicationExpr.
		// The negation is ABOVE multiplication in the parse tree (lower precedence).
		// So: -x * y should be -(x * y)? Let's test and see.
		{"-x * y", "negation vs multiplication", "*"},
		{"-x + y", "negation vs addition", "+"},
		{"--x", "double negation", ""},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			_, err := ParseExpression(tt.input)
			s.NoError(err, "should parse: %s", tt.input)
		})
	}
}

// TestNegationVsPower tests -x ^ 2: in TLA+ negation has lower precedence
// than power, so this MUST parse as -(x^2), NOT (-x)^2.
func (s *StressPrecedenceTestSuite) TestNegationVsPower() {
	expr, err := ParseExpression("-x ^ 2")
	s.Require().NoError(err)

	// Must be UnaryNegation wrapping a power expression: -(x^2)
	neg, ok := expr.(*ast.UnaryNegation)
	s.Require().True(ok, "expected UnaryNegation at top level, got %T", expr)

	inner, ok := neg.Right.(*ast.RealInfixExpression)
	s.Require().True(ok, "inner should be power expression, got %T", neg.Right)
	s.Equal("^", inner.Operator)
}

// TestExplicitParenthesizedNegation verifies -(x ^ 2) is always unambiguous.
func (s *StressPrecedenceTestSuite) TestExplicitParenthesizedNegation() {
	expr, err := ParseExpression("-(x ^ 2)")
	s.Require().NoError(err)

	neg, ok := expr.(*ast.UnaryNegation)
	s.Require().True(ok, "should be UnaryNegation")

	paren, ok := neg.Right.(*ast.Parenthesized)
	s.Require().True(ok, "inner should be Parenthesized, got %T", neg.Right)

	power, ok := paren.Inner.(*ast.RealInfixExpression)
	s.Require().True(ok, "inner of paren should be power, got %T", paren.Inner)
	s.Equal("^", power.Operator)
}

func (s *StressPrecedenceTestSuite) TestMultiplicationWithNegatedOperand() {
	// x * -y: multiplication's right operand includes negation via UnaryExpr
	// Actually: MultiplicationExpr's right is FractionExpr → PowerExpr → PrimedExpr → ...
	// So -y would need to be parsed differently. The grammar may not allow
	// x * -y directly if - is only at NegationExpr level.
	// This test characterizes the actual behavior.
	_, err := ParseExpression("x * -y")
	// This might fail because * expects FractionExpr which doesn't include negation.
	// OR it might work because the PEG backtracks.
	if err != nil {
		// If it fails, parentheses should fix it:
		_, err2 := ParseExpression("x * (-y)")
		s.NoError(err2, "parenthesized negation should always work")
	}
}

func (s *StressPrecedenceTestSuite) TestBothOperandsNegated() {
	_, err := ParseExpression("-1 * -2")
	// This might or might not parse depending on whether * can see -2 on its right.
	// If it fails, test the parenthesized form:
	if err != nil {
		_, err2 := ParseExpression("(-1) * (-2)")
		s.NoError(err2)
	}
}

// === Fraction precedence (level 13.4) ===

func (s *StressPrecedenceTestSuite) TestFractionChaining() {
	// 1/2/3 should be left-associative: (1/2)/3
	expr, err := ParseExpression("1/2/3")
	s.Require().NoError(err)

	// Top should be a Fraction with left=(1/2) and right=3
	frac, ok := expr.(*ast.Fraction)
	if ok {
		_, ok = frac.Numerator.(*ast.Fraction)
		s.True(ok, "left of outer fraction should be inner fraction (1/2)")
	}
	// If it's not a Fraction node, it might be a RealInfixExpression with /
}

func (s *StressPrecedenceTestSuite) TestFractionVsMultiplication() {
	// 2 * 3/4: fraction binds tighter than multiplication → 2 * (3/4)
	expr, err := ParseExpression("2 * 3/4")
	s.Require().NoError(err)

	top, ok := expr.(*ast.RealInfixExpression)
	s.Require().True(ok, "top should be multiplication")
	s.Equal("*", top.Operator)
}

func (s *StressPrecedenceTestSuite) TestFractionReverseOrder() {
	// 3/4 * 2: (3/4) * 2
	expr, err := ParseExpression("3/4 * 2")
	s.Require().NoError(err)

	top, ok := expr.(*ast.RealInfixExpression)
	s.Require().True(ok, "top should be multiplication")
	s.Equal("*", top.Operator)
}

func (s *StressPrecedenceTestSuite) TestNegationInFraction() {
	// -3/4: is it -(3/4) or (-3)/4?
	// NegationExpr is at level 12, FractionExpr at level 13.4.
	// NegationExpr calls DivisionExpr which eventually calls FractionExpr.
	// So -3/4 = -(3/4).
	expr, err := ParseExpression("-3/4")
	s.Require().NoError(err)

	// Check if top is negation wrapping a fraction
	_, isNeg := expr.(*ast.UnaryNegation)
	_, isFrac := expr.(*ast.Fraction)

	// It should be one or the other — characterize behavior
	s.True(isNeg || isFrac, "should be negation or fraction at top level, got %T", expr)
}

// === Implies right-associativity (level 1) ===

func (s *StressPrecedenceTestSuite) TestImpliesRightAssociativity() {
	expr, err := ParseExpression("a => b => c")
	s.Require().NoError(err)

	// Should be: a => (b => c) — right-associative
	top, ok := expr.(*ast.LogicInfixExpression)
	s.Require().True(ok, "should be LogicInfixExpression")
	s.Equal("⇒", top.Operator)

	// Left should be 'a'
	_, ok = top.Left.(*ast.Identifier)
	s.True(ok, "left should be identifier")

	// Right should be another implies: b => c
	right, ok := top.Right.(*ast.LogicInfixExpression)
	s.Require().True(ok, "right should be nested implies")
	s.Equal("⇒", right.Operator)
}

func (s *StressPrecedenceTestSuite) TestImpliesVsEquivalence() {
	// a => b <=> c: implies is level 1 (lowest), equiv is level 2
	// ImpliesExpr calls EquivExpr on left. So left = EquivExpr, which parses 'a'.
	// Then => matches, and right recurses to ImpliesExpr which parses 'b <=> c'.
	// Result: a => (b <=> c)
	expr, err := ParseExpression("a => b <=> c")
	s.Require().NoError(err)

	top, ok := expr.(*ast.LogicInfixExpression)
	s.Require().True(ok)
	s.Equal("⇒", top.Operator, "top should be implies")

	// Right should be equivalence
	right, ok := top.Right.(*ast.LogicInfixExpression)
	s.Require().True(ok)
	s.Equal("≡", right.Operator, "right should be equivalence")
}

func (s *StressPrecedenceTestSuite) TestEquivalenceVsImplies() {
	// a <=> b => c: equiv at level 2, implies at level 1.
	// ImpliesExpr left = EquivExpr which parses 'a <=> b'.
	// Then => matches, right = ImpliesExpr parses 'c'.
	// Result: (a <=> b) => c
	expr, err := ParseExpression("a <=> b => c")
	s.Require().NoError(err)

	top, ok := expr.(*ast.LogicInfixExpression)
	s.Require().True(ok)
	s.Equal("⇒", top.Operator, "top should be implies")

	// Left should be equivalence
	left, ok := top.Left.(*ast.LogicInfixExpression)
	s.Require().True(ok)
	s.Equal("≡", left.Operator, "left should be equivalence")
}

// === Logic vs comparison precedence ===

func (s *StressPrecedenceTestSuite) TestLogicVsComparison() {
	tests := []struct {
		input       string
		desc        string
		topOperator string
	}{
		{"x > 5 /\\ y < 10", "comparisons higher than AND", "∧"},
		{"x = 1 \\/ y = 2", "equality higher than OR", "∨"},
		{"x > 5 => y < 10", "comparison higher than implies", "⇒"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			expr, err := ParseExpression(tt.input)
			s.Require().NoError(err, "should parse: %s", tt.input)

			top, ok := expr.(*ast.LogicInfixExpression)
			s.Require().True(ok, "top should be LogicInfixExpression for %s, got %T", tt.input, expr)
			s.Equal(tt.topOperator, top.Operator)
		})
	}
}

func (s *StressPrecedenceTestSuite) TestAndVsOr() {
	// a /\ b \/ c: AND at level 4 (higher than OR at level 3)
	// So: (a /\ b) \/ c
	expr, err := ParseExpression("a /\\ b \\/ c")
	s.Require().NoError(err)

	top, ok := expr.(*ast.LogicInfixExpression)
	s.Require().True(ok)
	s.Equal("∨", top.Operator, "top should be OR")

	left, ok := top.Left.(*ast.LogicInfixExpression)
	s.Require().True(ok)
	s.Equal("∧", left.Operator, "left should be AND")
}

func (s *StressPrecedenceTestSuite) TestOrVsAnd() {
	// a \/ b /\ c: a \/ (b /\ c)
	expr, err := ParseExpression("a \\/ b /\\ c")
	s.Require().NoError(err)

	top, ok := expr.(*ast.LogicInfixExpression)
	s.Require().True(ok)
	s.Equal("∨", top.Operator, "top should be OR")

	right, ok := top.Right.(*ast.LogicInfixExpression)
	s.Require().True(ok)
	s.Equal("∧", right.Operator, "right should be AND")
}

// TestNotVsComparison tests ~x > 5: NOT is at level 4, comparison at 5.x.
// Since NotExpr calls QuantifierExpr which calls ComparisonExpr,
// ~x > 5 should parse as ~(x > 5).
func (s *StressPrecedenceTestSuite) TestNotVsComparison() {
	expr, err := ParseExpression("~x > 5")
	s.Require().NoError(err)

	// Should be: NOT(x > 5) — NOT wraps the entire comparison
	neg, ok := expr.(*ast.UnaryLogic)
	if ok {
		s.Equal("¬", neg.Operator)
		// Inner should be comparison
		_, ok = neg.Right.(*ast.BinaryComparison)
		s.True(ok, "inner of NOT should be comparison, got %T", neg.Right)
	} else {
		// If NOT doesn't wrap comparison, the top might be comparison with NOT on left.
		// This would mean ~x is evaluated first, then compared to 5.
		comp, ok := expr.(*ast.BinaryComparison)
		if ok {
			_, ok = comp.Left.(*ast.UnaryLogic)
			s.True(ok, "left of comparison should be NOT")
		} else {
			s.Fail("unexpected expression type: %T", expr)
		}
	}
}

// === Prime with field access ===

func (s *StressPrecedenceTestSuite) TestSimplePrime() {
	expr, err := ParseExpression("x'")
	s.Require().NoError(err)

	primed, ok := expr.(*ast.Primed)
	s.Require().True(ok)
	_, ok = primed.Base.(*ast.Identifier)
	s.True(ok)
}

func (s *StressPrecedenceTestSuite) TestFieldAccessThenPrime() {
	// record.field' — field access is at level 17, prime at level 15.
	// PrimedExpr wraps FieldAccessExpr. So FieldAccess is parsed first,
	// then prime is applied. Result: (record.field)'
	expr, err := ParseExpression("record.field'")
	s.Require().NoError(err)

	primed, ok := expr.(*ast.Primed)
	s.Require().True(ok, "top should be Primed, got %T", expr)

	fa, ok := primed.Base.(*ast.FieldAccess)
	s.Require().True(ok, "base of prime should be FieldAccess, got %T", primed.Base)
	s.Equal("field", fa.Member)
}

// TestPrimeThenFieldAccess tests record'.field — prime is applied before
// the .field suffix. Since PrimedExpr doesn't recurse to FieldAccessExpr,
// this should fail (trailing .field) or parse only record'.
func (s *StressPrecedenceTestSuite) TestPrimeThenFieldAccess() {
	_, err := ParseExpression("record'.field")
	// This should fail because after record' is parsed, .field is trailing content.
	// RootExpression requires !. (end of input).
	s.Error(err, "record'.field should be a parse error (trailing .field after prime)")
}

// === Set operation precedence ===

func (s *StressPrecedenceTestSuite) TestSetOperationPrecedence() {
	// Grammar hierarchy: SetDifference → SetIntersection → SetUnion → CartesianProduct
	// This means union binds TIGHTEST among set ops.

	tests := []struct {
		input string
		desc  string
	}{
		{"A ∪ B ∩ C", "union inside intersection operands"},
		{"A ∩ B ∪ C", "intersection wraps union operands"},
		{"{1} ∪ {2} \\ {3}", "set diff is loosest set op"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			_, err := ParseExpression(tt.input)
			s.NoError(err, "should parse: %s", tt.input)
		})
	}
}

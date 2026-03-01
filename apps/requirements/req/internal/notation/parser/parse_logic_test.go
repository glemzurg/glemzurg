package parser

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/ast"
	"github.com/stretchr/testify/suite"
)

func TestLogicSuite(t *testing.T) {
	suite.Run(t, new(LogicSuite))
}

type LogicSuite struct {
	suite.Suite
}

// =============================================================================
// Logical NOT (¬, ~)
// =============================================================================

func (s *LogicSuite) TestParseNotUnicode() {
	expr, err := ParseExpression("¬TRUE")
	s.NoError(err)

	not := expr.(*ast.LogicPrefixExpression)
	s.Equal("¬", not.Operator)

	inner := not.Right.(*ast.BooleanLiteral)
	s.True(inner.Value)
}

func (s *LogicSuite) TestParseNotAscii() {
	expr, err := ParseExpression("~FALSE")
	s.NoError(err)

	not := expr.(*ast.LogicPrefixExpression)
	s.Equal("¬", not.Operator) // Normalized to Unicode

	inner := not.Right.(*ast.BooleanLiteral)
	s.False(inner.Value)
}

func (s *LogicSuite) TestParseDoubleNot() {
	expr, err := ParseExpression("~~TRUE")
	s.NoError(err)

	outer := expr.(*ast.LogicPrefixExpression)
	s.Equal("¬", outer.Operator)

	inner := outer.Right.(*ast.LogicPrefixExpression)
	s.Equal("¬", inner.Operator)

	val := inner.Right.(*ast.BooleanLiteral)
	s.True(val.Value)
}

// =============================================================================
// Logical AND (∧, /\)
// =============================================================================

func (s *LogicSuite) TestParseAndUnicode() {
	expr, err := ParseExpression("TRUE ∧ FALSE")
	s.NoError(err)

	and := expr.(*ast.LogicInfixExpression)
	s.Equal("∧", and.Operator)

	left := and.Left.(*ast.BooleanLiteral)
	s.True(left.Value)

	right := and.Right.(*ast.BooleanLiteral)
	s.False(right.Value)
}

func (s *LogicSuite) TestParseAndAscii() {
	expr, err := ParseExpression(`TRUE /\ FALSE`)
	s.NoError(err)

	and := expr.(*ast.LogicInfixExpression)
	s.Equal("∧", and.Operator) // Normalized to Unicode
}

func (s *LogicSuite) TestParseAndChain() {
	// a /\ b /\ c = (a /\ b) /\ c (left-associative)
	expr, err := ParseExpression(`TRUE /\ FALSE /\ TRUE`)
	s.NoError(err)

	outer := expr.(*ast.LogicInfixExpression)
	s.Equal("∧", outer.Operator)

	// Right is the last TRUE
	right := outer.Right.(*ast.BooleanLiteral)
	s.True(right.Value)

	// Left is (TRUE /\ FALSE)
	inner := outer.Left.(*ast.LogicInfixExpression)
	s.Equal("∧", inner.Operator)
}

// =============================================================================
// Logical OR (∨, \/)
// =============================================================================

func (s *LogicSuite) TestParseOrUnicode() {
	expr, err := ParseExpression("TRUE ∨ FALSE")
	s.NoError(err)

	or := expr.(*ast.LogicInfixExpression)
	s.Equal("∨", or.Operator)
}

func (s *LogicSuite) TestParseOrAscii() {
	expr, err := ParseExpression(`TRUE \/ FALSE`)
	s.NoError(err)

	or := expr.(*ast.LogicInfixExpression)
	s.Equal("∨", or.Operator) // Normalized to Unicode
}

func (s *LogicSuite) TestParseOrChain() {
	// a \/ b \/ c = (a \/ b) \/ c (left-associative)
	expr, err := ParseExpression(`TRUE \/ FALSE \/ TRUE`)
	s.NoError(err)

	outer := expr.(*ast.LogicInfixExpression)
	s.Equal("∨", outer.Operator)

	// Left is (TRUE \/ FALSE)
	inner := outer.Left.(*ast.LogicInfixExpression)
	s.Equal("∨", inner.Operator)
}

// =============================================================================
// Logical IMPLIES (⇒, =>)
// =============================================================================

func (s *LogicSuite) TestParseImpliesUnicode() {
	expr, err := ParseExpression("TRUE ⇒ FALSE")
	s.NoError(err)

	implies := expr.(*ast.LogicInfixExpression)
	s.Equal("⇒", implies.Operator)
}

func (s *LogicSuite) TestParseImpliesAscii() {
	expr, err := ParseExpression("TRUE => FALSE")
	s.NoError(err)

	implies := expr.(*ast.LogicInfixExpression)
	s.Equal("⇒", implies.Operator) // Normalized to Unicode
}

func (s *LogicSuite) TestParseImpliesRightAssociative() {
	// a => b => c = a => (b => c) (right-associative)
	expr, err := ParseExpression("TRUE => FALSE => TRUE")
	s.NoError(err)

	outer := expr.(*ast.LogicInfixExpression)
	s.Equal("⇒", outer.Operator)

	// Left is the first TRUE
	left := outer.Left.(*ast.BooleanLiteral)
	s.True(left.Value)

	// Right is (FALSE => TRUE)
	inner := outer.Right.(*ast.LogicInfixExpression)
	s.Equal("⇒", inner.Operator)

	innerLeft := inner.Left.(*ast.BooleanLiteral)
	s.False(innerLeft.Value)

	innerRight := inner.Right.(*ast.BooleanLiteral)
	s.True(innerRight.Value)
}

// =============================================================================
// Logical EQUIVALENCE (≡, <=>)
// =============================================================================

func (s *LogicSuite) TestParseEquivUnicode() {
	expr, err := ParseExpression("TRUE ≡ FALSE")
	s.NoError(err)

	equiv := expr.(*ast.LogicInfixExpression)
	s.Equal("≡", equiv.Operator)
}

func (s *LogicSuite) TestParseEquivAscii() {
	expr, err := ParseExpression("TRUE <=> FALSE")
	s.NoError(err)

	equiv := expr.(*ast.LogicInfixExpression)
	s.Equal("≡", equiv.Operator) // Normalized to Unicode
}

// =============================================================================
// Precedence: NOT > AND > OR > EQUIV > IMPLIES
// =============================================================================

func (s *LogicSuite) TestPrecedenceNotOverAnd() {
	// ~a /\ b = (~a) /\ b
	expr, err := ParseExpression("~TRUE /\\ FALSE")
	s.NoError(err)

	and := expr.(*ast.LogicInfixExpression)
	s.Equal("∧", and.Operator)

	// Left is ~TRUE
	not := and.Left.(*ast.LogicPrefixExpression)
	s.Equal("¬", not.Operator)
}

func (s *LogicSuite) TestPrecedenceAndOverOr() {
	// a /\ b \/ c = (a /\ b) \/ c
	expr, err := ParseExpression(`TRUE /\ FALSE \/ TRUE`)
	s.NoError(err)

	or := expr.(*ast.LogicInfixExpression)
	s.Equal("∨", or.Operator)

	// Left is (TRUE /\ FALSE)
	and := or.Left.(*ast.LogicInfixExpression)
	s.Equal("∧", and.Operator)
}

func (s *LogicSuite) TestPrecedenceOrOverEquiv() {
	// a \/ b <=> c = (a \/ b) <=> c
	expr, err := ParseExpression(`TRUE \/ FALSE <=> TRUE`)
	s.NoError(err)

	equiv := expr.(*ast.LogicInfixExpression)
	s.Equal("≡", equiv.Operator)

	// Left is (TRUE \/ FALSE)
	or := equiv.Left.(*ast.LogicInfixExpression)
	s.Equal("∨", or.Operator)
}

func (s *LogicSuite) TestPrecedenceEquivOverImplies() {
	// a <=> b => c = (a <=> b) => c
	expr, err := ParseExpression("TRUE <=> FALSE => TRUE")
	s.NoError(err)

	implies := expr.(*ast.LogicInfixExpression)
	s.Equal("⇒", implies.Operator)

	// Left is (TRUE <=> FALSE)
	equiv := implies.Left.(*ast.LogicInfixExpression)
	s.Equal("≡", equiv.Operator)
}

func (s *LogicSuite) TestPrecedenceComplexExpression() {
	// ~a /\ b \/ c => d = ((~a /\ b) \/ c) => d
	expr, err := ParseExpression("~TRUE /\\ FALSE \\/ TRUE => FALSE")
	s.NoError(err)

	implies := expr.(*ast.LogicInfixExpression)
	s.Equal("⇒", implies.Operator)

	// Right is FALSE
	right := implies.Right.(*ast.BooleanLiteral)
	s.False(right.Value)

	// Left is ((~TRUE /\ FALSE) \/ TRUE)
	or := implies.Left.(*ast.LogicInfixExpression)
	s.Equal("∨", or.Operator)

	// or.Right is TRUE
	orRight := or.Right.(*ast.BooleanLiteral)
	s.True(orRight.Value)

	// or.Left is (~TRUE /\ FALSE)
	and := or.Left.(*ast.LogicInfixExpression)
	s.Equal("∧", and.Operator)

	// and.Left is ~TRUE
	not := and.Left.(*ast.LogicPrefixExpression)
	s.Equal("¬", not.Operator)
}

// =============================================================================
// Parentheses override precedence
// =============================================================================

func (s *LogicSuite) TestParenthesesOverridePrecedence() {
	// a /\ (b \/ c) - parentheses force OR to be evaluated first
	expr, err := ParseExpression(`TRUE /\ (FALSE \/ TRUE)`)
	s.NoError(err)

	and := expr.(*ast.LogicInfixExpression)
	s.Equal("∧", and.Operator)

	// Right is (FALSE \/ TRUE) wrapped in paren
	paren := and.Right.(*ast.ParenExpr)
	or := paren.Inner.(*ast.LogicInfixExpression)
	s.Equal("∨", or.Operator)
}

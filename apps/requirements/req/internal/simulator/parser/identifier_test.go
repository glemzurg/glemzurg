package parser

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/ast"
	"github.com/stretchr/testify/suite"
)

func TestIdentifierSuite(t *testing.T) {
	suite.Run(t, new(IdentifierSuite))
}

type IdentifierSuite struct {
	suite.Suite
}

// =============================================================================
// Simple Identifiers
// =============================================================================

func (s *IdentifierSuite) TestIdentifier_Simple() {
	tests := []struct {
		input    string
		expected string
	}{
		{"x", "x"},
		{"myVar", "myVar"},
		{"_private", "_private"},
		{"count123", "count123"},
		{"CamelCase", "CamelCase"},
		{"snake_case", "snake_case"},
		{"_", "_"},
		{"a", "a"},
		{"Z", "Z"},
	}

	for _, tt := range tests {
		s.Run(tt.input, func() {
			expr, err := ParseExpression(tt.input)
			s.NoError(err)

			ident, ok := expr.(*ast.Identifier)
			s.True(ok, "expected *ast.Identifier, got %T", expr)
			s.Equal(tt.expected, ident.Value)
		})
	}
}

func (s *IdentifierSuite) TestIdentifier_WithWhitespace() {
	expr, err := ParseExpression("  myVar  ")
	s.NoError(err)

	ident, ok := expr.(*ast.Identifier)
	s.True(ok, "expected *ast.Identifier, got %T", expr)
	s.Equal("myVar", ident.Value)
}

func (s *IdentifierSuite) TestIdentifier_NotReservedKeyword() {
	// These should fail because they are reserved keywords
	reservedWords := []string{"TRUE", "FALSE", "IF", "THEN", "ELSE", "CASE", "OTHER"}

	for _, keyword := range reservedWords {
		s.Run(keyword, func() {
			expr, err := ParseExpression(keyword)
			// Keywords should either parse as their own type (TRUE/FALSE)
			// or fail to parse as identifiers
			if err == nil {
				_, isIdent := expr.(*ast.Identifier)
				s.False(isIdent, "keyword %s should not parse as Identifier", keyword)
			}
		})
	}
}

func (s *IdentifierSuite) TestIdentifier_KeywordPrefixOK() {
	// Identifiers that start with keywords should be allowed
	tests := []struct {
		input    string
		expected string
	}{
		{"TRUENESS", "TRUENESS"},
		{"FALSE_alarm", "FALSE_alarm"},
		{"IFfy", "IFfy"},
		{"THENA", "THENA"},
		{"ELSEwhere", "ELSEwhere"},
		{"CASEy", "CASEy"},
		{"OTHERWISE", "OTHERWISE"},
		{"LETTER", "LETTER"},
		{"INSIDE", "INSIDE"},
	}

	for _, tt := range tests {
		s.Run(tt.input, func() {
			expr, err := ParseExpression(tt.input)
			s.NoError(err)

			ident, ok := expr.(*ast.Identifier)
			s.True(ok, "expected *ast.Identifier, got %T", expr)
			s.Equal(tt.expected, ident.Value)
		})
	}
}

// =============================================================================
// Existing Value (@)
// =============================================================================

func (s *IdentifierSuite) TestExistingValue_Simple() {
	expr, err := ParseExpression("@")
	s.NoError(err)

	_, ok := expr.(*ast.ExistingValue)
	s.True(ok, "expected *ast.ExistingValue, got %T", expr)
}

func (s *IdentifierSuite) TestExistingValue_WithWhitespace() {
	expr, err := ParseExpression("  @  ")
	s.NoError(err)

	_, ok := expr.(*ast.ExistingValue)
	s.True(ok, "expected *ast.ExistingValue, got %T", expr)
}

// =============================================================================
// Field Access
// =============================================================================

func (s *IdentifierSuite) TestFieldAccess_Simple() {
	expr, err := ParseExpression("record.field")
	s.NoError(err)

	fa, ok := expr.(*ast.FieldAccess)
	s.True(ok, "expected *ast.FieldAccess, got %T", expr)

	// Check the base is an identifier
	base, ok := fa.Base.(*ast.Identifier)
	s.True(ok, "expected base to be *ast.Identifier, got %T", fa.Base)
	s.Equal("record", base.Value)
	s.Equal("field", fa.Member)
}

func (s *IdentifierSuite) TestFieldAccess_VariousNames() {
	tests := []struct {
		input          string
		expectedIdent  string
		expectedMember string
	}{
		{"person.name", "person", "name"},
		{"order.total", "order", "total"},
		{"_private.value", "_private", "value"},
		{"MyClass.property", "MyClass", "property"},
		{"x.y", "x", "y"},
	}

	for _, tt := range tests {
		s.Run(tt.input, func() {
			expr, err := ParseExpression(tt.input)
			s.NoError(err)

			fa, ok := expr.(*ast.FieldAccess)
			s.True(ok, "expected *ast.FieldAccess, got %T", expr)

			base, ok := fa.Base.(*ast.Identifier)
			s.True(ok, "expected base to be *ast.Identifier, got %T", fa.Base)
			s.Equal(tt.expectedIdent, base.Value)
			s.Equal(tt.expectedMember, fa.Member)
		})
	}
}

func (s *IdentifierSuite) TestFieldAccess_ExistingValue() {
	// @.field - field access on existing value
	expr, err := ParseExpression("@.status")
	s.NoError(err)

	fa, ok := expr.(*ast.FieldAccess)
	s.True(ok, "expected *ast.FieldAccess, got %T", expr)

	// Base should be ExistingValue
	_, ok = fa.Base.(*ast.ExistingValue)
	s.True(ok, "expected base to be *ast.ExistingValue, got %T", fa.Base)
	s.Equal("status", fa.Member)
}

func (s *IdentifierSuite) TestFieldAccess_WithWhitespace() {
	// Note: TLA+ typically doesn't allow whitespace around the dot
	// but our parser should handle input with surrounding whitespace
	expr, err := ParseExpression("  record.field  ")
	s.NoError(err)

	fa, ok := expr.(*ast.FieldAccess)
	s.True(ok, "expected *ast.FieldAccess, got %T", expr)

	base, ok := fa.Base.(*ast.Identifier)
	s.True(ok, "expected base to be *ast.Identifier, got %T", fa.Base)
	s.Equal("record", base.Value)
	s.Equal("field", fa.Member)
}

// =============================================================================
// Chained Field Access (a.b.c)
// =============================================================================

func (s *IdentifierSuite) TestFieldAccess_Chained() {
	// a.b.c should parse as ((a).b).c
	expr, err := ParseExpression("a.b.c")
	s.NoError(err)

	// Outermost should be FieldAccess with member "c"
	outer, ok := expr.(*ast.FieldAccess)
	s.True(ok, "expected *ast.FieldAccess, got %T", expr)
	s.Equal("c", outer.Member)

	// Its base should be another FieldAccess with member "b"
	inner, ok := outer.Base.(*ast.FieldAccess)
	s.True(ok, "expected base to be *ast.FieldAccess, got %T", outer.Base)
	s.Equal("b", inner.Member)

	// The innermost base should be identifier "a"
	base, ok := inner.Base.(*ast.Identifier)
	s.True(ok, "expected base to be *ast.Identifier, got %T", inner.Base)
	s.Equal("a", base.Value)
}

func (s *IdentifierSuite) TestFieldAccess_ChainedLong() {
	// person.address.city.name
	expr, err := ParseExpression("person.address.city.name")
	s.NoError(err)

	// Build up from inside out: person -> .address -> .city -> .name
	fa1, ok := expr.(*ast.FieldAccess)
	s.True(ok, "expected outermost *ast.FieldAccess")
	s.Equal("name", fa1.Member)

	fa2, ok := fa1.Base.(*ast.FieldAccess)
	s.True(ok, "expected *ast.FieldAccess for .city")
	s.Equal("city", fa2.Member)

	fa3, ok := fa2.Base.(*ast.FieldAccess)
	s.True(ok, "expected *ast.FieldAccess for .address")
	s.Equal("address", fa3.Member)

	base, ok := fa3.Base.(*ast.Identifier)
	s.True(ok, "expected *ast.Identifier for person")
	s.Equal("person", base.Value)
}

func (s *IdentifierSuite) TestFieldAccess_ChainedWithExistingValue() {
	// @.a.b - chained access starting from existing value
	expr, err := ParseExpression("@.a.b")
	s.NoError(err)

	outer, ok := expr.(*ast.FieldAccess)
	s.True(ok, "expected *ast.FieldAccess, got %T", expr)
	s.Equal("b", outer.Member)

	inner, ok := outer.Base.(*ast.FieldAccess)
	s.True(ok, "expected base to be *ast.FieldAccess, got %T", outer.Base)
	s.Equal("a", inner.Member)

	_, ok = inner.Base.(*ast.ExistingValue)
	s.True(ok, "expected base to be *ast.ExistingValue, got %T", inner.Base)
}

func (s *IdentifierSuite) TestFieldAccess_ChainedString() {
	expr, err := ParseExpression("a.b.c")
	s.NoError(err)
	s.Equal("a.b.c", expr.String())
}

// =============================================================================
// Primed Expressions (x')
// =============================================================================

func (s *IdentifierSuite) TestPrimed_SimpleIdentifier() {
	expr, err := ParseExpression("x'")
	s.NoError(err)

	primed, ok := expr.(*ast.Primed)
	s.True(ok, "expected *ast.Primed, got %T", expr)

	ident, ok := primed.Base.(*ast.Identifier)
	s.True(ok, "expected base to be *ast.Identifier, got %T", primed.Base)
	s.Equal("x", ident.Value)
}

func (s *IdentifierSuite) TestPrimed_LongerIdentifier() {
	expr, err := ParseExpression("counter'")
	s.NoError(err)

	primed, ok := expr.(*ast.Primed)
	s.True(ok, "expected *ast.Primed, got %T", expr)

	ident, ok := primed.Base.(*ast.Identifier)
	s.True(ok, "expected base to be *ast.Identifier, got %T", primed.Base)
	s.Equal("counter", ident.Value)
}

func (s *IdentifierSuite) TestPrimed_FieldAccess() {
	// record.field' - prime applies to the whole field access
	expr, err := ParseExpression("record.field'")
	s.NoError(err)

	primed, ok := expr.(*ast.Primed)
	s.True(ok, "expected *ast.Primed, got %T", expr)

	fa, ok := primed.Base.(*ast.FieldAccess)
	s.True(ok, "expected base to be *ast.FieldAccess, got %T", primed.Base)
	s.Equal("field", fa.Member)

	ident, ok := fa.Base.(*ast.Identifier)
	s.True(ok, "expected field base to be *ast.Identifier, got %T", fa.Base)
	s.Equal("record", ident.Value)
}

func (s *IdentifierSuite) TestPrimed_ChainedFieldAccess() {
	// a.b.c' - prime applies to the whole chain
	expr, err := ParseExpression("a.b.c'")
	s.NoError(err)

	primed, ok := expr.(*ast.Primed)
	s.True(ok, "expected *ast.Primed, got %T", expr)

	fa1, ok := primed.Base.(*ast.FieldAccess)
	s.True(ok, "expected base to be *ast.FieldAccess, got %T", primed.Base)
	s.Equal("c", fa1.Member)

	fa2, ok := fa1.Base.(*ast.FieldAccess)
	s.True(ok, "expected field base to be *ast.FieldAccess, got %T", fa1.Base)
	s.Equal("b", fa2.Member)
}

func (s *IdentifierSuite) TestPrimed_InExpression() {
	// x' + 1
	expr, err := ParseExpression("x' + 1")
	s.NoError(err)

	arith, ok := expr.(*ast.BinaryArithmetic)
	s.True(ok, "expected *ast.BinaryArithmetic, got %T", expr)

	primed, ok := arith.Left.(*ast.Primed)
	s.True(ok, "expected left to be *ast.Primed, got %T", arith.Left)

	ident, ok := primed.Base.(*ast.Identifier)
	s.True(ok, "expected primed base to be *ast.Identifier, got %T", primed.Base)
	s.Equal("x", ident.Value)
}

func (s *IdentifierSuite) TestPrimed_InComparison() {
	// x' = y
	expr, err := ParseExpression("x' = y")
	s.NoError(err)

	eq, ok := expr.(*ast.BinaryEquality)
	s.True(ok, "expected *ast.BinaryEquality, got %T", expr)

	primed, ok := eq.Left.(*ast.Primed)
	s.True(ok, "expected left to be *ast.Primed, got %T", eq.Left)

	ident, ok := primed.Base.(*ast.Identifier)
	s.True(ok, "expected primed base to be *ast.Identifier, got %T", primed.Base)
	s.Equal("x", ident.Value)
}

func (s *IdentifierSuite) TestPrimed_String() {
	expr, err := ParseExpression("counter'")
	s.NoError(err)
	s.Equal("counter'", expr.String())
}

func (s *IdentifierSuite) TestPrimed_FieldAccessString() {
	expr, err := ParseExpression("record.field'")
	s.NoError(err)
	s.Equal("record.field'", expr.String())
}

// =============================================================================
// Identifiers in Expressions
// =============================================================================

func (s *IdentifierSuite) TestIdentifier_InArithmetic() {
	expr, err := ParseExpression("x + 1")
	s.NoError(err)

	arith, ok := expr.(*ast.BinaryArithmetic)
	s.True(ok, "expected *ast.BinaryArithmetic, got %T", expr)

	ident, ok := arith.Left.(*ast.Identifier)
	s.True(ok, "expected left to be *ast.Identifier, got %T", arith.Left)
	s.Equal("x", ident.Value)
}

func (s *IdentifierSuite) TestIdentifier_InComparison() {
	expr, err := ParseExpression("x > 5")
	s.NoError(err)

	comp, ok := expr.(*ast.BinaryComparison)
	s.True(ok, "expected *ast.BinaryComparison, got %T", expr)

	ident, ok := comp.Left.(*ast.Identifier)
	s.True(ok, "expected left to be *ast.Identifier, got %T", comp.Left)
	s.Equal("x", ident.Value)
}

func (s *IdentifierSuite) TestIdentifier_InLogic() {
	expr, err := ParseExpression("a /\\ b")
	s.NoError(err)

	logic, ok := expr.(*ast.BinaryLogic)
	s.True(ok, "expected *ast.BinaryLogic, got %T", expr)

	leftIdent, ok := logic.Left.(*ast.Identifier)
	s.True(ok, "expected left to be *ast.Identifier, got %T", logic.Left)
	s.Equal("a", leftIdent.Value)

	rightIdent, ok := logic.Right.(*ast.Identifier)
	s.True(ok, "expected right to be *ast.Identifier, got %T", logic.Right)
	s.Equal("b", rightIdent.Value)
}

func (s *IdentifierSuite) TestFieldAccess_InComparison() {
	expr, err := ParseExpression("person.age > 18")
	s.NoError(err)

	comp, ok := expr.(*ast.BinaryComparison)
	s.True(ok, "expected *ast.BinaryComparison, got %T", expr)

	fa, ok := comp.Left.(*ast.FieldAccess)
	s.True(ok, "expected left to be *ast.FieldAccess, got %T", comp.Left)

	base, ok := fa.Base.(*ast.Identifier)
	s.True(ok, "expected base to be *ast.Identifier, got %T", fa.Base)
	s.Equal("person", base.Value)
	s.Equal("age", fa.Member)
}

func (s *IdentifierSuite) TestExistingValue_InArithmetic() {
	expr, err := ParseExpression("@ + 1")
	s.NoError(err)

	arith, ok := expr.(*ast.BinaryArithmetic)
	s.True(ok, "expected *ast.BinaryArithmetic, got %T", expr)

	_, ok = arith.Left.(*ast.ExistingValue)
	s.True(ok, "expected left to be *ast.ExistingValue, got %T", arith.Left)
}

func (s *IdentifierSuite) TestFieldAccess_ExistingValueInExpression() {
	// @.count + 1
	expr, err := ParseExpression("@.count + 1")
	s.NoError(err)

	arith, ok := expr.(*ast.BinaryArithmetic)
	s.True(ok, "expected *ast.BinaryArithmetic, got %T", expr)

	fa, ok := arith.Left.(*ast.FieldAccess)
	s.True(ok, "expected left to be *ast.FieldAccess, got %T", arith.Left)

	_, ok = fa.Base.(*ast.ExistingValue)
	s.True(ok, "expected base to be *ast.ExistingValue, got %T", fa.Base)
	s.Equal("count", fa.Member)
}

// =============================================================================
// String Representation (round-trip)
// =============================================================================

func (s *IdentifierSuite) TestIdentifier_String() {
	expr, err := ParseExpression("myVariable")
	s.NoError(err)
	s.Equal("myVariable", expr.String())
}

func (s *IdentifierSuite) TestExistingValue_String() {
	expr, err := ParseExpression("@")
	s.NoError(err)
	s.Equal("@", expr.String())
}

func (s *IdentifierSuite) TestFieldAccess_String() {
	expr, err := ParseExpression("record.field")
	s.NoError(err)
	s.Equal("record.field", expr.String())
}

func (s *IdentifierSuite) TestFieldAccess_ExistingValue_String() {
	expr, err := ParseExpression("@.field")
	s.NoError(err)
	s.Equal("@.field", expr.String()) // Base is ExistingValue which renders as @
}

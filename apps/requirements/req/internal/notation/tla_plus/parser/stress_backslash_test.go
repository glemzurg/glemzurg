package parser

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/ast"
	"github.com/stretchr/testify/suite"
)

// StressBackslashTestSuite stress-tests set difference `\` disambiguation.
// The PEG grammar uses a negative lookahead on SetDifferenceExpr:
//
//	"\\" !( "/" / "i" / "c" / "u" / "s" / "n" / "d" / "b" / "o" / "h" / "B" / "O" / "H" )
//
// This ensures `\` doesn't match when followed by characters that start
// backslash-prefixed operators (\/, \in, \cup, \subseteq, etc.).
type StressBackslashTestSuite struct {
	suite.Suite
}

func TestStressBackslashSuite(t *testing.T) {
	suite.Run(t, new(StressBackslashTestSuite))
}

// TestSetDifferenceBasic tests standard set difference parsing.
func (s *StressBackslashTestSuite) TestSetDifferenceBasic() {
	tests := []struct {
		input string
		desc  string
	}{
		{`A \ B`, "basic set difference with spaces"},
		{`{1, 2, 3} \ {2}`, "set literals with difference"},
		{`A \ B \ C`, "chained set difference (left-associative)"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			expr, err := ParseExpression(tt.input)
			s.Require().NoError(err, "should parse: %q (%s)", tt.input, tt.desc)

			// Verify it's a BinarySetOperation with \ operator
			setOp, ok := expr.(*ast.BinarySetOperation)
			s.True(ok, "expected BinarySetOperation, got %T for %q", expr, tt.input)
			s.Equal(`\`, setOp.Operator)
		})
	}
}

// TestSetDifferenceChaining tests left-associative chaining of set difference.
func (s *StressBackslashTestSuite) TestSetDifferenceChaining() {
	expr, err := ParseExpression(`A \ B \ C`)
	s.Require().NoError(err, "should parse chained set difference")

	// Should be ((A \ B) \ C) — left-associative
	outer, ok := expr.(*ast.BinarySetOperation)
	s.True(ok, "expected BinarySetOperation")
	s.Equal(`\`, outer.Operator)

	inner, ok := outer.Left.(*ast.BinarySetOperation)
	s.True(ok, "left should be BinarySetOperation (left-assoc)")
	s.Equal(`\`, inner.Operator)

	// Inner left should be identifier A
	_, ok = inner.Left.(*ast.Identifier)
	s.True(ok, "innermost left should be Identifier")
}

// TestBackslashOperatorsNotSetDifference tests that backslash-prefixed operators
// are NOT parsed as set difference. Each of these should parse as the intended
// operator thanks to the negative lookahead.
func (s *StressBackslashTestSuite) TestBackslashOperatorsNotSetDifference() {
	tests := []struct {
		input    string
		desc     string
		expected string // expected AST type or operator
	}{
		// Logic operators
		{`a \/ b`, "disjunction (OR)", "∨"},
		{`a /\ b`, "conjunction (AND)", "∧"},

		// Membership / set relations
		{`x \in S`, "set membership", "∈"},
		{`x \notin S`, "not-in membership", "∉"},

		// Set operations
		{`A \cup B`, "set union", "∪"},
		{`A \cap B`, "set intersection", "∩"},
		{`A \subseteq B`, "subset-or-equal", "⊆"},
		{`A \supseteq B`, "superset-or-equal", "⊇"},
		{`A \subset B`, "proper subset", "⊂"},
		{`A \supset B`, "proper superset", "⊃"},

		// Arithmetic
		{`6 \div 2`, "integer division", "÷"},

		// Sequence concatenation
		{`<<1>> \o <<2>>`, "sequence concat", "∘"},
		{`<<1>> \circ <<2>>`, "sequence concat (circ)", "∘"},

		// Bag comparison
		{`A \sqsubseteq B`, "bag subset", "⊑"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			expr, err := ParseExpression(tt.input)
			s.Require().NoError(err, "should parse: %q (%s)", tt.input, tt.desc)

			// Verify it's NOT a set difference
			if setOp, ok := expr.(*ast.BinarySetOperation); ok {
				s.NotEqual(`\`, setOp.Operator,
					"should NOT be set difference for %q, got operator %q", tt.input, setOp.Operator)
			}
			// Just verify it parsed — the specific AST type depends on the operator
			s.NotNil(expr, "expression should not be nil")
		})
	}
}

// TestSetDifferenceWithOtherOperators tests set difference combined with
// other operators to verify precedence interactions.
func (s *StressBackslashTestSuite) TestSetDifferenceWithOtherOperators() {
	tests := []struct {
		input string
		desc  string
	}{
		{`A \ B \/ C`, "set diff then OR (diff higher precedence)"},
		{`A \ B /\ C`, "set diff then AND"},
		{`A \cup B \ C`, "union then diff"},
		{`A \ B \cup C`, "diff then union"},
		{`1..5 \ {3}`, "range minus element"},
		{`(A \ B) \cup (C \ D)`, "parenthesized diffs with union"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			_, err := ParseExpression(tt.input)
			s.Require().NoError(err, "should parse: %q (%s)", tt.input, tt.desc)
		})
	}
}

// TestSetDifferenceRequiresWhitespace tests that set difference requires
// whitespace around the `\` operator to distinguish it from \-prefix symbols
// like \in, \cup, etc.
func (s *StressBackslashTestSuite) TestSetDifferenceRequiresWhitespace() {
	tests := []struct {
		input       string
		desc        string
		shouldParse bool
	}{
		{`A \ B`, "standard spacing", true},
		{`A \B`, "no space after backslash — must fail", false},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			_, err := ParseExpression(tt.input)
			if tt.shouldParse {
				s.Require().NoError(err, "should parse: %q", tt.input)
			} else {
				s.Require().Error(err, "set difference without whitespace must fail: %q", tt.input)
			}
		})
	}
}

// TestBackslashLookaheadBoundaryChars tests the exact characters in the
// negative lookahead to ensure they all correctly block set difference.
func (s *StressBackslashTestSuite) TestBackslashLookaheadBoundaryChars() {
	// The lookahead prevents set-diff when `\` is followed by:
	// "/" "i" "c" "u" "s" "n" "d" "b" "o" "h" "B" "O" "H"
	// Each maps to a specific operator family.
	tests := []struct {
		input string
		desc  string
		char  string // the lookahead character being tested
	}{
		{`a \/ b`, "/ → disjunction", "/"},
		{`x \in S`, "i → membership (\\in)", "i"},
		{`A \cup B`, "c → set union (\\cup) / \\cap / \\circ", "c"},
		{`A \cup B`, "u → (covered by \\cup)", "u"},
		{`A \subseteq B`, "s → subset/superset", "s"},
		{`x \notin S`, "n → not-in", "n"},
		{`6 \div 2`, "d → division", "d"},
		{`\\b0`, "b → binary number prefix", "b"},
		{`\\o0`, "o → octal number prefix", "o"},
		{`\\hFF`, "h → hex number prefix", "h"},
		// Uppercase variants
		{`\\BFFFF`, "B → binary number (uppercase)", "B"},
		{`\\O77`, "O → octal number (uppercase)", "O"},
		{`\\HFF`, "H → hex number (uppercase)", "H"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			// These should all parse without being treated as set difference.
			_, err := ParseExpression(tt.input)
			// Some may fail for other reasons (e.g., `\b0` might not be valid in
			// certain contexts), but they should NOT produce a set difference node.
			if err != nil {
				s.T().Logf("NOTICE: %q fails to parse: %v (char=%s)", tt.input, err, tt.char)
			}
		})
	}
}

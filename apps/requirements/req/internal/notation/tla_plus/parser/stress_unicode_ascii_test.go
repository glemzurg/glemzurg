package parser

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// StressUnicodeAsciiTestSuite stress-tests mixed Unicode/ASCII operator variants.
// Each operator has both Unicode and ASCII forms. Mixed use within a single
// expression is untested elsewhere.
type StressUnicodeAsciiTestSuite struct {
	suite.Suite
}

func TestStressUnicodeAsciiSuite(t *testing.T) {
	suite.Run(t, new(StressUnicodeAsciiTestSuite))
}

// TestMixedOperatorsInSingleExpression tests mixing Unicode and ASCII
// operator variants within the same expression.
func (s *StressUnicodeAsciiTestSuite) TestMixedOperatorsInSingleExpression() {
	tests := []struct {
		input string
		desc  string
	}{
		// Logic operators
		{`a /\ b ∨ c`, "ASCII AND + Unicode OR"},
		{`a ∧ b \/ c`, "Unicode AND + ASCII OR"},

		// Comparison operators
		{`x <= 5 /\ y ≥ 10`, "ASCII <= + Unicode ≥"},
		{`x ≤ 5 /\ y >= 10`, "Unicode ≤ + ASCII >="},

		// Membership
		{`x \in S /\ y ∈ T`, "ASCII \\in + Unicode ∈"},

		// Set operations
		{`A \cup B ∩ C`, "ASCII union + Unicode intersection"},
		{`A ∪ B \cap C`, "Unicode union + ASCII intersection"},

		// Subset relations
		{`A \subseteq B /\ C ⊆ D`, "ASCII subset + Unicode subset"},

		// Sequence concat
		{`<<1>> \o <<2>> ∘ <<3>>`, "ASCII \\o + Unicode ∘"},

		// Division
		{`6 \div 2 + 3 ÷ 1`, "ASCII div + Unicode ÷"},

		// Quantifiers mixed
		{`\A x ∈ S : \E y \in T : x = y`, "ASCII \\A + Unicode ∈, ASCII \\E + ASCII \\in"},

		// Implies and equiv
		{`a => b ∧ c`, "ASCII implies with Unicode AND"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			_, err := ParseExpression(tt.input)
			s.Require().NoError(err, "should parse: %q (%s)", tt.input, tt.desc)
		})
	}
}

// TestRoundTripASCIItoUnicode tests that ASCII input is stringified to
// Unicode output via the AST's String() method.
func (s *StressUnicodeAsciiTestSuite) TestRoundTripASCIItoUnicode() {
	tests := []struct {
		input    string
		expected string
		desc     string
	}{
		// Logic operators — String() does not add outer parentheses
		{`a /\ b`, "a ∧ b", "AND: /\\ → ∧"},
		{`a \/ b`, "a ∨ b", "OR: \\/ → ∨"},
		{`~a`, "¬a", "NOT: ~ → ¬"},

		// Quantifiers — String() uses (∀x ...) format: no space after quantifier symbol,
		// parens around entire expression, no parens around predicate
		{`\A x \in S : x > 0`, "(∀x ∈ S : x > 0)", "universal quantifier"},
		{`\E x \in S : x > 0`, "(∃x ∈ S : x > 0)", "existential quantifier"},

		// Membership
		{`x \in S`, "x ∈ S", "membership: \\in → ∈"},
		{`x \notin S`, "x ∉ S", "not-in: \\notin → ∉"},

		// Comparison
		{`1 /= 2`, "1 ≠ 2", "not-equal: /= → ≠"},
		{`1 <= 2`, "1 ≤ 2", "lte: <= → ≤"},
		{`1 >= 2`, "1 ≥ 2", "gte: >= → ≥"},

		// Set operations
		{`A \cup B`, "A ∪ B", "union: \\cup → ∪"},
		{`A \cap B`, "A ∩ B", "intersection: \\cap → ∩"},
		{`A \subseteq B`, "A ⊆ B", "subset: \\subseteq → ⊆"},
		{`A \supseteq B`, "A ⊇ B", "superset: \\supseteq → ⊇"},
		{`A \subset B`, "A ⊂ B", "proper subset: \\subset → ⊂"},
		{`A \supset B`, "A ⊃ B", "proper superset: \\supset → ⊃"},

		// Arithmetic
		{`6 \div 2`, "6 ÷ 2", "div: \\div → ÷"},

		// Sequence — String() uses Unicode angle brackets ⟨⟩ instead of <<>>
		{`<<1>> \o <<2>>`, "⟨1⟩ ∘ ⟨2⟩", "concat: \\o → ∘"},
		{`<<1>> \circ <<2>>`, "⟨1⟩ ∘ ⟨2⟩", "concat: \\circ → ∘"},

		// Implies and equiv
		{`a => b`, "a ⇒ b", "implies: => → ⇒"},
		{`a <=> b`, "a ≡ b", "equiv: <=> → ≡"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			expr, err := ParseExpression(tt.input)
			s.Require().NoError(err, "should parse: %q", tt.input)
			s.Equal(tt.expected, expr.String(), "String() should produce Unicode for %q", tt.input)
		})
	}
}

// TestUnicodeInputParsesIdentically tests that Unicode input produces
// the same AST String() output as equivalent ASCII input.
func (s *StressUnicodeAsciiTestSuite) TestUnicodeInputParsesIdentically() {
	pairs := []struct {
		ascii   string
		unicode string
		desc    string
	}{
		{`a /\ b`, `a ∧ b`, "AND"},
		{`a \/ b`, `a ∨ b`, "OR"},
		{`~a`, `¬a`, "NOT"},
		{`x \in S`, `x ∈ S`, "membership"},
		{`x \notin S`, `x ∉ S`, "not-in"},
		{`1 /= 2`, `1 ≠ 2`, "not-equal"},
		{`A \cup B`, `A ∪ B`, "union"},
		{`A \cap B`, `A ∩ B`, "intersection"},
		{`A \subseteq B`, `A ⊆ B`, "subset-or-equal"},
		{`6 \div 2`, `6 ÷ 2`, "integer division"},
		{`<<1>> \o <<2>>`, `<<1>> ∘ <<2>>`, "concat"},
		{`a => b`, `a ⇒ b`, "implies"},
		{`a <=> b`, `a ≡ b`, "equiv"},
		{`\A x \in S : x > 0`, `∀ x ∈ S : x > 0`, "universal quantifier"},
		{`\E x \in S : x > 0`, `∃ x ∈ S : x > 0`, "existential quantifier"},
	}

	for _, tt := range pairs {
		s.Run(tt.desc, func() {
			asciiExpr, err := ParseExpression(tt.ascii)
			s.Require().NoError(err, "ASCII should parse: %q", tt.ascii)

			unicodeExpr, err := ParseExpression(tt.unicode)
			s.Require().NoError(err, "Unicode should parse: %q", tt.unicode)

			s.Equal(asciiExpr.String(), unicodeExpr.String(),
				"ASCII %q and Unicode %q should produce identical String() output",
				tt.ascii, tt.unicode)
		})
	}
}

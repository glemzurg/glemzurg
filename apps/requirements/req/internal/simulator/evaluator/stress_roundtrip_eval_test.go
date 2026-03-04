package evaluator

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/parser"
	"github.com/stretchr/testify/suite"
)

// StressRoundtripEvalTestSuite tests parse→stringify→re-parse→eval round-trips.
// For each expression: parse to AST, stringify back to text, re-parse,
// evaluate both ASTs, and compare results. This catches:
//   - Stringification bugs that produce unparseable output
//   - Precedence/associativity information lost in stringification
//   - Semantic drift between parse and re-parse
type StressRoundtripEvalTestSuite struct {
	suite.Suite
}

func TestStressRoundtripEvalSuite(t *testing.T) {
	suite.Run(t, new(StressRoundtripEvalTestSuite))
}

// roundTripCheck is a helper that verifies parse→stringify→re-parse→eval produces
// identical results.
func (s *StressRoundtripEvalTestSuite) roundTripCheck(input string) {
	// Step 1: Parse original input
	ast1, err := parser.ParseExpression(input)
	s.Require().NoError(err, "first parse of %q", input)

	// Step 2: Stringify
	stringified := ast1.String()

	// Step 3: Re-parse the stringified output
	ast2, err := parser.ParseExpression(stringified)
	s.Require().NoError(err, "re-parse of stringified %q (stringified: %q)", input, stringified)

	// Step 4: Evaluate both ASTs
	bindings := NewBindings()
	result1 := EvalAST(ast1, bindings)
	result2 := EvalAST(ast2, bindings)

	// Step 5: Compare results
	if result1.IsError() {
		s.True(result2.IsError(), "both should error for %q; original errored but re-parse didn't", input)
		return
	}
	s.False(result2.IsError(), "re-parsed %q errored: %v (stringified: %q)", input, result2.Error, stringified)
	s.Equal(result1.Value.Inspect(), result2.Value.Inspect(),
		"round-trip value mismatch for %q (stringified: %q)", input, stringified)
}

// TestArithmeticRoundTrip tests round-trip for arithmetic expressions.
func (s *StressRoundtripEvalTestSuite) TestArithmeticRoundTrip() {
	tests := []struct {
		input string
		desc  string
	}{
		{"1 + 2 * 3", "precedence: add vs mul"},
		{"(1 + 2) * 3", "explicit parens"},
		{"2 ^ 3", "power"},
		{"2 ^ 3 ^ 2", "right-associative power"},
		{"-5 + 3", "negation in addition"},
		{"7 % 3", "modulo"},
		{"1/2 + 3/4", "fraction addition"},
		{"2 * 3/4", "multiplication with fraction"},
		{"10 - 3 - 2", "left-associative subtraction"},
		{"1 + 2 * 3 - 4", "mixed arithmetic"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			s.roundTripCheck(tt.input)
		})
	}
}

// TestLogicRoundTrip tests round-trip for logic expressions.
func (s *StressRoundtripEvalTestSuite) TestLogicRoundTrip() {
	tests := []struct {
		input string
		desc  string
	}{
		{`TRUE /\ FALSE`, "AND"},
		{`TRUE \/ FALSE`, "OR"},
		{`TRUE => FALSE`, "implies"},
		{`~TRUE`, "NOT"},
		{`TRUE <=> FALSE`, "equiv"},
		{`TRUE /\ TRUE /\ FALSE`, "chained AND"},
		{`FALSE \/ FALSE \/ TRUE`, "chained OR"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			s.roundTripCheck(tt.input)
		})
	}
}

// TestComparisonRoundTrip tests round-trip for comparison expressions.
func (s *StressRoundtripEvalTestSuite) TestComparisonRoundTrip() {
	tests := []struct {
		input string
		desc  string
	}{
		{"1 = 1", "equality true"},
		{"1 = 2", "equality false"},
		{"1 /= 2", "inequality"},
		{"1 < 2", "less than"},
		{"2 > 1", "greater than"},
		{"1 <= 1", "less or equal"},
		{"1 >= 1", "greater or equal"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			s.roundTripCheck(tt.input)
		})
	}
}

// TestComplexRoundTrip tests round-trip for complex expressions.
func (s *StressRoundtripEvalTestSuite) TestComplexRoundTrip() {
	tests := []struct {
		input string
		desc  string
	}{
		{`\A x \in {1, 2, 3} : x > 0`, "universal quantifier"},
		{`\E x \in {1, 2, 3} : x > 3`, "existential quantifier (false)"},
		{"IF 1 > 0 THEN 1 ELSE 2", "IF-THEN-ELSE"},
		{"{1, 2, 3} ∪ {4, 5}", "set union"},
		{"{1, 2, 3} ∩ {2, 3, 4}", "set intersection"},
		{"<<1, 2, 3>>[2]", "tuple indexing"},
		{"CASE 1 > 0 -> 1 [] OTHER -> 0", "CASE expression"},
		{`1 \in {1, 2, 3}`, "membership"},
		{`{1, 2} \subseteq {1, 2, 3}`, "subset"},
		{"1..5", "range"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			s.roundTripCheck(tt.input)
		})
	}
}

// TestSetOperationRoundTrip tests round-trip for set operations with precedence.
func (s *StressRoundtripEvalTestSuite) TestSetOperationRoundTrip() {
	tests := []struct {
		input string
		desc  string
	}{
		{`{1, 2} ∪ {3}`, "simple union"},
		{`{1, 2, 3} ∩ {2, 3}`, "simple intersection"},
		{`{1, 2, 3} \ {2}`, "set difference"},
		{`({1, 2} ∪ {3}) ∩ {2, 3}`, "union then intersection with parens"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			s.roundTripCheck(tt.input)
		})
	}
}

// TestRecordRoundTrip tests round-trip for record expressions.
func (s *StressRoundtripEvalTestSuite) TestRecordRoundTrip() {
	tests := []struct {
		input string
		desc  string
	}{
		{"[a |-> 1, b |-> 2]", "record literal"},
		{"[a |-> 1].a", "record field access"},
		{"[a |-> 1, b |-> 2].b", "multi-field record access"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			s.roundTripCheck(tt.input)
		})
	}
}

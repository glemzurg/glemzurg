package parser

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

// StressNestingTestSuite stress-tests deep nesting scenarios.
// PEG parsers can have exponential backtracking or stack overflow with
// deep nesting. These tests verify the parser handles depth gracefully.
type StressNestingTestSuite struct {
	suite.Suite
}

func TestStressNestingSuite(t *testing.T) {
	suite.Run(t, new(StressNestingTestSuite))
}

// TestDeepParentheses tests deeply nested parenthesized expressions.
func (s *StressNestingTestSuite) TestDeepParentheses() {
	tests := []struct {
		depth int
		desc  string
	}{
		{5, "5-level nested parens"},
		{10, "10-level nested parens"},
		{15, "15-level nested parens"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			// Build ((((... 1 ...))))
			input := strings.Repeat("(", tt.depth) + "1" + strings.Repeat(")", tt.depth)
			_, err := ParseExpression(input)
			s.Require().NoError(err, "should parse %d-level nested parens", tt.depth)
		})
	}
}

// TestDeepArithmeticChains tests long chains of arithmetic operators.
func (s *StressNestingTestSuite) TestDeepArithmeticChains() {
	tests := []struct {
		count int
		op    string
		desc  string
	}{
		{20, "+", "20-term addition chain"},
		{20, "*", "20-term multiplication chain"},
		{10, "-", "10-term subtraction chain"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			// Build: 1 op 2 op 3 op ... op N
			parts := make([]string, tt.count)
			for i := range tt.count {
				parts[i] = fmt.Sprintf("%d", i+1)
			}
			input := strings.Join(parts, " "+tt.op+" ")
			_, err := ParseExpression(input)
			s.Require().NoError(err, "should parse %d-term %s chain", tt.count, tt.op)
		})
	}
}

// TestDeepLogicChains tests long chains of logic operators.
func (s *StressNestingTestSuite) TestDeepLogicChains() {
	tests := []struct {
		count int
		op    string
		desc  string
	}{
		{20, "/\\", "20-term conjunction"},
		{20, "\\/", "20-term disjunction"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			// Build: a /\ b /\ c /\ ...
			parts := make([]string, tt.count)
			for i := range tt.count {
				parts[i] = fmt.Sprintf("x%d", i)
			}
			input := strings.Join(parts, " "+tt.op+" ")
			_, err := ParseExpression(input)
			s.Require().NoError(err, "should parse %d-term logic chain", tt.count)
		})
	}
}

// TestDeepNestedIF tests deeply nested IF-THEN-ELSE expressions.
func (s *StressNestingTestSuite) TestDeepNestedIF() {
	tests := []struct {
		depth int
		desc  string
	}{
		{3, "3-level nested IF"},
		{5, "5-level nested IF"},
		{10, "10-level nested IF"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			// Build: IF TRUE THEN IF TRUE THEN ... THEN 1 ELSE 2 ... ELSE n+1
			var b strings.Builder
			for range tt.depth {
				b.WriteString("IF TRUE THEN ")
			}
			b.WriteString("1")
			for i := range tt.depth {
				fmt.Fprintf(&b, " ELSE %d", i+2)
			}
			_, err := ParseExpression(b.String())
			s.Require().NoError(err, "should parse %d-level nested IF", tt.depth)
		})
	}
}

// TestDeepNestedTuples tests deeply nested tuple expressions.
func (s *StressNestingTestSuite) TestDeepNestedTuples() {
	tests := []struct {
		depth int
		desc  string
	}{
		{3, "3-level nested tuples"},
		{5, "5-level nested tuples"},
		{10, "10-level nested tuples"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			// Build: <<1, <<2, <<3, <<4>>>>>>
			var b strings.Builder
			for i := range tt.depth {
				fmt.Fprintf(&b, "<<%d, ", i+1)
			}
			fmt.Fprintf(&b, "%d", tt.depth+1)
			for range tt.depth {
				b.WriteString(">>")
			}
			_, err := ParseExpression(b.String())
			s.Require().NoError(err, "should parse %d-level nested tuples", tt.depth)
		})
	}
}

// TestDeepNestedSets tests deeply nested set expressions.
func (s *StressNestingTestSuite) TestDeepNestedSets() {
	tests := []struct {
		depth int
		desc  string
	}{
		{3, "3-level nested sets"},
		{5, "5-level nested sets"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			// Build: {1, {2, {3, {4}}}}
			var b strings.Builder
			for i := range tt.depth {
				fmt.Fprintf(&b, "{%d, ", i+1)
			}
			fmt.Fprintf(&b, "%d", tt.depth+1)
			for range tt.depth {
				b.WriteString("}")
			}
			_, err := ParseExpression(b.String())
			s.Require().NoError(err, "should parse %d-level nested sets", tt.depth)
		})
	}
}

// TestDeepNestedRecords tests deeply nested record expressions.
func (s *StressNestingTestSuite) TestDeepNestedRecords() {
	expr := `[a |-> [b |-> [c |-> 1]]]`
	_, err := ParseExpression(expr)
	s.Require().NoError(err, "should parse nested records")
}

// TestDeepSetUnionChain tests a long chain of set union operations.
func (s *StressNestingTestSuite) TestDeepSetUnionChain() {
	// Build: {1} ∪ {2} ∪ {3} ∪ ... ∪ {10}
	parts := make([]string, 10)
	for i := range 10 {
		parts[i] = fmt.Sprintf("{%d}", i+1)
	}
	input := strings.Join(parts, " ∪ ")
	_, err := ParseExpression(input)
	s.Require().NoError(err, "should parse 10-term set union chain")
}

// TestDeepNestedArithmetic tests deeply left-nested arithmetic.
// NOTE: Depth 20+ causes the PEG parser to hang due to exponential
// backtracking in state cloning. This is a known limitation.
func (s *StressNestingTestSuite) TestDeepNestedArithmetic() {
	// Build: ((((1 + 2) + 3) + 4) + 5)
	depths := []int{5, 10}
	for _, depth := range depths {
		s.Run(fmt.Sprintf("depth_%d", depth), func() {
			var b strings.Builder
			b.WriteString(strings.Repeat("(", depth-1))
			b.WriteString("1 + 2")
			for i := 3; i <= depth; i++ {
				fmt.Fprintf(&b, ") + %d", i)
			}
			b.WriteString(")")
			_, err := ParseExpression(b.String())
			s.Require().NoError(err, "should parse %d-level nested arithmetic", depth)
		})
	}
}

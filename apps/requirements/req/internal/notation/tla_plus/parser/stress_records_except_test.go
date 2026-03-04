package parser

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/ast"
	"github.com/stretchr/testify/suite"
)

// StressRecordExceptTestSuite stress-tests record literal and EXCEPT edge cases.
// EXCEPT patterns with `@`, nested field updates, and complex value expressions
// are thinly tested elsewhere.
type StressRecordExceptTestSuite struct {
	suite.Suite
}

func TestStressRecordExceptSuite(t *testing.T) {
	suite.Run(t, new(StressRecordExceptTestSuite))
}

// TestRecordLiteralVariations tests various record literal constructions.
func (s *StressRecordExceptTestSuite) TestRecordLiteralVariations() {
	tests := []struct {
		input    string
		desc     string
		nFields  int
	}{
		{`[a |-> 1]`, "single field record", 1},
		{`[a |-> 1, b |-> 2]`, "two field record", 2},
		{`[a |-> 1, b |-> 2, c |-> 3]`, "three field record", 3},
		{`[x |-> TRUE, y |-> FALSE]`, "boolean values", 2},
		{`[name |-> "hello"]`, "string value", 1},
		{`[x |-> <<1, 2>>]`, "tuple as value", 1},
		{`[x |-> {1, 2, 3}]`, "set as value", 1},
		{`[x |-> IF TRUE THEN 1 ELSE 2]`, "IF as value", 1},
		{`[a |-> [b |-> 1]]`, "nested record", 1},
		{`[a |-> [b |-> [c |-> 1]]]`, "deeply nested record", 1},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			expr, err := ParseExpression(tt.input)
			s.NoError(err, "should parse: %q (%s)", tt.input, tt.desc)

			rec, ok := expr.(*ast.RecordInstance)
			s.True(ok, "expected RecordInstance, got %T", expr)
			s.Len(rec.Bindings, tt.nFields, "field count for %q", tt.input)
		})
	}
}

// TestRecordExceptBasic tests basic EXCEPT expressions.
func (s *StressRecordExceptTestSuite) TestRecordExceptBasic() {
	tests := []struct {
		input       string
		desc        string
		nAlterations int
	}{
		{`[r EXCEPT !.x = 1]`, "single field update", 1},
		{`[r EXCEPT !.x = 1, !.y = 2]`, "two field updates", 2},
		{`[r EXCEPT !.x = 1, !.y = 2, !.z = 3]`, "three field updates", 3},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			expr, err := ParseExpression(tt.input)
			s.NoError(err, "should parse: %q (%s)", tt.input, tt.desc)

			rec, ok := expr.(*ast.RecordAltered)
			s.True(ok, "expected RecordAltered, got %T", expr)
			s.Len(rec.Alterations, tt.nAlterations, "alteration count for %q", tt.input)
		})
	}
}

// TestRecordExceptWithAt tests EXCEPT expressions using the @ (existing value) reference.
func (s *StressRecordExceptTestSuite) TestRecordExceptWithAt() {
	tests := []struct {
		input string
		desc  string
	}{
		{`[r EXCEPT !.x = @ + 1]`, "@ in addition"},
		{`[r EXCEPT !.x = @ * 2]`, "@ in multiplication"},
		{`[r EXCEPT !.x = @ - 1]`, "@ in subtraction"},
		{`[r EXCEPT !.x = IF @ > 0 THEN @ ELSE 0]`, "@ in IF condition and result"},
		{`[r EXCEPT !.x = @ ∪ {newItem}]`, "@ in set union"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			expr, err := ParseExpression(tt.input)
			s.NoError(err, "should parse: %q (%s)", tt.input, tt.desc)

			_, ok := expr.(*ast.RecordAltered)
			s.True(ok, "expected RecordAltered, got %T", expr)
		})
	}
}

// TestRecordExceptWithComplexValues tests EXCEPT with complex value expressions.
func (s *StressRecordExceptTestSuite) TestRecordExceptWithComplexValues() {
	tests := []struct {
		input string
		desc  string
	}{
		{`[r EXCEPT !.count = @ + 1, !.total = @ + amount]`, "different @ contexts per field"},
		{`[r EXCEPT !.name = "updated"]`, "string value in EXCEPT"},
		{`[r EXCEPT !.flag = TRUE]`, "boolean value in EXCEPT"},
		{`[r EXCEPT !.items = <<1, 2, 3>>]`, "tuple value in EXCEPT"},
		{`[r EXCEPT !.data = {1, 2, 3}]`, "set value in EXCEPT"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			expr, err := ParseExpression(tt.input)
			s.NoError(err, "should parse: %q (%s)", tt.input, tt.desc)

			_, ok := expr.(*ast.RecordAltered)
			s.True(ok, "expected RecordAltered, got %T", expr)
		})
	}
}

// TestChainedExceptFails verifies that chaining EXCEPT expressions is NOT
// supported. The grammar requires an Identifier (not an arbitrary expression)
// as the base record in EXCEPT, so `[[r EXCEPT ...] EXCEPT ...]` fails.
func (s *StressRecordExceptTestSuite) TestChainedExceptFails() {
	input := `[[r EXCEPT !.x = 1] EXCEPT !.y = 2]`
	_, err := ParseExpression(input)
	s.Error(err, "chained EXCEPT should fail: base must be an identifier, not an expression")
}

// TestRecordFieldAccess tests field access on records.
func (s *StressRecordExceptTestSuite) TestRecordFieldAccess() {
	tests := []struct {
		input string
		desc  string
	}{
		{`r.field`, "simple field access"},
		{`r.a.b`, "chained field access"},
		{`r.a.b.c`, "triple chained field access"},
		{`[x |-> 1, y |-> 2].x`, "field access on record literal"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			_, err := ParseExpression(tt.input)
			s.NoError(err, "should parse: %q (%s)", tt.input, tt.desc)
		})
	}
}

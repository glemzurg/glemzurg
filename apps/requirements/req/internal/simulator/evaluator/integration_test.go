package evaluator

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/parser"
	"github.com/stretchr/testify/suite"
)

// IntegrationTestSuite tests parsing all expression types and evaluating them.
type IntegrationTestSuite struct {
	suite.Suite
}

func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

// =============================================================================
// Literals
// =============================================================================

func (s *IntegrationTestSuite) TestLiterals() {
	tests := []struct {
		input    string
		expected string
	}{
		// Booleans (object.Boolean.Inspect() returns lowercase)
		{"TRUE", "true"},
		{"FALSE", "false"},
		// Natural numbers
		{"0", "0"},
		{"42", "42"},
		{"123", "123"},
		// Negative integers
		{"-1", "-1"},
		{"-42", "-42"},
		// Fractions (become rationals)
		{"1/2", "1/2"},
		{"3/4", "3/4"},
		// Decimals
		{"3.14", "157/50"},
		{"0.5", "1/2"},
		// Strings
		{`"hello"`, `"hello"`},
		{`"world"`, `"world"`},
		{`""`, `""`},
	}

	for _, tt := range tests {
		expr, err := parser.ParseExpression(tt.input)
		s.NoError(err, "parsing %q", tt.input)

		bindings := NewBindings()
		result := Eval(expr, bindings)
		s.False(result.IsError(), "evaluating %q: %v", tt.input, result.Error)
		s.Equal(tt.expected, result.Value.Inspect(), "evaluating %q", tt.input)
	}
}

// =============================================================================
// Arithmetic
// =============================================================================

func (s *IntegrationTestSuite) TestArithmetic() {
	tests := []struct {
		input    string
		expected string
	}{
		// Addition
		{"1 + 2", "3"},
		{"10 + 20 + 30", "60"},
		// Subtraction
		{"5 - 3", "2"},
		{"10 - 20", "-10"},
		// Multiplication
		{"2 * 3", "6"},
		{"4 * 5 * 2", "40"},
		// Division (Unicode)
		{"6 ÷ 2", "3"},
		{"10 ÷ 4", "5/2"},
		// Division (ASCII)
		{`10 \div 4`, "5/2"},
		// Modulo
		{"7 % 3", "1"},
		{"10 % 5", "0"},
		// Exponentiation
		{"2 ^ 3", "8"},
		{"3 ^ 2", "9"},
		// Precedence
		{"1 + 2 * 3", "7"},
		{"(1 + 2) * 3", "9"},
		// Negation
		{"-5", "-5"},
		{"--5", "5"},
	}

	for _, tt := range tests {
		expr, err := parser.ParseExpression(tt.input)
		s.NoError(err, "parsing %q", tt.input)

		bindings := NewBindings()
		result := Eval(expr, bindings)
		s.False(result.IsError(), "evaluating %q: %v", tt.input, result.Error)
		s.Equal(tt.expected, result.Value.Inspect(), "evaluating %q", tt.input)
	}
}

// =============================================================================
// Comparison
// =============================================================================

func (s *IntegrationTestSuite) TestComparison() {
	tests := []struct {
		input    string
		expected bool
	}{
		// Equality
		{"1 = 1", true},
		{"1 = 2", false},
		// Inequality (Unicode)
		{"1 ≠ 2", true},
		{"1 ≠ 1", false},
		// Inequality (ASCII)
		{"1 # 2", true},
		{`1 /= 2`, true},
		// Less than
		{"1 < 2", true},
		{"2 < 1", false},
		// Greater than
		{"2 > 1", true},
		{"1 > 2", false},
		// Less than or equal (Unicode)
		{"1 ≤ 1", true},
		{"1 ≤ 2", true},
		{"2 ≤ 1", false},
		// Less than or equal (ASCII)
		{"1 <= 2", true},
		// Greater than or equal (Unicode)
		{"2 ≥ 1", true},
		{"2 ≥ 2", true},
		{"1 ≥ 2", false},
		// Greater than or equal (ASCII)
		{"2 >= 1", true},
	}

	for _, tt := range tests {
		expr, err := parser.ParseExpression(tt.input)
		s.NoError(err, "parsing %q", tt.input)

		bindings := NewBindings()
		result := Eval(expr, bindings)
		s.False(result.IsError(), "evaluating %q: %v", tt.input, result.Error)

		b, ok := result.Value.(*object.Boolean)
		s.True(ok, "expected Boolean for %q", tt.input)
		s.Equal(tt.expected, b.Value(), "evaluating %q", tt.input)
	}
}

// =============================================================================
// Logic
// =============================================================================

func (s *IntegrationTestSuite) TestLogic() {
	tests := []struct {
		input    string
		expected bool
	}{
		// AND (Unicode)
		{"TRUE ∧ TRUE", true},
		{"TRUE ∧ FALSE", false},
		// AND (ASCII)
		{`TRUE /\ TRUE`, true},
		// OR (Unicode)
		{"FALSE ∨ TRUE", true},
		{"FALSE ∨ FALSE", false},
		// OR (ASCII)
		{`FALSE \/ TRUE`, true},
		// NOT (Unicode)
		{"¬TRUE", false},
		{"¬FALSE", true},
		// NOT (ASCII)
		{"~TRUE", false},
		// Implies (Unicode)
		{"TRUE ⇒ TRUE", true},
		{"TRUE ⇒ FALSE", false},
		{"FALSE ⇒ TRUE", true},
		{"FALSE ⇒ FALSE", true},
		// Implies (ASCII)
		{"TRUE => FALSE", false},
		// Equivalence (ASCII)
		{"TRUE <=> TRUE", true},
		{"TRUE <=> FALSE", false},
		{"FALSE <=> FALSE", true},
	}

	for _, tt := range tests {
		expr, err := parser.ParseExpression(tt.input)
		s.NoError(err, "parsing %q", tt.input)

		bindings := NewBindings()
		result := Eval(expr, bindings)
		s.False(result.IsError(), "evaluating %q: %v", tt.input, result.Error)

		b, ok := result.Value.(*object.Boolean)
		s.True(ok, "expected Boolean for %q", tt.input)
		s.Equal(tt.expected, b.Value(), "evaluating %q", tt.input)
	}
}

// =============================================================================
// Sets
// =============================================================================

func (s *IntegrationTestSuite) TestSets() {
	tests := []struct {
		input    string
		expected string
	}{
		// Set literals
		{"{}", "{}"},
		{"{1}", "{1}"},
		{"{1, 2, 3}", "{1, 2, 3}"},
		// Set range
		{"1..3", "{1, 2, 3}"},
		{"0..0", "{0}"},
		// Set union (Unicode)
		{"{1, 2} ∪ {2, 3}", "{1, 2, 3}"},
		// Set union (ASCII)
		{`{1} \union {2}`, "{1, 2}"},
		// Set intersection (Unicode)
		{"{1, 2, 3} ∩ {2, 3, 4}", "{2, 3}"},
		// Set intersection (ASCII)
		{`{1, 2} \intersect {2, 3}`, "{2}"},
		// Set difference
		{"{1, 2, 3} \\ {2}", "{1, 3}"},
	}

	for _, tt := range tests {
		expr, err := parser.ParseExpression(tt.input)
		s.NoError(err, "parsing %q", tt.input)

		bindings := NewBindings()
		result := Eval(expr, bindings)
		s.False(result.IsError(), "evaluating %q: %v", tt.input, result.Error)
		s.Equal(tt.expected, result.Value.Inspect(), "evaluating %q", tt.input)
	}
}

func (s *IntegrationTestSuite) TestSetMembership() {
	tests := []struct {
		input    string
		expected bool
	}{
		// Membership (Unicode)
		{"1 ∈ {1, 2, 3}", true},
		{"4 ∈ {1, 2, 3}", false},
		// Membership (ASCII)
		{`1 \in {1, 2}`, true},
		// Non-membership (Unicode)
		{"4 ∉ {1, 2, 3}", true},
		{"1 ∉ {1, 2, 3}", false},
		// Non-membership (ASCII)
		{`4 \notin {1, 2, 3}`, true},
		// Subset (Unicode)
		{"{1, 2} ⊆ {1, 2, 3}", true},
		{"{1, 4} ⊆ {1, 2, 3}", false},
		// Subset (ASCII)
		{`{1} \subseteq {1, 2}`, true},
	}

	for _, tt := range tests {
		expr, err := parser.ParseExpression(tt.input)
		s.NoError(err, "parsing %q", tt.input)

		bindings := NewBindings()
		result := Eval(expr, bindings)
		s.False(result.IsError(), "evaluating %q: %v", tt.input, result.Error)

		b, ok := result.Value.(*object.Boolean)
		s.True(ok, "expected Boolean for %q", tt.input)
		s.Equal(tt.expected, b.Value(), "evaluating %q", tt.input)
	}
}

// =============================================================================
// Quantifiers
// =============================================================================

func (s *IntegrationTestSuite) TestQuantifiers() {
	tests := []struct {
		input    string
		expected bool
	}{
		// Universal (Unicode)
		{"∀ x ∈ {1, 2, 3} : x > 0", true},
		{"∀ x ∈ {1, 2, 3} : x > 2", false},
		// Universal (ASCII)
		{`\A x \in {1, 2} : x > 0`, true},
		// Existential (Unicode)
		{"∃ x ∈ {1, 2, 3} : x = 2", true},
		{"∃ x ∈ {1, 2, 3} : x > 5", false},
		// Existential (ASCII)
		{`\E x \in {1, 2, 3} : x = 1`, true},
	}

	for _, tt := range tests {
		expr, err := parser.ParseExpression(tt.input)
		s.NoError(err, "parsing %q", tt.input)

		bindings := NewBindings()
		result := Eval(expr, bindings)
		s.False(result.IsError(), "evaluating %q: %v", tt.input, result.Error)

		b, ok := result.Value.(*object.Boolean)
		s.True(ok, "expected Boolean for %q", tt.input)
		s.Equal(tt.expected, b.Value(), "evaluating %q", tt.input)
	}
}

// =============================================================================
// Tuples
// =============================================================================

func (s *IntegrationTestSuite) TestTuples() {
	tests := []struct {
		input    string
		expected string
	}{
		// Tuple literals (ASCII)
		{"<<>>", "<<>>"},
		{"<<1>>", "<<1>>"},
		{"<<1, 2, 3>>", "<<1, 2, 3>>"},
		// Tuple literals (Unicode)
		{"⟨1, 2⟩", "<<1, 2>>"},
	}

	for _, tt := range tests {
		expr, err := parser.ParseExpression(tt.input)
		s.NoError(err, "parsing %q", tt.input)

		bindings := NewBindings()
		result := Eval(expr, bindings)
		s.False(result.IsError(), "evaluating %q: %v", tt.input, result.Error)
		s.Equal(tt.expected, result.Value.Inspect(), "evaluating %q", tt.input)
	}
}

func (s *IntegrationTestSuite) TestTupleIndex() {
	expr, err := parser.ParseExpression("<<10, 20, 30>>[2]")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)
	s.False(result.IsError(), "unexpected error: %v", result.Error)
	s.Equal("20", result.Value.Inspect())
}

// =============================================================================
// Records
// =============================================================================

func (s *IntegrationTestSuite) TestRecords() {
	// Record literal
	expr, err := parser.ParseExpression("[name |-> \"Alice\", age |-> 30]")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)
	s.False(result.IsError(), "unexpected error: %v", result.Error)

	record, ok := result.Value.(*object.Record)
	s.True(ok)
	s.Equal(`"Alice"`, record.Get("name").Inspect())
	s.Equal("30", record.Get("age").Inspect())
}

func (s *IntegrationTestSuite) TestRecordFieldAccess() {
	bindings := NewBindings()
	bindings.Set("r", object.NewRecordFromFields(map[string]object.Object{
		"x": object.NewInteger(10),
		"y": object.NewInteger(20),
	}), NamespaceGlobal)

	expr, err := parser.ParseExpression("r.x + r.y")
	s.NoError(err)

	result := Eval(expr, bindings)
	s.False(result.IsError(), "unexpected error: %v", result.Error)
	s.Equal("30", result.Value.Inspect())
}

func (s *IntegrationTestSuite) TestRecordExcept() {
	bindings := NewBindings()
	bindings.Set("r", object.NewRecordFromFields(map[string]object.Object{
		"count": object.NewInteger(5),
	}), NamespaceGlobal)

	expr, err := parser.ParseExpression("[r EXCEPT !.count = @ + 1]")
	s.NoError(err)

	result := Eval(expr, bindings)
	s.False(result.IsError(), "unexpected error: %v", result.Error)

	record, ok := result.Value.(*object.Record)
	s.True(ok)
	s.Equal("6", record.Get("count").Inspect())
}

// =============================================================================
// Control Flow
// =============================================================================

func (s *IntegrationTestSuite) TestIfThenElse() {
	tests := []struct {
		input    string
		expected string
	}{
		{"IF TRUE THEN 1 ELSE 2", "1"},
		{"IF FALSE THEN 1 ELSE 2", "2"},
		{"IF 1 > 0 THEN 10 ELSE 20", "10"},
		{"IF 1 < 0 THEN 10 ELSE 20", "20"},
	}

	for _, tt := range tests {
		expr, err := parser.ParseExpression(tt.input)
		s.NoError(err, "parsing %q", tt.input)

		bindings := NewBindings()
		result := Eval(expr, bindings)
		s.False(result.IsError(), "evaluating %q: %v", tt.input, result.Error)
		s.Equal(tt.expected, result.Value.Inspect(), "evaluating %q", tt.input)
	}
}

func (s *IntegrationTestSuite) TestCaseExpression() {
	bindings := NewBindings()
	bindings.Set("x", object.NewInteger(0), NamespaceGlobal)

	expr, err := parser.ParseExpression("CASE x > 0 -> 1 [] x < 0 -> -1 [] OTHER -> 0")
	s.NoError(err)

	result := Eval(expr, bindings)
	s.False(result.IsError(), "unexpected error: %v", result.Error)
	s.Equal("0", result.Value.Inspect())
}

// =============================================================================
// Function Calls (Built-in modules)
// =============================================================================

func (s *IntegrationTestSuite) TestBuiltinFunctionCalls() {
	tests := []struct {
		input    string
		expected string
	}{
		// _Seq module
		{"_Seq!Len(<<1, 2, 3>>)", "3"},
		{"_Seq!Head(<<10, 20, 30>>)", "10"},
		{"_Seq!Tail(<<1, 2, 3>>)", "<<2, 3>>"},
		{"_Seq!Append(<<1, 2>>, 3)", "<<1, 2, 3>>"},
		// _Stack module
		{"_Stack!Push(<<1, 2>>, 0)", "<<0, 1, 2>>"},
		{"_Stack!Pop(<<1, 2, 3>>)", "1"},
		// _Queue module
		{"_Queue!Enqueue(<<1, 2>>, 3)", "<<1, 2, 3>>"},
		{"_Queue!Dequeue(<<1, 2, 3>>)", "1"},
	}

	for _, tt := range tests {
		expr, err := parser.ParseExpression(tt.input)
		s.NoError(err, "parsing %q", tt.input)

		bindings := NewBindings()
		result := Eval(expr, bindings)
		s.False(result.IsError(), "evaluating %q: %v", tt.input, result.Error)
		s.Equal(tt.expected, result.Value.Inspect(), "evaluating %q", tt.input)
	}
}

// =============================================================================
// Complex Expressions
// =============================================================================

func (s *IntegrationTestSuite) TestComplexExpressions() {
	tests := []struct {
		input    string
		expected string
	}{
		// Nested function calls
		{"_Seq!Len(_Seq!Tail(<<1, 2, 3, 4>>))", "3"},
		// IF with function call
		{"IF _Seq!Len(<<1, 2>>) > 0 THEN _Seq!Head(<<1, 2>>) ELSE 0", "1"},
		// Quantifier with arithmetic
		{"∀ x ∈ {1, 2, 3} : x * 2 > x", "true"},
		// Nested IF
		{"IF TRUE THEN IF FALSE THEN 1 ELSE 2 ELSE 3", "2"},
		// Arithmetic in CASE
		{"CASE 1 + 1 = 2 -> 100 [] OTHER -> 0", "100"},
	}

	for _, tt := range tests {
		expr, err := parser.ParseExpression(tt.input)
		s.NoError(err, "parsing %q", tt.input)

		bindings := NewBindings()
		result := Eval(expr, bindings)
		s.False(result.IsError(), "evaluating %q: %v", tt.input, result.Error)
		s.Equal(tt.expected, result.Value.Inspect(), "evaluating %q", tt.input)
	}
}

// =============================================================================
// AST Validation
// =============================================================================

func (s *IntegrationTestSuite) TestASTValidation() {
	expressions := []string{
		"TRUE",
		"42",
		`"hello"`,
		"1 + 2 * 3",
		"{1, 2, 3}",
		"<<1, 2, 3>>",
		"[x |-> 1]",
		"∀ x ∈ {1} : x > 0",
		"IF TRUE THEN 1 ELSE 2",
		"CASE TRUE -> 1 [] OTHER -> 0",
		"_Seq!Len(<<1, 2>>)",
	}

	for _, input := range expressions {
		expr, err := parser.ParseExpression(input)
		s.NoError(err, "parsing %q", input)

		err = expr.Validate()
		s.NoError(err, "validating AST for %q", input)
	}
}

// =============================================================================
// Error Cases
// =============================================================================

func (s *IntegrationTestSuite) TestParseErrors() {
	badInputs := []struct {
		input string
		desc  string
	}{
		{"", "empty input"},
		{"(", "unmatched paren"},
		{"1 +", "incomplete expression"},
		{`"unclosed`, "unclosed string"},
		{"@#$%", "invalid tokens"},
		{"{1, 2,}", "trailing comma in set"},
	}

	for _, tt := range badInputs {
		_, err := parser.ParseExpression(tt.input)
		s.Error(err, "expected error for %s: %q", tt.desc, tt.input)
	}
}

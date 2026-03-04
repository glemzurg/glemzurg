package evaluator

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/parser"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/stretchr/testify/suite"
)

// StressIntegrationTestSuite tests the full pipeline: parse → eval.
// These tests catch bugs that only manifest when the AST is evaluated.
type StressIntegrationTestSuite struct {
	suite.Suite
}

func TestStressIntegrationSuite(t *testing.T) {
	suite.Run(t, new(StressIntegrationTestSuite))
}

// =============================================================================
// Precedence correctness through evaluation
// =============================================================================

func (s *StressIntegrationTestSuite) TestPrecedenceThroughEval() {
	tests := []struct {
		input    string
		expected string
		desc     string
	}{
		// Power right-associativity
		{"2 ^ 3 ^ 2", "512", "2^(3^2) = 2^9 = 512, NOT (2^3)^2 = 64"},
		{"2 ^ 3", "8", "simple power"},
		{"2 ^ 10", "1024", "large power"},
		{"3 ^ 3", "27", "3 cubed"},

		// Fraction chaining (left-associative)
		{"1/2/3", "1/6", "(1/2)/3 = 1/6"},

		// Fraction vs multiplication
		{"2 * 3/4", "3/2", "2 * (3/4) = 6/4 = 3/2"},
		{"3/4 * 2", "3/2", "(3/4) * 2 = 3/2"},

		// Mixed arithmetic precedence
		{"1 + 2 * 3", "7", "1 + (2*3) = 7"},
		{"(1 + 2) * 3", "9", "explicit parens"},
		{"1 + 2 * 3 - 4", "3", "1 + 6 - 4 = 3"},
		{"2 * 3 + 4 * 5", "26", "6 + 20 = 26"},
		{"(1 + 2) * (3 + 4)", "21", "3 * 7 = 21"},

		// Left-associative subtraction and division
		{"10 - 3 - 2", "5", "(10-3)-2 = 5, NOT 10-(3-2) = 9"},
		{"10 ÷ 2 ÷ 5", "1", "(10÷2)÷5 = 1"},

		// Power with other arithmetic
		{"2 ^ 3 * 4", "32", "(2^3) * 4 = 32"},
		{"4 * 2 ^ 3", "32", "4 * (2^3) = 32"},

		// Negation
		{"--5", "5", "double negation"},
		{"-(-5)", "5", "parenthesized double negation"},

		// Modulo
		{"7 % 3", "1", "modulo"},
		{"10 % 3 + 1", "2", "(10%3) + 1 = 2"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			expr, err := parser.ParseExpression(tt.input)
			s.Require().NoError(err, "parsing %q", tt.input)

			bindings := NewBindings()
			result := EvalAST(expr, bindings)
			s.False(result.IsError(), "evaluating %q: %v", tt.input, result.Error)
			s.Equal(tt.expected, result.Value.Inspect(), "evaluating %q", tt.input)
		})
	}
}

// TestNegationVsPowerEval verifies -2^2 = -(2^2) = -4.
// In TLA+ negation has lower precedence than power.
func (s *StressIntegrationTestSuite) TestNegationVsPowerEval() {
	expr, err := parser.ParseExpression("-2 ^ 2")
	s.Require().NoError(err, "parsing -2 ^ 2")

	bindings := NewBindings()
	result := EvalAST(expr, bindings)
	s.False(result.IsError(), "evaluating -2 ^ 2: %v", result.Error)
	s.Equal("-4", result.Value.Inspect(), "-2^2 must be -(2^2) = -4")
}

// TestNegationInFractionEval characterizes whether -3/4 is -(3/4) or (-3)/4.
func (s *StressIntegrationTestSuite) TestNegationInFractionEval() {
	expr, err := parser.ParseExpression("-3/4")
	s.Require().NoError(err, "parsing -3/4")

	bindings := NewBindings()
	result := EvalAST(expr, bindings)
	s.False(result.IsError(), "evaluating -3/4: %v", result.Error)

	// Both -(3/4) and (-3)/4 should give -3/4
	s.Equal("-3/4", result.Value.Inspect(), "evaluating -3/4")
}

// =============================================================================
// Nested quantifier evaluation
// =============================================================================

func (s *StressIntegrationTestSuite) TestNestedQuantifiers() {
	tests := []struct {
		input    string
		expected bool
		desc     string
	}{
		{`\A x \in {1, 2} : \A y \in {3, 4} : x < y`, true, "all x < all y"},
		{`\E x \in {1, 2} : \E y \in {1, 2} : x + y = 3`, true, "exists pair summing to 3"},
		{`\A x \in {1, 2, 3} : \E y \in {0, 1, 2, 3} : y = x - 1`, true, "for each x, exists predecessor"},
		{`\E x \in {1, 2} : \A y \in {1, 2} : x <= y`, true, "exists min element (x=1)"},
		{`\A x \in {} : FALSE`, true, "vacuously true (empty domain)"},
		{`\E x \in {} : TRUE`, false, "no elements in empty domain"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			expr, err := parser.ParseExpression(tt.input)
			s.Require().NoError(err, "parsing %q", tt.input)

			bindings := NewBindings()
			result := EvalAST(expr, bindings)
			s.False(result.IsError(), "evaluating %q: %v", tt.input, result.Error)

			b, ok := result.Value.(*object.Boolean)
			s.Require().True(ok, "expected Boolean for %q, got %T", tt.input, result.Value)
			s.Equal(tt.expected, b.Value(), "evaluating %q", tt.input)
		})
	}
}

// =============================================================================
// Complex set operations through evaluation
// =============================================================================

func (s *StressIntegrationTestSuite) TestComplexSetOperations() {
	tests := []struct {
		input    string
		expected string
		desc     string
	}{
		{"{1, 2, 3} ∪ {3, 4, 5}", "{1, 2, 3, 4, 5}", "set union"},
		{"{1, 2, 3} ∩ {2, 3, 4}", "{2, 3}", "set intersection"},
		{"{1, 2, 3} \\ {2}", "{1, 3}", "set difference"},
		{"({1, 2} ∪ {3}) ∩ {2, 3}", "{2, 3}", "union then intersect"},
		{"1..5", "{1, 2, 3, 4, 5}", "set range"},
		{"0..0", "{0}", "single element range"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			expr, err := parser.ParseExpression(tt.input)
			s.Require().NoError(err, "parsing %q", tt.input)

			bindings := NewBindings()
			result := EvalAST(expr, bindings)
			s.False(result.IsError(), "evaluating %q: %v", tt.input, result.Error)
			s.Equal(tt.expected, result.Value.Inspect(), "evaluating %q", tt.input)
		})
	}
}

func (s *StressIntegrationTestSuite) TestSetMembershipOperations() {
	tests := []struct {
		input    string
		expected bool
		desc     string
	}{
		{"{1, 2} ⊆ {1, 2, 3}", true, "subset"},
		{"{1, 4} ⊆ {1, 2, 3}", false, "not subset"},
		{`1 \in ({1, 2} ∪ {3, 4})`, true, "membership after union"},
		{`5 \notin {1, 2, 3}`, true, "non-membership"},
		{`1 \in 1..5`, true, "membership in range"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			expr, err := parser.ParseExpression(tt.input)
			s.Require().NoError(err, "parsing %q", tt.input)

			bindings := NewBindings()
			result := EvalAST(expr, bindings)
			s.False(result.IsError(), "evaluating %q: %v", tt.input, result.Error)

			b, ok := result.Value.(*object.Boolean)
			s.Require().True(ok, "expected Boolean for %q", tt.input)
			s.Equal(tt.expected, b.Value(), "evaluating %q", tt.input)
		})
	}
}

// =============================================================================
// Set filter (set comprehension) evaluation
// =============================================================================

func (s *StressIntegrationTestSuite) TestSetFilterEval() {
	tests := []struct {
		input    string
		expected string
		desc     string
	}{
		{`{x \in {1, 2, 3, 4, 5} : x > 3}`, "{4, 5}", "filter elements > 3"},
		{`{x \in {1, 2, 3} : x = 2}`, "{2}", "filter for equality"},
		{`{x \in {1, 2, 3} : FALSE}`, "{}", "filter with FALSE predicate (empty result)"},
		{`{x \in {1, 2, 3} : TRUE}`, "{1, 2, 3}", "filter with TRUE predicate (all elements)"},
		{`{x \in {} : x > 0}`, "{}", "filter over empty set"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			expr, err := parser.ParseExpression(tt.input)
			s.Require().NoError(err, "parsing %q", tt.input)

			bindings := NewBindings()
			result := EvalAST(expr, bindings)
			s.False(result.IsError(), "evaluating %q: %v", tt.input, result.Error)
			s.Equal(tt.expected, result.Value.Inspect(), "evaluating %q", tt.input)
		})
	}
}

// =============================================================================
// Record and tuple evaluation
// =============================================================================

func (s *StressIntegrationTestSuite) TestRecordAndTupleEval() {
	tests := []struct {
		input    string
		expected string
		desc     string
	}{
		{"<<1, 2, 3>>[2]", "2", "tuple index (1-based)"},
		{"<<10, 20, 30>>[1]", "10", "tuple first element"},
		{"<<10, 20, 30>>[3]", "30", "tuple last element"},
		{"_Seq!Len(<<1, 2, 3>>)", "3", "sequence length"},
		{"_Seq!Append(<<1, 2>>, 3)", "<<1, 2, 3>>", "sequence append"},
		{"_Seq!Head(_Seq!Tail(<<1, 2, 3>>))", "2", "nested function calls"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			expr, err := parser.ParseExpression(tt.input)
			s.Require().NoError(err, "parsing %q", tt.input)

			bindings := NewBindings()
			result := EvalAST(expr, bindings)
			s.False(result.IsError(), "evaluating %q: %v", tt.input, result.Error)
			s.Equal(tt.expected, result.Value.Inspect(), "evaluating %q", tt.input)
		})
	}
}

func (s *StressIntegrationTestSuite) TestRecordFieldAccessEval() {
	bindings := NewBindings()
	bindings.Set("r", object.NewRecordFromFields(map[string]object.Object{
		"x": object.NewInteger(10),
		"y": object.NewInteger(20),
	}), NamespaceGlobal)

	tests := []struct {
		input    string
		expected string
		desc     string
	}{
		{"r.x", "10", "simple field access"},
		{"r.y", "20", "simple field access y"},
		{"r.x + r.y", "30", "field access in arithmetic"},
		{"r.x * 2", "20", "field access with multiplication"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			expr, err := parser.ParseExpression(tt.input)
			s.Require().NoError(err, "parsing %q", tt.input)

			result := EvalAST(expr, bindings)
			s.False(result.IsError(), "evaluating %q: %v", tt.input, result.Error)
			s.Equal(tt.expected, result.Value.Inspect(), "evaluating %q", tt.input)
		})
	}
}

func (s *StressIntegrationTestSuite) TestRecordExceptEval() {
	bindings := NewBindings()
	bindings.Set("r", object.NewRecordFromFields(map[string]object.Object{
		"count": object.NewInteger(5),
		"name":  object.NewString("test"),
	}), NamespaceGlobal)

	tests := []struct {
		input    string
		field    string
		expected string
		desc     string
	}{
		{"[r EXCEPT !.count = @ + 1]", "count", "6", "EXCEPT with @ arithmetic"},
		{`[r EXCEPT !.name = "updated"]`, "name", `"updated"`, "EXCEPT with string value"},
		{"[r EXCEPT !.count = 0]", "count", "0", "EXCEPT with literal value"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			expr, err := parser.ParseExpression(tt.input)
			s.Require().NoError(err, "parsing %q", tt.input)

			result := EvalAST(expr, bindings)
			s.False(result.IsError(), "evaluating %q: %v", tt.input, result.Error)

			record, ok := result.Value.(*object.Record)
			s.Require().True(ok, "expected Record for %q", tt.input)
			s.Equal(tt.expected, record.Get(tt.field).Inspect(), "field %s of %q", tt.field, tt.input)
		})
	}
}

// =============================================================================
// Logic short-circuit verification
// =============================================================================

func (s *StressIntegrationTestSuite) TestLogicShortCircuit() {
	tests := []struct {
		input    string
		expected bool
		desc     string
	}{
		// These test short-circuit behavior: the undefined_var would cause
		// an error if evaluated, but short-circuit should skip it.
		{"TRUE /\\ TRUE /\\ FALSE", false, "three-way AND"},
		{"FALSE \\/ FALSE \\/ TRUE", true, "three-way OR"},
		{"¬¬TRUE", true, "double NOT"},
		{"FALSE => TRUE", true, "false implies anything"},
		{"FALSE => FALSE", true, "false implies false"},
		{"TRUE => TRUE", true, "true implies true"},
		{"TRUE => FALSE", false, "true does not imply false"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			expr, err := parser.ParseExpression(tt.input)
			s.Require().NoError(err, "parsing %q", tt.input)

			bindings := NewBindings()
			result := EvalAST(expr, bindings)
			s.False(result.IsError(), "evaluating %q: %v", tt.input, result.Error)

			b, ok := result.Value.(*object.Boolean)
			s.Require().True(ok, "expected Boolean for %q", tt.input)
			s.Equal(tt.expected, b.Value(), "evaluating %q", tt.input)
		})
	}
}

// TestShortCircuitSkipsUndefined verifies that short-circuit evaluation
// actually avoids evaluating the right side when the left determines the result.
func (s *StressIntegrationTestSuite) TestShortCircuitSkipsUndefined() {
	// TRUE \/ undefined_var: should return TRUE without looking up undefined_var
	expr, err := parser.ParseExpression("TRUE \\/ undefined_var")
	s.Require().NoError(err)

	bindings := NewBindings()
	// Don't set undefined_var in bindings
	result := EvalAST(expr, bindings)
	s.False(result.IsError(), "short-circuit OR should not evaluate right side")

	b, ok := result.Value.(*object.Boolean)
	s.Require().True(ok)
	s.True(b.Value())

	// FALSE /\ undefined_var: should return FALSE without looking up undefined_var
	expr2, err := parser.ParseExpression("FALSE /\\ undefined_var")
	s.Require().NoError(err)

	result2 := EvalAST(expr2, bindings)
	s.False(result2.IsError(), "short-circuit AND should not evaluate right side")

	b2, ok := result2.Value.(*object.Boolean)
	s.Require().True(ok)
	s.False(b2.Value())
}

// =============================================================================
// Control flow evaluation
// =============================================================================

func (s *StressIntegrationTestSuite) TestControlFlowEval() {
	tests := []struct {
		input    string
		expected string
		desc     string
	}{
		{"IF TRUE THEN 1 ELSE 2", "1", "basic IF true"},
		{"IF FALSE THEN 1 ELSE 2", "2", "basic IF false"},
		{"IF 2 > 1 THEN 2 ^ 3 ELSE 0", "8", "IF with comparison and power"},
		{"IF TRUE THEN IF FALSE THEN 1 ELSE 2 ELSE 3", "2", "nested IF"},
		{"IF FALSE THEN 1 ELSE IF TRUE THEN 2 ELSE 3", "2", "IF-ELSE-IF chain"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			expr, err := parser.ParseExpression(tt.input)
			s.Require().NoError(err, "parsing %q", tt.input)

			bindings := NewBindings()
			result := EvalAST(expr, bindings)
			s.False(result.IsError(), "evaluating %q: %v", tt.input, result.Error)
			s.Equal(tt.expected, result.Value.Inspect(), "evaluating %q", tt.input)
		})
	}
}

func (s *StressIntegrationTestSuite) TestCaseEval() {
	bindings := NewBindings()

	tests := []struct {
		input    string
		xVal     int64
		expected string
		desc     string
	}{
		{"CASE x > 0 -> 1 [] x < 0 -> -1 [] OTHER -> 0", 5, "1", "positive"},
		{"CASE x > 0 -> 1 [] x < 0 -> -1 [] OTHER -> 0", -3, "-1", "negative"},
		{"CASE x > 0 -> 1 [] x < 0 -> -1 [] OTHER -> 0", 0, "0", "zero (OTHER)"},
		{"CASE 1 + 1 = 2 -> 100 [] OTHER -> 0", 0, "100", "arithmetic in condition"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			bindings.Set("x", object.NewInteger(tt.xVal), NamespaceGlobal)

			expr, err := parser.ParseExpression(tt.input)
			s.Require().NoError(err, "parsing %q", tt.input)

			result := EvalAST(expr, bindings)
			s.False(result.IsError(), "evaluating %q: %v", tt.input, result.Error)
			s.Equal(tt.expected, result.Value.Inspect(), "evaluating %q (x=%d)", tt.input, tt.xVal)
		})
	}
}

// =============================================================================
// Quantifier with complex predicates
// =============================================================================

func (s *StressIntegrationTestSuite) TestQuantifierComplexPredicates() {
	tests := []struct {
		input    string
		expected bool
		desc     string
	}{
		{`\A x \in {1, 2, 3} : x > 0 /\ x < 10`, true, "conjunction in predicate"},
		{`\A x \in {1, 2, 3} : x > 0 => x * x > 0`, true, "implies in predicate"},
		{`\E x \in {1, 2, 3} : x \in {2, 3, 4}`, true, "membership in predicate"},
		{`\A x \in {1, 2, 3} : ¬(x = 0)`, true, "negation in predicate"},
		{`\E x \in {1, 2, 3, 4, 5} : x * x = 9`, true, "exists perfect square"},
		{`\A x \in {2, 4, 6} : x % 2 = 0`, true, "all even"},
		{`\E x \in {1, 3, 5} : x % 2 = 0`, false, "no even in odd set"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			expr, err := parser.ParseExpression(tt.input)
			s.Require().NoError(err, "parsing %q", tt.input)

			bindings := NewBindings()
			result := EvalAST(expr, bindings)
			s.False(result.IsError(), "evaluating %q: %v", tt.input, result.Error)

			b, ok := result.Value.(*object.Boolean)
			s.Require().True(ok, "expected Boolean for %q", tt.input)
			s.Equal(tt.expected, b.Value(), "evaluating %q", tt.input)
		})
	}
}

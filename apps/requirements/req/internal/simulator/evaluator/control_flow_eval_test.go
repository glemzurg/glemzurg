package evaluator

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/parser"
	"github.com/stretchr/testify/suite"
)

func TestControlFlowEvalSuite(t *testing.T) {
	suite.Run(t, new(ControlFlowEvalSuite))
}

type ControlFlowEvalSuite struct {
	suite.Suite
}

// =============================================================================
// IF-THEN-ELSE Evaluation
// =============================================================================

func (s *ControlFlowEvalSuite) TestIfThenElse_TrueCondition() {
	expr, err := parser.ParseExpression("IF TRUE THEN 1 ELSE 2")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	s.Equal("1", result.Value.Inspect())
}

func (s *ControlFlowEvalSuite) TestIfThenElse_FalseCondition() {
	expr, err := parser.ParseExpression("IF FALSE THEN 1 ELSE 2")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	s.Equal("2", result.Value.Inspect())
}

func (s *ControlFlowEvalSuite) TestIfThenElse_WithComparison() {
	expr, err := parser.ParseExpression("IF x > 0 THEN x ELSE -x")
	s.NoError(err)

	// Test with positive x
	bindings := NewBindings()
	bindings.Set("x", object.NewInteger(5), NamespaceGlobal)
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	s.Equal("5", result.Value.Inspect())

	// Test with negative x
	bindings2 := NewBindings()
	bindings2.Set("x", object.NewInteger(-5), NamespaceGlobal)
	result2 := Eval(expr, bindings2)

	s.False(result2.IsError(), "unexpected error: %v", result2.Error)
	s.Equal("5", result2.Value.Inspect()) // -(-5) = 5
}

func (s *ControlFlowEvalSuite) TestIfThenElse_Nested() {
	// Signum function: returns -1, 0, or 1
	expr, err := parser.ParseExpression("IF x > 0 THEN 1 ELSE IF x < 0 THEN -1 ELSE 0")
	s.NoError(err)

	// Positive
	bindings := NewBindings()
	bindings.Set("x", object.NewInteger(5), NamespaceGlobal)
	result := Eval(expr, bindings)
	s.Equal("1", result.Value.Inspect())

	// Negative
	bindings2 := NewBindings()
	bindings2.Set("x", object.NewInteger(-5), NamespaceGlobal)
	result2 := Eval(expr, bindings2)
	s.Equal("-1", result2.Value.Inspect())

	// Zero
	bindings3 := NewBindings()
	bindings3.Set("x", object.NewInteger(0), NamespaceGlobal)
	result3 := Eval(expr, bindings3)
	s.Equal("0", result3.Value.Inspect())
}

func (s *ControlFlowEvalSuite) TestIfThenElse_NonBooleanCondition() {
	expr, err := parser.ParseExpression("IF 1 THEN 2 ELSE 3")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.True(result.IsError())
	s.Contains(result.Error.Message, "Boolean")
}

// =============================================================================
// CASE Expression Evaluation
// =============================================================================

func (s *ControlFlowEvalSuite) TestCaseExpr_FirstMatch() {
	expr, err := parser.ParseExpression("CASE x > 0 -> 1 [] x < 0 -> 2 [] OTHER -> 0")
	s.NoError(err)

	bindings := NewBindings()
	bindings.Set("x", object.NewInteger(5), NamespaceGlobal)
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	s.Equal("1", result.Value.Inspect())
}

func (s *ControlFlowEvalSuite) TestCaseExpr_SecondMatch() {
	expr, err := parser.ParseExpression("CASE x > 0 -> 1 [] x < 0 -> 2 [] OTHER -> 0")
	s.NoError(err)

	bindings := NewBindings()
	bindings.Set("x", object.NewInteger(-5), NamespaceGlobal)
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	s.Equal("2", result.Value.Inspect())
}

func (s *ControlFlowEvalSuite) TestCaseExpr_OtherMatch() {
	expr, err := parser.ParseExpression("CASE x > 0 -> 1 [] x < 0 -> 2 [] OTHER -> 0")
	s.NoError(err)

	bindings := NewBindings()
	bindings.Set("x", object.NewInteger(0), NamespaceGlobal)
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	s.Equal("0", result.Value.Inspect())
}

func (s *ControlFlowEvalSuite) TestCaseExpr_NoMatchNoOther() {
	expr, err := parser.ParseExpression("CASE x > 0 -> 1 [] x < 0 -> 2")
	s.NoError(err)

	bindings := NewBindings()
	bindings.Set("x", object.NewInteger(0), NamespaceGlobal)
	result := Eval(expr, bindings)

	s.True(result.IsError())
	s.Contains(result.Error.Message, "no branch matched")
}

func (s *ControlFlowEvalSuite) TestCaseExpr_WithExpressions() {
	expr, err := parser.ParseExpression("CASE n >= 0 -> n * 2 [] OTHER -> -n")
	s.NoError(err)

	// Positive
	bindings := NewBindings()
	bindings.Set("n", object.NewInteger(5), NamespaceGlobal)
	result := Eval(expr, bindings)
	s.Equal("10", result.Value.Inspect()) // 5 * 2 = 10

	// Negative
	bindings2 := NewBindings()
	bindings2.Set("n", object.NewInteger(-5), NamespaceGlobal)
	result2 := Eval(expr, bindings2)
	s.Equal("5", result2.Value.Inspect()) // -(-5) = 5
}

// =============================================================================
// Function Call Evaluation
// =============================================================================

func (s *ControlFlowEvalSuite) TestFunctionCall_Seq_Len() {
	expr, err := parser.ParseExpression("_Seq!Len(<<1, 2, 3>>)")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	s.Equal("3", result.Value.Inspect())
}

func (s *ControlFlowEvalSuite) TestFunctionCall_Seq_Len_Variable() {
	expr, err := parser.ParseExpression("_Seq!Len(seq)")
	s.NoError(err)

	bindings := NewBindings()
	bindings.Set("seq", object.NewTupleFromElements([]object.Object{
		object.NewInteger(1),
		object.NewInteger(2),
		object.NewInteger(3),
		object.NewInteger(4),
		object.NewInteger(5),
	}), NamespaceGlobal)
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	s.Equal("5", result.Value.Inspect())
}

func (s *ControlFlowEvalSuite) TestFunctionCall_Seq_Head() {
	expr, err := parser.ParseExpression("_Seq!Head(<<1, 2, 3>>)")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	s.Equal("1", result.Value.Inspect())
}

func (s *ControlFlowEvalSuite) TestFunctionCall_Seq_Tail() {
	expr, err := parser.ParseExpression("_Seq!Tail(<<1, 2, 3>>)")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	tuple, ok := result.Value.(*object.Tuple)
	s.True(ok)
	s.Equal(2, tuple.Len())
	s.Equal("2", tuple.At(1).Inspect())
	s.Equal("3", tuple.At(2).Inspect())
}

func (s *ControlFlowEvalSuite) TestFunctionCall_Seq_Append() {
	expr, err := parser.ParseExpression("_Seq!Append(<<1, 2>>, 3)")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	tuple, ok := result.Value.(*object.Tuple)
	s.True(ok)
	s.Equal(3, tuple.Len())
	s.Equal("3", tuple.At(3).Inspect())
}

func (s *ControlFlowEvalSuite) TestFunctionCall_Nested() {
	// _Seq!Len(_Seq!Tail(seq)) where seq = <<1, 2, 3>>
	expr, err := parser.ParseExpression("_Seq!Len(_Seq!Tail(<<1, 2, 3>>))")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	s.Equal("2", result.Value.Inspect()) // Tail has 2 elements
}

func (s *ControlFlowEvalSuite) TestFunctionCall_Stack_Push() {
	expr, err := parser.ParseExpression("_Stack!Push(<<1, 2>>, 0)")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	tuple, ok := result.Value.(*object.Tuple)
	s.True(ok)
	s.Equal(3, tuple.Len())
	s.Equal("0", tuple.At(1).Inspect()) // 0 is pushed to front
}

func (s *ControlFlowEvalSuite) TestFunctionCall_Stack_Pop() {
	expr, err := parser.ParseExpression("_Stack!Pop(<<1, 2, 3>>)")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	s.Equal("1", result.Value.Inspect()) // Returns top element
}

func (s *ControlFlowEvalSuite) TestFunctionCall_Queue_Enqueue() {
	expr, err := parser.ParseExpression("_Queue!Enqueue(<<1, 2>>, 3)")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	tuple, ok := result.Value.(*object.Tuple)
	s.True(ok)
	s.Equal(3, tuple.Len())
	s.Equal("3", tuple.At(3).Inspect()) // 3 is appended to end
}

func (s *ControlFlowEvalSuite) TestFunctionCall_Queue_Dequeue() {
	expr, err := parser.ParseExpression("_Queue!Dequeue(<<1, 2, 3>>)")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	s.Equal("1", result.Value.Inspect()) // Returns front element
}

func (s *ControlFlowEvalSuite) TestFunctionCall_UnknownFunction() {
	expr, err := parser.ParseExpression("UnknownFunc(1)")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.True(result.IsError())
	s.Contains(result.Error.Message, "unknown function")
}

// =============================================================================
// Combined Tests
// =============================================================================

func (s *ControlFlowEvalSuite) TestCombined_IfWithFunctionCall() {
	expr, err := parser.ParseExpression("IF _Seq!Len(seq) > 0 THEN _Seq!Head(seq) ELSE 0")
	s.NoError(err)

	// Non-empty sequence
	bindings := NewBindings()
	bindings.Set("seq", object.NewTupleFromElements([]object.Object{
		object.NewInteger(42),
		object.NewInteger(43),
	}), NamespaceGlobal)
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	s.Equal("42", result.Value.Inspect())

	// Empty sequence
	bindings2 := NewBindings()
	bindings2.Set("seq", object.NewTupleFromElements([]object.Object{}), NamespaceGlobal)
	result2 := Eval(expr, bindings2)

	s.False(result2.IsError(), "unexpected error: %v", result2.Error)
	s.Equal("0", result2.Value.Inspect())
}

func (s *ControlFlowEvalSuite) TestCombined_FunctionCallInQuantifier() {
	// All sequences have length > 0
	expr, err := parser.ParseExpression("∀ seq ∈ {<<1>>, <<1, 2>>, <<1, 2, 3>>} : _Seq!Len(seq) > 0")
	s.NoError(err)

	bindings := NewBindings()
	result := Eval(expr, bindings)

	s.False(result.IsError(), "unexpected error: %v", result.Error)
	b, ok := result.Value.(*object.Boolean)
	s.True(ok)
	s.True(b.Value())
}

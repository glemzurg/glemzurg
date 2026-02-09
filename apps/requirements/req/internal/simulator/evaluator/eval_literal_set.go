package evaluator

import (
	"github.com/glemzurg/go-tlaplus/internal/simulator/ast"
	"github.com/glemzurg/go-tlaplus/internal/simulator/object"
)

// evalSetLiteral evaluates a general set literal {expr1, expr2, ...}.
func evalSetLiteral(node *ast.SetLiteral, bindings *Bindings) *EvalResult {
	elements := make([]object.Object, 0, len(node.Elements))

	for _, elemExpr := range node.Elements {
		result := Eval(elemExpr, bindings)
		if result.IsError() {
			return result
		}
		elements = append(elements, result.Value)
	}

	return NewEvalResult(object.NewSetFromElements(elements))
}

// evalSetRangeExpr evaluates a dynamic set range expression (start..end).
func evalSetRangeExpr(node *ast.SetRangeExpr, bindings *Bindings) *EvalResult {
	// Evaluate start
	startResult := Eval(node.Start, bindings)
	if startResult.IsError() {
		return startResult
	}

	startNum, ok := startResult.Value.(*object.Number)
	if !ok {
		return NewEvalError("set range start must be a number, got %s", startResult.Value.Type())
	}

	// Check if it's an integer (denominator must be 1)
	if !startNum.Rat().IsInt() {
		return NewEvalError("set range start must be an integer, got %s", startNum.Inspect())
	}

	// Evaluate end
	endResult := Eval(node.End, bindings)
	if endResult.IsError() {
		return endResult
	}

	endNum, ok := endResult.Value.(*object.Number)
	if !ok {
		return NewEvalError("set range end must be a number, got %s", endResult.Value.Type())
	}

	if !endNum.Rat().IsInt() {
		return NewEvalError("set range end must be an integer, got %s", endNum.Inspect())
	}

	// Get integer values
	start := startNum.Rat().Num().Int64()
	end := endNum.Rat().Num().Int64()

	// Create set of integers from start to end (inclusive)
	elements := make([]object.Object, 0)
	for i := start; i <= end; i++ {
		elements = append(elements, object.NewInteger(i))
	}

	return NewEvalResult(object.NewSetFromElements(elements))
}

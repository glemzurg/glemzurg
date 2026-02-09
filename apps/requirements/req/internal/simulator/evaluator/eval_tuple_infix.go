package evaluator

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// evalTupleInfix evaluates tuple concatenation (∘).
func evalTupleInfix(node *ast.TupleInfixExpression, bindings *Bindings) *EvalResult {
	if len(node.Operands) < 2 {
		return NewEvalError("tuple concatenation requires at least 2 operands, got %d", len(node.Operands))
	}

	// Evaluate first operand to start the result
	firstResult := Eval(node.Operands[0], bindings)
	if firstResult.IsError() {
		return firstResult
	}

	result, ok := firstResult.Value.(*object.Tuple)
	if !ok {
		return NewEvalError("operand 1 must be Tuple, got %s", firstResult.Value.Type())
	}

	// Concatenate remaining operands
	for i := 1; i < len(node.Operands); i++ {
		opResult := Eval(node.Operands[i], bindings)
		if opResult.IsError() {
			return opResult
		}

		tuple, ok := opResult.Value.(*object.Tuple)
		if !ok {
			return NewEvalError("operand %d must be Tuple, got %s", i+1, opResult.Value.Type())
		}

		result = result.Concat(tuple)
	}

	return NewEvalResult(result)
}

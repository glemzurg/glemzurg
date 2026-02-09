package evaluator

import (
	"github.com/glemzurg/go-tlaplus/internal/simulator/ast"
	"github.com/glemzurg/go-tlaplus/internal/simulator/object"
)

// evalStringInfix evaluates string concatenation (∘).
func evalStringInfix(node *ast.StringInfixExpression, bindings *Bindings) *EvalResult {
	if len(node.Operands) < 2 {
		return NewEvalError("string concatenation requires at least 2 operands, got %d", len(node.Operands))
	}

	// Evaluate and concatenate all operands
	var result string
	for i, operand := range node.Operands {
		opResult := Eval(operand, bindings)
		if opResult.IsError() {
			return opResult
		}

		str, ok := opResult.Value.(*object.String)
		if !ok {
			return NewEvalError("operand %d must be String, got %s", i+1, opResult.Value.Type())
		}

		result += str.Value()
	}

	return NewEvalResult(object.NewString(result))
}

package evaluator

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// evalTupleLiteral evaluates a tuple literal.
func evalTupleLiteral(node *ast.TupleLiteral, bindings *Bindings) *EvalResult {
	elements := make([]object.Object, 0, len(node.Elements))

	for _, elemExpr := range node.Elements {
		result := Eval(elemExpr, bindings)
		if result.IsError() {
			return result
		}
		elements = append(elements, result.Value)
	}

	return NewEvalResult(object.NewTupleFromElements(elements))
}

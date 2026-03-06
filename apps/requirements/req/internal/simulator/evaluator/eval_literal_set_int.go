package evaluator

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// evalSetLiteralInt evaluates a set of integer literals.
func evalSetLiteralInt(node *ast.SetLiteralInt) *EvalResult {
	elements := make([]object.Object, 0, len(node.Values))

	for _, v := range node.Values {
		elements = append(elements, object.NewInteger(int64(v)))
	}

	return NewEvalResult(object.NewSetFromElements(elements))
}

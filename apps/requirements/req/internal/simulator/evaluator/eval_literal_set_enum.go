package evaluator

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// evalSetLiteralEnum evaluates a set of string enumeration values.
func evalSetLiteralEnum(node *ast.SetLiteralEnum) *EvalResult {
	elements := make([]object.Object, 0, len(node.Values))

	for _, v := range node.Values {
		elements = append(elements, object.NewString(v))
	}

	return NewEvalResult(object.NewSetFromElements(elements))
}

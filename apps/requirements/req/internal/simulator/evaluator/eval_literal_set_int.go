package evaluator

import (
	"github.com/glemzurg/go-tlaplus/internal/simulator/ast"
	"github.com/glemzurg/go-tlaplus/internal/simulator/object"
)

// evalSetLiteralInt evaluates a set of integer literals.
func evalSetLiteralInt(node *ast.SetLiteralInt) *EvalResult {
	elements := make([]object.Object, 0, len(node.Values))

	for _, v := range node.Values {
		elements = append(elements, object.NewInteger(int64(v)))
	}

	return NewEvalResult(object.NewSetFromElements(elements))
}

package evaluator

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// evalSetRange evaluates a range set (e.g., 1..10).
func evalSetRange(node *ast.SetRange, bindings *Bindings) *EvalResult {
	elements := make([]object.Object, 0, node.End-node.Start+1)

	for i := node.Start; i <= node.End; i++ {
		elements = append(elements, object.NewInteger(int64(i)))
	}

	return NewEvalResult(object.NewSetFromElements(elements))
}

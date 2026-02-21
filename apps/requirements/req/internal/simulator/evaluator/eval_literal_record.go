package evaluator

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// evalRecordInstance evaluates a record literal.
func evalRecordInstance(node *ast.RecordInstance, bindings *Bindings) *EvalResult {
	fields := make(map[string]object.Object, len(node.Bindings))

	for _, binding := range node.Bindings {
		fieldName := binding.Field.Value

		result := Eval(binding.Expression, bindings)
		if result.IsError() {
			return result
		}

		fields[fieldName] = result.Value
	}

	return NewEvalResult(object.NewRecordFromFields(fields))
}

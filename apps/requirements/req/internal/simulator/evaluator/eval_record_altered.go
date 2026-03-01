package evaluator

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// evalRecordAltered evaluates an EXCEPT expression [id EXCEPT !.field = expr, ...].
func evalRecordAltered(node *ast.RecordAltered, bindings *Bindings) *EvalResult {
	// Evaluate the base record
	baseResult := Eval(node.Identifier, bindings)
	if baseResult.IsError() {
		return baseResult
	}

	baseRecord, ok := baseResult.Value.(*object.Record)
	if !ok {
		return NewEvalError("EXCEPT requires Record, got %s", baseResult.Value.Type())
	}

	// Clone the record to avoid mutating the original
	result := baseRecord.Clone().(*object.Record)

	// Apply each alteration
	for _, alt := range node.Alterations {
		fieldName := alt.Field.Member

		// Get the current value for the @ reference
		currentValue := result.Get(fieldName)
		if currentValue == nil && alt.Field.Identifier == nil {
			// !.field where field doesn't exist
			return NewEvalError("field not found for EXCEPT: %s", fieldName)
		}

		// Create a child bindings with the existing value set
		childBindings := NewEnclosedBindings(bindings)
		if currentValue != nil {
			childBindings.SetExistingValue(currentValue)
		}

		// Evaluate the new value
		newValueResult := Eval(alt.Expression, childBindings)
		if newValueResult.IsError() {
			return newValueResult
		}

		// Update the field
		result.Set(fieldName, newValueResult.Value)
	}

	return NewEvalResult(result)
}

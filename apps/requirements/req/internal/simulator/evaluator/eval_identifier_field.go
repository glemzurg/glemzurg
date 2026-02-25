package evaluator

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// evalFieldIdentifier evaluates a field access expression (base.member or @.member).
// The base can be any expression: identifier, existing value (@), or another field access (chaining).
func evalFieldIdentifier(node *ast.FieldIdentifier, bindings *Bindings) *EvalResult {
	var record *object.Record

	// Get the effective base expression
	base := node.GetBase()

	if base != nil {
		// Evaluate the base expression
		result := Eval(base, bindings)
		if result.IsError() {
			return result
		}

		var ok bool
		record, ok = result.Value.(*object.Record)
		if !ok {
			return NewEvalError("field access requires Record, got %s", result.Value.Type())
		}
	} else {
		// nil base means use the existing value from EXCEPT context (@)
		existingValue := bindings.GetExistingValue()
		if existingValue == nil {
			return NewEvalError("@.member used outside of EXCEPT context")
		}

		var ok bool
		record, ok = existingValue.(*object.Record)
		if !ok {
			return NewEvalError("@.member requires Record, got %s", existingValue.Type())
		}
	}

	// First, check if this is a relation field
	classKey := bindings.SelfClassKey()
	relCtx := bindings.RelationContext()

	if classKey != "" && relCtx != nil {
		if relInfo := lookupRelation(classKey, node.Member, relCtx); relInfo != nil {
			return evalRelationTraversal(record, relInfo, relCtx)
		}
	}

	// Otherwise, do regular field access
	value := record.Get(node.Member)
	if value == nil {
		return NewEvalError("field not found: %s", node.Member)
	}

	return NewEvalResult(value)
}

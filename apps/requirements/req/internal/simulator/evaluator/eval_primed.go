package evaluator

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// evalPrimed evaluates a primed expression (x').
// A primed expression refers to the "next state" value of a variable.
// If the variable has been primed (assigned via x' = ...), return that value.
// Otherwise, return the current value (for reading in guards/conditions).
//
// For field access like record.field':
// - First get the root variable name (record)
// - Look up its primed value if available
// - Then evaluate the field access chain on that value
func evalPrimed(node *ast.Primed, bindings *Bindings) *EvalResult {
	switch base := node.Base.(type) {
	case *ast.Identifier:
		// Simple identifier: x'
		name := base.Value

		// Check if this variable has been primed - return the primed value
		if val, found := bindings.GetPrimedValue(name); found {
			return NewEvalResult(val)
		}

		// Otherwise, look up the current value (for reading before priming)
		val, found := bindings.GetValue(name)
		if !found {
			return NewEvalError("identifier not found: %s", name)
		}
		return NewEvalResult(val)

	case *ast.FieldAccess:
		// Field access: record.field' or record.a.b.c'
		// We need to find the root identifier, get its primed value,
		// then apply the field access chain to it.
		return evalPrimedFieldAccess(base, bindings)

	default:
		return NewEvalError("primed expression requires an identifier or field access, got %T", node.Base)
	}
}

// evalPrimedFieldAccess evaluates a field access chain with priming.
// For record.field', it gets record' and then accesses .field on it.
// For record.a.b', it gets record' and then accesses .a.b on it.
func evalPrimedFieldAccess(fa *ast.FieldAccess, bindings *Bindings) *EvalResult {
	// Collect the field access chain and find the root
	var fields []string
	var rootExpr ast.Expression

	current := fa
	for {
		fields = append([]string{current.Member}, fields...) // prepend to maintain order
		base := current.GetBase()

		if base == nil {
			// Base is nil, meaning @ (existing value) - priming doesn't apply to @
			// Just evaluate normally
			return evalFieldIdentifier(fa, bindings)
		}

		switch b := base.(type) {
		case *ast.Identifier:
			rootExpr = b
			goto done
		case *ast.FieldAccess:
			current = b
		case *ast.ExistingValue:
			// @ - just evaluate normally, priming doesn't apply
			return evalFieldIdentifier(fa, bindings)
		default:
			// For other expression types, evaluate the whole thing normally
			return evalFieldIdentifier(fa, bindings)
		}
	}

done:
	// rootExpr is the root identifier
	rootIdent := rootExpr.(*ast.Identifier)
	rootName := rootIdent.Value

	// Get the primed value if available, otherwise fall back to current value
	var rootValue object.Object
	if val, found := bindings.GetPrimedValue(rootName); found {
		rootValue = val
	} else if rootName == "self" {
		// Special case: "self" is stored separately in bindings, not in the regular store
		self := bindings.Self()
		if self == nil {
			return NewEvalError("self is not defined in this scope")
		}
		rootValue = self
	} else {
		val, found := bindings.GetValue(rootName)
		if !found {
			return NewEvalError("identifier not found: %s", rootName)
		}
		rootValue = val
	}

	// Now apply the field access chain
	currentValue := rootValue
	for _, field := range fields {
		record, ok := currentValue.(*object.Record)
		if !ok {
			return NewEvalError("field access requires Record, got %s", currentValue.Type())
		}

		fieldValue := record.Get(field)
		if fieldValue == nil {
			return NewEvalError("field not found: %s", field)
		}
		currentValue = fieldValue
	}

	return NewEvalResult(currentValue)
}

package evaluator

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// evalLogicEquality evaluates a generic equality comparison (=, ≠).
// This handles equality for all types: numbers, strings, booleans, sets, tuples, records, etc.
func evalLogicEquality(node *ast.LogicEquality, bindings *Bindings) *EvalResult {
	leftResult := Eval(node.Left, bindings)
	if leftResult.IsError() {
		return leftResult
	}

	rightResult := Eval(node.Right, bindings)
	if rightResult.IsError() {
		return rightResult
	}

	// Use the object's Equals method for type-appropriate comparison
	equals := objectsEqual(leftResult.Value, rightResult.Value)

	var result bool
	switch node.Operator {
	case "=":
		result = equals
	case "≠", "/=", "#":
		result = !equals
	default:
		return NewEvalError("unknown equality operator: %s", node.Operator)
	}

	return NewEvalResult(nativeBoolToBoolean(result))
}

// objectsEqual compares two objects for equality.
// It handles all object types appropriately.
func objectsEqual(left, right object.Object) bool {
	// Check for type mismatch - different types are never equal
	// Exception: Numbers are compared by value regardless of kind
	if left.Type() != right.Type() {
		return false
	}

	switch l := left.(type) {
	case *object.Number:
		r, ok := right.(*object.Number)
		if !ok {
			return false
		}
		return l.Equals(r)

	case *object.String:
		r, ok := right.(*object.String)
		if !ok {
			return false
		}
		return l.Value() == r.Value()

	case *object.Boolean:
		r, ok := right.(*object.Boolean)
		if !ok {
			return false
		}
		return l.Value() == r.Value()

	case *object.Set:
		r, ok := right.(*object.Set)
		if !ok {
			return false
		}
		return l.Equals(r)

	case *object.Tuple:
		r, ok := right.(*object.Tuple)
		if !ok {
			return false
		}
		return l.Equals(r)

	case *object.Record:
		r, ok := right.(*object.Record)
		if !ok {
			return false
		}
		return l.Equals(r)

	case *object.Bag:
		r, ok := right.(*object.Bag)
		if !ok {
			return false
		}
		return l.Equals(r)

	default:
		// For unknown types, use Inspect() as fallback
		return left.Inspect() == right.Inspect()
	}
}

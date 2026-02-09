package evaluator

import (
	"github.com/glemzurg/go-tlaplus/internal/simulator/ast"
	"github.com/glemzurg/go-tlaplus/internal/simulator/object"
)

// evalSetConstant evaluates a built-in set constant.
// Note: Infinite sets (Nat, Int, Real) cannot be fully enumerated,
// so we return a special representation or error for those.
func evalSetConstant(node *ast.SetConstant) *EvalResult {
	switch node.Value {
	case ast.SetConstantBoolean:
		// BOOLEAN = {TRUE, FALSE}
		elements := []object.Object{
			object.NewBoolean(true),
			object.NewBoolean(false),
		}
		return NewEvalResult(object.NewSetFromElements(elements))

	case ast.SetConstantNat:
		// Nat is infinite - cannot enumerate
		return NewEvalError("cannot enumerate infinite set: Nat")

	case ast.SetConstantInt:
		// Int is infinite - cannot enumerate
		return NewEvalError("cannot enumerate infinite set: Int")

	case ast.SetConstantReal:
		// Real is infinite - cannot enumerate
		return NewEvalError("cannot enumerate infinite set: Real")

	default:
		return NewEvalError("unknown set constant: %s", node.Value)
	}
}

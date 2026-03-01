package evaluator

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// evalLogicBoundQuantifier evaluates a bound quantifier (∀, ∃).
func evalLogicBoundQuantifier(node *ast.LogicBoundQuantifier, bindings *Bindings) *EvalResult {
	// Get the membership expression to extract the bound variable and set
	membership, ok := node.Membership.(*ast.LogicMembership)
	if !ok {
		return NewEvalError("quantifier requires membership expression")
	}

	// Get the bound variable name
	varIdent, ok := membership.Left.(*ast.Identifier)
	if !ok {
		return NewEvalError("quantifier variable must be Identifier")
	}
	varName := varIdent.Value

	// Evaluate the set
	setResult := Eval(membership.Right, bindings)
	if setResult.IsError() {
		return setResult
	}

	set, ok := setResult.Value.(*object.Set)
	if !ok {
		return NewEvalError("quantifier requires Set, got %s", setResult.Value.Type())
	}

	elements := set.Elements()

	switch node.Quantifier {
	case "∀", "\\A":
		// For all: return true if predicate is true for ALL elements
		for _, elem := range elements {
			childBindings := NewEnclosedBindings(bindings)
			childBindings.Set(varName, elem, NamespaceLocal)

			predResult := Eval(node.Predicate, childBindings)
			if predResult.IsError() {
				return predResult
			}

			predBool, ok := predResult.Value.(*object.Boolean)
			if !ok {
				return NewEvalError("predicate must return Boolean, got %s", predResult.Value.Type())
			}

			if !predBool.Value() {
				return NewEvalResult(FALSE)
			}
		}
		return NewEvalResult(TRUE)

	case "∃", "\\E":
		// Exists: return true if predicate is true for ANY element
		for _, elem := range elements {
			childBindings := NewEnclosedBindings(bindings)
			childBindings.Set(varName, elem, NamespaceLocal)

			predResult := Eval(node.Predicate, childBindings)
			if predResult.IsError() {
				return predResult
			}

			predBool, ok := predResult.Value.(*object.Boolean)
			if !ok {
				return NewEvalError("predicate must return Boolean, got %s", predResult.Value.Type())
			}

			if predBool.Value() {
				return NewEvalResult(TRUE)
			}
		}
		return NewEvalResult(FALSE)

	default:
		return NewEvalError("unknown quantifier: %s", node.Quantifier)
	}
}

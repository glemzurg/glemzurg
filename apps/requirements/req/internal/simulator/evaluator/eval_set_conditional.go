package evaluator

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// evalSetConditional evaluates a set comprehension {x ∈ S : predicate}.
func evalSetConditional(node *ast.SetFilter, bindings *Bindings) *EvalResult {
	// Get the membership expression
	membership, ok := node.Membership.(*ast.Membership)
	if !ok {
		return NewEvalError("set conditional requires membership expression")
	}

	// Get the bound variable name
	varIdent, ok := membership.Left.(*ast.Identifier)
	if !ok {
		return NewEvalError("set conditional variable must be Identifier")
	}
	varName := varIdent.Value

	// Evaluate the source set
	setResult := EvalAST(membership.Right, bindings)
	if setResult.IsError() {
		return setResult
	}

	sourceSet, ok := setResult.Value.(*object.Set)
	if !ok {
		return NewEvalError("set conditional requires Set, got %s", setResult.Value.Type())
	}

	// Filter elements based on predicate
	resultElements := make([]object.Object, 0)
	for _, elem := range sourceSet.Elements() {
		childBindings := NewEnclosedBindings(bindings)
		childBindings.Set(varName, elem, NamespaceLocal)

		predResult := EvalAST(node.Predicate, childBindings)
		if predResult.IsError() {
			return predResult
		}

		predBool, ok := predResult.Value.(*object.Boolean)
		if !ok {
			return NewEvalError("predicate must return Boolean, got %s", predResult.Value.Type())
		}

		if predBool.Value() {
			resultElements = append(resultElements, elem)
		}
	}

	return NewEvalResult(object.NewSetFromElements(resultElements))
}

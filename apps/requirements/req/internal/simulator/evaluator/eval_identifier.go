package evaluator

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/ast"
)

// evalIdentifier evaluates a variable lookup.
func evalIdentifier(node *ast.Identifier, bindings *Bindings) *EvalResult {
	// Special case: "self" refers to the current model_class record
	if node.Value == "self" {
		self := bindings.Self()
		if self == nil {
			return NewEvalError("self is not defined in this scope")
		}
		return NewEvalResult(self)
	}

	// Look up the identifier in bindings
	value, found := bindings.GetValue(node.Value)
	if !found {
		return NewEvalError("identifier not found: %s", node.Value)
	}

	return NewEvalResult(value)
}

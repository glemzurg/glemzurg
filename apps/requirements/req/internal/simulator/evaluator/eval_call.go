package evaluator

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/ast"
)

// evalCallExpression evaluates a function call.
// TODO: This needs integration with the custom TLA+ block registry.
func evalCallExpression(node *ast.ScopedCall) *EvalResult {
	// Build the full function name
	var fullName string
	switch {
	case node.Domain != nil:
		fullName = node.Domain.Value + "!" + node.Subdomain.Value + "!" + node.Class.Value + "!" + node.FunctionName.Value
	case node.Subdomain != nil:
		fullName = node.Subdomain.Value + "!" + node.Class.Value + "!" + node.FunctionName.Value
	case node.Class != nil:
		fullName = node.Class.Value + "!" + node.FunctionName.Value
	default:
		fullName = node.FunctionName.Value
	}

	if node.ModelScope {
		fullName = "_" + fullName
	}

	// TODO: Look up the function in the builtin registry and execute
	return NewEvalError("function calls not yet implemented: %s", fullName)
}

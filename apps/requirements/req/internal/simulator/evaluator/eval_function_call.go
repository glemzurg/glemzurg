package evaluator

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// evalFunctionCall evaluates a function call with optional scope path.
// Supports patterns:
//   - _Module!FunctionName(args...) - built-in module call (e.g., _Seq!Len)
//   - _FunctionName(args...) - global function call
//   - ActionName(args...) - class-scoped action (current class)
//   - Class!ActionName(args...) - class-scoped action (from subdomain)
//   - Subdomain!Class!ActionName(args...) - class-scoped action (from domain)
//   - Domain!Subdomain!Class!ActionName(args...) - fully scoped class action
func evalFunctionCall(node *ast.FunctionCall, bindings *Bindings) *EvalResult {
	// Build the full function name from scope path
	funcName := node.FullName()

	// Evaluate all arguments
	args := make([]object.Object, len(node.Args))
	for i, argExpr := range node.Args {
		result := Eval(argExpr, bindings)
		if result.IsError() {
			return result
		}
		args[i] = result.Value
	}

	// For global/built-in functions (leading underscore), look up in builtins
	if node.IsGlobalOrBuiltin() {
		fn, ok := LookupBuiltin(funcName)
		if !ok {
			return NewEvalError("unknown function: %s", funcName)
		}
		return fn(args)
	}

	// For class-scoped actions, we need to look up in the registry
	// This will be implemented in Phase 2 (Model-to-Simulator Bridge)
	return NewEvalError("unknown function: %s (class-scoped actions not yet implemented)", funcName)
}

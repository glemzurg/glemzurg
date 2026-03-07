package evaluator

import (
	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_expression"
)

// IRRegistryInterface allows the IR evaluator to look up custom function definitions.
// This interface is implemented by the registry's RuntimeAdapter.
type IRRegistryInterface interface {
	// LookupGlobal looks up a global function by its local name (without underscore).
	// Returns the IR body and parameter names, or nil if not found.
	LookupGlobal(localName string) (body me.Expression, params []string, found bool)
}

// EvalContext holds context for registry-based evaluation.
type EvalContext struct {
	IRRegistry IRRegistryInterface // IR-based registry interface
	ScopeLevel int
	Domain     string
	Subdomain  string
	Class      string
}

// Global eval context (set by pipeline before evaluation).
var globalEvalContext *EvalContext

// SetEvalContext sets the global eval context for registry-based evaluation.
func SetEvalContext(ctx *EvalContext) {
	globalEvalContext = ctx
}

// ClearEvalContext clears the global eval context.
func ClearEvalContext() {
	globalEvalContext = nil
}

// GetEvalContext returns the current eval context, or nil if not set.
func GetEvalContext() *EvalContext {
	return globalEvalContext
}

// EvalWithContext evaluates an IR expression with registry context.
func EvalWithContext(expr me.Expression, bindings *Bindings, ctx *EvalContext) *EvalResult {
	oldCtx := globalEvalContext
	globalEvalContext = ctx
	defer func() { globalEvalContext = oldCtx }()

	return Eval(expr, bindings)
}

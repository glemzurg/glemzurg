package evaluator

import (
	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/typechecker"
)

// IRRegistryInterface allows the IR evaluator to look up custom function definitions.
// This interface is implemented by the registry's RuntimeAdapter.
type IRRegistryInterface interface {
	// LookupGlobal looks up a global function by its local name (without underscore).
	// Returns the IR body and parameter names, or nil if not found.
	LookupGlobal(localName string) (body me.Expression, params []string, found bool)
}

// RegistryEvalInterface allows the evaluator to call back to the registry for function evaluation.
// This interface is implemented by the registry package.
// LEGACY: Used by the AST evaluation path. New code should use IRRegistryInterface.
type RegistryEvalInterface interface {
	ResolveAndEval(
		call *ast.ScopedCall,
		typedArgs []*typechecker.TypedNode,
		bindings any,
		scopeLevel int,
		domain, subdomain, class string,
	) (object.Object, error)
}

// EvalContext holds context for registry-based evaluation.
type EvalContext struct {
	Registry   RegistryEvalInterface // LEGACY: AST-based registry interface
	IRRegistry IRRegistryInterface   // IR-based registry interface
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

// EvalTypedWithContext evaluates a typed node with registry context.
func EvalTypedWithContext(typed *typechecker.TypedNode, bindings *Bindings, ctx *EvalContext) *EvalResult {
	// Save and restore context
	oldCtx := globalEvalContext
	globalEvalContext = ctx
	defer func() { globalEvalContext = oldCtx }()

	return evalTypedNode(typed, bindings)
}

// EvalWithContext evaluates an IR expression with registry context.
func EvalWithContext(expr me.Expression, bindings *Bindings, ctx *EvalContext) *EvalResult {
	oldCtx := globalEvalContext
	globalEvalContext = ctx
	defer func() { globalEvalContext = oldCtx }()

	return Eval(expr, bindings)
}

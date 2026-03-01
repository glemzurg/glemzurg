package evaluator

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/typechecker"
)

// RegistryEvalInterface allows the evaluator to call back to the registry for function evaluation.
// This interface is implemented by the registry package.
// Note: bindings is interface{} to avoid import cycles - implementations should type assert to *Bindings.
type RegistryEvalInterface interface {
	// ResolveAndEval resolves a call expression and evaluates it.
	// Returns the result or an error.
	// The bindings parameter should be *Bindings but is interface{} to avoid import cycles.
	ResolveAndEval(
		call *ast.CallExpression,
		typedArgs []*typechecker.TypedNode,
		bindings interface{},
		scopeLevel int,
		domain, subdomain, class string,
	) (object.Object, error)
}

// EvalContext holds context for registry-based evaluation.
type EvalContext struct {
	Registry   RegistryEvalInterface
	ScopeLevel int
	Domain     string
	Subdomain  string
	Class      string
}

// Global eval context (set by pipeline before evaluation)
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

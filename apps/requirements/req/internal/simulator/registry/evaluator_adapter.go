package registry

import (
	"fmt"

	"github.com/glemzurg/go-tlaplus/internal/simulator/ast"
	"github.com/glemzurg/go-tlaplus/internal/simulator/object"
	"github.com/glemzurg/go-tlaplus/internal/simulator/typechecker"
)

// EvalTypedFunc is the function signature for evaluating a typed node.
// This matches the evaluator.EvalTyped function signature.
type EvalTypedFunc func(typed *typechecker.TypedNode, bindings BindingsAdapter) *EvalResultAdapter

// BindingsAdapter wraps evaluator.Bindings for use by the registry.
type BindingsAdapter interface {
	// GetValue retrieves a value by name
	GetValue(name string) (object.Object, bool)
	// Set creates or updates a binding in local namespace
	SetLocal(name string, value object.Object)
	// CreateChild creates a new child scope
	CreateChild() BindingsAdapter
	// SetSelf sets the self record for class methods
	SetSelf(self *object.Record)
	// GetSelf returns the self record
	GetSelf() *object.Record
}

// EvalResultAdapter wraps evaluator.EvalResult for use by the registry.
type EvalResultAdapter interface {
	// Value returns the result value
	Value() object.Object
	// IsError returns true if this is an error result
	IsError() bool
	// ErrorMessage returns the error message if IsError is true
	ErrorMessage() string
}

// RuntimeAdapter implements evaluator.RegistryEvalInterface.
// It bridges the evaluator to the registry for function calls.
type RuntimeAdapter struct {
	registry *Registry
	evalFn   func(typed *typechecker.TypedNode, bindings interface{}) *EvalResultAdapter
}

// NewRuntimeAdapter creates a new runtime adapter for registry-based evaluation.
// The evalFn should be a wrapper around evaluator.evalTypedNode that handles
// the scope context switching.
func NewRuntimeAdapter(r *Registry) *RuntimeAdapter {
	return &RuntimeAdapter{
		registry: r,
	}
}

// ResolveAndEval implements evaluator.RegistryEvalInterface.
// It resolves a call expression and evaluates it using the registry.
func (a *RuntimeAdapter) ResolveAndEval(
	call *ast.CallExpression,
	typedArgs []*typechecker.TypedNode,
	bindings interface{},
	scopeLevel int,
	domain, subdomain, class string,
) (object.Object, error) {
	// Create scope context for resolution
	var scopeCtx *ScopeContext

	switch ScopeLevel(scopeLevel) {
	case ScopeLevelGlobal:
		scopeCtx = NewGlobalScopeContext(a.registry)
	case ScopeLevelDomain:
		scopeCtx = NewDomainScopeContext(a.registry, domain)
	case ScopeLevelSubdomain:
		scopeCtx = NewSubdomainScopeContext(a.registry, domain, subdomain)
	case ScopeLevelClass:
		scopeCtx = NewClassScopeContext(a.registry, domain, subdomain, class)
	default:
		return nil, fmt.Errorf("invalid scope level: %d", scopeLevel)
	}

	// Resolve the function
	key, def, err := scopeCtx.ResolveCall(call)
	if err != nil {
		return nil, err
	}

	// Ensure definition is type-checked
	if def.TypedBody == nil {
		return nil, &EvalError{Key: key, Message: "function is not type-checked"}
	}

	// The actual evaluation will be done by the evaluator through
	// a callback mechanism set up in the pipeline.
	// For now, we return the definition info so the pipeline can handle it.

	// This is a placeholder - the actual implementation requires
	// the pipeline to set up the callback properly.
	return nil, &EvalError{
		Key:     key,
		Message: "registry evaluation requires pipeline integration",
	}
}

// GetDefinitionForCall resolves a call and returns the definition.
// This is used by the pipeline to get the definition before evaluation.
func (a *RuntimeAdapter) GetDefinitionForCall(
	call *ast.CallExpression,
	scopeLevel int,
	domain, subdomain, class string,
) (DefinitionKey, *Definition, error) {
	var scopeCtx *ScopeContext

	switch ScopeLevel(scopeLevel) {
	case ScopeLevelGlobal:
		scopeCtx = NewGlobalScopeContext(a.registry)
	case ScopeLevelDomain:
		scopeCtx = NewDomainScopeContext(a.registry, domain)
	case ScopeLevelSubdomain:
		scopeCtx = NewSubdomainScopeContext(a.registry, domain, subdomain)
	case ScopeLevelClass:
		scopeCtx = NewClassScopeContext(a.registry, domain, subdomain, class)
	default:
		return "", nil, fmt.Errorf("invalid scope level: %d", scopeLevel)
	}

	return scopeCtx.ResolveCall(call)
}

// EvalError represents an evaluation error for a specific definition.
type EvalError struct {
	Key     DefinitionKey
	Message string
}

func (e *EvalError) Error() string {
	return string(e.Key) + ": " + e.Message
}

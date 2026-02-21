package registry

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/types"
)

// TypeCheckerAdapter implements typechecker.RegistryInterface.
// It bridges the registry to the type checker for function resolution.
type TypeCheckerAdapter struct {
	registry *Registry
}

// NewTypeCheckerAdapter creates a new adapter for the given registry.
func NewTypeCheckerAdapter(r *Registry) *TypeCheckerAdapter {
	return &TypeCheckerAdapter{registry: r}
}

// ResolveCallExpression resolves a call expression using the registry.
// It implements typechecker.RegistryInterface.
func (a *TypeCheckerAdapter) ResolveCallExpression(
	call *ast.CallExpression,
	scopeLevel int,
	domain, subdomain, class string,
) (key string, paramTypes []types.Type, returnType types.Type, err error) {
	// Create a scope context for resolution
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
		return "", nil, nil, fmt.Errorf("invalid scope level: %d", scopeLevel)
	}

	// Resolve the call
	defKey, def, err := scopeCtx.ResolveCall(call)
	if err != nil {
		return "", nil, nil, err
	}

	// Extract parameter types
	paramTypes = make([]types.Type, len(def.Parameters))
	for i, param := range def.Parameters {
		paramTypes[i] = param.Type
	}

	// Return type may be nil if not yet type-checked
	returnType = def.ReturnType
	if returnType == nil {
		// Use a fresh type variable as placeholder
		returnType = types.Any{}
	}

	return string(defKey), paramTypes, returnType, nil
}

// DependencyRecorder implements typechecker.DependencyTracker.
// It records dependencies to the registry during type checking.
type DependencyRecorder struct {
	registry *Registry
}

// NewDependencyRecorder creates a new dependency recorder.
func NewDependencyRecorder(r *Registry) *DependencyRecorder {
	return &DependencyRecorder{registry: r}
}

// RecordDependency records that fromKey depends on toKey.
func (d *DependencyRecorder) RecordDependency(fromKey, toKey string) {
	d.registry.AddDependency(DefinitionKey(fromKey), DefinitionKey(toKey))
}

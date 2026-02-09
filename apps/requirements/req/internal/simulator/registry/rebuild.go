package registry

import (
	"fmt"

	"github.com/glemzurg/go-tlaplus/internal/simulator/typechecker"
	"github.com/glemzurg/go-tlaplus/internal/simulator/types"
)

// RebuildStrategy determines how to handle type-checking after changes.
type RebuildStrategy int

const (
	// IncrementalRebuild only re-type-checks changed definitions and dependents.
	IncrementalRebuild RebuildStrategy = iota

	// FullRebuild type-checks all definitions from scratch.
	FullRebuild
)

func (s RebuildStrategy) String() string {
	switch s {
	case IncrementalRebuild:
		return "incremental"
	case FullRebuild:
		return "full"
	default:
		return "unknown"
	}
}

// TypeCheckFunc is the function signature for type-checking a definition.
// It receives the definition and a type checker, and should:
// 1. Type-check the definition body
// 2. Set def.TypedBody and def.ReturnType
// 3. Record dependencies via registry.AddDependency
// 4. Return any type error encountered
type TypeCheckFunc func(def *Definition, tc *typechecker.TypeChecker, scopeCtx *ScopeContext) error

// RebuildError contains all type errors encountered during rebuild.
type RebuildError struct {
	Errors []DefinitionError
}

func (e *RebuildError) Error() string {
	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}
	return fmt.Sprintf("%d type errors during rebuild", len(e.Errors))
}

// DefinitionError is a type error in a specific definition.
type DefinitionError struct {
	Key     DefinitionKey
	Message string
}

func (e DefinitionError) Error() string {
	return fmt.Sprintf("%s: %s", e.Key, e.Message)
}

// Rebuild performs type-checking based on the strategy.
// Uses fail-fast approach: collects all errors before returning.
func (r *Registry) Rebuild(
	strategy RebuildStrategy,
	invalidated *InvalidationSet,
	tc *typechecker.TypeChecker,
	typeCheckFn TypeCheckFunc,
) *RebuildError {
	switch strategy {
	case IncrementalRebuild:
		return r.incrementalRebuild(invalidated, tc, typeCheckFn)
	case FullRebuild:
		return r.fullRebuild(tc, typeCheckFn)
	default:
		return &RebuildError{
			Errors: []DefinitionError{{Key: "", Message: "unknown rebuild strategy"}},
		}
	}
}

func (r *Registry) incrementalRebuild(
	invalidated *InvalidationSet,
	tc *typechecker.TypeChecker,
	typeCheckFn TypeCheckFunc,
) *RebuildError {
	if invalidated == nil || len(invalidated.Keys) == 0 {
		return nil
	}

	// Sort by dependency order (definitions with no deps first)
	sorted := r.topologicalSort(invalidated.Keys)

	var errors []DefinitionError
	for _, key := range sorted {
		def, ok := r.Get(key)
		if !ok {
			continue
		}

		if !def.NeedsTypeCheck() {
			continue
		}

		// Clear old dependencies before re-type-checking
		r.ClearDependencies(key)

		// Create scope context for this definition
		scopeCtx := r.createScopeContextForDef(def)

		// Type-check the definition
		if err := typeCheckFn(def, tc, scopeCtx); err != nil {
			errors = append(errors, DefinitionError{Key: key, Message: err.Error()})
		}
	}

	if len(errors) > 0 {
		return &RebuildError{Errors: errors}
	}
	return nil
}

func (r *Registry) fullRebuild(
	tc *typechecker.TypeChecker,
	typeCheckFn TypeCheckFunc,
) *RebuildError {
	// Clear all typed bodies and dependencies
	r.mu.Lock()
	var allKeys []DefinitionKey
	for key, def := range r.definitions {
		def.TypedBody = nil
		def.ReturnType = nil
		def.DependsOn = nil
		def.DependedBy = nil
		allKeys = append(allKeys, key)
	}
	r.mu.Unlock()

	// Sort topologically and type-check
	sorted := r.topologicalSort(allKeys)

	var errors []DefinitionError
	for _, key := range sorted {
		def, ok := r.Get(key)
		if !ok {
			continue
		}

		// Create scope context for this definition
		scopeCtx := r.createScopeContextForDef(def)

		// Type-check the definition
		if err := typeCheckFn(def, tc, scopeCtx); err != nil {
			errors = append(errors, DefinitionError{Key: key, Message: err.Error()})
		}
	}

	if len(errors) > 0 {
		return &RebuildError{Errors: errors}
	}
	return nil
}

// createScopeContextForDef creates the appropriate scope context for type-checking a definition.
func (r *Registry) createScopeContextForDef(def *Definition) *ScopeContext {
	if def.Kind == KindGlobalFunction {
		return NewGlobalScopeContext(r)
	}

	// For class functions, create a class-level scope context
	parts := def.Scope.Parts()
	if len(parts) != 3 {
		return NewGlobalScopeContext(r)
	}

	return NewClassScopeContext(r, parts[0], parts[1], parts[2])
}

// topologicalSort sorts definition keys so that dependencies come before dependents.
// This ensures that when type-checking, all dependencies are already type-checked.
func (r *Registry) topologicalSort(keys []DefinitionKey) []DefinitionKey {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Build set of keys we're sorting
	keySet := make(map[DefinitionKey]struct{})
	for _, k := range keys {
		keySet[k] = struct{}{}
	}

	// Track visited and in-progress for cycle detection
	visited := make(map[DefinitionKey]struct{})
	inProgress := make(map[DefinitionKey]struct{})
	var result []DefinitionKey

	var visit func(key DefinitionKey) bool
	visit = func(key DefinitionKey) bool {
		if _, done := visited[key]; done {
			return true
		}
		if _, cycling := inProgress[key]; cycling {
			// Cycle detected - just continue (cyclic deps will error during type-check)
			return true
		}

		inProgress[key] = struct{}{}

		def, ok := r.definitions[key]
		if ok {
			// Visit dependencies first (but only if they're in our key set)
			for _, depKey := range def.DependsOn {
				if _, inSet := keySet[depKey]; inSet {
					visit(depKey)
				}
			}
		}

		delete(inProgress, key)
		visited[key] = struct{}{}
		result = append(result, key)
		return true
	}

	for _, key := range keys {
		visit(key)
	}

	return result
}

// SetTypedBody updates a definition's typed body and return type after successful type-checking.
func (r *Registry) SetTypedBody(key DefinitionKey, typed *typechecker.TypedNode, returnType types.Type) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	def, ok := r.definitions[key]
	if !ok {
		return fmt.Errorf("definition not found: %s", key)
	}

	def.TypedBody = typed
	def.ReturnType = returnType
	return nil
}

// GetUntypedDefinitions returns all definitions that need type-checking.
func (r *Registry) GetUntypedDefinitions() []DefinitionKey {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []DefinitionKey
	for key, def := range r.definitions {
		if def.NeedsTypeCheck() {
			result = append(result, key)
		}
	}
	return result
}

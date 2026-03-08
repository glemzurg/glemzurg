package registry

import (
	"fmt"
)

// RebuildStrategy determines how to handle validation after changes.
type RebuildStrategy int

const (
	// IncrementalRebuild only re-validates changed definitions and dependents.
	IncrementalRebuild RebuildStrategy = iota

	// FullRebuild validates all definitions from scratch.
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

// ValidateFunc is the function signature for validating a definition.
type ValidateFunc func(def *Definition, scopeCtx *ScopeContext) error

// RebuildError contains all errors encountered during rebuild.
type RebuildError struct {
	Errors []DefinitionError
}

func (e *RebuildError) Error() string {
	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}
	return fmt.Sprintf("%d errors during rebuild", len(e.Errors))
}

// DefinitionError is an error in a specific definition.
type DefinitionError struct {
	Key     DefinitionKey
	Message string
}

func (e DefinitionError) Error() string {
	return fmt.Sprintf("%s: %s", e.Key, e.Message)
}

// Rebuild performs validation based on the strategy.
func (r *Registry) Rebuild(
	strategy RebuildStrategy,
	invalidated *InvalidationSet,
	validateFn ValidateFunc,
) *RebuildError {
	switch strategy {
	case IncrementalRebuild:
		return r.incrementalRebuild(invalidated, validateFn)
	case FullRebuild:
		return r.fullRebuild(validateFn)
	default:
		return &RebuildError{
			Errors: []DefinitionError{{Key: "", Message: "unknown rebuild strategy"}},
		}
	}
}

func (r *Registry) incrementalRebuild(
	invalidated *InvalidationSet,
	validateFn ValidateFunc,
) *RebuildError {
	if invalidated == nil || len(invalidated.Keys) == 0 {
		return nil
	}

	sorted := r.topologicalSort(invalidated.Keys)

	var errors []DefinitionError
	for _, key := range sorted {
		def, ok := r.Get(key)
		if !ok {
			continue
		}

		if def.Validated {
			continue
		}

		// Clear old dependencies before re-validating
		r.ClearDependencies(key)

		scopeCtx := r.createScopeContextForDef(def)

		if err := validateFn(def, scopeCtx); err != nil {
			errors = append(errors, DefinitionError{Key: key, Message: err.Error()})
		} else {
			def.Validated = true
		}
	}

	if len(errors) > 0 {
		return &RebuildError{Errors: errors}
	}
	return nil
}

func (r *Registry) fullRebuild(
	validateFn ValidateFunc,
) *RebuildError {
	r.mu.Lock()
	var allKeys []DefinitionKey
	for key, def := range r.definitions {
		def.Validated = false
		def.ReturnType = nil
		def.DependsOn = nil
		def.DependedBy = nil
		allKeys = append(allKeys, key)
	}
	r.mu.Unlock()

	sorted := r.topologicalSort(allKeys)

	var errors []DefinitionError
	for _, key := range sorted {
		def, ok := r.Get(key)
		if !ok {
			continue
		}

		scopeCtx := r.createScopeContextForDef(def)

		if err := validateFn(def, scopeCtx); err != nil {
			errors = append(errors, DefinitionError{Key: key, Message: err.Error()})
		} else {
			def.Validated = true
		}
	}

	if len(errors) > 0 {
		return &RebuildError{Errors: errors}
	}
	return nil
}

// createScopeContextForDef creates the appropriate scope context for a definition.
func (r *Registry) createScopeContextForDef(def *Definition) *ScopeContext {
	if def.Kind == KindGlobalFunction {
		return NewGlobalScopeContext(r)
	}

	parts := def.Scope.Parts()
	if len(parts) != 3 {
		return NewGlobalScopeContext(r)
	}

	return NewClassScopeContext(r, parts[0], parts[1], parts[2])
}

// topologicalSort sorts definition keys so that dependencies come before dependents.
func (r *Registry) topologicalSort(keys []DefinitionKey) []DefinitionKey {
	r.mu.RLock()
	defer r.mu.RUnlock()

	keySet := make(map[DefinitionKey]struct{})
	for _, k := range keys {
		keySet[k] = struct{}{}
	}

	visited := make(map[DefinitionKey]struct{})
	inProgress := make(map[DefinitionKey]struct{})
	var result []DefinitionKey

	var visit func(key DefinitionKey)
	visit = func(key DefinitionKey) {
		if _, done := visited[key]; done {
			return
		}
		if _, cycling := inProgress[key]; cycling {
			return
		}

		inProgress[key] = struct{}{}

		def, ok := r.definitions[key]
		if ok {
			for _, depKey := range def.DependsOn {
				if _, inSet := keySet[depKey]; inSet {
					visit(depKey)
				}
			}
		}

		delete(inProgress, key)
		visited[key] = struct{}{}
		result = append(result, key)
	}

	for _, key := range keys {
		visit(key)
	}

	return result
}

// GetUntypedDefinitions returns all definitions that need validation.
func (r *Registry) GetUntypedDefinitions() []DefinitionKey {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []DefinitionKey
	for key, def := range r.definitions {
		if !def.Validated {
			result = append(result, key)
		}
	}
	return result
}

// Package typechecker implements type checking for TLA+ AST.
//
// The type checker performs Hindley-Milner type inference with:
// - Type environment for tracking variable types
// - Unification algorithm for solving type constraints
// - Type schemes for let-polymorphism
//
// Usage:
//
//	env := typechecker.NewEnv()
//	typed, err := typechecker.Check(node, env)
//	if err != nil {
//	    // Type error
//	}
//	// typed is the AST with type annotations
package typechecker

import (
	"github.com/glemzurg/go-tlaplus/internal/simulator/types"
)

// TypeEnv maps identifiers to their type schemes.
// Environments are hierarchical: each environment has an optional parent.
type TypeEnv struct {
	bindings map[string]types.Scheme
	parent   *TypeEnv
}

// NewEnv creates a new empty type environment.
func NewEnv() *TypeEnv {
	return &TypeEnv{
		bindings: make(map[string]types.Scheme),
		parent:   nil,
	}
}

// Extend creates a child environment with the given parent.
// Lookups check the child first, then the parent chain.
func (env *TypeEnv) Extend() *TypeEnv {
	return &TypeEnv{
		bindings: make(map[string]types.Scheme),
		parent:   env,
	}
}

// Bind adds a binding for a name to a type scheme.
// If the name already exists in this environment, it is shadowed.
func (env *TypeEnv) Bind(name string, scheme types.Scheme) {
	env.bindings[name] = scheme
}

// BindMono adds a binding for a name to a monomorphic type.
func (env *TypeEnv) BindMono(name string, t types.Type) {
	env.bindings[name] = types.Monotype(t)
}

// Lookup finds a type scheme for a name.
// Returns the scheme and true if found, or zero value and false if not.
func (env *TypeEnv) Lookup(name string) (types.Scheme, bool) {
	if scheme, ok := env.bindings[name]; ok {
		return scheme, true
	}
	if env.parent != nil {
		return env.parent.Lookup(name)
	}
	return types.Scheme{}, false
}

// Contains checks if a name is bound in this environment or its parents.
func (env *TypeEnv) Contains(name string) bool {
	_, ok := env.Lookup(name)
	return ok
}

// FreeTypeVars returns all free type variables in the environment.
func (env *TypeEnv) FreeTypeVars() map[int]struct{} {
	result := make(map[int]struct{})

	for current := env; current != nil; current = current.parent {
		for _, scheme := range current.bindings {
			for id := range scheme.FreeTypeVars() {
				result[id] = struct{}{}
			}
		}
	}

	return result
}

// Apply applies a substitution to all types in the environment.
func (env *TypeEnv) Apply(subst types.Substitution) *TypeEnv {
	newEnv := &TypeEnv{
		bindings: make(map[string]types.Scheme, len(env.bindings)),
		parent:   nil,
	}

	// Apply to parent first (if any)
	if env.parent != nil {
		newEnv.parent = env.parent.Apply(subst)
	}

	// Apply to bindings in this level
	for name, scheme := range env.bindings {
		newType := subst.Apply(scheme.Type)
		newEnv.bindings[name] = types.Scheme{
			TypeVars: scheme.TypeVars,
			Type:     newType,
		}
	}

	return newEnv
}

// Clone creates a shallow copy of the environment.
func (env *TypeEnv) Clone() *TypeEnv {
	newBindings := make(map[string]types.Scheme, len(env.bindings))
	for k, v := range env.bindings {
		newBindings[k] = v
	}
	return &TypeEnv{
		bindings: newBindings,
		parent:   env.parent,
	}
}

// Names returns all bound names in this environment (not parents).
func (env *TypeEnv) Names() []string {
	names := make([]string, 0, len(env.bindings))
	for name := range env.bindings {
		names = append(names, name)
	}
	return names
}

// AllNames returns all bound names in this environment and all parents.
func (env *TypeEnv) AllNames() []string {
	seen := make(map[string]bool)
	var names []string

	for current := env; current != nil; current = current.parent {
		for name := range current.bindings {
			if !seen[name] {
				seen[name] = true
				names = append(names, name)
			}
		}
	}

	return names
}

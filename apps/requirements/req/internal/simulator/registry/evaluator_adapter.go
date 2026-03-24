package registry

import (
	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
)

// RuntimeAdapter implements evaluator.IRRegistryInterface.
// It bridges the evaluator to the registry for function calls.
type RuntimeAdapter struct {
	registry *Registry
}

// NewRuntimeAdapter creates a new runtime adapter for registry-based evaluation.
func NewRuntimeAdapter(r *Registry) *RuntimeAdapter {
	return &RuntimeAdapter{
		registry: r,
	}
}

// LookupGlobal implements evaluator.IRRegistryInterface.
// It looks up a global function by local name (without underscore prefix).
func (a *RuntimeAdapter) LookupGlobal(localName string) (me.Expression, []string, bool) {
	def, ok := a.registry.GetGlobal(localName)
	if !ok {
		return nil, nil, false
	}
	params := make([]string, len(def.Parameters))
	for i, p := range def.Parameters {
		params[i] = p.Name
	}
	return def.Body, params, true
}

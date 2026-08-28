package schema

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"

// RunScope is the set of classes included in a simulation run.
// It is intake-only: pass it to [New], then discard it — schema answers scope
// questions via lookup triples and scoped bulk APIs.
//
// A zero value (nil classKeys) means every class in the model is in scope.
// A non-nil map is the exact in-scope set (keys absent from the model are ignored).
type RunScope struct {
	classKeys map[identity.Key]struct{}
}

// RunScopeAll returns a scope that includes every class in the model.
func RunScopeAll() RunScope {
	return RunScope{}
}

// NewRunScope builds a scope from an explicit class key list.
// An empty list yields RunScopeAll (same as zero value) so callers that
// resolved "simulate everything" need not special-case.
func NewRunScope(classKeys []identity.Key) RunScope {
	if len(classKeys) == 0 {
		return RunScopeAll()
	}
	m := make(map[identity.Key]struct{}, len(classKeys))
	for _, k := range classKeys {
		m[k] = struct{}{}
	}
	return RunScope{classKeys: m}
}

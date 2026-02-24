package surface

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
)

// Diagnostic represents a non-fatal issue found during surface analysis.
type Diagnostic struct {
	Level    string        // "warning" or "info"
	Message  string        // Human-readable description.
	ClassKey *identity.Key // Related class, if applicable.
	AssocKey *identity.Key // Related association, if applicable.
}

// Diagnose analyzes a resolved surface and produces diagnostic messages
// about potential issues: broken creation chains, unreachable states,
// orphaned multiplicity constraints, etc.
func Diagnose(resolved *ResolvedSurface, model *req_model.Model) []Diagnostic {
	var diagnostics []Diagnostic

	allAssocs := model.GetClassAssociations()

	// 1. Broken creation chain: mandatory outbound association to excluded class.
	for classKey := range resolved.Classes {
		for _, assoc := range allAssocs {
			if assoc.FromClassKey == classKey && assoc.ToMultiplicity.LowerBound >= 1 {
				if _, inScope := resolved.Classes[assoc.ToClassKey]; !inScope {
					ak := assoc.Key
					ck := classKey
					diagnostics = append(diagnostics, Diagnostic{
						Level: "warning",
						Message: fmt.Sprintf("broken creation chain: %s has mandatory association '%s' to excluded class %s",
							classKey.String(), assoc.Name, assoc.ToClassKey.String()),
						ClassKey: &ck,
						AssocKey: &ak,
					})
				}
			}
		}
	}

	// 2. Isolated class: class in scope with no creation events.
	// Only warn if the class DID have creation events in the full model
	// (meaning the surface didn't cause the isolation).
	for classKey, class := range resolved.Classes {
		hasCreation := false
		for _, t := range class.Transitions {
			if t.FromStateKey == nil {
				hasCreation = true
				break
			}
		}
		if hasCreation {
			continue // Has creation events — not isolated.
		}
		// This class has no creation events. Check if it had them in the full model.
		fullClass, found := findClassInModel(classKey, model)
		if !found {
			continue // Can't find class in full model — skip.
		}
		fullHasCreation := false
		for _, t := range fullClass.Transitions {
			if t.FromStateKey == nil {
				fullHasCreation = true
				break
			}
		}
		if !fullHasCreation {
			continue // Also isolated in full model — not caused by surface.
		}
		ck := classKey
		diagnostics = append(diagnostics, Diagnostic{
			Level:    "warning",
			Message:  fmt.Sprintf("isolated class: %s has no creation events and cannot be instantiated", classKey.String()),
			ClassKey: &ck,
		})
	}

	// 3. Half-association: one endpoint in scope, other excluded.
	for _, assoc := range allAssocs {
		_, fromIn := resolved.Classes[assoc.FromClassKey]
		_, toIn := resolved.Classes[assoc.ToClassKey]
		if (fromIn && !toIn) || (!fromIn && toIn) {
			ak := assoc.Key
			diagnostics = append(diagnostics, Diagnostic{
				Level: "info",
				Message: fmt.Sprintf("half-association: association '%s' dropped because one endpoint is outside the surface",
					assoc.Name),
				AssocKey: &ak,
			})
		}
	}

	// 4. Excluded invariants (already reported as warnings in resolved.Warnings).
	// We don't duplicate those here.

	// 5. Realized domain included (already reported in resolver warnings).

	// 6. All events internal: class where ALL events have SentBy pointing to in-scope classes.
	//
	// TODO(CalledBy/SentBy): Event.SentBy was removed from the req_model/model_state.Event struct
	// because CalledBy/SentBy are simulator concerns, not part of the pure data model. When
	// re-enabling:
	//   1. Define a simulator-local SentBy []identity.Key on a wrapper struct or parallel map.
	//   2. Populate it from simulator config/annotations during surface resolution.
	//   3. Update this loop to read from the simulator-local location.

	// 7. SentBy/CalledBy referencing unknown class.
	//
	// TODO(CalledBy/SentBy): Event.SentBy, Action.CalledBy, and Query.CalledBy were removed from
	// the req_model structs because these are simulator concerns, not part of the pure data model.
	// When re-enabling:
	//   1. Define simulator-local SentBy/CalledBy fields on wrapper structs or a parallel map
	//      keyed by event/action/query identity.Key.
	//   2. Populate them from simulator config/annotations during surface resolution.
	//   3. Update these loops to read from the simulator-local location.

	return diagnostics
}

// findClassInModel looks up a class by key in the full model.
func findClassInModel(classKey identity.Key, model *req_model.Model) (*model_class.Class, bool) {
	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			if class, ok := subdomain.Classes[classKey]; ok {
				return &class, true
			}
		}
	}
	return nil, false
}

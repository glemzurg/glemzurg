package surface

import (
	"fmt"

	"github.com/glemzurg/go-tlaplus/internal/identity"
	"github.com/glemzurg/go-tlaplus/internal/req_model"
	"github.com/glemzurg/go-tlaplus/internal/req_model/model_class"
)

// Diagnostic represents a non-fatal issue found during surface analysis.
type Diagnostic struct {
	Level   string       // "warning" or "info"
	Message string       // Human-readable description.
	ClassKey *identity.Key // Related class, if applicable.
	AssocKey *identity.Key // Related association, if applicable.
}

// Diagnose analyzes a resolved surface and produces diagnostic messages
// about potential issues: broken creation chains, unreachable states,
// orphaned multiplicity constraints, etc.
func Diagnose(resolved *ResolvedSurface, model *req_model.Model) []Diagnostic {
	var diagnostics []Diagnostic

	// Build all class keys in model for reference checking.
	allClassKeys := make(map[identity.Key]bool)
	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			for classKey := range subdomain.Classes {
				allClassKeys[classKey] = true
			}
		}
	}

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
	for classKey, class := range resolved.Classes {
		if len(class.Events) == 0 {
			continue
		}
		allInternal := true
		for _, event := range class.Events {
			if len(event.SentBy) == 0 {
				allInternal = false
				break
			}
			// Check if at least one sender is in scope.
			senderInScope := false
			for _, senderKey := range event.SentBy {
				if _, inScope := resolved.Classes[senderKey]; inScope {
					senderInScope = true
					break
				}
			}
			if !senderInScope {
				allInternal = false
				break
			}
		}
		if allInternal {
			ck := classKey
			diagnostics = append(diagnostics, Diagnostic{
				Level: "warning",
				Message: fmt.Sprintf("all events internal: class %s has all events with in-scope senders — simulator will never fire them directly",
					class.Name),
				ClassKey: &ck,
			})
		}
	}

	// 7. SentBy/CalledBy referencing unknown class.
	for _, class := range resolved.Classes {
		for _, event := range class.Events {
			for _, senderKey := range event.SentBy {
				if !allClassKeys[senderKey] {
					diagnostics = append(diagnostics, Diagnostic{
						Level: "warning",
						Message: fmt.Sprintf("event '%s' SentBy references unknown class: %s",
							event.Name, senderKey.String()),
					})
				}
			}
		}
		for _, action := range class.Actions {
			for _, callerKey := range action.CalledBy {
				if !allClassKeys[callerKey] {
					diagnostics = append(diagnostics, Diagnostic{
						Level: "warning",
						Message: fmt.Sprintf("action '%s' CalledBy references unknown class: %s",
							action.Name, callerKey.String()),
					})
				}
			}
		}
		for _, query := range class.Queries {
			for _, callerKey := range query.CalledBy {
				if !allClassKeys[callerKey] {
					diagnostics = append(diagnostics, Diagnostic{
						Level: "warning",
						Message: fmt.Sprintf("query '%s' CalledBy references unknown class: %s",
							query.Name, callerKey.String()),
					})
				}
			}
		}
	}

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

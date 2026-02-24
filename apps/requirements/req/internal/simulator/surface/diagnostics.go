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

// CallerData holds simulator-local SentBy/CalledBy metadata.
// This data is a simulator concern, not part of the core req_model.
type CallerData struct {
	// EventSentBy maps event keys to class keys that send them.
	EventSentBy map[identity.Key][]identity.Key
	// ActionCalledBy maps action keys to class keys that call them.
	ActionCalledBy map[identity.Key][]identity.Key
	// QueryCalledBy maps query keys to class keys that call them.
	QueryCalledBy map[identity.Key][]identity.Key
}

// Diagnose analyzes a resolved surface and produces diagnostic messages
// about potential issues: broken creation chains, unreachable states,
// orphaned multiplicity constraints, etc.
//
// callerData is optional (nil means no caller data available).
func Diagnose(resolved *ResolvedSurface, model *req_model.Model, callerData *CallerData) []Diagnostic {
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

	// Diagnostics 6 and 7 require CallerData.
	if callerData != nil {
		diagnostics = append(diagnostics, diagnoseCallerData(resolved, model, callerData)...)
	}

	return diagnostics
}

// diagnoseCallerData produces diagnostics related to SentBy/CalledBy metadata.
func diagnoseCallerData(resolved *ResolvedSurface, model *req_model.Model, cd *CallerData) []Diagnostic {
	var diagnostics []Diagnostic

	// 6. All events internal: class where ALL events have SentBy pointing to in-scope classes.
	for classKey, class := range resolved.Classes {
		if len(class.Events) == 0 {
			continue
		}
		allInternal := true
		for _, event := range class.Events {
			senders := cd.EventSentBy[event.Key]
			if len(senders) == 0 {
				allInternal = false
				break
			}
			hasInScopeSender := false
			for _, senderKey := range senders {
				if _, inScope := resolved.Classes[senderKey]; inScope {
					hasInScopeSender = true
					break
				}
			}
			if !hasInScopeSender {
				allInternal = false
				break
			}
		}
		if allInternal {
			ck := classKey
			diagnostics = append(diagnostics, Diagnostic{
				Level:    "warning",
				Message:  fmt.Sprintf("all events internal: %s has no externally-fireable events", classKey.String()),
				ClassKey: &ck,
			})
		}
	}

	// 7. SentBy/CalledBy referencing unknown class.
	for eventKey, senders := range cd.EventSentBy {
		for _, senderKey := range senders {
			if _, found := findClassInModel(senderKey, model); !found {
				diagnostics = append(diagnostics, Diagnostic{
					Level:   "warning",
					Message: fmt.Sprintf("SentBy references unknown class: event %s references %s", eventKey.String(), senderKey.String()),
				})
			}
		}
	}
	for actionKey, callers := range cd.ActionCalledBy {
		for _, callerKey := range callers {
			if _, found := findClassInModel(callerKey, model); !found {
				diagnostics = append(diagnostics, Diagnostic{
					Level:   "warning",
					Message: fmt.Sprintf("CalledBy references unknown class: action %s references %s", actionKey.String(), callerKey.String()),
				})
			}
		}
	}
	for queryKey, callers := range cd.QueryCalledBy {
		for _, callerKey := range callers {
			if _, found := findClassInModel(callerKey, model); !found {
				diagnostics = append(diagnostics, Diagnostic{
					Level:   "warning",
					Message: fmt.Sprintf("CalledBy references unknown class: query %s references %s", queryKey.String(), callerKey.String()),
				})
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

package surface

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// Diagnostic represents a non-fatal issue found during surface analysis.
type Diagnostic struct {
	Level    string        // "warning" or "info"
	Message  string        // Human-readable description.
	ClassKey *identity.Key // Related class, if applicable.
	AssocKey *identity.Key // Related association, if applicable.
}

// CallerData holds simulator-local SentBy/CalledBy metadata.
// This data is a simulator concern, not part of the core core.
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
func Diagnose(resolved *ResolvedSurface, model *core.Model, callerData *CallerData) []Diagnostic {
	var diagnostics []Diagnostic

	allAssocs := model.GetClassAssociations()

	diagnostics = append(diagnostics, diagnoseBrokenCreationChains(resolved, allAssocs)...)
	diagnostics = append(diagnostics, diagnoseIsolatedClasses(resolved, model)...)
	diagnostics = append(diagnostics, diagnoseHalfAssociations(resolved, allAssocs)...)

	// 4. Excluded invariants (already reported as warnings in resolved.Warnings).
	// We don't duplicate those here.

	// 5. Realized domain included (already reported in resolver warnings).

	// Diagnostics 6 and 7 require CallerData.
	if callerData != nil {
		diagnostics = append(diagnostics, diagnoseCallerData(resolved, model, callerData)...)
	}

	return diagnostics
}

// diagnoseBrokenCreationChains finds mandatory outbound associations to excluded classes.
func diagnoseBrokenCreationChains(resolved *ResolvedSurface, allAssocs map[identity.Key]model_class.Association) []Diagnostic {
	var diagnostics []Diagnostic
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
	return diagnostics
}

// diagnoseIsolatedClasses finds classes in scope with no creation events,
// but only when they had creation events in the full model.
func diagnoseIsolatedClasses(resolved *ResolvedSurface, model *core.Model) []Diagnostic {
	var diagnostics []Diagnostic
	for classKey, class := range resolved.Classes {
		if classHasCreationEvent(class.Transitions) {
			continue
		}
		fullClass, found := findClassInModel(classKey, model)
		if !found || !classHasCreationEvent(fullClass.Transitions) {
			continue
		}
		ck := classKey
		diagnostics = append(diagnostics, Diagnostic{
			Level:    "warning",
			Message:  fmt.Sprintf("isolated class: %s has no creation events and cannot be instantiated", classKey.String()),
			ClassKey: &ck,
		})
	}
	return diagnostics
}

// classHasCreationEvent checks if any transition has a nil FromStateKey.
func classHasCreationEvent(transitions map[identity.Key]model_state.Transition) bool {
	for _, t := range transitions {
		if t.FromStateKey == nil {
			return true
		}
	}
	return false
}

// diagnoseHalfAssociations finds associations where one endpoint is in scope and the other is not.
func diagnoseHalfAssociations(resolved *ResolvedSurface, allAssocs map[identity.Key]model_class.Association) []Diagnostic {
	var diagnostics []Diagnostic
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
	return diagnostics
}

// diagnoseCallerData produces diagnostics related to SentBy/CalledBy metadata.
func diagnoseCallerData(resolved *ResolvedSurface, model *core.Model, cd *CallerData) []Diagnostic {
	var diagnostics []Diagnostic

	diagnostics = append(diagnostics, diagnoseAllEventsInternal(resolved, cd)...)
	diagnostics = append(diagnostics, diagnoseUnknownClassRefs(model, cd)...)

	return diagnostics
}

// diagnoseAllEventsInternal finds classes where ALL events have SentBy pointing to in-scope classes.
func diagnoseAllEventsInternal(resolved *ResolvedSurface, cd *CallerData) []Diagnostic {
	var diagnostics []Diagnostic
	for classKey, class := range resolved.Classes {
		if len(class.Events) == 0 {
			continue
		}
		allInternal := true
		for _, event := range class.Events {
			if !eventHasInScopeSender(event.Key, resolved, cd) {
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
	return diagnostics
}

// eventHasInScopeSender checks if an event has at least one sender that is in scope.
func eventHasInScopeSender(eventKey identity.Key, resolved *ResolvedSurface, cd *CallerData) bool {
	senders := cd.EventSentBy[eventKey]
	if len(senders) == 0 {
		return false
	}
	for _, senderKey := range senders {
		if _, inScope := resolved.Classes[senderKey]; inScope {
			return true
		}
	}
	return false
}

// diagnoseUnknownClassRefs finds SentBy/CalledBy entries referencing classes not in the model.
func diagnoseUnknownClassRefs(model *core.Model, cd *CallerData) []Diagnostic {
	var diagnostics []Diagnostic

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
func findClassInModel(classKey identity.Key, model *core.Model) (*model_class.Class, bool) {
	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			if class, ok := subdomain.Classes[classKey]; ok {
				return &class, true
			}
		}
	}
	return nil, false
}

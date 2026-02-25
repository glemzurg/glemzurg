package engine

import (
	"sort"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/surface"
)

// ClassInfo holds pre-computed simulation metadata for one class.
type ClassInfo struct {
	Class          model_class.Class
	ClassKey       identity.Key
	CreationEvents []model_state.Event             // Events that have creation transitions (FromStateKey==nil).
	StateEvents    map[string][]EventInfo          // stateName → eligible events from that state.
	DoActions      map[string][]model_state.Action // stateName → "do" actions available while in state.
	HasStates      bool
}

// EventInfo pairs an event with the transitions it can trigger from a specific state.
type EventInfo struct {
	Event       model_state.Event
	Transitions []model_state.Transition
}

// AssociationInfo holds pre-computed metadata for one class association.
type AssociationInfo struct {
	Association   model_class.Association
	FromClassKey  identity.Key
	ToClassKey    identity.Key
	MandatoryTo   bool // ToMultiplicity.LowerBound >= 1
	MandatoryFrom bool // FromMultiplicity.LowerBound >= 1
	MinTo         uint // ToMultiplicity.LowerBound
	MinFrom       uint // FromMultiplicity.LowerBound
}

// ClassCatalog pre-computes per-class simulation metadata from the model.
type ClassCatalog struct {
	classes      map[identity.Key]*ClassInfo
	associations []AssociationInfo
	classAssocs  map[identity.Key][]AssociationInfo // classKey → associations involving it

	// Simulator-local SentBy/CalledBy data.
	eventSentBy    map[identity.Key][]identity.Key // event key → sender class keys
	actionCalledBy map[identity.Key][]identity.Key // action key → caller class keys
	queryCalledBy  map[identity.Key][]identity.Key // query key → caller class keys
}

// NewClassCatalog builds a class catalog from the model.
func NewClassCatalog(model *req_model.Model) *ClassCatalog {
	catalog := &ClassCatalog{
		classes:        make(map[identity.Key]*ClassInfo),
		classAssocs:    make(map[identity.Key][]AssociationInfo),
		eventSentBy:    make(map[identity.Key][]identity.Key),
		actionCalledBy: make(map[identity.Key][]identity.Key),
		queryCalledBy:  make(map[identity.Key][]identity.Key),
	}

	// Walk all classes in the model.
	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			for _, class := range subdomain.Classes {
				if len(class.States) == 0 {
					continue // Skip classes without state machines.
				}

				info := &ClassInfo{
					Class:       class,
					ClassKey:    class.Key,
					StateEvents: make(map[string][]EventInfo),
					DoActions:   make(map[string][]model_state.Action),
					HasStates:   true,
				}

				// Build event lookup by key.
				eventByKey := make(map[identity.Key]model_state.Event)
				for _, e := range class.Events {
					eventByKey[e.Key] = e
				}

				// Find creation events: transitions where FromStateKey==nil.
				creationEventKeys := make(map[identity.Key]bool)
				for _, t := range class.Transitions {
					if t.FromStateKey == nil {
						creationEventKeys[t.EventKey] = true
					}
				}
				for ek := range creationEventKeys {
					if ev, ok := eventByKey[ek]; ok {
						info.CreationEvents = append(info.CreationEvents, ev)
					}
				}
				// Sort creation events for deterministic ordering.
				sort.Slice(info.CreationEvents, func(i, j int) bool {
					return info.CreationEvents[i].Key.String() < info.CreationEvents[j].Key.String()
				})

				// Build per-state event info.
				for _, s := range class.States {
					var eventInfos []EventInfo
					// Group transitions by event key for this state.
					eventTransitions := make(map[identity.Key][]model_state.Transition)
					for _, t := range class.Transitions {
						if t.FromStateKey != nil && *t.FromStateKey == s.Key {
							eventTransitions[t.EventKey] = append(eventTransitions[t.EventKey], t)
						}
					}
					for ek, transitions := range eventTransitions {
						if ev, ok := eventByKey[ek]; ok {
							eventInfos = append(eventInfos, EventInfo{
								Event:       ev,
								Transitions: transitions,
							})
						}
					}
					// Sort for determinism.
					sort.Slice(eventInfos, func(i, j int) bool {
						return eventInfos[i].Event.Key.String() < eventInfos[j].Event.Key.String()
					})
					if len(eventInfos) > 0 {
						info.StateEvents[s.Name] = eventInfos
					}

					// Build "do" actions for this state.
					var doActions []model_state.Action
					for _, sa := range s.Actions {
						if sa.When == "do" {
							if action, ok := class.Actions[sa.ActionKey]; ok {
								doActions = append(doActions, action)
							}
						}
					}
					if len(doActions) > 0 {
						sort.Slice(doActions, func(i, j int) bool {
							return doActions[i].Key.String() < doActions[j].Key.String()
						})
						info.DoActions[s.Name] = doActions
					}
				}

				catalog.classes[class.Key] = info
			}
		}
	}

	// Build association info.
	allAssocs := model.GetClassAssociations()
	for _, assoc := range allAssocs {
		ai := AssociationInfo{
			Association:   assoc,
			FromClassKey:  assoc.FromClassKey,
			ToClassKey:    assoc.ToClassKey,
			MandatoryTo:   assoc.ToMultiplicity.LowerBound >= 1,
			MandatoryFrom: assoc.FromMultiplicity.LowerBound >= 1,
			MinTo:         assoc.ToMultiplicity.LowerBound,
			MinFrom:       assoc.FromMultiplicity.LowerBound,
		}
		catalog.associations = append(catalog.associations, ai)
		catalog.classAssocs[assoc.FromClassKey] = append(catalog.classAssocs[assoc.FromClassKey], ai)
		if assoc.FromClassKey != assoc.ToClassKey {
			catalog.classAssocs[assoc.ToClassKey] = append(catalog.classAssocs[assoc.ToClassKey], ai)
		}
	}
	// Sort associations for determinism.
	sort.Slice(catalog.associations, func(i, j int) bool {
		return catalog.associations[i].Association.Key.String() < catalog.associations[j].Association.Key.String()
	})

	return catalog
}

// SetEventSentBy records which classes send a given event.
func (c *ClassCatalog) SetEventSentBy(eventKey identity.Key, senderClassKeys []identity.Key) {
	c.eventSentBy[eventKey] = senderClassKeys
}

// SetActionCalledBy records which classes call a given action.
func (c *ClassCatalog) SetActionCalledBy(actionKey identity.Key, callerClassKeys []identity.Key) {
	c.actionCalledBy[actionKey] = callerClassKeys
}

// SetQueryCalledBy records which classes call a given query.
func (c *ClassCatalog) SetQueryCalledBy(queryKey identity.Key, callerClassKeys []identity.Key) {
	c.queryCalledBy[queryKey] = callerClassKeys
}

// CallerData exports the SentBy/CalledBy metadata as a surface.CallerData
// for use with surface.Diagnose.
func (c *ClassCatalog) CallerData() *surface.CallerData {
	return &surface.CallerData{
		EventSentBy:    c.eventSentBy,
		ActionCalledBy: c.actionCalledBy,
		QueryCalledBy:  c.queryCalledBy,
	}
}

// GetClassInfo returns the pre-computed info for a class, or nil if not found.
func (c *ClassCatalog) GetClassInfo(classKey identity.Key) *ClassInfo {
	return c.classes[classKey]
}

// AllSimulatableClasses returns all classes with state machines, sorted by key.
func (c *ClassCatalog) AllSimulatableClasses() []*ClassInfo {
	result := make([]*ClassInfo, 0, len(c.classes))
	for _, info := range c.classes {
		result = append(result, info)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].ClassKey.String() < result[j].ClassKey.String()
	})
	return result
}

// GetMandatoryOutboundAssociations returns associations where the given class is
// the "from" side and the "to" side requires at least one instance (LowerBound >= 1).
func (c *ClassCatalog) GetMandatoryOutboundAssociations(classKey identity.Key) []AssociationInfo {
	var result []AssociationInfo
	for _, ai := range c.classAssocs[classKey] {
		if ai.FromClassKey == classKey && ai.MandatoryTo {
			result = append(result, ai)
		}
	}
	return result
}

// GetCreationEvent returns the first creation event for a class (if any).
func (c *ClassCatalog) GetCreationEvent(classKey identity.Key) (*model_state.Event, bool) {
	info := c.classes[classKey]
	if info == nil || len(info.CreationEvents) == 0 {
		return nil, false
	}
	return &info.CreationEvents[0], true
}

// AllAssociations returns all associations in the catalog.
func (c *ClassCatalog) AllAssociations() []AssociationInfo {
	return c.associations
}

// GetAssociationsForClass returns all associations involving the given class.
func (c *ClassCatalog) GetAssociationsForClass(classKey identity.Key) []AssociationInfo {
	return c.classAssocs[classKey]
}

// ExternalCreationEvents returns creation events that are NOT triggered by
// another in-scope class through an association. A creation event is "internal"
// if another class in the catalog has a mandatory outbound association that
// targets this class — meaning the other class creates this one as part of
// its own creation chain.
func (c *ClassCatalog) ExternalCreationEvents(classKey identity.Key) []model_state.Event {
	info := c.classes[classKey]
	if info == nil {
		return nil
	}

	// Check if any other class in scope has a mandatory association pointing to this class.
	for otherKey, otherInfo := range c.classes {
		if otherKey == classKey {
			continue
		}
		_ = otherInfo
		for _, ai := range c.classAssocs[otherKey] {
			if ai.FromClassKey == otherKey && ai.ToClassKey == classKey && ai.MandatoryTo {
				// This class's creation is driven by another class — not external.
				return nil
			}
		}
	}

	return info.CreationEvents
}

// ExternalStateEvents returns events eligible for external (top-level) firing
// on an instance in a given state. An event is "internal" if its SentBy list
// contains any class that is in scope (i.e., in the catalog). Only truly
// external events are returned for top-level simulation selection.
func (c *ClassCatalog) ExternalStateEvents(classKey identity.Key, stateName string) []EventInfo {
	info := c.classes[classKey]
	if info == nil {
		return nil
	}
	allEvents := info.StateEvents[stateName]
	if len(allEvents) == 0 {
		return nil
	}

	var external []EventInfo
	for _, ei := range allEvents {
		if c.isEventExternal(ei.Event) {
			external = append(external, ei)
		}
	}
	return external
}

// ExternalDoActions returns "do" actions eligible for top-level firing in a state.
// A "do" action is "internal" if its CalledBy list contains an in-scope class.
func (c *ClassCatalog) ExternalDoActions(classKey identity.Key, stateName string) []model_state.Action {
	info := c.classes[classKey]
	if info == nil {
		return nil
	}
	allDo := info.DoActions[stateName]
	if len(allDo) == 0 {
		return nil
	}

	var external []model_state.Action
	for _, action := range allDo {
		if c.isActionExternal(action) {
			external = append(external, action)
		}
	}
	return external
}

// isEventExternal returns true if the event has no SentBy classes in scope.
func (c *ClassCatalog) isEventExternal(event model_state.Event) bool {
	senders := c.eventSentBy[event.Key]
	if len(senders) == 0 {
		return true // No senders declared — always external.
	}
	for _, senderKey := range senders {
		if _, inScope := c.classes[senderKey]; inScope {
			return false // A sender is in scope — this event is internal.
		}
	}
	return true // No senders are in scope — external.
}

// isActionExternal returns true if the action has no CalledBy classes in scope.
func (c *ClassCatalog) isActionExternal(action model_state.Action) bool {
	callers := c.actionCalledBy[action.Key]
	if len(callers) == 0 {
		return true // No callers declared — always external.
	}
	for _, callerKey := range callers {
		if _, inScope := c.classes[callerKey]; inScope {
			return false // A caller is in scope — this action is internal.
		}
	}
	return true // No callers are in scope — external.
}

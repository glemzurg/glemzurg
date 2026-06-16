package engine

import (
	"slices"
	"sort"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
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
	HasEvents      bool
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
func NewClassCatalog(model *core.Model) *ClassCatalog {
	catalog := &ClassCatalog{
		classes:        make(map[identity.Key]*ClassInfo),
		classAssocs:    make(map[identity.Key][]AssociationInfo),
		eventSentBy:    make(map[identity.Key][]identity.Key),
		actionCalledBy: make(map[identity.Key][]identity.Key),
		queryCalledBy:  make(map[identity.Key][]identity.Key),
	}

	// Walk all in-scope classes; stateless classes are liveness-only metadata.
	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			for _, class := range subdomain.Classes {
				catalog.classes[class.Key] = buildScopedClassInfo(class)
			}
		}
	}

	// Build association info.
	catalog.buildAssociationInfo(model)

	return catalog
}

// buildScopedClassInfo creates catalog metadata for any in-scope class.
// Stateless classes stay in the catalog for liveness even though they cannot simulate.
func buildScopedClassInfo(class model_class.Class) *ClassInfo {
	if len(class.States) == 0 {
		return &ClassInfo{
			Class:       class,
			ClassKey:    class.Key,
			StateEvents: make(map[string][]EventInfo),
			DoActions:   make(map[string][]model_state.Action),
			HasEvents:   len(class.Events) > 0,
		}
	}
	return buildClassInfo(class)
}

// buildClassInfo creates pre-computed simulation metadata for a simulatable class.
func buildClassInfo(class model_class.Class) *ClassInfo {
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

	info.CreationEvents = findCreationEvents(class, eventByKey)
	info.HasEvents = len(class.Events) > 0
	buildPerStateInfo(info, class, eventByKey)

	return info
}

// findCreationEvents finds events that trigger creation transitions (FromStateKey==nil).
func findCreationEvents(class model_class.Class, eventByKey map[identity.Key]model_state.Event) []model_state.Event {
	creationEventKeys := make(map[identity.Key]bool)
	for _, t := range class.Transitions {
		if t.FromStateKey == nil {
			creationEventKeys[t.EventKey] = true
		}
	}

	var events []model_state.Event
	for ek := range creationEventKeys {
		if ev, ok := eventByKey[ek]; ok {
			events = append(events, ev)
		}
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].Key.String() < events[j].Key.String()
	})
	return events
}

// buildPerStateInfo populates StateEvents and DoActions for each state in the class.
func buildPerStateInfo(info *ClassInfo, class model_class.Class, eventByKey map[identity.Key]model_state.Event) {
	for _, s := range class.States {
		eventInfos := buildStateEventInfos(class, s, eventByKey)
		if len(eventInfos) > 0 {
			info.StateEvents[s.Name] = eventInfos
		}

		doActions := buildDoActions(class, s)
		if len(doActions) > 0 {
			info.DoActions[s.Name] = doActions
		}
	}
}

// buildStateEventInfos builds the event infos for transitions from a specific state.
func buildStateEventInfos(
	class model_class.Class,
	s model_state.State,
	eventByKey map[identity.Key]model_state.Event,
) []EventInfo {
	// Group transitions by event key for this state.
	eventTransitions := make(map[identity.Key][]model_state.Transition)
	for _, t := range class.Transitions {
		if t.FromStateKey != nil && *t.FromStateKey == s.Key {
			eventTransitions[t.EventKey] = append(eventTransitions[t.EventKey], t)
		}
	}

	var eventInfos []EventInfo
	for ek, transitions := range eventTransitions {
		if ev, ok := eventByKey[ek]; ok {
			eventInfos = append(eventInfos, EventInfo{
				Event:       ev,
				Transitions: transitions,
			})
		}
	}

	sort.Slice(eventInfos, func(i, j int) bool {
		return eventInfos[i].Event.Key.String() < eventInfos[j].Event.Key.String()
	})
	return eventInfos
}

// buildDoActions builds "do" actions for a specific state.
func buildDoActions(class model_class.Class, s model_state.State) []model_state.Action {
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
	}
	return doActions
}

// buildAssociationInfo builds association metadata from the model.
func (c *ClassCatalog) buildAssociationInfo(model *core.Model) {
	allAssocs := model.GetClassAssociations()
	for _, assoc := range allAssocs {
		if _, fromIn := c.classes[assoc.FromClassKey]; !fromIn {
			continue
		}
		if _, toIn := c.classes[assoc.ToClassKey]; !toIn {
			continue
		}
		ai := AssociationInfo{
			Association:   assoc,
			FromClassKey:  assoc.FromClassKey,
			ToClassKey:    assoc.ToClassKey,
			MandatoryTo:   assoc.ToMultiplicity.LowerBound >= 1,
			MandatoryFrom: assoc.FromMultiplicity.LowerBound >= 1,
			MinTo:         assoc.ToMultiplicity.LowerBound,
			MinFrom:       assoc.FromMultiplicity.LowerBound,
		}
		c.associations = append(c.associations, ai)
		c.classAssocs[assoc.FromClassKey] = append(c.classAssocs[assoc.FromClassKey], ai)
		if assoc.FromClassKey != assoc.ToClassKey {
			c.classAssocs[assoc.ToClassKey] = append(c.classAssocs[assoc.ToClassKey], ai)
		}
	}
	// Sort associations for determinism.
	sort.Slice(c.associations, func(i, j int) bool {
		return c.associations[i].Association.Key.String() < c.associations[j].Association.Key.String()
	})
}

// SetEventSentBy records which classes send a given event.
func (c *ClassCatalog) SetEventSentBy(eventKey identity.Key, senderClassKeys []identity.Key) {
	c.eventSentBy[eventKey] = senderClassKeys
}

func (c *ClassCatalog) addEventSender(eventKey, senderClassKey identity.Key) {
	if slices.Contains(c.eventSentBy[eventKey], senderClassKey) {
		return
	}
	c.eventSentBy[eventKey] = append(c.eventSentBy[eventKey], senderClassKey)
}

// SetActionCalledBy records which classes call a given action.
func (c *ClassCatalog) SetActionCalledBy(actionKey identity.Key, callerClassKeys []identity.Key) {
	c.actionCalledBy[actionKey] = callerClassKeys
}

// SetQueryCalledBy records which classes call a given query.
func (c *ClassCatalog) SetQueryCalledBy(queryKey identity.Key, callerClassKeys []identity.Key) {
	c.queryCalledBy[queryKey] = callerClassKeys
}

func (c *ClassCatalog) addQueryCaller(queryKey, callerClassKey identity.Key) {
	if slices.Contains(c.queryCalledBy[queryKey], callerClassKey) {
		return
	}
	c.queryCalledBy[queryKey] = append(c.queryCalledBy[queryKey], callerClassKey)
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

// AllScopedClasses returns every in-scope class (simulatable and stateless), sorted by key.
func (c *ClassCatalog) AllScopedClasses() []*ClassInfo {
	result := make([]*ClassInfo, 0, len(c.classes))
	for _, info := range c.classes {
		result = append(result, info)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].ClassKey.String() < result[j].ClassKey.String()
	})
	return result
}

// AllSimulatableClasses returns classes with state machines, sorted by key.
func (c *ClassCatalog) AllSimulatableClasses() []*ClassInfo {
	result := make([]*ClassInfo, 0, len(c.classes))
	for _, info := range c.classes {
		if info.HasStates {
			result = append(result, info)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].ClassKey.String() < result[j].ClassKey.String()
	})
	return result
}

// AllEventBearingClasses returns simulatable classes that declare at least one event.
func (c *ClassCatalog) AllEventBearingClasses() []*ClassInfo {
	result := make([]*ClassInfo, 0, len(c.classes))
	for _, info := range c.classes {
		if info.HasEvents {
			result = append(result, info)
		}
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

// GetActionForEvent resolves the action wired to a transition for the given event and instance state.
// When multiple transitions share the event, the first matching transition with an action is returned.
func (c *ClassCatalog) GetActionForEvent(
	classKey identity.Key,
	eventKey identity.Key,
	instanceStateName string,
) (*model_state.Action, bool) {
	info := c.classes[classKey]
	if info == nil {
		return nil, false
	}

	class := info.Class
	var fromStateKey *identity.Key
	if instanceStateName != "" {
		for _, s := range class.States {
			if s.Name == instanceStateName {
				key := s.Key
				fromStateKey = &key
				break
			}
		}
	}

	var matches []model_state.Transition
	for _, t := range class.Transitions {
		if t.EventKey != eventKey {
			continue
		}
		if instanceStateName == "" {
			if t.FromStateKey != nil {
				continue
			}
		} else if t.FromStateKey == nil || fromStateKey == nil || *t.FromStateKey != *fromStateKey {
			continue
		}
		matches = append(matches, t)
	}

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Key.String() < matches[j].Key.String()
	})

	for _, t := range matches {
		if t.ActionKey == nil {
			return nil, true
		}
		if action, ok := class.Actions[*t.ActionKey]; ok {
			return &action, true
		}
	}
	return nil, false
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

// ExternalCreationEvents returns creation events eligible for top-level firing.
// An event is excluded when a simulatable in-scope class sends it (SentBy) or
// when another class's mandatory outbound association targets this class.
func (c *ClassCatalog) ExternalCreationEvents(classKey identity.Key) []model_state.Event {
	info := c.classes[classKey]
	if info == nil || len(info.CreationEvents) == 0 {
		return nil
	}

	if c.isMandatoryAssociationCreationTarget(classKey) {
		return nil
	}

	var external []model_state.Event
	for _, ev := range info.CreationEvents {
		if c.isEventExternal(ev) {
			external = append(external, ev)
		}
	}
	return external
}

func (c *ClassCatalog) isMandatoryAssociationCreationTarget(classKey identity.Key) bool {
	for otherKey, otherInfo := range c.classes {
		if otherKey == classKey || !otherInfo.HasStates {
			continue
		}
		for _, ai := range c.classAssocs[otherKey] {
			if ai.FromClassKey == otherKey && ai.ToClassKey == classKey && ai.MandatoryTo {
				return true
			}
		}
	}
	return false
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

// ExternalQueries returns queries eligible for top-level firing on existing instances.
// A query is "internal" if its CalledBy list contains a simulatable in-scope class.
func (c *ClassCatalog) ExternalQueries(classKey identity.Key) []model_state.Query {
	info := c.classes[classKey]
	if info == nil || len(info.Class.Queries) == 0 {
		return nil
	}

	queries := make([]model_state.Query, 0, len(info.Class.Queries))
	for _, query := range info.Class.Queries {
		if c.isQueryExternal(query) {
			queries = append(queries, query)
		}
	}
	sort.Slice(queries, func(i, j int) bool {
		return queries[i].Key.String() < queries[j].Key.String()
	})
	return queries
}

// SurfaceDoActions returns all "do" state actions for top-level simulation.
// Do-actions are surface-level by nature — they are not filtered by CalledBy.
func (c *ClassCatalog) SurfaceDoActions(classKey identity.Key, stateName string) []model_state.Action {
	info := c.classes[classKey]
	if info == nil {
		return nil
	}
	return info.DoActions[stateName]
}

// isEventExternal returns true when no simulatable in-scope class sends the event.
func (c *ClassCatalog) isEventExternal(event model_state.Event) bool {
	return !c.hasSimulatableSender(c.eventSentBy[event.Key])
}

func (c *ClassCatalog) isQueryExternal(query model_state.Query) bool {
	return !c.hasSimulatableSender(c.queryCalledBy[query.Key])
}

func (c *ClassCatalog) hasSimulatableSender(senders []identity.Key) bool {
	for _, senderKey := range senders {
		if info, ok := c.classes[senderKey]; ok && info.HasStates {
			return true
		}
	}
	return false
}

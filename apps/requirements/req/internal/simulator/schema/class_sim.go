package schema

import (
	"fmt"
	"sort"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// EventInfo pairs an event with the transitions it can trigger from a specific state.
type EventInfo struct {
	Event       model_state.Event
	Transitions []model_state.Transition
}

// ClassSimInfo holds pre-computed simulation metadata for one in-scope class.
type ClassSimInfo struct {
	Class          model_class.Class
	ClassKey       identity.Key
	CreationEvents []model_state.Event
	StateEvents    map[string][]EventInfo
	DoActions      map[string][]model_state.Action
	HasStates      bool
	HasEvents      bool
}

func (s *Schema) reindexClassSim() {
	s.classSim = make(map[identity.Key]*ClassSimInfo)
	s.forEachClassInScope(func(c *model_class.Class) {
		s.classSim[c.Key] = buildClassSimInfo(*c)
	})
}

func buildClassSimInfo(class model_class.Class) *ClassSimInfo {
	if len(class.States) == 0 {
		return &ClassSimInfo{
			Class:       class,
			ClassKey:    class.Key,
			StateEvents: make(map[string][]EventInfo),
			DoActions:   make(map[string][]model_state.Action),
			HasEvents:   len(class.Events) > 0,
		}
	}
	info := &ClassSimInfo{
		Class:       class,
		ClassKey:    class.Key,
		StateEvents: make(map[string][]EventInfo),
		DoActions:   make(map[string][]model_state.Action),
		HasStates:   true,
	}
	eventByKey := make(map[identity.Key]model_state.Event)
	for _, e := range class.Events {
		eventByKey[e.Key] = e
	}
	info.CreationEvents = findCreationEvents(class, eventByKey)
	info.HasEvents = len(class.Events) > 0
	buildPerStateInfo(info, class, eventByKey)
	return info
}

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

func buildPerStateInfo(info *ClassSimInfo, class model_class.Class, eventByKey map[identity.Key]model_state.Event) {
	for _, st := range class.States {
		eventInfos := buildStateEventInfos(class, st, eventByKey)
		if len(eventInfos) > 0 {
			info.StateEvents[st.Name] = eventInfos
		}
		doActions := buildDoActions(class, st)
		if len(doActions) > 0 {
			info.DoActions[st.Name] = doActions
		}
	}
}

func buildStateEventInfos(
	class model_class.Class,
	st model_state.State,
	eventByKey map[identity.Key]model_state.Event,
) []EventInfo {
	eventTransitions := make(map[identity.Key][]model_state.Transition)
	for _, t := range class.Transitions {
		if t.FromStateKey != nil && *t.FromStateKey == st.Key {
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

func buildDoActions(class model_class.Class, st model_state.State) []model_state.Action {
	var doActions []model_state.Action
	for _, sa := range st.Actions {
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

// ClassSim returns simulation metadata for an in-scope class.
func (s *Schema) ClassSim(classKey identity.Key) (*ClassSimInfo, bool, error) {
	if s == nil {
		return nil, false, fmt.Errorf("schema.ClassSim: nil schema")
	}
	if _, ok := s.classes[classKey]; !ok {
		return nil, false, fmt.Errorf("unknown class: %s", classKey.String())
	}
	if !s.inScope[classKey] {
		return nil, false, nil
	}
	info := s.classSim[classKey]
	return info, true, nil
}

// EachInScopeClassSim calls fn for every in-scope class simulation index (key order).
// Prefer keyed ClassSim for single-class questions; use this only for true full-surface scans.
func (s *Schema) EachInScopeClassSim(fn func(*ClassSimInfo)) {
	if s == nil || fn == nil || len(s.classSim) == 0 {
		return
	}
	keys := make([]identity.Key, 0, len(s.classSim))
	for k := range s.classSim {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].String() < keys[j].String()
	})
	for _, k := range keys {
		fn(s.classSim[k])
	}
}

// EachInScopeClass calls fn for every in-scope class model body (key order).
func (s *Schema) EachInScopeClass(fn func(model_class.Class)) {
	s.EachInScopeClassSim(func(sim *ClassSimInfo) {
		if sim != nil {
			fn(sim.Class)
		}
	})
}

// EachSimulatableClassSim calls fn for in-scope classes that have a state machine.
func (s *Schema) EachSimulatableClassSim(fn func(*ClassSimInfo)) {
	s.EachInScopeClassSim(func(sim *ClassSimInfo) {
		if sim != nil && sim.HasStates {
			fn(sim)
		}
	})
}

// EachEventBearingClassSim calls fn for in-scope classes that declare at least one event.
func (s *Schema) EachEventBearingClassSim(fn func(*ClassSimInfo)) {
	s.EachInScopeClassSim(func(sim *ClassSimInfo) {
		if sim != nil && sim.HasEvents {
			fn(sim)
		}
	})
}

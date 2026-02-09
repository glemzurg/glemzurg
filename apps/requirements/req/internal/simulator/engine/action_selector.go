package engine

import (
	"fmt"
	"math/rand"
	"sort"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
)

// PendingAction describes a single eligible simulation action.
type PendingAction struct {
	Class      *ClassInfo
	Event      *model_state.Event   // Non-nil for event-triggered actions.
	DoAction   *model_state.Action  // Non-nil for "do" state actions.
	Instance   *state.ClassInstance // nil for creation.
	IsCreation bool
	IsDo       bool // True when this is a "do" state action.
}

// ActionSelector randomly selects the next simulation action.
type ActionSelector struct {
	catalog *ClassCatalog
	rng     *rand.Rand
}

// NewActionSelector creates a new action selector.
func NewActionSelector(catalog *ClassCatalog, rng *rand.Rand) *ActionSelector {
	return &ActionSelector{
		catalog: catalog,
		rng:     rng,
	}
}

// SelectAction picks a random eligible action from all classes and instances.
// Returns error if no actions are available (deadlock).
func (s *ActionSelector) SelectAction(simState *state.SimulationState) (*PendingAction, error) {
	eligible := s.collectEligibleActions(simState)

	if len(eligible) == 0 {
		return nil, fmt.Errorf("deadlock: no eligible actions")
	}

	chosen := eligible[s.rng.Intn(len(eligible))]
	return &chosen, nil
}

// collectEligibleActions builds the list of all eligible actions across all classes.
func (s *ActionSelector) collectEligibleActions(simState *state.SimulationState) []PendingAction {
	var eligible []PendingAction

	for _, classInfo := range s.catalog.AllSimulatableClasses() {
		// External creation events — only events not driven by another class.
		externalCreationEvents := s.catalog.ExternalCreationEvents(classInfo.ClassKey)
		for i := range externalCreationEvents {
			eligible = append(eligible, PendingAction{
				Class:      classInfo,
				Event:      &externalCreationEvents[i],
				Instance:   nil,
				IsCreation: true,
			})
		}

		// Normal events and "do" actions on existing instances.
		// Sort instances by ID for deterministic ordering (map iteration is non-deterministic).
		instances := simState.InstancesByClass(classInfo.ClassKey)
		sort.Slice(instances, func(i, j int) bool {
			return instances[i].ID < instances[j].ID
		})
		for _, instance := range instances {
			currentState := getInstanceStateName(instance)
			if currentState == "" {
				continue
			}

			// Normal state transition events.
			stateEvents := s.catalog.ExternalStateEvents(classInfo.ClassKey, currentState)
			for i := range stateEvents {
				eligible = append(eligible, PendingAction{
					Class:    classInfo,
					Event:    &stateEvents[i].Event,
					Instance: instance,
				})
			}

			// "Do" actions — only external ones eligible for top-level firing.
			doActions := s.catalog.ExternalDoActions(classInfo.ClassKey, currentState)
			for i := range doActions {
				eligible = append(eligible, PendingAction{
					Class:    classInfo,
					DoAction: &doActions[i],
					Instance: instance,
					IsDo:     true,
				})
			}
		}
	}

	return eligible
}

// getInstanceStateName extracts the current state name from an instance's _state attribute.
func getInstanceStateName(instance *state.ClassInstance) string {
	stateAttr := instance.GetAttribute("_state")
	if stateAttr == nil {
		return ""
	}
	if strObj, ok := stateAttr.(*object.String); ok {
		return strObj.Value()
	}
	return ""
}

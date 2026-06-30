package engine

import (
	"fmt"
	"math/rand"
	"sort"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
)

// PendingAction describes a single eligible simulation action.
type PendingAction struct {
	Class      *ClassInfo
	Event      *model_state.Event   // Non-nil for event-triggered transitions.
	Query      *model_state.Query   // Non-nil for query invocations.
	DoAction   *model_state.Action  // Non-nil for "do" state actions.
	Instance   *state.ClassInstance // nil for creation.
	IsCreation bool
	IsQuery    bool
	IsDo       bool // True when this is a "do" state action.

	// Association-class Add binds both host-association endpoints.
	SourceAssocKey   *identity.Key
	SourceInstanceID *state.InstanceID
	TargetInstanceID *state.InstanceID
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
		if classInfo.HasEvents {
			externalCreationEvents := s.catalog.ExternalCreationEvents(classInfo.ClassKey)
			for i := range externalCreationEvents {
				eligible = append(eligible, PendingAction{
					Class:      classInfo,
					Event:      &externalCreationEvents[i],
					Instance:   nil,
					IsCreation: true,
				})
			}

			eligible = append(eligible, s.collectAssociationClassCreations(classInfo, simState)...)
		}

		instances := simState.InstancesByClass(classInfo.ClassKey)
		sort.Slice(instances, func(i, j int) bool {
			return instances[i].ID < instances[j].ID
		})
		for _, instance := range instances {
			currentState := getInstanceStateName(instance)
			if currentState == "" {
				continue
			}

			if classInfo.HasEvents {
				stateEvents := s.catalog.ExternalStateEvents(classInfo.ClassKey, currentState)
				for i := range stateEvents {
					eligible = append(eligible, PendingAction{
						Class:    classInfo,
						Event:    &stateEvents[i].Event,
						Instance: instance,
					})
				}

				externalQueries := s.catalog.ExternalQueries(classInfo.ClassKey)
				for i := range externalQueries {
					eligible = append(eligible, PendingAction{
						Class:    classInfo,
						Query:    &externalQueries[i],
						Instance: instance,
						IsQuery:  true,
					})
				}
			}

			// Do-actions are always surface-level on existing instances.
			doActions := s.catalog.SurfaceDoActions(classInfo.ClassKey, currentState)
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

func (s *ActionSelector) collectAssociationClassCreations(
	classInfo *ClassInfo,
	simState *state.SimulationState,
) []PendingAction {
	acInfo := s.catalog.LookupAssociationClass(classInfo.ClassKey)
	if acInfo == nil || len(classInfo.CreationEvents) == 0 {
		return nil
	}

	fromInstances := simState.InstancesByClass(acInfo.FromClassKey)
	toInstances := simState.InstancesByClass(acInfo.ToClassKey)
	sort.Slice(fromInstances, func(i, j int) bool { return fromInstances[i].ID < fromInstances[j].ID })
	sort.Slice(toInstances, func(i, j int) bool { return toInstances[i].ID < toInstances[j].ID })
	if len(fromInstances) == 0 || len(toInstances) == 0 {
		return nil
	}

	creationEvent := classInfo.CreationEvents[0]
	var eligible []PendingAction
	hostAssocKey := acInfo.HostAssociation.Key

	hostAssoc := acInfo.HostAssociation
	for _, fromInst := range fromInstances {
		for _, toInst := range toInstances {
			fromID := fromInst.ID
			toID := toInst.ID
			if !s.pairAllowsAnotherLink(hostAssoc, simState, fromID, toID) {
				continue
			}
			eligible = append(eligible, PendingAction{
				Class:            classInfo,
				Event:            &creationEvent,
				Instance:         nil,
				IsCreation:       true,
				SourceAssocKey:   &hostAssocKey,
				SourceInstanceID: &fromID,
				TargetInstanceID: &toID,
			})
		}
	}
	return eligible
}

func (s *ActionSelector) pairAllowsAnotherLink(
	hostAssoc model_class.Association,
	simState *state.SimulationState,
	fromID, toID state.InstanceID,
) bool {
	return simState.CountActivePairLinks(hostAssoc, fromID, toID) == 0
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

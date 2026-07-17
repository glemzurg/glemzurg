package engine

import (
	"fmt"
	"math/rand"
	"sort"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_data_type"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/actions"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
)

// PendingAction describes a single eligible simulation action.
type PendingAction struct {
	Class            *ClassInfo
	Event            *model_state.Event     // Non-nil for event-triggered transitions.
	Query            *model_state.Query     // Non-nil for query invocations.
	DerivedAttribute *model_class.Attribute // Non-nil for derived attribute reads.
	DoAction         *model_state.Action    // Non-nil for "do" state actions.
	Instance         *state.ClassInstance   // nil for creation.
	IsCreation       bool
	IsQuery          bool
	IsDerivedRead    bool // True when this reads an external derived attribute.
	IsDo             bool // True when this is a "do" state action.

	// Association-class Add binds both host-association endpoints.
	SourceAssocKey   *identity.Key
	SourceInstanceID *state.InstanceID
	TargetInstanceID *state.InstanceID
}

// ActionSelector randomly selects the next simulation action.
type ActionSelector struct {
	catalog         *ClassCatalog
	derivedEval     *DerivedAttributeEvaluator
	bindingsBuilder *state.BindingsBuilder
	rng             *rand.Rand
}

// NewActionSelector creates a new action selector.
func NewActionSelector(
	catalog *ClassCatalog,
	derivedEval *DerivedAttributeEvaluator,
	bindingsBuilder *state.BindingsBuilder,
	rng *rand.Rand,
) *ActionSelector {
	return &ActionSelector{
		catalog:         catalog,
		derivedEval:     derivedEval,
		bindingsBuilder: bindingsBuilder,
		rng:             rng,
	}
}

// SelectAction picks a random eligible action from all classes and instances.
// Returns error if no actions are available (deadlock).
func (s *ActionSelector) SelectAction(simState *state.SimulationState) (*PendingAction, error) {
	eligible := s.collectEligibleActions(simState)
	eligible = s.filterByObjectParamAvailability(eligible, simState)
	eligible = s.filterBySimulationRequires(eligible)

	if len(eligible) == 0 {
		return nil, fmt.Errorf("deadlock: no eligible actions")
	}

	chosen := eligible[s.rng.Intn(len(eligible))]
	return &chosen, nil
}

// filterByObjectParamAvailability drops events/actions whose object-of parameters
// name an in-scope class that has no instances yet. Out-of-scope object classes
// always pass (sampled as empty set). Model-agnostic.
func (s *ActionSelector) filterByObjectParamAvailability(
	eligible []PendingAction,
	simState *state.SimulationState,
) []PendingAction {
	if s.catalog == nil || simState == nil {
		return eligible
	}
	filtered := make([]PendingAction, 0, len(eligible))
	for _, pending := range eligible {
		if s.objectParamsHaveInstances(pending, simState) {
			filtered = append(filtered, pending)
		}
	}
	return filtered
}

func (s *ActionSelector) objectParamsHaveInstances(
	pending PendingAction,
	simState *state.SimulationState,
) bool {
	for _, classKey := range s.requiredObjectParamClasses(pending) {
		if !s.catalog.IsClassInScope(classKey) {
			continue
		}
		if len(simState.InstancesByClass(classKey)) == 0 {
			return false
		}
	}
	return true
}

// requiredObjectParamClasses lists class keys referenced by object-of parameters
// on the pending surface action or query.
func (s *ActionSelector) requiredObjectParamClasses(pending PendingAction) []identity.Key {
	var params []model_state.Parameter
	switch {
	case pending.IsQuery && pending.Query != nil:
		params = pending.Query.Parameters
	case pending.IsDo && pending.DoAction != nil:
		params = pending.DoAction.Parameters
	default:
		action := s.resolveSurfaceAction(pending)
		if action == nil {
			return nil
		}
		params = action.Parameters
		if pending.Event != nil && len(pending.Event.ParameterNames) > 0 {
			params = actions.MatchActionParametersByEventNames(pending.Event.ParameterNames, action)
		}
	}
	var keys []identity.Key
	seen := make(map[identity.Key]bool)
	for _, param := range params {
		for _, classKey := range objectClassKeysFromDataType(param.DataType, s.catalog) {
			if seen[classKey] {
				continue
			}
			seen[classKey] = true
			keys = append(keys, classKey)
		}
	}
	return keys
}

// objectClassKeysFromDataType collects in-catalog class keys referenced by object-of
// constraints anywhere in a parameter data type tree.
func objectClassKeysFromDataType(dt *model_data_type.DataType, catalog *ClassCatalog) []identity.Key {
	if dt == nil || catalog == nil {
		return nil
	}
	var keys []identity.Key
	if dt.Atomic != nil &&
		dt.Atomic.ConstraintType == model_data_type.CONSTRAINT_TYPE_OBJECT &&
		dt.Atomic.ObjectClassKey != nil {
		if classKey, ok := resolveObjectClassRef(*dt.Atomic.ObjectClassKey, catalog); ok {
			keys = append(keys, classKey)
		}
	}
	if dt.ElementDataType != nil {
		keys = append(keys, objectClassKeysFromDataType(dt.ElementDataType, catalog)...)
	}
	for i := range dt.RecordFields {
		keys = append(keys, objectClassKeysFromDataType(dt.RecordFields[i].FieldDataType, catalog)...)
	}
	return keys
}

// resolveObjectClassRef maps an object-of class reference to a catalog class key.
// Prefers in-scope classes; falls back to full extent names (out-of-scope) so callers
// can still distinguish "known but OOS" (allow empty) from "unknown".
func resolveObjectClassRef(objectClassRef string, catalog *ClassCatalog) (identity.Key, bool) {
	if catalog == nil || objectClassRef == "" {
		return identity.Key{}, false
	}
	want := identity.NormalizeSubKey(objectClassRef)
	for _, info := range catalog.AllScopedClasses() {
		if objectClassRefMatches(want, objectClassRef, info) {
			return info.ClassKey, true
		}
	}
	// Known only as out-of-scope extent: still return a key so OOS path can skip the gate.
	for classKey, tlaName := range catalog.ClassNameMap() {
		if classKey.SubKey == objectClassRef || classKey.String() == objectClassRef {
			return classKey, true
		}
		if identity.NormalizeSubKey(tlaName) == want || tlaName == objectClassRef {
			return classKey, true
		}
	}
	return identity.Key{}, false
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
			// Association-class _new is never surface: only cascade/peer association materialization.
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

			eligible = append(eligible, s.collectDerivedReadActions(classInfo, instance)...)
		}
	}

	return eligible
}

func (s *ActionSelector) filterBySimulationRequires(eligible []PendingAction) []PendingAction {
	if s.bindingsBuilder == nil {
		return eligible
	}
	classNameMap := s.catalog.ClassNameMap()
	filtered := make([]PendingAction, 0, len(eligible))
	for _, pending := range eligible {
		if pending.IsQuery || pending.IsDerivedRead {
			filtered = append(filtered, pending)
			continue
		}
		action := s.resolveSurfaceAction(pending)
		if action == nil || !actions.ActionHasParameterSimulation(*action) {
			filtered = append(filtered, pending)
			continue
		}
		bindings := actions.BuildSimulationBindings(s.bindingsBuilder, classNameMap, pending.Instance)
		ok, err := actions.ActionSimulationRequiresMet(*action, bindings)
		if err != nil || !ok {
			continue
		}
		filtered = append(filtered, pending)
	}
	return filtered
}

func (s *ActionSelector) resolveSurfaceAction(pending PendingAction) *model_state.Action {
	if pending.DoAction != nil {
		return pending.DoAction
	}
	if pending.Event == nil {
		return nil
	}
	instanceState := ""
	if pending.Instance != nil {
		instanceState = getInstanceStateName(pending.Instance)
	}
	action, found := s.catalog.GetActionForEvent(pending.Class.ClassKey, pending.Event.Key, instanceState)
	if !found {
		return nil
	}
	return action
}

func (s *ActionSelector) collectDerivedReadActions(
	classInfo *ClassInfo,
	instance *state.ClassInstance,
) []PendingAction {
	externalDerived := s.catalog.ExternalDerivedAttributes(classInfo.ClassKey)
	var eligible []PendingAction
	for i := range externalDerived {
		eligible = append(eligible, PendingAction{
			Class:            classInfo,
			DerivedAttribute: &externalDerived[i],
			Instance:         instance,
			IsDerivedRead:    true,
		})
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

package engine

import (
	"fmt"
	"math/rand"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/actions"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/invariants"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
)

const maxCascadeDepth = 20

// CreationChainHandler handles cascading creation of mandatory associated instances.
// When a new instance is created, this checks if the class has mandatory outbound
// associations and triggers creation events on the target classes.
type CreationChainHandler struct {
	catalog         *ClassCatalog
	actionExecutor  *actions.ActionExecutor
	stateActionExec *StateActionExecutor
	paramBinder     *actions.ParameterBinder
	rng             *rand.Rand
}

// NewCreationChainHandler creates a new creation chain handler.
func NewCreationChainHandler(
	catalog *ClassCatalog,
	actionExecutor *actions.ActionExecutor,
	stateActionExec *StateActionExecutor,
	paramBinder *actions.ParameterBinder,
	rng *rand.Rand,
) *CreationChainHandler {
	return &CreationChainHandler{
		catalog:         catalog,
		actionExecutor:  actionExecutor,
		stateActionExec: stateActionExec,
		paramBinder:     paramBinder,
		rng:             rng,
	}
}

// HandleCreationChain fires creation events on mandatory associated classes.
// Called after a creation transition completes. Returns all cascaded steps
// and accumulated violations.
func (h *CreationChainHandler) HandleCreationChain(
	createdInstanceID state.InstanceID,
	simState *state.SimulationState,
	depth int,
) ([]*SimulationStep, invariants.ViolationErrors, error) {
	if depth > maxCascadeDepth {
		return nil, nil, fmt.Errorf("creation chain cascade exceeded max depth of %d", maxCascadeDepth)
	}

	// Get the class key from the created instance.
	instance := simState.GetInstance(createdInstanceID)
	if instance == nil {
		return nil, nil, fmt.Errorf("instance %d not found for creation chain", createdInstanceID)
	}

	mandatory := h.catalog.GetMandatoryOutboundAssociations(instance.ClassKey)
	if len(mandatory) == 0 {
		return nil, nil, nil
	}

	var cascadedSteps []*SimulationStep
	var allViolations invariants.ViolationErrors

	for _, assocInfo := range mandatory {
		toClassInfo := h.catalog.GetClassInfo(assocInfo.ToClassKey)
		if toClassInfo == nil {
			continue // Target class has no state machine, skip.
		}

		creationEvent, found := h.catalog.GetCreationEvent(assocInfo.ToClassKey)
		if !found {
			return cascadedSteps, allViolations, fmt.Errorf(
				"class %s requires creation via association %s but has no creation transition",
				toClassInfo.Class.Name, assocInfo.Association.Name,
			)
		}

		if h.catalog.IsAssociationClass(assocInfo.ToClassKey) {
			steps, violations, err := h.createMandatoryAssociationClassInstances(
				toClassInfo, creationEvent, assocInfo, createdInstanceID, simState, depth,
			)
			if err != nil {
				return cascadedSteps, allViolations, err
			}
			cascadedSteps = append(cascadedSteps, steps...)
			allViolations = append(allViolations, violations...)
			continue
		}

		// Create MinTo instances of the target class.
		for range assocInfo.MinTo {
			step, stepViolations, err := h.createMandatoryInstance(
				toClassInfo, creationEvent, assocInfo, createdInstanceID, simState, depth,
			)
			if err != nil {
				return cascadedSteps, allViolations, err
			}
			cascadedSteps = append(cascadedSteps, step)
			allViolations = append(allViolations, stepViolations...)
		}
	}

	return cascadedSteps, allViolations, nil
}

func (h *CreationChainHandler) createMandatoryInstance(
	toClassInfo *ClassInfo,
	creationEvent *model_state.Event,
	assocInfo AssociationInfo,
	createdInstanceID state.InstanceID,
	simState *state.SimulationState,
	depth int,
) (*SimulationStep, invariants.ViolationErrors, error) {
	params, err := actions.SampleEventPayload(*creationEvent, nil, h.paramBinder, nil, h.rng)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"creation chain: event %s parameter sampling: %w",
			creationEvent.Name, err,
		)
	}

	assocKey := assocInfo.Association.Key
	result, err := h.actionExecutor.ExecuteTransition(
		toClassInfo.Class,
		*creationEvent,
		nil, // nil instance = creation
		params,
		&assocKey,
		&createdInstanceID,
		nil,
	)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"creation chain: failed to create %s via %s: %w",
			toClassInfo.Class.Name, assocInfo.Association.Name, err,
		)
	}

	step := &SimulationStep{
		Kind:             StepKindCreation,
		ClassKey:         toClassInfo.ClassKey,
		ClassName:        toClassInfo.Class.Name,
		EventKey:         creationEvent.Key,
		EventName:        creationEvent.Name,
		InstanceID:       result.InstanceID,
		ToState:          result.ToState,
		Parameters:       params,
		TransitionResult: result,
		Violations:       result.Violations,
	}

	// Execute entry actions on the new instance.
	if !result.WasDeletion && result.ToState != "" {
		toStateKey := stateNameToKey(result.ToState, toClassInfo.Class)
		if toStateKey != nil {
			newInstance := simState.GetInstance(result.InstanceID)
			if newInstance != nil {
				entryKeys, entryViolations, err := h.stateActionExec.ExecuteEntryActions(
					toClassInfo.Class, *toStateKey, newInstance,
				)
				if err != nil {
					return nil, nil, fmt.Errorf("creation chain entry actions error: %w", err)
				}
				step.ExecutedActionKeys = append(step.ExecutedActionKeys, entryKeys...)
				step.Violations = append(step.Violations, entryViolations...)
			}
		}
	}

	// Recursively handle creation chain for the new instance.
	childSteps, childViolations, err := h.HandleCreationChain(result.InstanceID, simState, depth+1)
	if err != nil {
		return nil, nil, err
	}
	step.CascadedSteps = childSteps
	step.Violations = append(step.Violations, childViolations...)

	return step, step.Violations, nil
}

func (h *CreationChainHandler) createMandatoryAssociationClassInstances(
	acClassInfo *ClassInfo,
	creationEvent *model_state.Event,
	assocInfo AssociationInfo,
	fromInstanceID state.InstanceID,
	simState *state.SimulationState,
	depth int,
) ([]*SimulationStep, invariants.ViolationErrors, error) {
	acMeta := h.catalog.LookupAssociationClass(acClassInfo.ClassKey)
	if acMeta == nil {
		return nil, nil, fmt.Errorf("association class %s: missing metadata", acClassInfo.Class.Name)
	}

	toInstances := h.activeToEndpointInstances(simState, acMeta.ToClassKey)
	if len(toInstances) == 0 {
		return nil, nil, nil
	}

	var cascadedSteps []*SimulationStep
	var allViolations invariants.ViolationErrors
	fromLegKey := acMeta.FromLegAssocKey

	for range assocInfo.MinTo {
		toID := toInstances[0].ID
		step, stepViolations, err := h.createAssociationClassInstance(
			acClassInfo, creationEvent, fromLegKey, fromInstanceID, toID, simState, depth,
		)
		if err != nil {
			return cascadedSteps, allViolations, err
		}
		cascadedSteps = append(cascadedSteps, step)
		allViolations = append(allViolations, stepViolations...)
	}

	return cascadedSteps, allViolations, nil
}

func (h *CreationChainHandler) createAssociationClassInstance(
	acClassInfo *ClassInfo,
	creationEvent *model_state.Event,
	fromLegKey identity.Key,
	fromInstanceID state.InstanceID,
	toInstanceID state.InstanceID,
	simState *state.SimulationState,
	depth int,
) (*SimulationStep, invariants.ViolationErrors, error) {
	params, err := actions.SampleEventPayload(*creationEvent, nil, h.paramBinder, nil, h.rng)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"creation chain: event %s parameter sampling: %w",
			creationEvent.Name, err,
		)
	}

	result, err := h.actionExecutor.ExecuteTransition(
		acClassInfo.Class,
		*creationEvent,
		nil,
		params,
		&fromLegKey,
		&fromInstanceID,
		&toInstanceID,
	)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"creation chain: failed to create %s: %w",
			acClassInfo.Class.Name, err,
		)
	}

	step := &SimulationStep{
		Kind:             StepKindCreation,
		ClassKey:         acClassInfo.ClassKey,
		ClassName:        acClassInfo.Class.Name,
		EventKey:         creationEvent.Key,
		EventName:        creationEvent.Name,
		InstanceID:       result.InstanceID,
		ToState:          result.ToState,
		Parameters:       params,
		TransitionResult: result,
		Violations:       result.Violations,
	}

	if !result.WasDeletion && result.ToState != "" {
		toStateKey := stateNameToKey(result.ToState, acClassInfo.Class)
		if toStateKey != nil {
			newInstance := simState.GetInstance(result.InstanceID)
			if newInstance != nil {
				entryKeys, entryViolations, err := h.stateActionExec.ExecuteEntryActions(
					acClassInfo.Class, *toStateKey, newInstance,
				)
				if err != nil {
					return nil, nil, fmt.Errorf("creation chain entry actions error: %w", err)
				}
				step.ExecutedActionKeys = append(step.ExecutedActionKeys, entryKeys...)
				step.Violations = append(step.Violations, entryViolations...)
			}
		}
	}

	childSteps, childViolations, err := h.HandleCreationChain(result.InstanceID, simState, depth+1)
	if err != nil {
		return nil, nil, err
	}
	step.CascadedSteps = childSteps
	step.Violations = append(step.Violations, childViolations...)

	return step, step.Violations, nil
}

func (h *CreationChainHandler) activeToEndpointInstances(
	simState *state.SimulationState,
	classKey identity.Key,
) []*state.ClassInstance {
	instances := simState.InstancesByClass(classKey)
	var active []*state.ClassInstance
	for _, inst := range instances {
		if !IsActiveAssociationClassInstance(h.catalog, inst.ClassKey, getInstanceStateName(inst)) {
			continue
		}
		active = append(active, inst)
	}
	return active
}

// stateNameToKey looks up a state name in the class and returns its key.
func stateNameToKey(stateName string, class model_class.Class) *identity.Key {
	for _, s := range class.States {
		if s.Name == stateName {
			key := s.Key
			return &key
		}
	}
	return nil
}

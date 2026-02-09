package engine

import (
	"fmt"
	"math/rand"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
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
) ([]*SimulationStep, invariants.ViolationList, error) {
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
	var allViolations invariants.ViolationList

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

		// Create MinTo instances of the target class.
		for i := uint(0); i < assocInfo.MinTo; i++ {
			params := h.paramBinder.GenerateRandomParameters(creationEvent.Parameters, h.rng)

			assocKey := assocInfo.Association.Key
			result, err := h.actionExecutor.ExecuteTransition(
				toClassInfo.Class,
				*creationEvent,
				nil, // nil instance = creation
				params,
				&assocKey,
				&createdInstanceID,
			)
			if err != nil {
				return cascadedSteps, allViolations, fmt.Errorf(
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
						entryViolations, err := h.stateActionExec.ExecuteEntryActions(
							toClassInfo.Class, *toStateKey, newInstance,
						)
						if err != nil {
							return cascadedSteps, allViolations, fmt.Errorf(
								"creation chain entry actions error: %w", err,
							)
						}
						step.Violations = append(step.Violations, entryViolations...)
					}
				}
			}

			// Recursively handle creation chain for the new instance.
			childSteps, childViolations, err := h.HandleCreationChain(
				result.InstanceID, simState, depth+1,
			)
			if err != nil {
				return cascadedSteps, allViolations, err
			}
			step.CascadedSteps = childSteps
			step.Violations = append(step.Violations, childViolations...)

			cascadedSteps = append(cascadedSteps, step)
			allViolations = append(allViolations, step.Violations...)
		}
	}

	return cascadedSteps, allViolations, nil
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

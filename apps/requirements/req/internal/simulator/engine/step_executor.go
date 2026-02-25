package engine

import (
	"fmt"
	"math/rand"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/actions"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/invariants"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
)

// StepExecutor executes a single simulation step end-to-end.
type StepExecutor struct {
	actionExecutor  *actions.ActionExecutor
	stateActionExec *StateActionExecutor
	chainHandler    *CreationChainHandler
	multChecker     *MultiplicityChecker
	paramBinder     *actions.ParameterBinder
	catalog         *ClassCatalog
	rng             *rand.Rand
}

// NewStepExecutor creates a new step executor.
func NewStepExecutor(
	actionExecutor *actions.ActionExecutor,
	stateActionExec *StateActionExecutor,
	chainHandler *CreationChainHandler,
	multChecker *MultiplicityChecker,
	paramBinder *actions.ParameterBinder,
	catalog *ClassCatalog,
	rng *rand.Rand,
) *StepExecutor {
	return &StepExecutor{
		actionExecutor:  actionExecutor,
		stateActionExec: stateActionExec,
		chainHandler:    chainHandler,
		multChecker:     multChecker,
		paramBinder:     paramBinder,
		catalog:         catalog,
		rng:             rng,
	}
}

// Execute runs a single simulation step for the given pending action.
func (e *StepExecutor) Execute(
	pending *PendingAction,
	simState *state.SimulationState,
	stepNumber int,
) (*SimulationStep, error) {
	// Handle "do" actions separately — they don't involve state transitions.
	if pending.IsDo {
		return e.executeDo(pending, simState, stepNumber)
	}

	return e.executeTransition(pending, simState, stepNumber)
}

// executeDo handles a "do" state action — runs the action on the instance.
func (e *StepExecutor) executeDo(
	pending *PendingAction,
	simState *state.SimulationState,
	stepNumber int,
) (*SimulationStep, error) {
	step := &SimulationStep{
		StepNumber: stepNumber,
		Kind:       StepKindNormal,
		ClassKey:   pending.Class.ClassKey,
		ClassName:  pending.Class.Class.Name,
		InstanceID: pending.Instance.ID,
	}

	if pending.DoAction == nil {
		return nil, fmt.Errorf("do action is nil")
	}

	result, err := e.actionExecutor.ExecuteAction(*pending.DoAction, pending.Instance, nil)
	if err != nil {
		return nil, fmt.Errorf("do action %s error: %w", pending.DoAction.Name, err)
	}

	step.DoActionResult = result
	step.Violations = append(step.Violations, result.Violations...)
	return step, nil
}

// executeTransition handles event-triggered transitions (creation, normal, deletion).
func (e *StepExecutor) executeTransition(
	pending *PendingAction,
	simState *state.SimulationState,
	stepNumber int,
) (*SimulationStep, error) {
	if pending.Event == nil {
		return nil, fmt.Errorf("event is nil for non-do action")
	}

	step := &SimulationStep{
		StepNumber: stepNumber,
		ClassKey:   pending.Class.ClassKey,
		ClassName:  pending.Class.Class.Name,
		EventKey:   pending.Event.Key,
		EventName:  pending.Event.Name,
	}

	// 1. Generate event parameters.
	params := e.paramBinder.GenerateRandomParameters(pending.Event.Parameters, e.rng)
	step.Parameters = params

	// 2. Execute exit StateActions (if not creation).
	if pending.Instance != nil {
		fromStateKey := getCurrentStateKey(pending.Instance, pending.Class)
		if fromStateKey != nil {
			exitViolations, err := e.stateActionExec.ExecuteExitActions(
				pending.Class.Class, *fromStateKey, pending.Instance,
			)
			if err != nil {
				return nil, fmt.Errorf("exit actions error: %w", err)
			}
			step.Violations = append(step.Violations, exitViolations...)
		}
	}

	// 3. Execute the transition.
	result, err := e.actionExecutor.ExecuteTransition(
		pending.Class.Class, *pending.Event, pending.Instance,
		params, nil, nil, // No source association for top-level steps.
	)
	if err != nil {
		return nil, fmt.Errorf("transition error: %w", err)
	}

	step.TransitionResult = result
	step.InstanceID = result.InstanceID
	step.FromState = result.FromState
	step.ToState = result.ToState
	step.Violations = append(step.Violations, result.Violations...)

	if result.WasCreation {
		step.Kind = StepKindCreation
	} else if result.WasDeletion {
		step.Kind = StepKindDeletion
	} else {
		step.Kind = StepKindNormal
	}

	// 4. Execute entry StateActions (if not deletion).
	if !result.WasDeletion && result.ToState != "" {
		toStateKey := stateNameToKey(result.ToState, pending.Class.Class)
		if toStateKey != nil {
			entryInstance := simState.GetInstance(result.InstanceID)
			if entryInstance != nil {
				entryViolations, err := e.stateActionExec.ExecuteEntryActions(
					pending.Class.Class, *toStateKey, entryInstance,
				)
				if err != nil {
					return nil, fmt.Errorf("entry actions error: %w", err)
				}
				step.Violations = append(step.Violations, entryViolations...)
			}
		}
	}

	// 5. Handle creation chains (if creation).
	if result.WasCreation {
		cascadedSteps, cascadeViolations, err := e.chainHandler.HandleCreationChain(
			result.InstanceID, simState, 0,
		)
		if err != nil {
			return nil, fmt.Errorf("creation chain error: %w", err)
		}
		step.CascadedSteps = cascadedSteps
		step.Violations = append(step.Violations, cascadeViolations...)
	}

	// 6. Check multiplicity constraints.
	instance := simState.GetInstance(result.InstanceID)
	if instance != nil {
		multViolations := e.multChecker.CheckInstance(instance, simState)
		for _, mv := range multViolations {
			step.Violations = append(step.Violations, invariants.NewMultiplicityViolation(
				mv.InstanceID,
				mv.ClassKey,
				mv.AssociationName,
				mv.Direction,
				mv.ActualCount,
				mv.RequiredMin,
				mv.RequiredMax,
				mv.Message,
			))
		}
	}

	return step, nil
}

// getCurrentStateKey looks up the instance's current state key from its _state attribute.
func getCurrentStateKey(instance *state.ClassInstance, classInfo *ClassInfo) *identity.Key {
	stateName := getInstanceStateName(instance)
	if stateName == "" {
		return nil
	}
	return stateNameToKey(stateName, classInfo.Class)
}

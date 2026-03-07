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
		return e.executeDo(pending, stepNumber)
	}

	return e.executeTransition(pending, simState, stepNumber)
}

// executeDo handles a "do" state action — runs the action on the instance.
func (e *StepExecutor) executeDo(
	pending *PendingAction,
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
	if err := e.executeExitActions(pending, step); err != nil {
		return nil, err
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

	switch {
	case result.WasCreation:
		step.Kind = StepKindCreation
	case result.WasDeletion:
		step.Kind = StepKindDeletion
	default:
		step.Kind = StepKindNormal
	}

	// 4. Execute entry StateActions (if not deletion).
	if err := e.executeEntryActions(pending, result, simState, step); err != nil {
		return nil, err
	}

	// 5. Handle creation chains (if creation).
	if err := e.handleCreationChain(result, simState, step); err != nil {
		return nil, err
	}

	// 6. Check multiplicity constraints.
	e.checkMultiplicityConstraints(result, simState, step)

	return step, nil
}

// executeExitActions runs exit state actions for a non-creation transition.
func (e *StepExecutor) executeExitActions(pending *PendingAction, step *SimulationStep) error {
	if pending.Instance == nil {
		return nil
	}
	fromStateKey := getCurrentStateKey(pending.Instance, pending.Class)
	if fromStateKey == nil {
		return nil
	}
	exitViolations, err := e.stateActionExec.ExecuteExitActions(
		pending.Class.Class, *fromStateKey, pending.Instance,
	)
	if err != nil {
		return fmt.Errorf("exit actions error: %w", err)
	}
	step.Violations = append(step.Violations, exitViolations...)
	return nil
}

// executeEntryActions runs entry state actions after a non-deletion transition.
func (e *StepExecutor) executeEntryActions(
	pending *PendingAction,
	result *actions.TransitionResult,
	simState *state.SimulationState,
	step *SimulationStep,
) error {
	if result.WasDeletion || result.ToState == "" {
		return nil
	}
	toStateKey := stateNameToKey(result.ToState, pending.Class.Class)
	if toStateKey == nil {
		return nil
	}
	entryInstance := simState.GetInstance(result.InstanceID)
	if entryInstance == nil {
		return nil
	}
	entryViolations, err := e.stateActionExec.ExecuteEntryActions(
		pending.Class.Class, *toStateKey, entryInstance,
	)
	if err != nil {
		return fmt.Errorf("entry actions error: %w", err)
	}
	step.Violations = append(step.Violations, entryViolations...)
	return nil
}

// handleCreationChain handles cascaded creation steps for a creation transition.
func (e *StepExecutor) handleCreationChain(
	result *actions.TransitionResult,
	simState *state.SimulationState,
	step *SimulationStep,
) error {
	if !result.WasCreation {
		return nil
	}
	cascadedSteps, cascadeViolations, err := e.chainHandler.HandleCreationChain(
		result.InstanceID, simState, 0,
	)
	if err != nil {
		return fmt.Errorf("creation chain error: %w", err)
	}
	step.CascadedSteps = cascadedSteps
	step.Violations = append(step.Violations, cascadeViolations...)
	return nil
}

// checkMultiplicityConstraints checks multiplicity constraints on the result instance.
func (e *StepExecutor) checkMultiplicityConstraints(
	result *actions.TransitionResult,
	simState *state.SimulationState,
	step *SimulationStep,
) {
	instance := simState.GetInstance(result.InstanceID)
	if instance == nil {
		return
	}
	multViolations := e.multChecker.CheckInstance(instance, simState)
	for _, mv := range multViolations {
		step.Violations = append(step.Violations, invariants.NewMultiplicityViolation(
			invariants.MultiplicityViolationParams{
				InstanceID:      mv.InstanceID,
				ClassKey:        mv.ClassKey,
				AssociationName: mv.AssociationName,
				Direction:       mv.Direction,
				ActualCount:     mv.ActualCount,
				RequiredMin:     mv.RequiredMin,
				RequiredMax:     mv.RequiredMax,
				Message:         mv.Message,
			},
		))
	}
}

// getCurrentStateKey looks up the instance's current state key from its _state attribute.
func getCurrentStateKey(instance *state.ClassInstance, classInfo *ClassInfo) *identity.Key {
	stateName := getInstanceStateName(instance)
	if stateName == "" {
		return nil
	}
	return stateNameToKey(stateName, classInfo.Class)
}

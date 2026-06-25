package engine

import (
	"errors"
	"fmt"
	"math/rand"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/actions"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
)

// StepParameterGenerator supplies surface and nested event parameter values.
type StepParameterGenerator struct {
	Binder  *actions.ParameterBinder
	Sampler *actions.ParameterSampler
}

// NewStepParameterGenerator wires type-only and requires-aware parameter generation.
func NewStepParameterGenerator(binder *actions.ParameterBinder, sampler *actions.ParameterSampler) *StepParameterGenerator {
	return &StepParameterGenerator{Binder: binder, Sampler: sampler}
}

// StepExecutor executes a single simulation step end-to-end.
type StepExecutor struct {
	actionExecutor  *actions.ActionExecutor
	stateActionExec *StateActionExecutor
	chainHandler    *CreationChainHandler
	paramGen        *StepParameterGenerator
	catalog         *ClassCatalog
	rng             *rand.Rand
}

// NewStepExecutor creates a new step executor.
func NewStepExecutor(
	actionExecutor *actions.ActionExecutor,
	stateActionExec *StateActionExecutor,
	chainHandler *CreationChainHandler,
	paramGen *StepParameterGenerator,
	catalog *ClassCatalog,
	rng *rand.Rand,
) *StepExecutor {
	return &StepExecutor{
		actionExecutor:  actionExecutor,
		stateActionExec: stateActionExec,
		chainHandler:    chainHandler,
		paramGen:        paramGen,
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
	if pending.IsQuery {
		return e.executeQuery(pending, stepNumber)
	}

	// Handle "do" actions separately — they don't involve state transitions.
	if pending.IsDo {
		return e.executeDo(pending, stepNumber)
	}

	return e.executeTransition(pending, simState, stepNumber)
}

// executeQuery runs a read-only query on an existing instance.
func (e *StepExecutor) executeQuery(
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

	if pending.Query == nil {
		return nil, fmt.Errorf("query is nil")
	}

	params, err := e.sampleQueryParameters(pending)
	if err != nil {
		return nil, fmt.Errorf("query %s parameter sampling: %w", pending.Query.Name, err)
	}
	step.Parameters = params

	result, err := e.actionExecutor.ExecuteQuery(*pending.Query, pending.Instance, params)
	if err != nil {
		return nil, fmt.Errorf("query %s error: %w", pending.Query.Name, err)
	}

	step.QueryKey = pending.Query.Key
	step.QueryName = pending.Query.Name
	step.QueryResult = result
	step.Violations = append(step.Violations, result.Violations...)
	return step, nil
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

	step.ExecutedActionKeys = append(step.ExecutedActionKeys, pending.DoAction.Key)
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

	// 1. Generate event parameters (surface steps sample from transition action requires).
	params, err := e.sampleEventParameters(pending)
	if err != nil {
		return nil, fmt.Errorf("event %s parameter sampling: %w", pending.Event.Name, err)
	}
	step.Parameters = params

	// 2. Execute exit StateActions (if not creation).
	if err := e.executeExitActions(pending, step); err != nil {
		return nil, err
	}

	// 3. Execute the transition.
	result, err := e.actionExecutor.ExecuteTransition(
		pending.Class.Class, *pending.Event, pending.Instance,
		params, actions.CreationLinkSource{
			SourceAssocKey: pending.SourceAssocKey,
			SourceID:       pending.SourceInstanceID,
		}, pending.TargetInstanceID,
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

	return step, nil
}

// sampleQueryParameters generates parameters for a query step.
func (e *StepExecutor) sampleQueryParameters(pending *PendingAction) (map[string]object.Object, error) {
	if pending.Query == nil || len(pending.Query.Parameters) == 0 {
		return map[string]object.Object{}, nil
	}
	if e.paramGen != nil && e.paramGen.Sampler != nil {
		return e.paramGen.Sampler.SampleQueryFromRequires(*pending.Query, e.rng)
	}
	binder := actions.NewParameterBinder()
	if e.paramGen != nil && e.paramGen.Binder != nil {
		binder = e.paramGen.Binder
	}
	return binder.GenerateRandomParameters(pending.Query.Parameters, e.rng), nil
}

// sampleEventParameters generates parameters for a top-level transition event.
func (e *StepExecutor) sampleEventParameters(pending *PendingAction) (map[string]object.Object, error) {
	instanceState := ""
	if pending.Instance != nil {
		instanceState = getInstanceStateName(pending.Instance)
	}

	action, found := e.catalog.GetActionForEvent(pending.Class.ClassKey, pending.Event.Key, instanceState)
	var actionPtr *model_state.Action
	if found && action != nil {
		actionPtr = action
	}

	params, err := sampleEventPayload(pending.Event, actionPtr, e.paramGen, e.rng)
	if err != nil {
		var unsupported *actions.UnsupportedRequiresSamplingError
		if errors.As(err, &unsupported) {
			unsupported.ClassName = pending.Class.Class.Name
		}
		return nil, err
	}
	return params, nil
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
	exitKeys, exitViolations, err := e.stateActionExec.ExecuteExitActions(
		pending.Class.Class, *fromStateKey, pending.Instance,
	)
	if err != nil {
		return fmt.Errorf("exit actions error: %w", err)
	}
	step.ExecutedActionKeys = append(step.ExecutedActionKeys, exitKeys...)
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
	entryKeys, entryViolations, err := e.stateActionExec.ExecuteEntryActions(
		pending.Class.Class, *toStateKey, entryInstance,
	)
	if err != nil {
		return fmt.Errorf("entry actions error: %w", err)
	}
	step.ExecutedActionKeys = append(step.ExecutedActionKeys, entryKeys...)
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

// getCurrentStateKey looks up the instance's current state key from its _state attribute.
func getCurrentStateKey(instance *state.ClassInstance, classInfo *ClassInfo) *identity.Key {
	stateName := getInstanceStateName(instance)
	if stateName == "" {
		return nil
	}
	return stateNameToKey(stateName, classInfo.Class)
}

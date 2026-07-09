package engine

import (
	"fmt"
	"math/rand"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/actions"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/invariants"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
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
// parentParams are the parent creation event parameters (e.g. Amounts) used to
// drive association-class row construction when present.
func (h *CreationChainHandler) HandleCreationChain(
	createdInstanceID state.InstanceID,
	simState *state.SimulationState,
	depth int,
	parentParams map[string]object.Object,
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
		cascadeClassKey := CreationCascadeClassKey(assocInfo)
		toClassInfo := h.catalog.GetClassInfo(cascadeClassKey)
		if toClassInfo == nil {
			continue // Target class has no state machine, skip.
		}

		creationEvent, found := h.catalog.GetCreationEvent(cascadeClassKey)
		if !found {
			return cascadedSteps, allViolations, fmt.Errorf(
				"class %s requires creation via association %s but has no creation transition",
				toClassInfo.Class.Name, assocInfo.Association.Name,
			)
		}

		if assocInfo.Association.AssociationClassKey != nil {
			steps, violations, err := h.createMandatoryAssociationClassInstances(acCascadeInput{
				acClassInfo:    toClassInfo,
				creationEvent:  creationEvent,
				assocInfo:      assocInfo,
				fromInstanceID: createdInstanceID,
				simState:       simState,
				depth:          depth,
				parentParams:   parentParams,
			})
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
	params, err := h.sampleCreationEventParams(toClassInfo, creationEvent)
	if err != nil {
		return nil, nil, err
	}

	assocKey := assocInfo.Association.Key
	result, err := h.actionExecutor.ExecuteTransition(
		toClassInfo.Class,
		*creationEvent,
		nil, // nil instance = creation
		params,
		actions.CreationLinkSource{SourceAssocKey: &assocKey, SourceID: &createdInstanceID}, nil,
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
	if !result.WasDestroy && result.ToState != "" {
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
	childSteps, childViolations, err := h.HandleCreationChain(result.InstanceID, simState, depth+1, nil)
	if err != nil {
		return nil, nil, err
	}
	step.CascadedSteps = childSteps
	step.Violations = append(step.Violations, childViolations...)

	return step, step.Violations, nil
}

// acCascadeInput groups association-class cascade construction inputs.
type acCascadeInput struct {
	acClassInfo    *ClassInfo
	creationEvent  *model_state.Event
	assocInfo      AssociationInfo
	fromInstanceID state.InstanceID
	simState       *state.SimulationState
	depth          int
	parentParams   map[string]object.Object
}

func (h *CreationChainHandler) createMandatoryAssociationClassInstances(
	in acCascadeInput,
) ([]*SimulationStep, invariants.ViolationErrors, error) {
	acMeta := h.catalog.LookupAssociationClass(in.acClassInfo.ClassKey)
	if acMeta == nil {
		return nil, nil, fmt.Errorf("association class %s: missing metadata", in.acClassInfo.Class.Name)
	}

	toClassInfo := h.catalog.GetClassInfo(acMeta.ToClassKey)
	if toClassInfo == nil {
		return nil, nil, fmt.Errorf(
			"association class %s: to endpoint class not in catalog",
			in.acClassInfo.Class.Name,
		)
	}
	toCreationEvent, found := h.catalog.GetCreationEvent(acMeta.ToClassKey)
	if !found {
		return nil, nil, fmt.Errorf(
			"class %s requires association-class rows via %s but to endpoint %s has no creation transition",
			in.acClassInfo.Class.Name, in.assocInfo.Association.Name, toClassInfo.Class.Name,
		)
	}

	var cascadedSteps []*SimulationStep
	var allViolations invariants.ViolationErrors
	hostAssocKey := acMeta.HostAssociation.Key

	// Parent Amounts (set of [account, amount] records) drive AC rows and amounts when present.
	if amountRows := parentAmountRows(in.parentParams); len(amountRows) > 0 {
		for _, row := range amountRows {
			toID, ok := instanceIDForRecord(in.simState, acMeta.ToClassKey, row.account)
			if !ok {
				return cascadedSteps, allViolations, fmt.Errorf(
					"creation chain: Amounts account does not match a live %s instance",
					toClassInfo.Class.Name,
				)
			}
			params := map[string]object.Object{"Amount": row.amount}
			step, stepViolations, err := h.createAssociationClassInstance(acCreateInput{
				acClassInfo:   in.acClassInfo,
				creationEvent: in.creationEvent,
				hostAssocKey:  hostAssocKey,
				endpoints:     instanceEndpointIDs{FromInstanceID: in.fromInstanceID, ToInstanceID: toID},
				simState:      in.simState,
				depth:         in.depth,
				paramOverride: params,
			})
			if err != nil {
				return cascadedSteps, allViolations, err
			}
			cascadedSteps = append(cascadedSteps, step)
			allViolations = append(allViolations, stepViolations...)
		}
		return cascadedSteps, allViolations, nil
	}

	toSteps, toViolations, toInstances, err := h.ensureActiveToEndpointInstances(
		toClassInfo, toCreationEvent, acMeta.ToClassKey, in.assocInfo.MinTo, in.simState, in.depth,
	)
	if err != nil {
		return cascadedSteps, allViolations, err
	}
	cascadedSteps = append(cascadedSteps, toSteps...)
	allViolations = append(allViolations, toViolations...)
	if len(toInstances) == 0 {
		return cascadedSteps, allViolations, nil
	}

	for range in.assocInfo.MinTo {
		toID := toInstances[0].ID
		step, stepViolations, err := h.createAssociationClassInstance(acCreateInput{
			acClassInfo:   in.acClassInfo,
			creationEvent: in.creationEvent,
			hostAssocKey:  hostAssocKey,
			endpoints:     instanceEndpointIDs{FromInstanceID: in.fromInstanceID, ToInstanceID: toID},
			simState:      in.simState,
			depth:         in.depth,
			paramOverride: nil,
		})
		if err != nil {
			return cascadedSteps, allViolations, err
		}
		cascadedSteps = append(cascadedSteps, step)
		allViolations = append(allViolations, stepViolations...)
	}

	return cascadedSteps, allViolations, nil
}

// amountRow is one parent Amounts entry: to-endpoint account record + amount value.
type amountRow struct {
	account *object.Record
	amount  object.Object
}

// parentAmountRows extracts Amounts as a set of records with account and amount fields.
func parentAmountRows(parentParams map[string]object.Object) []amountRow {
	if len(parentParams) == 0 {
		return nil
	}
	raw, ok := parentParams["Amounts"]
	if !ok || raw == nil {
		return nil
	}
	set, ok := raw.(*object.Set)
	if !ok {
		return nil
	}
	var rows []amountRow
	for _, elem := range set.Elements() {
		rec, ok := elem.(*object.Record)
		if !ok {
			continue
		}
		account, ok := rec.Get("account").(*object.Record)
		if !ok || account == nil {
			continue
		}
		amount := rec.Get("amount")
		if amount == nil {
			continue
		}
		rows = append(rows, amountRow{account: account, amount: amount})
	}
	return rows
}

func instanceIDForRecord(
	simState *state.SimulationState,
	classKey identity.Key,
	attrs *object.Record,
) (state.InstanceID, bool) {
	for _, inst := range simState.InstancesByClass(classKey) {
		if inst.Attributes == attrs || inst.Attributes.Equals(attrs) {
			return inst.ID, true
		}
	}
	return 0, false
}

func (h *CreationChainHandler) ensureActiveToEndpointInstances(
	toClassInfo *ClassInfo,
	creationEvent *model_state.Event,
	toClassKey identity.Key,
	minCount uint,
	simState *state.SimulationState,
	depth int,
) ([]*SimulationStep, invariants.ViolationErrors, []*state.ClassInstance, error) {
	var steps []*SimulationStep
	var violations invariants.ViolationErrors

	active := h.activeToEndpointInstances(simState, toClassKey)
	for uint(len(active)) < minCount {
		step, stepViolations, err := h.createPlainEndpointInstance(
			toClassInfo, creationEvent, simState, depth,
		)
		if err != nil {
			return steps, violations, active, err
		}
		steps = append(steps, step)
		violations = append(violations, stepViolations...)
		active = h.activeToEndpointInstances(simState, toClassKey)
	}
	return steps, violations, active, nil
}

func (h *CreationChainHandler) createPlainEndpointInstance(
	toClassInfo *ClassInfo,
	creationEvent *model_state.Event,
	simState *state.SimulationState,
	depth int,
) (*SimulationStep, invariants.ViolationErrors, error) {
	params, err := h.sampleCreationEventParams(toClassInfo, creationEvent)
	if err != nil {
		return nil, nil, err
	}

	result, err := h.actionExecutor.ExecuteTransition(
		toClassInfo.Class,
		*creationEvent,
		nil,
		params,
		actions.CreationLinkSource{}, nil,
	)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"creation chain: failed to create %s: %w",
			toClassInfo.Class.Name, err,
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

	if !result.WasDestroy && result.ToState != "" {
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

	childSteps, childViolations, err := h.HandleCreationChain(result.InstanceID, simState, depth+1, nil)
	if err != nil {
		return nil, nil, err
	}
	step.CascadedSteps = childSteps
	step.Violations = append(step.Violations, childViolations...)

	return step, step.Violations, nil
}

// sampleCreationEventParams samples creation-event parameters using the creation
// transition action when present so typed action parameters (spans, etc.) apply.
func (h *CreationChainHandler) sampleCreationEventParams(
	classInfo *ClassInfo,
	creationEvent *model_state.Event,
) (map[string]object.Object, error) {
	var actionPtr *model_state.Action
	if action, found := h.catalog.GetActionForEvent(classInfo.ClassKey, creationEvent.Key, ""); found {
		actionPtr = action
	}
	params, err := actions.SampleEventPayload(*creationEvent, actionPtr, h.paramBinder, nil, h.rng)
	if err != nil {
		return nil, fmt.Errorf(
			"creation chain: event %s parameter sampling: %w",
			creationEvent.Name, err,
		)
	}
	return params, nil
}

// instanceEndpointIDs holds the from and to endpoint instances for association-class materialization.
type instanceEndpointIDs struct {
	FromInstanceID state.InstanceID
	ToInstanceID   state.InstanceID
}

// acCreateInput groups one association-class instance creation.
type acCreateInput struct {
	acClassInfo   *ClassInfo
	creationEvent *model_state.Event
	hostAssocKey  identity.Key
	endpoints     instanceEndpointIDs
	simState      *state.SimulationState
	depth         int
	paramOverride map[string]object.Object
}

func (h *CreationChainHandler) createAssociationClassInstance(
	in acCreateInput,
) (*SimulationStep, invariants.ViolationErrors, error) {
	params := in.paramOverride
	if len(params) == 0 {
		var err error
		params, err = h.sampleCreationEventParams(in.acClassInfo, in.creationEvent)
		if err != nil {
			return nil, nil, err
		}
	}

	result, err := h.actionExecutor.ExecuteTransition(
		in.acClassInfo.Class,
		*in.creationEvent,
		nil,
		params,
		actions.CreationLinkSource{SourceAssocKey: &in.hostAssocKey, SourceID: &in.endpoints.FromInstanceID},
		&in.endpoints.ToInstanceID,
	)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"creation chain: failed to create %s: %w",
			in.acClassInfo.Class.Name, err,
		)
	}

	step := &SimulationStep{
		Kind:             StepKindCreation,
		ClassKey:         in.acClassInfo.ClassKey,
		ClassName:        in.acClassInfo.Class.Name,
		EventKey:         in.creationEvent.Key,
		EventName:        in.creationEvent.Name,
		InstanceID:       result.InstanceID,
		ToState:          result.ToState,
		Parameters:       params,
		TransitionResult: result,
		Violations:       result.Violations,
	}

	if !result.WasDestroy && result.ToState != "" {
		toStateKey := stateNameToKey(result.ToState, in.acClassInfo.Class)
		if toStateKey != nil {
			newInstance := in.simState.GetInstance(result.InstanceID)
			if newInstance != nil {
				entryKeys, entryViolations, err := h.stateActionExec.ExecuteEntryActions(
					in.acClassInfo.Class, *toStateKey, newInstance,
				)
				if err != nil {
					return nil, nil, fmt.Errorf("creation chain entry actions error: %w", err)
				}
				step.ExecutedActionKeys = append(step.ExecutedActionKeys, entryKeys...)
				step.Violations = append(step.Violations, entryViolations...)
			}
		}
	}

	childSteps, childViolations, err := h.HandleCreationChain(result.InstanceID, in.simState, in.depth+1, nil)
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
	return simState.InstancesByClass(classKey)
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

package engine

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/actions"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
)

func (e *StepExecutor) appendAssociationPeerCascades(
	step *SimulationStep,
	transitionResult *actions.TransitionResult,
	simState *state.SimulationState,
) error {
	if step == nil || transitionResult == nil || transitionResult.ActionResult == nil {
		return nil
	}
	for _, peer := range transitionResult.ActionResult.PeerTransitions {
		cascade, err := e.buildPeerEffectStep(peer, simState, 0)
		if err != nil {
			return err
		}
		step.CascadedSteps = append(step.CascadedSteps, cascade)
		step.Violations = append(step.Violations, cascade.Violations...)
	}
	return nil
}

func (e *StepExecutor) buildPeerEffectStep(
	peer actions.PeerTransitionRecord,
	simState *state.SimulationState,
	depth int,
) (*SimulationStep, error) {
	if depth > maxCascadeDepth {
		return nil, fmt.Errorf("association peer cascade exceeded max depth of %d", maxCascadeDepth)
	}
	if peer.Result == nil {
		return nil, fmt.Errorf("peer transition on class %s missing result", peer.ClassName)
	}

	step := simulationStepFromPeerTransition(peer)

	if peer.Result.ActionResult != nil {
		for _, nested := range peer.Result.ActionResult.PeerTransitions {
			child, err := e.buildPeerEffectStep(nested, simState, depth+1)
			if err != nil {
				return nil, err
			}
			step.CascadedSteps = append(step.CascadedSteps, child)
			step.Violations = append(step.Violations, child.Violations...)
		}
	}

	if peer.Result.WasCreation {
		chainSteps, chainViolations, err := e.chainHandler.HandleCreationChain(peer.Result.InstanceID, simState, depth+1)
		if err != nil {
			return nil, err
		}
		step.CascadedSteps = append(step.CascadedSteps, chainSteps...)
		step.Violations = append(step.Violations, chainViolations...)
	}

	return step, nil
}

func simulationStepFromPeerTransition(peer actions.PeerTransitionRecord) *SimulationStep {
	kind := StepKindNormal
	switch {
	case peer.Result.WasCreation:
		kind = StepKindCreation
	case peer.Result.WasDeletion:
		kind = StepKindDeletion
	}
	return &SimulationStep{
		Kind:             kind,
		ClassKey:         peer.ClassKey,
		ClassName:        peer.ClassName,
		EventKey:         peer.EventKey,
		EventName:        peer.EventName,
		InstanceID:       peer.Result.InstanceID,
		FromState:        peer.Result.FromState,
		ToState:          peer.Result.ToState,
		Parameters:       peer.Parameters,
		TransitionResult: peer.Result,
		Violations:       peer.Result.Violations,
	}
}

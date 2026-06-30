package actions

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
)

func (e *ActionExecutor) tryApplyAssociationStateChangeGuarantee(
	ctx *ExecutionContext,
	instance *state.ClassInstance,
	target string,
	expr me.Expression,
	bindings *evaluator.Bindings,
) (bool, error) {
	if e.peerCatalog == nil {
		return false, nil
	}
	assocKey, assoc, found := e.peerCatalog.OutgoingAssociationByTLAField(instance.ClassKey, target)
	if !found {
		return false, nil
	}

	rhsValue := evaluator.Eval(expr, bindings)
	if rhsValue.IsError() {
		return false, fmt.Errorf("association state_change on %q evaluation error: %s", target, rhsValue.Error.Inspect())
	}
	newSet, ok := evaluator.CoerceToSet(rhsValue.Value)
	if !ok {
		return false, fmt.Errorf("association state_change on %q: expression must evaluate to a set", target)
	}

	removed := associationPeersRemovedFromSet(e.bindingsBuilder.State(), instance.ID, assoc, newSet)
	ctx.SetAssociationRemovedPeers(instance.ID, assocKey, removed)
	return true, nil
}

func associationPeersRemovedFromSet(
	simState *state.SimulationState,
	ownerID state.InstanceID,
	assoc model_class.Association,
	newSet *object.Set,
) []state.InstanceID {
	linked := linkedAssociationPeerEndpoints(simState, ownerID, assoc)
	if len(linked) == 0 {
		return nil
	}
	var removed []state.InstanceID
	for _, peerID := range linked {
		peerInstance := simState.GetInstance(peerID)
		if peerInstance == nil {
			continue
		}
		if newSet.Contains(peerInstance.Attributes) {
			continue
		}
		removed = append(removed, peerID)
	}
	return removed
}

func (e *ActionExecutor) applyAssociationLinkRemovals(ctx *ExecutionContext) {
	simState := e.bindingsBuilder.State()
	for key, peerIDs := range ctx.associationRemovedPeerSets() {
		for _, peerID := range peerIDs {
			if ctx.AssociationDestroyCandidate(key, peerID) {
				continue
			}
			simState.RemoveLink(key.AssocKey, key.OwnerInstanceID, peerID)
		}
	}
}

package actions

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/instance"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

func (e *ActionExecutor) tryApplyAssociationStateChangeGuarantee(
	ctx *ExecutionContext,
	instance *instance.Instance,
	target string,
	expr me.Expression,
	bindings *evaluator.Bindings,
) (bool, error) {
	if e.sch == nil {
		return false, nil
	}
	assocKey, assoc, reverse, found := e.sch.AssociationByNavigableTLAField(instance.GetClassKey(), target)
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

	if err := e.applyAssociationStateChangeSet(associationStateChangeApply{
		ctx:      ctx,
		instance: instance,
		target:   target,
		assocKey: assocKey,
		assoc:    assoc,
		reverse:  reverse,
		newSet:   newSet,
	}); err != nil {
		return true, err
	}
	return true, nil
}

// associationStateChangeApply carries a resolved state_change guarantee ready to apply to live links.
type associationStateChangeApply struct {
	ctx      *ExecutionContext
	instance *instance.Instance
	target   string
	assocKey identity.Key
	assoc    model_class.Association
	reverse  bool
	newSet   *object.Set
}

// applyAssociationStateChangeSet records removed peers and adds missing plain links for the RHS set.
func (e *ActionExecutor) applyAssociationStateChangeSet(work associationStateChangeApply) error {
	simState := e.bindingsBuilder.State()
	ownerID := work.instance.GetID()
	peerClassKey := work.assoc.ToClassKey
	if work.reverse {
		peerClassKey = work.assoc.FromClassKey
	}
	removed := associationPeersRemovedFromSet(simState, ownerID, work.assoc, work.reverse, peerClassKey, work.newSet)
	work.ctx.SetAssociationRemovedPeers(ownerID, work.assocKey, work.reverse, removed)

	// Plain associations also establish missing links from the RHS set. Association-class
	// hosts materialize rows via reify; their endpoint image is derived from those rows.
	if work.assoc.AssociationClassKey != nil {
		return nil
	}
	if err := e.addMissingPlainAssociationLinks(plainAssocLinkWork{
		simState:     simState,
		ownerID:      ownerID,
		assocKey:     work.assocKey,
		assoc:        work.assoc,
		reverse:      work.reverse,
		peerClassKey: peerClassKey,
		newSet:       work.newSet,
	}); err != nil {
		return fmt.Errorf("association state_change on %q: %w", work.target, err)
	}
	return nil
}

// plainAssocLinkWork is the context for establishing plain association links from a state_change RHS set.
type plainAssocLinkWork struct {
	simState     *instance.State
	ownerID      instance.ID
	assocKey     identity.Key
	assoc        model_class.Association
	reverse      bool
	peerClassKey identity.Key
	newSet       *object.Set
}

func associationPeersRemovedFromSet(
	simState *instance.State,
	ownerID instance.ID,
	assoc model_class.Association,
	reverse bool,
	peerClassKey identity.Key,
	newSet *object.Set,
) []instance.ID {
	linked := linkedPeersForDirection(simState, ownerID, assoc, reverse)
	if len(linked) == 0 {
		return nil
	}
	var removed []instance.ID
	for _, peerID := range linked {
		if peerInRHSSet(simState, peerClassKey, peerID, newSet) {
			continue
		}
		removed = append(removed, peerID)
	}
	return removed
}

func peerInRHSSet(
	simState *instance.State,
	peerClassKey identity.Key,
	peerID instance.ID,
	newSet *object.Set,
) bool {
	peerInstance := simState.GetInstance(peerID)
	if peerInstance == nil {
		return false
	}
	if newSet.Contains(peerInstance.GetAttributes()) {
		return true
	}
	for _, elem := range newSet.Elements() {
		if id, ok := resolveToEndpointInstanceID(simState, peerClassKey, elem); ok && id == peerID {
			return true
		}
	}
	return false
}

func linkedPeersForDirection(
	simState *instance.State,
	ownerID instance.ID,
	assoc model_class.Association,
	reverse bool,
) []instance.ID {
	if reverse {
		// Owner is the to-endpoint; peers are from-endpoints.
		return simState.GetLinkedReverse(ownerID, assoc.Key)
	}
	return linkedAssociationPeerEndpoints(simState, ownerID, assoc)
}

// addMissingPlainAssociationLinks links each RHS set element that identifies a live peer.
// Forward: owner is from-end, peers are to-end. Reverse: owner is to-end, peers are from-end.
func (e *ActionExecutor) addMissingPlainAssociationLinks(work plainAssocLinkWork) error {
	linked := make(map[instance.ID]bool)
	for _, peerID := range linkedPeersForDirection(work.simState, work.ownerID, work.assoc, work.reverse) {
		linked[peerID] = true
	}
	for _, elem := range work.newSet.Elements() {
		peerID, ok := resolveToEndpointInstanceID(work.simState, work.peerClassKey, elem)
		if !ok {
			continue
		}
		if linked[peerID] {
			continue
		}
		var err error
		if work.reverse {
			// Peer is the from-endpoint; owner is the to-endpoint.
			err = work.simState.AddLink(work.assocKey, peerID, work.ownerID)
		} else {
			err = work.simState.AddLink(work.assocKey, work.ownerID, peerID)
		}
		if err != nil {
			return fmt.Errorf("link %s: %w", work.assoc.Name, err)
		}
		linked[peerID] = true
	}
	return nil
}

func (e *ActionExecutor) applyAssociationLinkRemovals(ctx *ExecutionContext) {
	simState := e.bindingsBuilder.State()
	for key, peerIDs := range ctx.associationRemovedPeerSets() {
		for _, peerID := range peerIDs {
			if ctx.AssociationDestroyCandidate(key, peerID) {
				continue
			}
			if key.Reverse {
				simState.RemoveLink(key.AssocKey, peerID, key.OwnerInstanceID)
			} else {
				simState.RemoveLink(key.AssocKey, key.OwnerInstanceID, peerID)
			}
		}
	}
}

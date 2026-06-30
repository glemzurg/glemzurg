package actions

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/invariants"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
)

type peerEventViolationContext struct {
	OwnerInstanceID state.InstanceID
	OwnerClassKey   identity.Key
	AssociationName string
}

func (e *ActionExecutor) peerEventAvailable(
	class model_class.Class,
	instance *state.ClassInstance,
	eventKey identity.Key,
) bool {
	if e.peerCatalog == nil {
		return false
	}
	event, ok := e.peerCatalog.PeerEvent(class.Key, eventKey)
	if !ok {
		return false
	}
	candidates := e.findCandidateTransitions(class, event, instance, getInstanceCurrentState(instance))
	return len(candidates) > 0
}

func findFinalDestroyEvent(class model_class.Class) (model_state.Event, bool) {
	for _, ev := range class.Events {
		if !model_state.IsSystemFinalEvent(ev.Name) {
			continue
		}
		return ev, true
	}
	for _, t := range class.Transitions {
		if t.ToStateKey != nil {
			continue
		}
		ev, ok := class.Events[t.EventKey]
		if ok {
			return ev, true
		}
	}
	return model_state.Event{}, false
}

func (e *ActionExecutor) recordPeerEventUnavailable(
	ctx *ExecutionContext,
	vctx peerEventViolationContext,
	peerClass model_class.Class,
	peerInstanceID state.InstanceID,
	eventKey identity.Key,
	eventName string,
) {
	ctx.AddPeerViolation(e.peerEventUnavailableViolation(vctx, peerClass, peerInstanceID, eventKey, eventName))
}

func (e *ActionExecutor) peerEventUnavailableViolation(
	vctx peerEventViolationContext,
	peerClass model_class.Class,
	peerInstanceID state.InstanceID,
	eventKey identity.Key,
	eventName string,
) *invariants.ViolationError {
	stateName := ""
	if peerInstanceID != 0 {
		if inst := e.bindingsBuilder.State().GetInstance(peerInstanceID); inst != nil {
			stateName = getInstanceCurrentState(inst)
		}
	}
	msg := fmt.Sprintf(
		"association %q sent event %s to class %s",
		vctx.AssociationName, eventName, peerClass.Name,
	)
	if peerInstanceID != 0 {
		if stateName != "" {
			msg = fmt.Sprintf(
				"%s but instance %d has no %s transition from state %s",
				msg, peerInstanceID, eventName, stateName,
			)
		} else {
			msg = fmt.Sprintf("%s but instance %d is not available", msg, peerInstanceID)
		}
	} else {
		msg = fmt.Sprintf("%s but the class has no %s creation transition", msg, eventName)
	}
	return invariants.NewPeerEventUnavailableViolation(invariants.PeerEventUnavailableParams{
		OwnerClassKey:   vctx.OwnerClassKey,
		OwnerInstanceID: vctx.OwnerInstanceID,
		AssociationName: vctx.AssociationName,
		PeerClassKey:    peerClass.Key,
		PeerInstanceID:  peerInstanceID,
		EventKey:        eventKey,
		EventName:       eventName,
		Message:         msg,
	})
}

func (e *ActionExecutor) ownerViolationContext(ownerInstanceID state.InstanceID, fallbackClassKey identity.Key, assocName string) peerEventViolationContext {
	ownerClassKey := fallbackClassKey
	if owner := e.bindingsBuilder.State().GetInstance(ownerInstanceID); owner != nil {
		ownerClassKey = owner.ClassKey
	}
	return peerEventViolationContext{
		OwnerInstanceID: ownerInstanceID,
		OwnerClassKey:   ownerClassKey,
		AssociationName: assocName,
	}
}

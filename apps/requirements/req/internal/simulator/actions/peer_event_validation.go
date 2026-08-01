package actions

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/instance"
)

type peerEventViolationContext struct {
	OwnerInstanceID instance.ID
	OwnerClassKey   identity.Key
	AssociationName string
}

func (e *ActionExecutor) peerEventAvailable(
	class model_class.Class,
	inst *instance.Instance,
	eventKey identity.Key,
) bool {
	if e.sch == nil {
		return false
	}
	event, ok := e.sch.PeerEvent(class.Key, eventKey)
	if !ok {
		return false
	}
	candidates := e.findCandidateTransitions(class, event, inst, getInstanceCurrentState(inst))
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
	peerInstanceID instance.ID,
	eventKey identity.Key,
	eventName string,
) {
	e.recordPeerEventUnavailableDetail(ctx, vctx, peerClass, peerInstanceID, eventKey, eventName, "")
}

func (e *ActionExecutor) recordPeerEventUnavailableDetail(
	ctx *ExecutionContext,
	vctx peerEventViolationContext,
	peerClass model_class.Class,
	peerInstanceID instance.ID,
	eventKey identity.Key,
	eventName string,
	detail string,
) {
	st := e.bindingsBuilder.State()
	if st == nil {
		return
	}
	v := st.CheckPeerEventUnavailable(instance.PeerEventUnavailableInput{
		OwnerClassKey:   vctx.OwnerClassKey,
		OwnerInstanceID: vctx.OwnerInstanceID,
		AssociationName: vctx.AssociationName,
		PeerClassKey:    peerClass.Key,
		PeerClassName:   peerClass.Name,
		PeerInstanceID:  peerInstanceID,
		EventKey:        eventKey,
		EventName:       eventName,
		Detail:          detail,
	})
	if v != nil {
		ctx.AddPeerViolation(v)
	}
}

func (e *ActionExecutor) ownerViolationContext(ownerInstanceID instance.ID, fallbackClassKey identity.Key, assocName string) peerEventViolationContext {
	ownerClassKey := fallbackClassKey
	if owner := e.bindingsBuilder.State().GetInstance(ownerInstanceID); owner != nil {
		ownerClassKey = owner.GetClassKey()
	}
	return peerEventViolationContext{
		OwnerInstanceID: ownerInstanceID,
		OwnerClassKey:   ownerClassKey,
		AssociationName: assocName,
	}
}

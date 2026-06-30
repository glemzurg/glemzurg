package actions

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
)

type destroyGuaranteeWork struct {
	mapTarget *associationSetMapTarget
	selection *me.SetFilter
	eventCall *me.EventCall
	event     model_state.Event
	linked    []state.InstanceID
}

func (e *ActionExecutor) tryQueueAssociationDestroyGuarantee(
	ctx *ExecutionContext,
	instance *state.ClassInstance,
	guar model_logic.Logic,
	bindings *evaluator.Bindings,
) (bool, error) {
	if guar.Type != model_logic.LogicTypeDelete {
		return false, nil
	}
	work, queuePeers, err := e.prepareDestroyGuaranteeWork(ctx, instance, guar, bindings)
	if err != nil {
		return false, err
	}
	if queuePeers {
		e.queueDestroyGuaranteePeers(ctx, instance, work, bindings)
	}
	return true, nil
}

func (e *ActionExecutor) prepareDestroyGuaranteeWork(
	ctx *ExecutionContext,
	instance *state.ClassInstance,
	guar model_logic.Logic,
	bindings *evaluator.Bindings,
) (*destroyGuaranteeWork, bool, error) {
	assocRef, selection, eventCall, ok := model_class.MatchAssociationDestroyGuarantee(guar)
	if !ok {
		return nil, false, fmt.Errorf("delete guarantee on %q: expression not in delete guarantee form", guar.Target)
	}
	if model_class.DestroyGuaranteeHasInlineStateChange(guar) {
		if handled, err := e.tryApplyAssociationStateChangeGuarantee(
			ctx, instance, guar.Target, guar.Spec.Expression, bindings,
		); err != nil {
			return nil, false, err
		} else if !handled {
			return nil, false, fmt.Errorf("delete guarantee on %q: inline association update did not apply", guar.Target)
		}
	}
	mapTarget, event, eventFound, err := e.resolveDestroyGuaranteeTarget(instance, guar.Target, assocRef, eventCall)
	if err != nil {
		return nil, false, err
	}
	removed := ctx.AssociationRemovedPeers(instance.ID, mapTarget.assocKey)
	if !eventFound {
		if len(removed) > 0 {
			e.recordDestroyGuaranteeUnavailable(ctx, instance, destroyGuaranteeUnavailableWork{
				mapTarget: mapTarget,
				selection: selection,
				eventCall: eventCall,
				removed:   removed,
				bindings:  bindings,
			})
		}
		return nil, false, nil
	}
	if len(removed) == 0 {
		return nil, false, nil
	}
	return &destroyGuaranteeWork{
		mapTarget: mapTarget,
		selection: selection,
		eventCall: eventCall,
		event:     event,
		linked:    removed,
	}, true, nil
}

func (e *ActionExecutor) resolveDestroyGuaranteeTarget(
	instance *state.ClassInstance,
	target string,
	assocRef *me.AssociationRef,
	eventCall *me.EventCall,
) (*associationSetMapTarget, model_state.Event, bool, error) {
	if e.peerCatalog == nil {
		return nil, model_state.Event{}, false, fmt.Errorf("delete guarantee on %q: peer catalog not configured", target)
	}
	assocKey, assoc, found := e.peerCatalog.OutgoingAssociationByTLAField(instance.ClassKey, target)
	if !found {
		return nil, model_state.Event{}, false, fmt.Errorf(
			"delete guarantee on %q: no outgoing association on class %s",
			target, instance.ClassKey.String(),
		)
	}
	if assocRef.AssociationKey != assocKey {
		return nil, model_state.Event{}, false, fmt.Errorf(
			"delete guarantee on %q: expression association %s does not match target",
			target, assocRef.AssociationKey.String(),
		)
	}
	toClass, ok := e.peerCatalog.PeerClass(assoc.ToClassKey)
	if !ok {
		return nil, model_state.Event{}, false, fmt.Errorf(
			"delete guarantee on %q: peer class %s not found",
			target, assoc.ToClassKey.String(),
		)
	}
	event, ok := e.peerCatalog.PeerEvent(assoc.ToClassKey, eventCall.EventKey)
	if !ok {
		return &associationSetMapTarget{assocKey: assocKey, assoc: assoc, toClass: toClass}, model_state.Event{}, false, nil
	}
	return &associationSetMapTarget{assocKey: assocKey, assoc: assoc, toClass: toClass}, event, true, nil
}

func (e *ActionExecutor) queueDestroyGuaranteePeers(
	ctx *ExecutionContext,
	instance *state.ClassInstance,
	work *destroyGuaranteeWork,
	bindings *evaluator.Bindings,
) {
	simState := e.bindingsBuilder.State()
	for _, peerID := range work.linked {
		peerInstance := simState.GetInstance(peerID)
		if peerInstance == nil {
			continue
		}
		e.queueDestroyGuaranteePeer(ctx, instance, work, bindings, peerID, peerInstance)
	}
}

func (e *ActionExecutor) queueDestroyGuaranteePeer(
	ctx *ExecutionContext,
	instance *state.ClassInstance,
	work *destroyGuaranteeWork,
	bindings *evaluator.Bindings,
	peerID state.InstanceID,
	peerInstance *state.ClassInstance,
) {
	if !destroyGuaranteeSelectsPeer(work.selection, peerInstance.Attributes, bindings) {
		return
	}
	ctx.MarkAssociationDestroyCandidate(instance.ID, work.mapTarget.assocKey, peerID)
	childBindings := evaluator.NewEnclosedBindings(bindings)
	childBindings.Set(work.selection.Variable, peerInstance.Attributes, evaluator.NamespaceLocal)
	params, err := resolvePositionalEventCallParams(
		destroyEventCallBoundVariable(work.eventCall),
		work.event.ParameterNames,
		work.eventCall,
		childBindings,
	)
	if err != nil {
		e.recordSetMapParamBindingError(ctx, instance, work.mapTarget, work.event, err)
		return
	}
	e.queueDestroyPeerUpdate(ctx, instance, work.mapTarget, work.event, params, peerID)
}

type destroyGuaranteeUnavailableWork struct {
	mapTarget *associationSetMapTarget
	selection *me.SetFilter
	eventCall *me.EventCall
	removed   []state.InstanceID
	bindings  *evaluator.Bindings
}

// recordDestroyGuaranteeUnavailable records PeerEventUnavailable for each removed peer
// selected for _destroy when the peer class cannot accept the delete event. Association
// links for those peers are retained until _destroy succeeds.
func (e *ActionExecutor) recordDestroyGuaranteeUnavailable(
	ctx *ExecutionContext,
	instance *state.ClassInstance,
	work destroyGuaranteeUnavailableWork,
) {
	simState := e.bindingsBuilder.State()
	vctx := peerEventViolationContext{
		OwnerInstanceID: instance.ID,
		OwnerClassKey:   instance.ClassKey,
		AssociationName: work.mapTarget.assoc.Name,
	}
	eventName := work.eventCall.EventKey.SubKey
	recorded := false
	for _, peerID := range work.removed {
		peerInstance := simState.GetInstance(peerID)
		if peerInstance == nil || !destroyGuaranteeSelectsPeer(work.selection, peerInstance.Attributes, work.bindings) {
			continue
		}
		ctx.MarkAssociationDestroyCandidate(instance.ID, work.mapTarget.assocKey, peerID)
		e.recordPeerEventUnavailable(ctx, vctx, work.mapTarget.toClass, peerID, work.eventCall.EventKey, eventName)
		recorded = true
	}
	if !recorded {
		e.recordPeerEventUnavailable(ctx, vctx, work.mapTarget.toClass, 0, work.eventCall.EventKey, eventName)
	}
}

// destroyEventCallBoundVariable is the first destroy_event argument name (ignored for binding).
func destroyEventCallBoundVariable(eventCall *me.EventCall) string {
	if len(eventCall.Args) == 0 {
		return ""
	}
	if local, ok := eventCall.Args[0].(*me.LocalVar); ok {
		return local.Name
	}
	return ""
}

func destroyGuaranteeSelectsPeer(
	selection *me.SetFilter,
	peerRecord *object.Record,
	bindings *evaluator.Bindings,
) bool {
	childBindings := evaluator.NewEnclosedBindings(bindings)
	childBindings.Set(selection.Variable, peerRecord, evaluator.NamespaceLocal)
	predResult := evaluator.Eval(selection.Predicate, childBindings)
	if predResult.IsError() {
		return false
	}
	predBool, ok := predResult.Value.(*object.Boolean)
	return ok && predBool.Value()
}

func (e *ActionExecutor) queueDestroyPeerUpdate(
	ctx *ExecutionContext,
	instance *state.ClassInstance,
	mapTarget *associationSetMapTarget,
	event model_state.Event,
	params map[string]object.Object,
	peerID state.InstanceID,
) {
	vctx := peerEventViolationContext{
		OwnerInstanceID: instance.ID,
		OwnerClassKey:   instance.ClassKey,
		AssociationName: mapTarget.assoc.Name,
	}
	peerInstance := e.bindingsBuilder.State().GetInstance(peerID)
	if peerInstance == nil {
		return
	}
	if !e.peerEventAvailable(mapTarget.toClass, peerInstance, event.Key) {
		e.recordPeerEventUnavailable(ctx, vctx, mapTarget.toClass, peerID, event.Key, event.Name)
		return
	}
	ctx.AddPeerUpdate(DeferredPeerUpdate{
		OwnerInstanceID: instance.ID,
		AssocKey:        mapTarget.assocKey,
		PeerInstanceID:  peerID,
		ToClassKey:      mapTarget.assoc.ToClassKey,
		EventKey:        event.Key,
		EventName:       event.Name,
		Params:          params,
		RemovesLink:     true,
	})
}

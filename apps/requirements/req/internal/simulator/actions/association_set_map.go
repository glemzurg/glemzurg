package actions

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/instance"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/invariants"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

type associationSetMapTarget struct {
	assocKey identity.Key
	assoc    model_class.Association
	toClass  model_class.Class
}

func (e *ActionExecutor) tryQueueAssociationSetMapGuarantee(
	ctx *ExecutionContext,
	instance *instance.Instance,
	target string,
	expr me.Expression,
	bindings *evaluator.Bindings,
) (bool, error) {
	setMap, ok := expr.(*me.SetMap)
	if !ok {
		return false, nil
	}
	return e.queueAssociationSetMap(ctx, instance, target, setMap, bindings)
}

func (e *ActionExecutor) tryQueueAssociationAddOrUpdateGuarantee(
	ctx *ExecutionContext,
	instance *instance.Instance,
	target string,
	expr me.Expression,
	bindings *evaluator.Bindings,
) (bool, error) {
	assocRef, createCall, updateCall, ok := model_class.MatchAssociationAddOrUpdateExpr(expr)
	if !ok {
		return false, nil
	}
	if e.peerCatalog == nil {
		return false, fmt.Errorf("association add-or-update guarantee on %q: peer catalog not configured", target)
	}
	assocKey, assoc, found := e.peerCatalog.OutgoingAssociationByTLAField(instance.ClassKey, target)
	if !found {
		return false, fmt.Errorf(
			"association add-or-update guarantee on %q: no outgoing association on class %s",
			target, instance.ClassKey.String(),
		)
	}
	if assocRef.AssociationKey != assocKey {
		return false, fmt.Errorf(
			"association add-or-update guarantee on %q: expression association %s does not match target",
			target, assocRef.AssociationKey.String(),
		)
	}

	linked := linkedAssociationPeerEndpoints(e.bindingsBuilder.State(), instance.ID, assoc)
	if len(linked) == 0 {
		setAddExpr := &me.SetOp{
			Op:   me.SetUnion,
			Left: assocRef,
			Right: &me.SetLiteral{
				Elements: []me.Expression{createCall},
			},
		}
		return e.tryQueueAssociationSetAddGuarantee(ctx, instance, target, setAddExpr, bindings, setAddLinkEnv{})
	}

	ifte, ok := expr.(*me.IfThenElse)
	if !ok {
		return false, fmt.Errorf("association add-or-update guarantee on %q: expected IF expression", target)
	}
	setMapExpr, ok := ifte.Else.(*me.SetMap)
	if !ok {
		return false, fmt.Errorf("association add-or-update guarantee on %q: ELSE branch must be set-map", target)
	}
	_ = updateCall
	_ = assoc
	return e.queueAssociationSetMap(ctx, instance, target, setMapExpr, bindings)
}

func (e *ActionExecutor) queueAssociationSetMap(
	ctx *ExecutionContext,
	instance *instance.Instance,
	target string,
	setMap *me.SetMap,
	bindings *evaluator.Bindings,
) (bool, error) {
	mapTarget, eventCall, err := e.resolveAssociationSetMapTarget(instance, target, setMap)
	if err != nil {
		return false, err
	}
	if mapTarget == nil {
		if eventCall != nil {
			// Set-map form matched but peer class is out of scope — ignore.
			return true, nil
		}
		// Not an association set-map (e.g. endpoint image set-comprehension).
		return false, nil
	}
	if eventCall == nil {
		return false, nil
	}
	linked := linkedAssociationPeerEndpoints(e.bindingsBuilder.State(), instance.ID, mapTarget.assoc)
	if len(linked) == 0 {
		// No peers to update (e.g. cascade Delete on an empty Defines) — not an error.
		return true, nil
	}
	event, ok, err := e.resolveSetMapPeerEvent(ctx, instance, target, mapTarget, eventCall)
	if err != nil || !ok {
		return ok, err
	}
	params, err := resolvePositionalEventCallParams(setMap.Variable, event.ParameterNames, eventCall, bindings)
	if err != nil {
		e.recordSetMapParamBindingError(ctx, instance, mapTarget, event, err)
		return true, nil
	}
	params = reifyOwnerSelfParams(params, instance)
	e.queueSetMapPeerUpdates(ctx, instance, mapTarget, event, params, linked)
	return true, nil
}

func (e *ActionExecutor) resolveSetMapPeerEvent(
	ctx *ExecutionContext,
	instance *instance.Instance,
	target string,
	mapTarget *associationSetMapTarget,
	eventCall *me.EventCall,
) (model_state.Event, bool, error) {
	event, ok := e.peerCatalog.PeerEvent(mapTarget.assoc.ToClassKey, eventCall.EventKey)
	if !ok {
		vctx := peerEventViolationContext{
			OwnerInstanceID: instance.ID,
			OwnerClassKey:   instance.ClassKey,
			AssociationName: mapTarget.assoc.Name,
		}
		e.recordPeerEventUnavailable(ctx, vctx, mapTarget.toClass, 0, eventCall.EventKey, eventCall.EventKey.SubKey)
		return model_state.Event{}, true, nil
	}
	if model_state.IsSystemFinalEvent(event.Name) {
		return model_state.Event{}, false, fmt.Errorf(
			"association set-map guarantee on %q: peer _destroy must use guarantee type destroy with destroy_event",
			target,
		)
	}
	return event, true, nil
}

func (e *ActionExecutor) resolveAssociationSetMapTarget(
	instance *instance.Instance,
	target string,
	setMap *me.SetMap,
) (*associationSetMapTarget, *me.EventCall, error) {
	assocRef, eventCall, ok := model_class.MatchAssociationSetMapExpr(setMap)
	if !ok {
		return nil, nil, nil
	}
	if e.peerCatalog == nil {
		return nil, nil, fmt.Errorf("association set-map guarantee on %q: peer catalog not configured", target)
	}
	assocKey, assoc, found := e.peerCatalog.OutgoingAssociationByTLAField(instance.ClassKey, target)
	if !found {
		return nil, nil, fmt.Errorf(
			"association set-map guarantee on %q: no outgoing association on class %s",
			target, instance.ClassKey.String(),
		)
	}
	if assocRef.AssociationKey != assocKey {
		return nil, nil, fmt.Errorf(
			"association set-map guarantee on %q: expression association %s does not match target",
			target, assocRef.AssociationKey.String(),
		)
	}
	toClass, ok := e.peerCatalog.PeerClass(assoc.ToClassKey)
	if !ok {
		// Known association, peer class outside surface: no-op (eventCall non-nil signals handled).
		return nil, eventCall, nil
	}
	return &associationSetMapTarget{assocKey: assocKey, assoc: assoc, toClass: toClass}, eventCall, nil
}

func (e *ActionExecutor) queueSetMapPeerUpdates(
	ctx *ExecutionContext,
	instance *instance.Instance,
	mapTarget *associationSetMapTarget,
	event model_state.Event,
	params map[string]object.Object,
	linked []instance.ID,
) {
	vctx := peerEventViolationContext{
		OwnerInstanceID: instance.ID,
		OwnerClassKey:   instance.ClassKey,
		AssociationName: mapTarget.assoc.Name,
	}
	simState := e.bindingsBuilder.State()
	for _, peerID := range linked {
		peerInstance := simState.GetInstance(peerID)
		if peerInstance == nil {
			continue
		}
		if !e.peerEventAvailable(mapTarget.toClass, peerInstance, event.Key) {
			e.recordPeerEventUnavailable(ctx, vctx, mapTarget.toClass, peerID, event.Key, event.Name)
			continue
		}
		ctx.AddPeerUpdate(DeferredPeerUpdate{
			OwnerInstanceID: instance.ID,
			AssocKey:        mapTarget.assocKey,
			PeerInstanceID:  peerID,
			ToClassKey:      mapTarget.assoc.ToClassKey,
			EventKey:        event.Key,
			EventName:       event.Name,
			Params:          params,
			RemovesLink:     model_state.IsSystemFinalEvent(event.Name),
		})
	}
}

func (e *ActionExecutor) recordSetMapParamBindingError(
	ctx *ExecutionContext,
	instance *instance.Instance,
	mapTarget *associationSetMapTarget,
	event model_state.Event,
	err error,
) {
	msg := fmt.Sprintf(
		"association %q set-map event %s parameter binding failed: %s",
		mapTarget.assoc.Name, event.Name, err.Error(),
	)
	ctx.AddPeerViolation(invariants.NewPeerEventUnavailableViolation(invariants.PeerEventUnavailableParams{
		OwnerClassKey:   instance.ClassKey,
		OwnerInstanceID: instance.ID,
		AssociationName: mapTarget.assoc.Name,
		PeerClassKey:    mapTarget.toClass.Key,
		EventKey:        event.Key,
		EventName:       event.Name,
		Message:         msg,
	}))
}

func (e *ActionExecutor) applyPeerUpdates(ctx *ExecutionContext) error {
	for _, pu := range ctx.GetPeerUpdates() {
		if err := e.applyPeerUpdate(ctx, pu); err != nil {
			return err
		}
	}
	return nil
}

func (e *ActionExecutor) applyPeerUpdate(ctx *ExecutionContext, pu DeferredPeerUpdate) error {
	toClass, event, ok := e.resolvePeerUpdateEvent(pu)
	if !ok {
		return fmt.Errorf("peer update: to-class %s not found", pu.ToClassKey.String())
	}
	instance := e.bindingsBuilder.State().GetInstance(pu.PeerInstanceID)
	if instance == nil {
		return fmt.Errorf("peer update: instance %d not found", pu.PeerInstanceID)
	}
	if !e.peerEventAvailable(toClass, instance, pu.EventKey) {
		e.recordPeerUpdateUnavailable(ctx, pu, toClass, event.Name)
		return nil
	}
	if pu.RemovesLink {
		if err := e.deleteAssociationClassBeforePeer(ctx, pu, toClass); err != nil {
			return err
		}
	}
	return e.executePeerUpdateTransition(ctx, pu, toClass, event, instance)
}

func (e *ActionExecutor) recordPeerUpdateUnavailable(
	ctx *ExecutionContext,
	pu DeferredPeerUpdate,
	toClass model_class.Class,
	eventName string,
) {
	assocName := associationNameForKey(e.peerCatalog, pu.AssocKey)
	vctx := e.ownerViolationContext(pu.OwnerInstanceID, toClass.Key, assocName)
	e.recordPeerEventUnavailable(ctx, vctx, toClass, pu.PeerInstanceID, pu.EventKey, eventName)
}

func (e *ActionExecutor) executePeerUpdateTransition(
	ctx *ExecutionContext,
	pu DeferredPeerUpdate,
	toClass model_class.Class,
	event model_state.Event,
	instance *instance.Instance,
) error {
	result, err := e.ExecuteTransition(toClass, event, instance, pu.Params, CreationLinkSource{}, nil)
	if err != nil {
		e.recordPeerUpdateUnavailable(ctx, pu, toClass, event.Name)
		return nil
	}
	if pu.RemovesLink {
		e.removeAssociationLinkAfterPeerDelete(pu, result)
	}
	e.recordPeerTransition(ctx, toClass, event, pu.Params, result)
	return nil
}

func (e *ActionExecutor) resolvePeerUpdateEvent(pu DeferredPeerUpdate) (model_class.Class, model_state.Event, bool) {
	if e.peerCatalog == nil {
		return model_class.Class{}, model_state.Event{}, false
	}
	toClass, ok := e.peerCatalog.PeerClass(pu.ToClassKey)
	if !ok {
		return model_class.Class{}, model_state.Event{}, false
	}
	event, ok := e.peerCatalog.PeerEvent(pu.ToClassKey, pu.EventKey)
	return toClass, event, ok
}

func associationNameForKey(catalog PeerCreationCatalog, assocKey identity.Key) string {
	if assoc, ok := catalog.AssociationByKey(assocKey); ok {
		return assoc.Name
	}
	return ""
}

func (e *ActionExecutor) deleteAssociationClassBeforePeer(
	ctx *ExecutionContext,
	pu DeferredPeerUpdate,
	toClass model_class.Class,
) error {
	assoc, found := e.peerCatalog.AssociationByKey(pu.AssocKey)
	if !found || assoc.AssociationClassKey == nil {
		return nil
	}
	simState := e.bindingsBuilder.State()
	link, ok := associationLinkForPair(simState, assoc, pu.OwnerInstanceID, pu.PeerInstanceID)
	if !ok {
		return nil
	}
	return e.fireAssociationClassDestroy(ctx, pu, toClass, assoc, link)
}

type associationClassDestroyWork struct {
	pu             DeferredPeerUpdate
	toClass        model_class.Class
	assoc          model_class.Association
	acClass        model_class.Class
	linkInstanceID instance.ID
}

func (e *ActionExecutor) fireAssociationClassDestroy(
	ctx *ExecutionContext,
	pu DeferredPeerUpdate,
	toClass model_class.Class,
	assoc model_class.Association,
	link instance.AssociationLink,
) error {
	acClass, ok := e.peerCatalog.PeerClass(*assoc.AssociationClassKey)
	if !ok {
		return fmt.Errorf("association class %s not found", assoc.AssociationClassKey.String())
	}
	work := associationClassDestroyWork{
		pu: pu, toClass: toClass, assoc: assoc, acClass: acClass, linkInstanceID: link.LinkInstanceID,
	}
	deleteEvent, ok := findFinalDestroyEvent(acClass)
	if !ok {
		work.recordUnavailable(ctx, e, identity.Key{}, model_state.EventNameDestroy)
		return nil
	}
	acInstance := e.bindingsBuilder.State().GetInstance(link.LinkInstanceID)
	if acInstance == nil {
		return nil
	}
	return e.executeAssociationClassDestroy(ctx, work, deleteEvent, acInstance)
}

func (w associationClassDestroyWork) recordUnavailable(
	ctx *ExecutionContext,
	e *ActionExecutor,
	eventKey identity.Key,
	eventName string,
) {
	vctx := e.ownerViolationContext(w.pu.OwnerInstanceID, w.toClass.Key, w.assoc.Name)
	e.recordPeerEventUnavailable(ctx, vctx, w.acClass, w.linkInstanceID, eventKey, eventName)
}

func (e *ActionExecutor) executeAssociationClassDestroy(
	ctx *ExecutionContext,
	work associationClassDestroyWork,
	deleteEvent model_state.Event,
	acInstance *instance.Instance,
) error {
	if !e.peerEventAvailable(work.acClass, acInstance, deleteEvent.Key) {
		work.recordUnavailable(ctx, e, deleteEvent.Key, deleteEvent.Name)
		return nil
	}
	result, err := e.ExecuteTransition(work.acClass, deleteEvent, acInstance, nil, CreationLinkSource{}, nil)
	if err != nil {
		work.recordUnavailable(ctx, e, deleteEvent.Key, deleteEvent.Name)
		return nil
	}
	e.recordPeerTransition(ctx, work.acClass, deleteEvent, nil, result)
	return nil
}

func (e *ActionExecutor) removeAssociationLinkAfterPeerDelete(
	pu DeferredPeerUpdate,
	result *TransitionResult,
) {
	if result != nil && result.WasDestroy {
		return
	}
	assoc, found := e.peerCatalog.AssociationByKey(pu.AssocKey)
	if !found || assoc.AssociationClassKey != nil {
		return
	}
	e.bindingsBuilder.State().RemoveLink(pu.AssocKey, pu.OwnerInstanceID, pu.PeerInstanceID)
}

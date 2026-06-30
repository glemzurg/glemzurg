package actions

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
)

func (e *ActionExecutor) tryQueueAssociationSetMapGuarantee(
	ctx *ExecutionContext,
	instance *state.ClassInstance,
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
	instance *state.ClassInstance,
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

	linked := e.bindingsBuilder.State().GetLinkedForward(instance.ID, assocKey)
	if len(linked) == 0 {
		setAddExpr := &me.SetOp{
			Op:   me.SetUnion,
			Left: assocRef,
			Right: &me.SetLiteral{
				Elements: []me.Expression{createCall},
			},
		}
		return e.tryQueueAssociationSetAddGuarantee(ctx, instance, target, setAddExpr, bindings)
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
	instance *state.ClassInstance,
	target string,
	setMap *me.SetMap,
	bindings *evaluator.Bindings,
) (bool, error) {
	assocRef, eventCall, ok := model_class.MatchAssociationSetMapExpr(setMap)
	if !ok {
		return false, nil
	}
	if e.peerCatalog == nil {
		return false, fmt.Errorf("association set-map guarantee on %q: peer catalog not configured", target)
	}
	assocKey, assoc, found := e.peerCatalog.OutgoingAssociationByTLAField(instance.ClassKey, target)
	if !found {
		return false, fmt.Errorf(
			"association set-map guarantee on %q: no outgoing association on class %s",
			target, instance.ClassKey.String(),
		)
	}
	if assocRef.AssociationKey != assocKey {
		return false, fmt.Errorf(
			"association set-map guarantee on %q: expression association %s does not match target",
			target, assocRef.AssociationKey.String(),
		)
	}

	linked := e.bindingsBuilder.State().GetLinkedForward(instance.ID, assocKey)
	if len(linked) == 0 {
		return false, fmt.Errorf("association set-map guarantee on %q: association is empty", target)
	}

	event, ok := e.peerCatalog.PeerEvent(assoc.ToClassKey, eventCall.EventKey)
	if !ok {
		return false, fmt.Errorf(
			"association set-map guarantee on %q: peer class %s has no event for key %s",
			target, assoc.ToClassKey.String(), eventCall.EventKey.String(),
		)
	}

	params, err := resolveSetMapEventParams(setMap.Variable, eventCall, bindings)
	if err != nil {
		return false, err
	}
	params, err = fillEventParamsFromBindings(event, params, bindings)
	if err != nil {
		return false, err
	}

	for _, peerID := range linked {
		ctx.AddPeerUpdate(DeferredPeerUpdate{
			PeerInstanceID: peerID,
			ToClassKey:     assoc.ToClassKey,
			EventKey:       event.Key,
			Params:         params,
		})
	}
	return true, nil
}

func resolveSetMapEventParams(
	boundVar string,
	eventCall *me.EventCall,
	bindings *evaluator.Bindings,
) (map[string]object.Object, error) {
	params := make(map[string]object.Object)
	for i, arg := range eventCall.Args {
		name, ok := eventCallArgName(arg)
		if !ok {
			return nil, fmt.Errorf("association set-map event arg[%d]: expected parameter reference", i)
		}
		if name == boundVar {
			continue
		}
		result := evaluator.Eval(arg, bindings)
		if result.IsError() {
			return nil, fmt.Errorf("association set-map event arg %q: %s", name, result.Error.Inspect())
		}
		params[name] = result.Value
	}
	return params, nil
}

func fillEventParamsFromBindings(
	event model_state.Event,
	params map[string]object.Object,
	bindings *evaluator.Bindings,
) (map[string]object.Object, error) {
	if params == nil {
		params = make(map[string]object.Object)
	}
	for _, name := range event.ParameterNames {
		if _, ok := params[name]; ok {
			continue
		}
		result := evaluator.Eval(&me.LocalVar{Name: name}, bindings)
		if result.IsError() {
			return nil, fmt.Errorf("event param %q: %s", name, result.Error.Inspect())
		}
		params[name] = result.Value
	}
	return params, nil
}

func (e *ActionExecutor) applyPeerUpdates(ctx *ExecutionContext) error {
	for _, pu := range ctx.GetPeerUpdates() {
		if err := e.applyPeerUpdate(pu); err != nil {
			return err
		}
	}
	return nil
}

func (e *ActionExecutor) applyPeerUpdate(pu DeferredPeerUpdate) error {
	if e.peerCatalog == nil {
		return fmt.Errorf("peer update: catalog not configured")
	}
	toClass, ok := e.peerCatalog.PeerClass(pu.ToClassKey)
	if !ok {
		return fmt.Errorf("peer update: to-class %s not found", pu.ToClassKey.String())
	}
	event, ok := e.peerCatalog.PeerEvent(pu.ToClassKey, pu.EventKey)
	if !ok {
		return fmt.Errorf("peer update: event %s not found on class %s", pu.EventKey.String(), toClass.Name)
	}
	instance := e.bindingsBuilder.State().GetInstance(pu.PeerInstanceID)
	if instance == nil {
		return fmt.Errorf("peer update: instance %d not found", pu.PeerInstanceID)
	}
	_, err := e.ExecuteTransition(
		toClass,
		event,
		instance,
		pu.Params,
		CreationLinkSource{SourceAssocKey: nil, SourceID: nil},
		nil,
	)
	return err
}

package actions

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
)

func (e *ActionExecutor) tryQueueAssociationSetAddGuarantee(
	ctx *ExecutionContext,
	instance *state.ClassInstance,
	target string,
	expr me.Expression,
	bindings *evaluator.Bindings,
	linkEnv setAddLinkEnv,
) (bool, error) {
	assocRef, eventCall, ok := model_class.MatchAssociationSetAddExpr(expr)
	if !ok {
		return false, nil
	}
	assocTarget, err := e.resolveAssociationSetAddTarget(instance, target, assocRef)
	if err != nil {
		return false, err
	}
	if assocTarget == nil {
		return false, nil
	}
	if !e.validateSetAddPeerEvents(ctx, instance, assocTarget, eventCall) {
		return true, nil
	}
	creationEvent, ok := e.peerCatalog.PeerCreationEvent(assocTarget.assoc.ToClassKey)
	if !ok {
		return false, fmt.Errorf("association set-add guarantee on %q: peer class has no creation event", target)
	}
	params, err := resolvePositionalEventCallParams("", creationEvent.ParameterNames, eventCall, bindings)
	if err != nil {
		return false, fmt.Errorf("association set-add guarantee on %q: %w", target, err)
	}
	ctx.AddPeerCreation(DeferredPeerCreation{
		FromInstanceID: instance.ID,
		AssocKey:       assocTarget.assoc.Key,
		ToClassKey:     assocTarget.assoc.ToClassKey,
		Params:         params,
		ActionParams:   linkEnv.actionParams,
	})
	return true, nil
}

type associationSetAddTarget struct {
	assoc   model_class.Association
	toClass model_class.Class
}

func (e *ActionExecutor) resolveAssociationSetAddTarget(
	instance *state.ClassInstance,
	target string,
	assocRef *me.AssociationRef,
) (*associationSetAddTarget, error) {
	if e.peerCatalog == nil {
		return nil, fmt.Errorf("association set-add guarantee on %q: peer catalog not configured", target)
	}
	assocKey, assoc, found := e.peerCatalog.OutgoingAssociationByTLAField(instance.ClassKey, target)
	if !found {
		return nil, fmt.Errorf(
			"association set-add guarantee on %q: no outgoing association on class %s",
			target, instance.ClassKey.String(),
		)
	}
	if assocRef.AssociationKey != assocKey {
		return nil, fmt.Errorf(
			"association set-add guarantee on %q: expression association %s does not match target",
			target, assocRef.AssociationKey.String(),
		)
	}
	toClass, ok := e.peerCatalog.PeerClass(assoc.ToClassKey)
	if !ok {
		return nil, fmt.Errorf(
			"association set-add guarantee on %q: peer class %s not found",
			target, assoc.ToClassKey.String(),
		)
	}
	return &associationSetAddTarget{assoc: assoc, toClass: toClass}, nil
}

func (e *ActionExecutor) validateSetAddPeerEvents(
	ctx *ExecutionContext,
	instance *state.ClassInstance,
	target *associationSetAddTarget,
	eventCall *me.EventCall,
) bool {
	vctx := peerEventViolationContext{
		OwnerInstanceID: instance.ID,
		OwnerClassKey:   instance.ClassKey,
		AssociationName: target.assoc.Name,
	}
	creationEvent, ok := e.peerCatalog.PeerCreationEvent(target.assoc.ToClassKey)
	if !ok || !e.peerEventAvailable(target.toClass, nil, creationEvent.Key) {
		e.recordPeerEventUnavailable(ctx, vctx, target.toClass, 0, eventCall.EventKey, eventCall.EventKey.SubKey)
		return false
	}
	if target.assoc.AssociationClassKey == nil {
		return true
	}
	acClass, ok := e.peerCatalog.PeerClass(*target.assoc.AssociationClassKey)
	if !ok {
		return true
	}
	acCreationEvent, ok := e.peerCatalog.PeerCreationEvent(*target.assoc.AssociationClassKey)
	if !ok || !e.peerEventAvailable(acClass, nil, acCreationEvent.Key) {
		e.recordPeerEventUnavailable(ctx, vctx, acClass, 0, eventCall.EventKey, eventCall.EventKey.SubKey)
		return false
	}
	return true
}

func (e *ActionExecutor) applyPeerCreations(ctx *ExecutionContext) error {
	for _, pc := range ctx.GetPeerCreations() {
		if err := e.applyPeerCreation(ctx, pc); err != nil {
			return err
		}
	}
	return nil
}

func (e *ActionExecutor) applyPeerCreation(ctx *ExecutionContext, pc DeferredPeerCreation) error {
	if e.peerCatalog == nil {
		return fmt.Errorf("peer creation for association %s: catalog not configured", pc.AssocKey.String())
	}
	assoc, found := e.peerCatalog.AssociationByKey(pc.AssocKey)
	if !found {
		return fmt.Errorf("peer creation for association %s: association metadata not found", pc.AssocKey.String())
	}
	// Association-class materialization only when the AC class is on the surface catalog.
	if assoc.AssociationClassKey != nil {
		if _, ok := e.peerCatalog.PeerClass(*assoc.AssociationClassKey); ok {
			return e.applyAssociationClassPeerCreation(ctx, pc, assoc)
		}
	}
	return e.applyPlainPeerCreation(ctx, pc, assoc)
}

func (e *ActionExecutor) applyPlainPeerCreation(
	ctx *ExecutionContext,
	pc DeferredPeerCreation,
	assoc model_class.Association,
) error {
	toClass, creationEvent, err := e.resolvePeerCreationEvent(pc)
	if err != nil {
		return err
	}
	assocKey := pc.AssocKey
	fromID := pc.FromInstanceID
	result, err := e.ExecuteTransition(
		toClass, creationEvent, nil, pc.Params,
		CreationLinkSource{SourceAssocKey: &assocKey, SourceID: &fromID}, nil,
	)
	if err != nil {
		vctx := e.ownerViolationContext(fromID, toClass.Key, assoc.Name)
		e.recordPeerEventUnavailable(ctx, vctx, toClass, 0, creationEvent.Key, creationEvent.Name)
		return nil
	}
	e.recordPeerTransition(ctx, toClass, creationEvent, pc.Params, result)
	return e.applyInferredSecondaryLinks(pc, result.InstanceID)
}

func (e *ActionExecutor) applyAssociationClassPeerCreation(
	ctx *ExecutionContext,
	pc DeferredPeerCreation,
	assoc model_class.Association,
) error {
	// Existing to-endpoint: only materialize the association-class row (params go to AC).
	if pc.ToInstanceID != nil {
		return e.materializeAssociationClassRow(ctx, pc, assoc, *pc.ToInstanceID, pc.Params)
	}

	toClass, creationEvent, err := e.resolvePeerCreationEvent(pc)
	if err != nil {
		return err
	}
	endpointResult, err := e.ExecuteTransition(
		toClass, creationEvent, nil, pc.Params, CreationLinkSource{}, nil,
	)
	if err != nil {
		vctx := e.ownerViolationContext(pc.FromInstanceID, toClass.Key, assoc.Name)
		e.recordPeerEventUnavailable(ctx, vctx, toClass, 0, creationEvent.Key, creationEvent.Name)
		return nil
	}
	e.recordPeerTransition(ctx, toClass, creationEvent, pc.Params, endpointResult)
	// AC row created without extra params when the event call targeted the to-class.
	return e.materializeAssociationClassRow(ctx, pc, assoc, endpointResult.InstanceID, nil)
}

func (e *ActionExecutor) resolvePeerCreationEvent(pc DeferredPeerCreation) (model_class.Class, model_state.Event, error) {
	toClass, ok := e.peerCatalog.PeerClass(pc.ToClassKey)
	if !ok {
		return model_class.Class{}, model_state.Event{}, fmt.Errorf(
			"peer creation for association %s: to-class %s not found", pc.AssocKey.String(), pc.ToClassKey.String(),
		)
	}
	creationEvent, ok := e.peerCatalog.PeerCreationEvent(pc.ToClassKey)
	if !ok {
		return model_class.Class{}, model_state.Event{}, fmt.Errorf(
			"peer creation for association %s: to-class %s has no creation event",
			pc.AssocKey.String(), toClass.Name,
		)
	}
	return toClass, creationEvent, nil
}

func (e *ActionExecutor) materializeAssociationClassRow(
	ctx *ExecutionContext,
	pc DeferredPeerCreation,
	assoc model_class.Association,
	targetID state.InstanceID,
	acParams map[string]object.Object,
) error {
	acClass, ok := e.peerCatalog.PeerClass(*assoc.AssociationClassKey)
	if !ok {
		return fmt.Errorf("peer creation for association %s: association class %s not found", pc.AssocKey.String(), assoc.AssociationClassKey.String())
	}
	acCreationEvent, ok := e.peerCatalog.PeerCreationEvent(*assoc.AssociationClassKey)
	if !ok {
		vctx := e.ownerViolationContext(pc.FromInstanceID, acClass.Key, assoc.Name)
		e.recordPeerEventUnavailable(ctx, vctx, acClass, 0, identity.Key{}, model_state.EventNameNew)
		return nil
	}
	assocKey := pc.AssocKey
	fromID := pc.FromInstanceID
	acResult, err := e.ExecuteTransition(
		acClass, acCreationEvent, nil, acParams,
		CreationLinkSource{SourceAssocKey: &assocKey, SourceID: &fromID}, &targetID,
	)
	if err != nil {
		vctx := e.ownerViolationContext(pc.FromInstanceID, acClass.Key, assoc.Name)
		e.recordPeerEventUnavailable(ctx, vctx, acClass, 0, acCreationEvent.Key, acCreationEvent.Name)
		return nil
	}
	e.recordPeerTransition(ctx, acClass, acCreationEvent, acParams, acResult)
	return nil
}

// applyInferredSecondaryLinks analyzes set-add peer creation against action parameters
// and the association graph. When a parameter is a live instance of class C and C has
// exactly one outgoing association to the created peer class, the simulator also links
// that parameter instance to the new peer. Pure runtime inference — not authored data.
func (e *ActionExecutor) applyInferredSecondaryLinks(pc DeferredPeerCreation, newPeerID state.InstanceID) error {
	if e.peerCatalog == nil || len(pc.ActionParams) == 0 {
		return nil
	}
	simState := e.bindingsBuilder.State()
	for _, paramVal := range pc.ActionParams {
		fromID, ok := instanceIDFromObject(simState, paramVal)
		if !ok {
			continue
		}
		fromInst := simState.GetInstance(fromID)
		if fromInst == nil {
			continue
		}
		// Skip the primary set-add endpoint (already linked).
		if fromID == pc.FromInstanceID {
			continue
		}
		candidates := e.peerCatalog.OutgoingAssociationsTo(fromInst.ClassKey, pc.ToClassKey)
		if len(candidates) != 1 {
			// Zero or ambiguous associations: do not invent links.
			continue
		}
		assoc := candidates[0]
		if err := simState.AddLink(assoc.Key, fromID, newPeerID); err != nil {
			return fmt.Errorf("inferred secondary link after set-add: %w", err)
		}
	}
	return nil
}

func instanceIDFromObject(simState *state.SimulationState, val object.Object) (state.InstanceID, bool) {
	rec, ok := val.(*object.Record)
	if !ok || rec == nil {
		return 0, false
	}
	if id, ok := state.InstanceIDFromExtentElement(rec); ok {
		if simState.GetInstance(id) != nil {
			return id, true
		}
	}
	// Bare attribute records: match by pointer first (self), then unique structural equality.
	// Structural equality alone is ambiguous when multiple instances share the same data shape
	// (e.g. wallets that only store _state); refuse to guess among duplicates.
	data := state.DataFromExtentElement(rec)
	var (
		found state.InstanceID
		n     int
	)
	for _, inst := range simState.AllInstances() {
		if inst.Attributes == rec || inst.Attributes == data {
			return inst.ID, true
		}
		if (data != nil && inst.Attributes.Equals(data)) || inst.Attributes.Equals(rec) {
			found = inst.ID
			n++
		}
	}
	if n == 1 {
		return found, true
	}
	return 0, false
}

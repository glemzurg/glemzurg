package actions

import (
	"errors"
	"fmt"
	"slices"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/instance"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// errPeerClassOutOfScope means the association is known but its peer class is not on the surface.
// Callers treat this as a successful no-op (empty-set / ignore-link default).
var errPeerClassOutOfScope = errors.New("peer class out of simulation scope")

func (e *ActionExecutor) tryQueueAssociationSetAddGuarantee(
	ctx *ExecutionContext,
	instance *instance.Instance,
	target string,
	expr me.Expression,
	bindings *evaluator.Bindings,
	linkEnv setAddLinkEnv,
) (bool, error) {
	// Multi set-add before single: Assoc \union { _new(...) : x \in Domain }.
	if handled, err := e.tryQueueAssociationMultiSetAddGuarantee(ctx, instance, target, expr, bindings, linkEnv); err != nil || handled {
		return handled, err
	}
	assocRef, eventCall, ok := model_class.MatchAssociationSetAddExpr(expr)
	if !ok {
		return false, nil
	}
	assocTarget, err := e.resolveAssociationSetAddTarget(instance, target, assocRef)
	if errors.Is(err, errPeerClassOutOfScope) {
		// Out-of-scope peer class: association matched but peer is not on the surface.
		return true, nil
	}
	if err != nil {
		return false, err
	}
	if assocTarget == nil {
		return false, nil
	}
	if !e.validateSetAddPeerEvents(ctx, instance, assocTarget, eventCall) {
		return true, nil
	}
	creationEvent, ok := e.sch.PeerCreationEvent(assocTarget.assoc.ToClassKey)
	if !ok {
		return false, fmt.Errorf("association set-add guarantee on %q: peer class has no creation event", target)
	}
	params, err := resolvePositionalEventCallParams("", creationEvent.ParameterNames, eventCall, bindings)
	if err != nil {
		return false, fmt.Errorf("association set-add guarantee on %q: %w", target, err)
	}
	ctx.AddPeerCreation(DeferredPeerCreation{
		FromInstanceID: instance.GetID(),
		AssocKey:       assocTarget.assoc.Key,
		ToClassKey:     assocTarget.assoc.ToClassKey,
		Params:         params,
		ActionParams:   linkEnv.actionParams,
	})
	return true, nil
}

// tryQueueAssociationMultiSetAddGuarantee queues one peer create+link per domain element for
// Assoc \union { _new(...) : x \in Domain }. _new construction args are resolved with x bound.
func (e *ActionExecutor) tryQueueAssociationMultiSetAddGuarantee(
	ctx *ExecutionContext,
	instance *instance.Instance,
	target string,
	expr me.Expression,
	bindings *evaluator.Bindings,
	linkEnv setAddLinkEnv,
) (bool, error) {
	assocRef, setMap, eventCall, ok := model_class.MatchAssociationMultiSetAddExpr(expr)
	if !ok {
		return false, nil
	}
	if !isSystemCreationEventCall(eventCall) {
		return false, fmt.Errorf(
			"association multi set-add guarantee on %q: transform must be _new / «new» (construction only)",
			target,
		)
	}
	assocTarget, creationEvent, handled, err := e.prepareMultiSetAddTarget(ctx, instance, target, assocRef, eventCall)
	if err != nil || handled {
		return handled, err
	}
	if assocTarget == nil {
		return false, nil
	}
	domainSet, err := evalMultiSetAddDomain(target, setMap.Set, bindings)
	if err != nil {
		return false, err
	}
	work := multiSetAddWork{
		ctx: ctx, instance: instance, assocTarget: assocTarget, creationEvent: creationEvent,
		eventCall: eventCall, setMap: setMap, domainSet: domainSet, bindings: bindings, linkEnv: linkEnv,
	}
	if err := e.queueMultiSetAddCreations(work); err != nil {
		return false, fmt.Errorf("association multi set-add guarantee on %q: %w", target, err)
	}
	return true, nil
}

// multiSetAddWork holds context for queuing one peer create per domain element.
type multiSetAddWork struct {
	ctx           *ExecutionContext
	instance      *instance.Instance
	assocTarget   *associationSetAddTarget
	creationEvent model_state.Event
	eventCall     *me.EventCall
	setMap        *me.SetMap
	domainSet     *object.Set
	bindings      *evaluator.Bindings
	linkEnv       setAddLinkEnv
}

func (e *ActionExecutor) prepareMultiSetAddTarget(
	ctx *ExecutionContext,
	instance *instance.Instance,
	target string,
	assocRef *me.AssociationRef,
	eventCall *me.EventCall,
) (*associationSetAddTarget, model_state.Event, bool, error) {
	assocTarget, err := e.resolveAssociationSetAddTarget(instance, target, assocRef)
	if errors.Is(err, errPeerClassOutOfScope) {
		return nil, model_state.Event{}, true, nil
	}
	if err != nil {
		return nil, model_state.Event{}, false, err
	}
	if assocTarget == nil {
		return nil, model_state.Event{}, false, nil
	}
	if !e.validateSetAddPeerEvents(ctx, instance, assocTarget, eventCall) {
		return nil, model_state.Event{}, true, nil
	}
	creationEvent, ok := e.sch.PeerCreationEvent(assocTarget.assoc.ToClassKey)
	if !ok {
		return nil, model_state.Event{}, false, fmt.Errorf(
			"association multi set-add guarantee on %q: peer class has no creation event", target,
		)
	}
	return assocTarget, creationEvent, false, nil
}

func evalMultiSetAddDomain(target string, domainExpr me.Expression, bindings *evaluator.Bindings) (*object.Set, error) {
	domainResult := evaluator.Eval(domainExpr, bindings)
	if domainResult.IsError() {
		return nil, fmt.Errorf("association multi set-add guarantee on %q: domain: %s", target, domainResult.Error.Inspect())
	}
	domainSet, ok := evaluator.CoerceToSet(domainResult.Value)
	if !ok {
		return nil, fmt.Errorf(
			"association multi set-add guarantee on %q: domain must be a set, got %s",
			target, domainResult.Value.Type(),
		)
	}
	return domainSet, nil
}

func (e *ActionExecutor) queueMultiSetAddCreations(work multiSetAddWork) error {
	for _, elem := range work.domainSet.Elements() {
		child := evaluator.NewEnclosedBindings(work.bindings)
		if work.setMap.Variable != "" {
			child.Set(work.setMap.Variable, elem, evaluator.NamespaceLocal)
		}
		// Domain binder is a construction argument (e.g. _new(p)); do not strip it as a set-map peer row.
		params, err := resolvePositionalEventCallParams(
			"", work.creationEvent.ParameterNames, work.eventCall, child,
		)
		if err != nil {
			return err
		}
		work.ctx.AddPeerCreation(DeferredPeerCreation{
			FromInstanceID: work.instance.GetID(),
			AssocKey:       work.assocTarget.assoc.Key,
			ToClassKey:     work.assocTarget.assoc.ToClassKey,
			Params:         params,
			ActionParams:   work.linkEnv.actionParams,
		})
	}
	return nil
}

type associationSetAddTarget struct {
	assoc   model_class.Association
	toClass model_class.Class
}

func (e *ActionExecutor) resolveAssociationSetAddTarget(
	instance *instance.Instance,
	target string,
	assocRef *me.AssociationRef,
) (*associationSetAddTarget, error) {
	if e.sch == nil {
		return nil, fmt.Errorf("association set-add guarantee on %q: peer catalog not configured", target)
	}
	assocKey, assoc, found := e.sch.OutgoingAssociationByTLAField(instance.GetClassKey(), target)
	if !found {
		return nil, fmt.Errorf(
			"association set-add guarantee on %q: no outgoing association on class %s",
			target, instance.GetClassKey().String(),
		)
	}
	if assocRef.AssociationKey != assocKey {
		return nil, fmt.Errorf(
			"association set-add guarantee on %q: expression association %s does not match target",
			target, assocRef.AssociationKey.String(),
		)
	}
	toClass, ok := e.sch.PeerClass(assoc.ToClassKey)
	if !ok {
		// Association is known but peer class is outside the simulation surface.
		return nil, errPeerClassOutOfScope
	}
	return &associationSetAddTarget{assoc: assoc, toClass: toClass}, nil
}

func (e *ActionExecutor) validateSetAddPeerEvents(
	ctx *ExecutionContext,
	instance *instance.Instance,
	target *associationSetAddTarget,
	eventCall *me.EventCall,
) bool {
	vctx := peerEventViolationContext{
		OwnerInstanceID: instance.GetID(),
		OwnerClassKey:   instance.GetClassKey(),
		AssociationName: target.assoc.Name,
	}
	creationEvent, ok := e.sch.PeerCreationEvent(target.assoc.ToClassKey)
	if !ok || !e.peerEventAvailable(target.toClass, nil, creationEvent.Key) {
		e.recordPeerEventUnavailable(ctx, vctx, target.toClass, 0, eventCall.EventKey, eventCall.EventKey.SubKey)
		return false
	}
	if target.assoc.AssociationClassKey == nil {
		return true
	}
	acClass, ok := e.sch.PeerClass(*target.assoc.AssociationClassKey)
	if !ok {
		return true
	}
	acCreationEvent, ok := e.sch.PeerCreationEvent(*target.assoc.AssociationClassKey)
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
	if e.sch == nil {
		return fmt.Errorf("peer creation for association %s: catalog not configured", pc.AssocKey.String())
	}
	assoc, found := e.sch.AssociationByKey(pc.AssocKey)
	if !found {
		return fmt.Errorf("peer creation for association %s: association metadata not found", pc.AssocKey.String())
	}
	// Association-class host: creating a *new* far endpoint via set-add is retired (C1 / E1).
	// Existing-endpoint bulk materialization (ToInstanceID set) still builds an AC row for
	// host guarantees that expand over known peers (e.g. transaction amounts). Out-of-scope
	// AC degrades to plain endpoint links.
	if assoc.AssociationClassKey != nil {
		if _, ok := e.sch.PeerClass(*assoc.AssociationClassKey); ok {
			return e.applyAssociationClassPeerCreation(ctx, pc, assoc)
		}
		if e.sch.IsClassInScope(*assoc.AssociationClassKey) {
			return fmt.Errorf(
				"peer creation for association %s: association class %s is in scope but not creatable as a peer; cannot materialize as a plain host link",
				pc.AssocKey.String(), assoc.AssociationClassKey.String(),
			)
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
		// PeerInstanceID 0 would otherwise imply "no creation transition"; pass the real error.
		e.recordPeerEventUnavailableDetail(ctx, vctx,
			peerEventTarget{Class: toClass, InstanceID: 0},
			peerEventRef{Key: creationEvent.Key, Name: creationEvent.Name},
			fmt.Sprintf("association %q failed to create %s via %s: %v", assoc.Name, toClass.Name, creationEvent.Name, err))
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
	// Existing to-endpoint: materialize the association-class row (host bulk-create style).
	if pc.ToInstanceID != nil {
		return e.materializeAssociationClassRow(ctx, pc, assoc, *pc.ToInstanceID, pc.Params)
	}

	// New far-endpoint + AC together is not supported: use surface AC _new(from, to, …).
	return fmt.Errorf(
		"peer creation for association %s: association class %s must be created via surface _new with host endpoint parameters (not host set-add of a new peer)",
		pc.AssocKey.String(), assoc.AssociationClassKey.String(),
	)
}

func (e *ActionExecutor) resolvePeerCreationEvent(pc DeferredPeerCreation) (model_class.Class, model_state.Event, error) {
	toClass, ok := e.sch.PeerClass(pc.ToClassKey)
	if !ok {
		return model_class.Class{}, model_state.Event{}, fmt.Errorf(
			"peer creation for association %s: to-class %s not found", pc.AssocKey.String(), pc.ToClassKey.String(),
		)
	}
	creationEvent, ok := e.sch.PeerCreationEvent(pc.ToClassKey)
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
	targetID instance.ID,
	acParams map[string]object.Object,
) error {
	acClass, ok := e.sch.PeerClass(*assoc.AssociationClassKey)
	if !ok {
		return fmt.Errorf("peer creation for association %s: association class %s not found", pc.AssocKey.String(), assoc.AssociationClassKey.String())
	}
	acCreationEvent, ok := e.sch.PeerCreationEvent(*assoc.AssociationClassKey)
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

// applyInferredSecondaryLinks analyzes set-add peer creation against action and
// creation parameters and the association graph. When a parameter is a live instance
// of class C and C has exactly one outgoing association to the created peer class,
// the simulator also links that parameter instance to the new peer.
func (e *ActionExecutor) applyInferredSecondaryLinks(pc DeferredPeerCreation, newPeerID instance.ID) error {
	if e.sch == nil {
		return nil
	}
	paramSources := make([]object.Object, 0, len(pc.ActionParams)+len(pc.Params))
	for _, v := range pc.ActionParams {
		paramSources = append(paramSources, v)
	}
	for _, v := range pc.Params {
		paramSources = append(paramSources, v)
	}
	if len(paramSources) == 0 {
		return nil
	}
	simState := e.bindingsBuilder.State()
	for _, paramVal := range paramSources {
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
		candidates := e.sch.OutgoingAssociationsTo(fromInst.GetClassKey(), pc.ToClassKey)
		if len(candidates) != 1 {
			// Zero or ambiguous associations: do not invent links.
			continue
		}
		assoc := candidates[0]
		// Skip when already linked (e.g. creation action reverse state_change established it).
		if slices.Contains(simState.GetLinkedForward(fromID, assoc.Key), newPeerID) {
			continue
		}
		if err := simState.AddLink(assoc.Key, fromID, newPeerID); err != nil {
			return fmt.Errorf("inferred secondary link after set-add: %w", err)
		}
	}
	return nil
}

func instanceIDFromObject(simState *instance.State, val object.Object) (instance.ID, bool) {
	rec, ok := val.(*object.Record)
	if !ok || rec == nil {
		return 0, false
	}
	return simState.LookupIDByRecord(rec)
}

package actions

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
)

// associationClassReifyWork holds resolved association-class reify targets for one guarantee.
// Target of the guarantee is the association class TLA name; host association is looked up
// because each class reifies at most one association.
type associationClassReifyWork struct {
	target          string
	assocKey        identity.Key
	assoc           model_class.Association
	acCreationEvent model_state.Event
	// eventCall is non-nil for singleton _new(...); setMap is non-nil for bulk set-map form.
	eventCall      *me.EventCall
	setMap         *me.SetMap
	endpointExpr   me.Expression
	endpointBinder string // set-map variable when bulk; empty for singleton
}

// tryQueueAssociationClassReifyGuarantee handles state_change with endpoint_selector set:
// Target is association class name; Spec is AC creation EventCall or { _new(...) : r \in Domain };
// endpoint_selector names the far side (with set-map binder in scope when Spec is a set-map).
func (e *ActionExecutor) tryQueueAssociationClassReifyGuarantee(
	ctx *ExecutionContext,
	instance *state.ClassInstance,
	guar model_logic.Logic,
	bindings *evaluator.Bindings,
) (bool, error) {
	if !model_logic.IsAssociationClassReify(guar) {
		return false, nil
	}
	work, err := e.resolveAssociationClassReifyWork(instance, guar)
	if err != nil {
		return false, err
	}
	if work.setMap != nil {
		return true, e.queueSetMapAssociationClassReify(ctx, instance, work, bindings)
	}
	return true, e.queueOneAssociationClassReify(ctx, instance, work, work.eventCall, bindings)
}

func (e *ActionExecutor) resolveAssociationClassReifyWork(
	instance *state.ClassInstance,
	guar model_logic.Logic,
) (associationClassReifyWork, error) {
	var empty associationClassReifyWork
	if e.peerCatalog == nil {
		return empty, fmt.Errorf("association-class reify on %q: peer catalog not configured", guar.Target)
	}
	assocKey, assoc, found := e.peerCatalog.OutgoingAssociationByAssociationClassTLAName(instance.ClassKey, guar.Target)
	if !found {
		return empty, fmt.Errorf(
			"association-class reify on %q: no outgoing association with association class named %q",
			guar.Target, guar.Target,
		)
	}
	if assoc.AssociationClassKey == nil {
		return empty, fmt.Errorf("association-class reify on %q: association has no association class", guar.Target)
	}
	acCreationEvent, ok := e.peerCatalog.PeerCreationEvent(*assoc.AssociationClassKey)
	if !ok {
		return empty, fmt.Errorf("association-class reify on %q: association class has no creation event", guar.Target)
	}
	endpointExpr := guar.EndpointSelectorSpec.Expression
	if endpointExpr == nil {
		return empty, fmt.Errorf("association-class reify on %q: endpoint_selector not lowered", guar.Target)
	}

	work := associationClassReifyWork{
		target:          guar.Target,
		assocKey:        assocKey,
		assoc:           assoc,
		acCreationEvent: acCreationEvent,
		endpointExpr:    endpointExpr,
	}

	expr := guar.Spec.Expression
	if expr == nil {
		return empty, fmt.Errorf("association-class reify on %q: creation specification not lowered", guar.Target)
	}
	switch e := expr.(type) {
	case *me.EventCall:
		if !isSystemCreationEventCall(e) {
			return empty, fmt.Errorf("association-class reify on %q: specification must be _new / «new»", guar.Target)
		}
		work.eventCall = e
	case *me.SetMap:
		eventCall, ok := e.Transform.(*me.EventCall)
		if !ok || !isSystemCreationEventCall(eventCall) {
			return empty, fmt.Errorf("association-class reify on %q: set-map transform must be _new / «new»", guar.Target)
		}
		work.setMap = e
		work.eventCall = eventCall
		work.endpointBinder = e.Variable
	default:
		return empty, fmt.Errorf("association-class reify on %q: specification must be _new(...) or a set-map of _new", guar.Target)
	}
	return work, nil
}

func (e *ActionExecutor) queueSetMapAssociationClassReify(
	ctx *ExecutionContext,
	instance *state.ClassInstance,
	work associationClassReifyWork,
	bindings *evaluator.Bindings,
) error {
	domainResult := evaluator.Eval(work.setMap.Set, bindings)
	if domainResult.IsError() {
		return fmt.Errorf("association-class reify on %q: domain: %s", work.target, domainResult.Error.Inspect())
	}
	domainSet, ok := evaluator.CoerceToSet(domainResult.Value)
	if !ok {
		return fmt.Errorf("association-class reify on %q: domain must be a set", work.target)
	}
	for _, elem := range domainSet.Elements() {
		child := evaluator.NewEnclosedBindings(bindings)
		child.Set(work.endpointBinder, elem, evaluator.NamespaceLocal)
		if err := e.queueOneAssociationClassReify(ctx, instance, work, work.eventCall, child); err != nil {
			return err
		}
	}
	return nil
}

func (e *ActionExecutor) queueOneAssociationClassReify(
	ctx *ExecutionContext,
	instance *state.ClassInstance,
	work associationClassReifyWork,
	eventCall *me.EventCall,
	bindings *evaluator.Bindings,
) error {
	toResult := evaluator.Eval(work.endpointExpr, bindings)
	if toResult.IsError() {
		return fmt.Errorf("endpoint_selector: %s", toResult.Error.Inspect())
	}
	toID, ok := resolveToEndpointInstanceID(e.bindingsBuilder.State(), work.assoc.ToClassKey, toResult.Value)
	if !ok {
		return fmt.Errorf("endpoint_selector does not identify a live instance of class %s", work.assoc.ToClassKey.String())
	}
	params, err := resolvePositionalEventCallParams(
		work.endpointBinder, work.acCreationEvent.ParameterNames, eventCall, bindings,
	)
	if err != nil {
		return fmt.Errorf("creation parameters: %w", err)
	}
	toIDCopy := toID
	ctx.AddPeerCreation(DeferredPeerCreation{
		FromInstanceID: instance.ID,
		AssocKey:       work.assocKey,
		ToClassKey:     work.assoc.ToClassKey,
		ToInstanceID:   &toIDCopy,
		Params:         params,
	})
	return nil
}

// resolveToEndpointInstanceID resolves a to-side value to a live instance id.
func resolveToEndpointInstanceID(
	simState *state.SimulationState,
	toClassKey identity.Key,
	val object.Object,
) (state.InstanceID, bool) {
	rec, ok := val.(*object.Record)
	if !ok {
		return 0, false
	}
	if id, ok := liveInstanceIDFromExtent(simState, toClassKey, rec); ok {
		return id, true
	}
	if id, ok := matchLiveInstanceByData(simState, toClassKey, rec); ok {
		return id, true
	}
	return discoverToEndpointFromRow(simState, toClassKey, rec)
}

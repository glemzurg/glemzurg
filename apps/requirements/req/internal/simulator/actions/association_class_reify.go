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
// Target is the association class TLA name. endpoint_selector is LET-like: a set-map
// { endpoint : r \in Domain } supplies the binder and domain; Spec is a bare _new(...).
type associationClassReifyWork struct {
	target          string
	assocKey        identity.Key
	assoc           model_class.Association
	acCreationEvent model_state.Event
	eventCall       *me.EventCall
	// selectorMap non-nil when endpoint_selector is a set-map over a domain.
	selectorMap    *me.SetMap
	endpointExpr   me.Expression // transform of set-map, or whole selector when singleton
	endpointBinder string
}

// tryQueueAssociationClassReifyGuarantee handles state_change with endpoint_selector set.
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
	if work.selectorMap != nil {
		return true, e.queueSelectorMapAssociationClassReify(ctx, instance, work, bindings)
	}
	return true, e.queueOneAssociationClassReify(ctx, instance, work, bindings)
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

	eventCall, err := associationClassCreationEventCall(guar)
	if err != nil {
		return empty, fmt.Errorf("association-class reify on %q: %w", guar.Target, err)
	}
	selectorExpr := guar.EndpointSelectorSpec.Expression
	if selectorExpr == nil {
		return empty, fmt.Errorf("association-class reify on %q: endpoint_selector not lowered", guar.Target)
	}

	work := associationClassReifyWork{
		target:          guar.Target,
		assocKey:        assocKey,
		assoc:           assoc,
		acCreationEvent: acCreationEvent,
		eventCall:       eventCall,
	}

	// LET-like set-map: { endpointExpr : r \in Domain } — domain and binder from selector.
	if setMap, ok := selectorExpr.(*me.SetMap); ok {
		work.selectorMap = setMap
		work.endpointExpr = setMap.Transform
		work.endpointBinder = setMap.Variable
		return work, nil
	}
	// Singleton: endpoint_selector is a single peer expression.
	work.endpointExpr = selectorExpr
	return work, nil
}

func associationClassCreationEventCall(guar model_logic.Logic) (*me.EventCall, error) {
	expr := guar.Spec.Expression
	if expr == nil {
		return nil, fmt.Errorf("creation specification not lowered")
	}
	eventCall, ok := expr.(*me.EventCall)
	if !ok {
		return nil, fmt.Errorf("creation specification must be an event call (_new(...)), got %T", expr)
	}
	if !isSystemCreationEventCall(eventCall) {
		return nil, fmt.Errorf("creation specification must be a system creation event (_new / «new»)")
	}
	return eventCall, nil
}

func (e *ActionExecutor) queueSelectorMapAssociationClassReify(
	ctx *ExecutionContext,
	instance *state.ClassInstance,
	work associationClassReifyWork,
	bindings *evaluator.Bindings,
) error {
	domainResult := evaluator.Eval(work.selectorMap.Set, bindings)
	if domainResult.IsError() {
		return fmt.Errorf("association-class reify on %q: selector domain: %s", work.target, domainResult.Error.Inspect())
	}
	domainSet, ok := evaluator.CoerceToSet(domainResult.Value)
	if !ok {
		return fmt.Errorf("association-class reify on %q: selector domain must be a set", work.target)
	}
	for _, elem := range domainSet.Elements() {
		child := evaluator.NewEnclosedBindings(bindings)
		child.Set(work.endpointBinder, elem, evaluator.NamespaceLocal)
		if err := e.queueOneAssociationClassReify(ctx, instance, work, child); err != nil {
			return err
		}
	}
	return nil
}

func (e *ActionExecutor) queueOneAssociationClassReify(
	ctx *ExecutionContext,
	instance *state.ClassInstance,
	work associationClassReifyWork,
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
		work.endpointBinder, work.acCreationEvent.ParameterNames, work.eventCall, bindings,
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

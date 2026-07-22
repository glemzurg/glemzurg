package actions

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/instance"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
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
// When the association class is outside the simulation surface, reify is a no-op: the host
// association is plain and endpoint links come from the host field's state_change guarantee.
func (e *ActionExecutor) tryQueueAssociationClassReifyGuarantee(
	ctx *ExecutionContext,
	instance *instance.Instance,
	guar model_logic.Logic,
	bindings *evaluator.Bindings,
) (bool, error) {
	if !model_logic.IsAssociationClassReify(guar) {
		return false, nil
	}
	work, active, err := e.resolveAssociationClassReifyWork(instance, guar)
	if err != nil {
		return false, err
	}
	if !active {
		// Association class not on surface — host association degraded to plain links.
		return true, nil
	}
	if work.selectorMap != nil {
		return true, e.queueSelectorMapAssociationClassReify(ctx, instance, work, bindings)
	}
	return true, e.queueOneAssociationClassReify(ctx, instance, work, bindings)
}

// resolveAssociationClassReifyWork returns active=false when the association class is not
// present (or not creatable) on the surface catalog — reify becomes a no-op.
func (e *ActionExecutor) resolveAssociationClassReifyWork(
	instance *instance.Instance,
	guar model_logic.Logic,
) (work associationClassReifyWork, active bool, err error) {
	if e.peerCatalog == nil {
		return work, false, fmt.Errorf("association-class reify on %q: peer catalog not configured", guar.Target)
	}
	assocKey, assoc, found := e.peerCatalog.OutgoingAssociationByAssociationClassTLAName(instance.ClassKey, guar.Target)
	if !found {
		// Host association may be plain after surface strip of the association class.
		return work, false, nil
	}
	if assoc.AssociationClassKey == nil {
		return work, false, fmt.Errorf("association-class reify on %q: association has no association class", guar.Target)
	}
	acCreationEvent, ok := e.peerCatalog.PeerCreationEvent(*assoc.AssociationClassKey)
	if !ok {
		// Association class present as key but not creatable on this surface — skip reify.
		return work, false, nil
	}

	eventCall, err := associationClassCreationEventCall(guar)
	if err != nil {
		return work, false, fmt.Errorf("association-class reify on %q: %w", guar.Target, err)
	}
	selectorExpr := guar.EndpointSelectorSpec.Expression
	if selectorExpr == nil {
		return work, false, fmt.Errorf("association-class reify on %q: endpoint_selector not lowered", guar.Target)
	}

	work = associationClassReifyWork{
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
		return work, true, nil
	}
	// Singleton: endpoint_selector is a single peer expression.
	work.endpointExpr = selectorExpr
	return work, true, nil
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
	instance *instance.Instance,
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
	instance *instance.Instance,
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
	simState *instance.State,
	toClassKey identity.Key,
	val object.Object,
) (instance.ID, bool) {
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

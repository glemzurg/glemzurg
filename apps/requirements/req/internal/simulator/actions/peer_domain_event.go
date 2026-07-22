package actions

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/instance"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
)

// matchPeerDomainEventSetMap recognizes { Event(args) : x \in Domain } where Domain is
// not an association ref (parameter set, set comprehension, etc.).
func matchPeerDomainEventSetMap(expr me.Expression) (*me.SetMap, *me.EventCall, bool) {
	setMap, ok := expr.(*me.SetMap)
	if !ok {
		return nil, nil, false
	}
	if _, isAssoc := setMap.Set.(*me.AssociationRef); isAssoc {
		return nil, nil, false
	}
	eventCall, ok := setMap.Transform.(*me.EventCall)
	if !ok {
		return nil, nil, false
	}
	return setMap, eventCall, true
}

// tryQueuePeerDomainEventSetMap queues peer updates for each instance in an evaluated domain set.
func (e *ActionExecutor) tryQueuePeerDomainEventSetMap(
	ctx *ExecutionContext,
	instance *instance.Instance,
	expr me.Expression,
	bindings *evaluator.Bindings,
) (bool, error) {
	setMap, eventCall, ok := matchPeerDomainEventSetMap(expr)
	if !ok {
		return false, nil
	}
	if e.peerCatalog == nil {
		return false, fmt.Errorf("peer domain event: peer catalog not configured")
	}

	domainResult := evaluator.Eval(setMap.Set, bindings)
	if domainResult.IsError() {
		return false, fmt.Errorf("peer domain event domain: %s", domainResult.Error.Inspect())
	}
	domainSet, ok := evaluator.CoerceToSet(domainResult.Value)
	if !ok {
		return false, fmt.Errorf("peer domain event: domain must evaluate to a set")
	}

	toClass, event, found := e.findPeerEventByKey(eventCall.EventKey)
	if !found {
		return false, fmt.Errorf("peer domain event: cannot resolve event %s", eventCall.EventKey.String())
	}
	if model_state.IsSystemFinalEvent(event.Name) {
		return false, fmt.Errorf("peer domain event: _destroy must use guarantee type destroy")
	}
	return e.queuePeerDomainUpdates(ctx, instance, peerDomainEventWork{
		toClass: toClass, event: event, setMap: setMap, eventCall: eventCall,
	}, domainSet, bindings)
}

func (e *ActionExecutor) findPeerEventByKey(eventKey identity.Key) (model_class.Class, model_state.Event, bool) {
	classKey, err := identity.ParseKey(eventKey.ParentKey)
	if err != nil {
		return model_class.Class{}, model_state.Event{}, false
	}
	toClass, ok := e.peerCatalog.PeerClass(classKey)
	if !ok {
		return model_class.Class{}, model_state.Event{}, false
	}
	event, ok := e.peerCatalog.PeerEvent(classKey, eventKey)
	return toClass, event, ok
}

// peerDomainEventWork is the resolved peer class event and set-map over a domain.
type peerDomainEventWork struct {
	toClass   model_class.Class
	event     model_state.Event
	setMap    *me.SetMap
	eventCall *me.EventCall
}

func (e *ActionExecutor) queuePeerDomainUpdates(
	ctx *ExecutionContext,
	owner *instance.Instance,
	work peerDomainEventWork,
	domainSet *object.Set,
	bindings *evaluator.Bindings,
) (bool, error) {
	vctx := peerEventViolationContext{
		OwnerInstanceID: owner.ID,
		OwnerClassKey:   owner.ClassKey,
		AssociationName: "",
	}
	for _, elem := range domainSet.Elements() {
		e.queueOnePeerDomainUpdate(ctx, owner, work, elem, bindings, vctx)
	}
	return true, nil
}

func (e *ActionExecutor) queueOnePeerDomainUpdate(
	ctx *ExecutionContext,
	owner *instance.Instance,
	work peerDomainEventWork,
	elem object.Object,
	bindings *evaluator.Bindings,
	vctx peerEventViolationContext,
) {
	simState := e.bindingsBuilder.State()
	peerID, ok := instanceIDFromObject(simState, elem)
	if !ok {
		return
	}
	peerInstance := simState.GetInstance(peerID)
	if peerInstance == nil || peerInstance.ClassKey != work.toClass.Key {
		return
	}
	params, ok := e.bindPeerDomainEventParams(work, elem, bindings, owner)
	if !ok {
		e.recordPeerEventUnavailable(ctx, vctx, work.toClass, peerID, work.event.Key, work.event.Name)
		return
	}
	if !e.peerEventAvailable(work.toClass, peerInstance, work.event.Key) {
		e.recordPeerEventUnavailable(ctx, vctx, work.toClass, peerID, work.event.Key, work.event.Name)
		return
	}
	ctx.AddPeerUpdate(DeferredPeerUpdate{
		OwnerInstanceID: owner.ID,
		PeerInstanceID:  peerID,
		ToClassKey:      work.toClass.Key,
		EventKey:        work.event.Key,
		EventName:       work.event.Name,
		Params:          params,
	})
}

func (e *ActionExecutor) bindPeerDomainEventParams(
	work peerDomainEventWork,
	elem object.Object,
	bindings *evaluator.Bindings,
	owner *instance.Instance,
) (map[string]object.Object, bool) {
	child := evaluator.NewEnclosedBindings(bindings)
	if work.setMap.Variable != "" {
		child.Set(work.setMap.Variable, elem, evaluator.NamespaceLocal)
	}
	params, err := resolvePositionalEventCallParams(work.setMap.Variable, work.event.ParameterNames, work.eventCall, child)
	if err != nil {
		return nil, false
	}
	// self is bare attributes; reify owner identity for secondary-link inference.
	return reifyOwnerSelfParams(params, owner), true
}

// reifyOwnerSelfParams replaces bare records equal to the owner instance's attributes
// with [id, data] extent elements so later identity-sensitive work sees the owner id.
func reifyOwnerSelfParams(
	params map[string]object.Object,
	owner *instance.Instance,
) map[string]object.Object {
	if owner == nil || len(params) == 0 {
		return params
	}
	out := make(map[string]object.Object, len(params))
	for name, val := range params {
		rec, ok := val.(*object.Record)
		if !ok || rec == nil || object.IsExtentElement(rec) {
			out[name] = val
			continue
		}
		if rec == owner.Attributes ||
			(owner.Attributes != nil && owner.Attributes.Equals(rec)) ||
			(state.DataFromExtentElement(rec) != nil && owner.Attributes != nil &&
				owner.Attributes.Equals(state.DataFromExtentElement(rec))) {
			out[name] = state.ClassExtentElement(owner.ID, owner.Attributes)
			continue
		}
		out[name] = val
	}
	return out
}

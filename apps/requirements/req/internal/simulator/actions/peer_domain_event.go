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
// Receiver is the first object argument of Event (not the domain element).
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

// tryQueuePeerDomainEventSetMap queues peer updates for each domain element using
// receiver-first messaging: Event(receiver, param…). Domain only supplies binders.
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
	if e.sch == nil {
		return false, fmt.Errorf("event set-map: peer catalog not configured")
	}
	if model_state.IsSystemCreationEvent(eventCall.EventKey.SubKey) ||
		eventCall.EventKey.SubKey == model_state.EventTLANameNew {
		// _new belongs in association set-add / multi set-add, not receiver-first messaging.
		return false, nil
	}

	domainResult := evaluator.Eval(setMap.Set, bindings)
	if domainResult.IsError() {
		return false, fmt.Errorf("event set-map domain: %s", domainResult.Error.Inspect())
	}
	domainSet, ok := evaluator.CoerceToSet(domainResult.Value)
	if !ok {
		return false, fmt.Errorf("event set-map: domain must evaluate to a set")
	}

	// Event key from lower may be a same-named event on another class; each receiver
	// re-resolves the event by name on its own class at queue time.
	_, event, found := e.findPeerEventByKey(eventCall.EventKey)
	if !found {
		// Fall back to SubKey as event name when key is only a name hint.
		event = model_state.Event{Key: eventCall.EventKey, Name: eventCall.EventKey.SubKey}
	}
	if model_state.IsSystemFinalEvent(event.Name) {
		return false, fmt.Errorf("event set-map: _destroy must use guarantee type destroy")
	}
	if model_state.IsSystemCreationEvent(event.Name) {
		return false, nil
	}
	if len(eventCall.Args) < 1 {
		return false, fmt.Errorf(
			"event set-map: %s requires a receiver as the first argument (Event(receiver, …params))",
			event.Name,
		)
	}
	return e.queueReceiverFirstEventUpdates(ctx, instance, peerDomainEventWork{
		eventName: event.Name, setMap: setMap, eventCall: eventCall,
	}, domainSet, bindings)
}

func (e *ActionExecutor) findPeerEventByKey(eventKey identity.Key) (model_class.Class, model_state.Event, bool) {
	classKey, err := identity.ParseKey(eventKey.ParentKey)
	if err != nil {
		return model_class.Class{}, model_state.Event{}, false
	}
	toClass, ok := e.sch.PeerClass(classKey)
	if !ok {
		return model_class.Class{}, model_state.Event{}, false
	}
	event, ok := e.sch.PeerEvent(classKey, eventKey)
	return toClass, event, ok
}

// peerDomainEventWork is an event set-map; the event is resolved per receiver by name.
type peerDomainEventWork struct {
	eventName string
	setMap    *me.SetMap
	eventCall *me.EventCall
}

func (e *ActionExecutor) queueReceiverFirstEventUpdates(
	ctx *ExecutionContext,
	owner *instance.Instance,
	work peerDomainEventWork,
	domainSet *object.Set,
	bindings *evaluator.Bindings,
) (bool, error) {
	vctx := peerEventViolationContext{
		OwnerInstanceID: owner.GetID(),
		OwnerClassKey:   owner.GetClassKey(),
		AssociationName: "",
	}
	for _, elem := range domainSet.Elements() {
		if err := e.queueOneReceiverFirstEventUpdate(ctx, owner, work, elem, bindings, vctx); err != nil {
			return false, err
		}
	}
	return true, nil
}

func (e *ActionExecutor) queueOneReceiverFirstEventUpdate(
	ctx *ExecutionContext,
	owner *instance.Instance,
	work peerDomainEventWork,
	elem object.Object,
	bindings *evaluator.Bindings,
	vctx peerEventViolationContext,
) error {
	child := evaluator.NewEnclosedBindings(bindings)
	if work.setMap.Variable != "" {
		child.Set(work.setMap.Variable, elem, evaluator.NamespaceLocal)
	}
	receiverInst, err := e.resolveEventSetMapReceiver(work, child)
	if err != nil {
		return err
	}
	if receiverInst == nil {
		return nil
	}
	toClass, event, found := e.findEventByNameOnClass(receiverInst.GetClassKey(), work.eventName)
	if !found {
		return nil
	}
	return e.queueResolvedReceiverFirstUpdate(receiverFirstResolved{
		ctx: ctx, owner: owner, work: work, child: child,
		receiverInst: receiverInst, toClass: toClass, event: event, vctx: vctx,
	})
}

type receiverFirstResolved struct {
	ctx          *ExecutionContext
	owner        *instance.Instance
	work         peerDomainEventWork
	child        *evaluator.Bindings
	receiverInst *instance.Instance
	toClass      model_class.Class
	event        model_state.Event
	vctx         peerEventViolationContext
}

func (e *ActionExecutor) queueResolvedReceiverFirstUpdate(r receiverFirstResolved) error {
	params, err := bindEventParamsAfterReceiver(r.event, r.work.eventCall, r.child, r.owner)
	if err != nil {
		e.recordPeerEventUnavailable(r.ctx, r.vctx, r.toClass, r.receiverInst.GetID(), r.event.Key, r.event.Name)
		return nil
	}
	if !e.peerEventAvailable(r.toClass, r.receiverInst, r.event.Key) {
		e.recordPeerEventUnavailable(r.ctx, r.vctx, r.toClass, r.receiverInst.GetID(), r.event.Key, r.event.Name)
		return nil
	}
	r.ctx.AddPeerUpdate(DeferredPeerUpdate{
		OwnerInstanceID: r.owner.GetID(),
		PeerInstanceID:  r.receiverInst.GetID(),
		ToClassKey:      r.toClass.Key,
		EventKey:        r.event.Key,
		EventName:       r.event.Name,
		Params:          params,
	})
	return nil
}

func (e *ActionExecutor) resolveEventSetMapReceiver(
	work peerDomainEventWork,
	bindings *evaluator.Bindings,
) (*instance.Instance, error) {
	simState := e.bindingsBuilder.State()
	receiverVal, err := evalEventCallArg(work.eventCall.Args[0], bindings, "receiver")
	if err != nil {
		return nil, fmt.Errorf("event set-map %s: %w", work.eventName, err)
	}
	receiverID, ok := instanceIDFromObject(simState, receiverVal)
	if !ok {
		return nil, fmt.Errorf(
			"event set-map %s: first argument must be an object instance (receiver)",
			work.eventName,
		)
	}
	receiverInst := simState.GetInstance(receiverID)
	if receiverInst == nil {
		return nil, fmt.Errorf("event set-map %s: receiver instance %d not found", work.eventName, receiverID)
	}
	return receiverInst, nil
}

func (e *ActionExecutor) findEventByNameOnClass(classKey identity.Key, eventName string) (model_class.Class, model_state.Event, bool) {
	toClass, ok := e.sch.PeerClass(classKey)
	if !ok {
		return model_class.Class{}, model_state.Event{}, false
	}
	for _, ev := range toClass.Events {
		if ev.Name == eventName {
			return toClass, ev, true
		}
	}
	return model_class.Class{}, model_state.Event{}, false
}

// bindEventParamsAfterReceiver evaluates EventCall args after the leading receiver
// and binds them positionally to the event's ParameterNames.
func bindEventParamsAfterReceiver(
	event model_state.Event,
	eventCall *me.EventCall,
	bindings *evaluator.Bindings,
	owner *instance.Instance,
) (map[string]object.Object, error) {
	paramNames := event.ParameterNames
	valueArgs := eventCall.Args[1:]
	if len(valueArgs) != len(paramNames) {
		return nil, fmt.Errorf(
			"event call supplies %d parameters after receiver but event declares %d",
			len(valueArgs), len(paramNames),
		)
	}
	params := make(map[string]object.Object, len(paramNames))
	for i, paramName := range paramNames {
		result := evaluator.Eval(valueArgs[i], bindings)
		if result.IsError() {
			return nil, fmt.Errorf("event argument %d for parameter %q: %s", i, paramName, result.Error.Inspect())
		}
		params[paramName] = result.Value
	}
	return reifyOwnerSelfParams(params, owner), nil
}

func evalEventCallArg(arg me.Expression, bindings *evaluator.Bindings, label string) (object.Object, error) {
	result := evaluator.Eval(arg, bindings)
	if result.IsError() {
		return nil, fmt.Errorf("%s: %s", label, result.Error.Inspect())
	}
	return result.Value, nil
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
		if rec == owner.GetAttributes() ||
			(owner.GetAttributes() != nil && owner.GetAttributes().Equals(rec)) ||
			(state.DataFromExtentElement(rec) != nil && owner.GetAttributes() != nil &&
				owner.GetAttributes().Equals(state.DataFromExtentElement(rec))) {
			out[name] = state.ClassExtentElement(owner.GetID(), owner.GetAttributes())
			continue
		}
		out[name] = val
	}
	return out
}

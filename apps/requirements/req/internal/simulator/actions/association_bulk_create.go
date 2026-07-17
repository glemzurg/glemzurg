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

// tryQueueAssociationBulkCreateFromSet recognizes association guarantees of the form
//
//	{ _new(…) : r \in Rows }
//
// where Rows evaluates to a set of records, the association has an association class,
// and _new is a system creation constructor. Each row supplies:
//   - to-endpoint: a field whose value is a class-extent element [id |-> …, data |-> …]
//     for the association's to-class (or a bare data record matching a live instance)
//   - AC creation parameters: remaining EventCall args evaluated with r bound
//
// Model-agnostic: domain fields and event parameter names come from the expression and
// the association class creation event, not from any fixed domain vocabulary.
//
// EventCall keys from lower often carry the owning class's _new identity; the association
// class creation event is used for materialization when the call is a system creation.
func (e *ActionExecutor) tryQueueAssociationBulkCreateFromSet(
	ctx *ExecutionContext,
	instance *state.ClassInstance,
	target string,
	expr me.Expression,
	bindings *evaluator.Bindings,
) (bool, error) {
	plan, ok := e.matchAssociationBulkCreate(instance, target, expr)
	if !ok {
		return false, nil
	}
	domainSet, err := evalBulkCreateDomain(target, plan.setMap.Set, bindings)
	if err != nil {
		return false, err
	}
	if err := e.queueBulkCreatePeerCreations(ctx, instance, target, plan, domainSet, bindings); err != nil {
		return false, err
	}
	return true, nil
}

// associationBulkCreatePlan is the matched shape for AC bulk-create from a set of rows.
type associationBulkCreatePlan struct {
	setMap          *me.SetMap
	eventCall       *me.EventCall
	assocKey        identity.Key
	assoc           model_class.Association
	acCreationEvent model_state.Event
}

func (e *ActionExecutor) matchAssociationBulkCreate(
	instance *state.ClassInstance,
	target string,
	expr me.Expression,
) (associationBulkCreatePlan, bool) {
	var empty associationBulkCreatePlan
	setMap, ok := expr.(*me.SetMap)
	if !ok {
		return empty, false
	}
	if _, isAssoc := setMap.Set.(*me.AssociationRef); isAssoc {
		return empty, false
	}
	eventCall, ok := setMap.Transform.(*me.EventCall)
	if !ok || e.peerCatalog == nil || !isSystemCreationEventCall(eventCall) {
		return empty, false
	}
	assocKey, assoc, found := e.peerCatalog.OutgoingAssociationByTLAField(instance.ClassKey, target)
	if !found || assoc.AssociationClassKey == nil {
		return empty, false
	}
	// Materialization always uses the association class creation event from the catalog.
	acCreationEvent, ok := e.peerCatalog.PeerCreationEvent(*assoc.AssociationClassKey)
	if !ok {
		return empty, false
	}
	return associationBulkCreatePlan{
		setMap:          setMap,
		eventCall:       eventCall,
		assocKey:        assocKey,
		assoc:           assoc,
		acCreationEvent: acCreationEvent,
	}, true
}

func evalBulkCreateDomain(
	target string,
	domainExpr me.Expression,
	bindings *evaluator.Bindings,
) (*object.Set, error) {
	domainResult := evaluator.Eval(domainExpr, bindings)
	if domainResult.IsError() {
		return nil, fmt.Errorf("association bulk create on %q: domain: %s", target, domainResult.Error.Inspect())
	}
	domainSet, ok := evaluator.CoerceToSet(domainResult.Value)
	if !ok {
		return nil, fmt.Errorf("association bulk create on %q: domain must be a set, got %s", target, domainResult.Value.Type())
	}
	return domainSet, nil
}

// bulkCreateQueueEnv holds fixed context while iterating domain rows.
type bulkCreateQueueEnv struct {
	ctx      *ExecutionContext
	instance *state.ClassInstance
	target   string
	plan     associationBulkCreatePlan
	simState *state.SimulationState
	bindings *evaluator.Bindings
}

func (e *ActionExecutor) queueBulkCreatePeerCreations(
	ctx *ExecutionContext,
	instance *state.ClassInstance,
	target string,
	plan associationBulkCreatePlan,
	domainSet *object.Set,
	bindings *evaluator.Bindings,
) error {
	env := bulkCreateQueueEnv{
		ctx:      ctx,
		instance: instance,
		target:   target,
		plan:     plan,
		simState: e.bindingsBuilder.State(),
		bindings: bindings,
	}
	for _, elem := range domainSet.Elements() {
		if err := queueOneBulkCreateRow(env, elem); err != nil {
			return err
		}
	}
	return nil
}

func queueOneBulkCreateRow(env bulkCreateQueueEnv, elem object.Object) error {
	row, ok := elem.(*object.Record)
	if !ok {
		return fmt.Errorf("association bulk create on %q: domain elements must be records", env.target)
	}
	child := evaluator.NewEnclosedBindings(env.bindings)
	child.Set(env.plan.setMap.Variable, elem, evaluator.NamespaceLocal)

	toID, ok := discoverToEndpointFromRow(env.simState, env.plan.assoc.ToClassKey, row)
	if !ok {
		return fmt.Errorf(
			"association bulk create on %q: row has no to-endpoint for class %s",
			env.target, env.plan.assoc.ToClassKey.String(),
		)
	}
	params, err := resolvePositionalEventCallParams(
		env.plan.setMap.Variable, env.plan.acCreationEvent.ParameterNames, env.plan.eventCall, child,
	)
	if err != nil {
		return fmt.Errorf("association bulk create on %q: %w", env.target, err)
	}
	toIDCopy := toID
	env.ctx.AddPeerCreation(DeferredPeerCreation{
		FromInstanceID: env.instance.ID,
		AssocKey:       env.plan.assocKey,
		ToClassKey:     env.plan.assoc.ToClassKey,
		ToInstanceID:   &toIDCopy,
		Params:         params,
	})
	return nil
}

// isSystemCreationEventCall reports whether the EventCall is a system creation (_new / «new»).
// Event keys from lower may belong to the owning class; only the event name is significant here.
func isSystemCreationEventCall(eventCall *me.EventCall) bool {
	if eventCall == nil {
		return false
	}
	name := eventCall.EventKey.SubKey
	return model_state.IsSystemCreationEvent(name) || name == model_state.EventTLANameNew
}

// discoverToEndpointFromRow finds a live to-class instance referenced by a bulk-create row.
// Prefers class-extent elements [id |-> N, data |-> …]; falls back to structural data match.
func discoverToEndpointFromRow(
	simState *state.SimulationState,
	toClassKey identity.Key,
	row *object.Record,
) (state.InstanceID, bool) {
	if id, ok := liveInstanceIDFromExtent(simState, toClassKey, row); ok {
		return id, true
	}
	for _, name := range row.FieldNames() {
		val := row.Get(name)
		rec, ok := val.(*object.Record)
		if !ok {
			continue
		}
		if id, ok := liveInstanceIDFromExtent(simState, toClassKey, rec); ok {
			return id, true
		}
		if id, ok := matchLiveInstanceByData(simState, toClassKey, rec); ok {
			return id, true
		}
	}
	return 0, false
}

func liveInstanceIDFromExtent(
	simState *state.SimulationState,
	toClassKey identity.Key,
	rec *object.Record,
) (state.InstanceID, bool) {
	id, ok := state.InstanceIDFromExtentElement(rec)
	if !ok {
		return 0, false
	}
	inst := simState.GetInstance(id)
	if inst == nil || inst.ClassKey != toClassKey {
		return 0, false
	}
	return id, true
}

func matchLiveInstanceByData(
	simState *state.SimulationState,
	toClassKey identity.Key,
	rec *object.Record,
) (state.InstanceID, bool) {
	data := state.DataFromExtentElement(rec)
	for _, inst := range simState.InstancesByClass(toClassKey) {
		if inst.Attributes == rec || inst.Attributes == data ||
			(data != nil && inst.Attributes.Equals(data)) ||
			inst.Attributes.Equals(rec) {
			return inst.ID, true
		}
	}
	return 0, false
}

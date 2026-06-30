package actions

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
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
) (bool, error) {
	assocRef, eventCall, ok := model_class.MatchAssociationSetAddExpr(expr)
	if !ok {
		return false, nil
	}
	if e.peerCatalog == nil {
		return false, fmt.Errorf("association set-add guarantee on %q: peer catalog not configured", target)
	}
	assocKey, assoc, found := e.peerCatalog.OutgoingAssociationByTLAField(instance.ClassKey, target)
	if !found {
		return false, fmt.Errorf(
			"association set-add guarantee on %q: no outgoing association on class %s",
			target, instance.ClassKey.String(),
		)
	}
	if assocRef.AssociationKey != assocKey {
		return false, fmt.Errorf(
			"association set-add guarantee on %q: expression association %s does not match target",
			target, assocRef.AssociationKey.String(),
		)
	}
	params, err := resolveEventCallParams(eventCall, bindings)
	if err != nil {
		return false, err
	}
	ctx.AddPeerCreation(DeferredPeerCreation{
		FromInstanceID: instance.ID,
		AssocKey:       assoc.Key,
		ToClassKey:     assoc.ToClassKey,
		Params:         params,
	})
	return true, nil
}

func resolveEventCallParams(eventCall *me.EventCall, bindings *evaluator.Bindings) (map[string]object.Object, error) {
	params := make(map[string]object.Object, len(eventCall.Args))
	for i, arg := range eventCall.Args {
		name, ok := eventCallArgName(arg)
		if !ok {
			return nil, fmt.Errorf("association set-add _new arg[%d]: expected parameter reference", i)
		}
		result := evaluator.Eval(arg, bindings)
		if result.IsError() {
			return nil, fmt.Errorf("association set-add _new arg %q: %s", name, result.Error.Inspect())
		}
		params[name] = result.Value
	}
	return params, nil
}

func eventCallArgName(arg me.Expression) (string, bool) {
	switch a := arg.(type) {
	case *me.LocalVar:
		return a.Name, true
	default:
		return "", false
	}
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
	toClass, ok := e.peerCatalog.PeerClass(pc.ToClassKey)
	if !ok {
		return fmt.Errorf("peer creation for association %s: to-class %s not found", pc.AssocKey.String(), pc.ToClassKey.String())
	}
	creationEvent, ok := e.peerCatalog.PeerCreationEvent(pc.ToClassKey)
	if !ok {
		return fmt.Errorf(
			"peer creation for association %s: to-class %s has no creation event",
			pc.AssocKey.String(), toClass.Name,
		)
	}
	assocKey := pc.AssocKey
	fromID := pc.FromInstanceID
	result, err := e.ExecuteTransition(
		toClass,
		creationEvent,
		nil,
		pc.Params,
		CreationLinkSource{SourceAssocKey: &assocKey, SourceID: &fromID},
		nil,
	)
	if err != nil {
		return err
	}
	e.recordPeerTransition(ctx, toClass, creationEvent, pc.Params, result)
	return nil
}

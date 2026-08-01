package actions

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/instance"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// tryQueueAssociationClassReifyGuarantee recognizes retired host reify guarantees (C1).
// Association-class rows are created only via surface AC _new with host endpoint parameters
// (see design-association-class-surface-new.md). endpoint_selector reify is no longer executed.
func (e *ActionExecutor) tryQueueAssociationClassReifyGuarantee(
	ctx *ExecutionContext,
	instance *instance.Instance,
	guar model_logic.Logic,
	bindings *evaluator.Bindings,
) (bool, error) {
	_ = ctx
	_ = instance
	_ = bindings
	if !model_logic.IsAssociationClassReify(guar) {
		return false, nil
	}
	return true, fmt.Errorf(
		"association-class reify on %q is not supported: create the association class via surface _new with host from/to endpoint parameters",
		guar.Target,
	)
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

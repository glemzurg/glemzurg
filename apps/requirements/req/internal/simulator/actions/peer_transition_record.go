package actions

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// PeerTransitionRecord captures a peer-class transition materialized by an association
// set-add or set-map guarantee on the owning action.
type PeerTransitionRecord struct {
	ClassKey   identity.Key
	ClassName  string
	EventKey   identity.Key
	EventName  string
	Parameters map[string]object.Object
	Result     *TransitionResult
}

func (e *ActionExecutor) recordPeerTransition(
	ctx *ExecutionContext,
	class model_class.Class,
	event model_state.Event,
	params map[string]object.Object,
	result *TransitionResult,
) {
	if ctx == nil || result == nil {
		return
	}
	ctx.AddPeerTransition(PeerTransitionRecord{
		ClassKey:   class.Key,
		ClassName:  class.Name,
		EventKey:   event.Key,
		EventName:  event.Name,
		Parameters: params,
		Result:     result,
	})
}

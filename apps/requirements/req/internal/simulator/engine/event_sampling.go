package engine

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
)

// eventSamplingParameterDefs returns typed parameter defs used for event payload sampling.
// Full event-name-to-action matching is deferred; bound transitions use action parameters.
func eventSamplingParameterDefs(action *model_state.Action) []model_state.Parameter {
	if action != nil && len(action.Parameters) > 0 {
		return action.Parameters
	}
	return nil
}

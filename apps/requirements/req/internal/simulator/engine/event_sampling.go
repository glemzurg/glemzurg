package engine

import (
	"math/rand"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/actions"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// sampleEventPayload generates parameter values for an event, including event-only names.
func sampleEventPayload(
	event *model_state.Event,
	action *model_state.Action,
	gen *StepParameterGenerator,
	rng *rand.Rand,
) (map[string]object.Object, error) {
	if event == nil {
		return map[string]object.Object{}, nil
	}
	return actions.SampleEventPayload(*event, action, gen.Binder, gen.Sampler, rng)
}

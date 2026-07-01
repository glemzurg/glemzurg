package engine

import (
	"math/rand"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/actions"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
)

type surfaceEventSamplingDeps struct {
	paramGen           *StepParameterGenerator
	bindingsBuilder    *state.BindingsBuilder
	catalog            *ClassCatalog
	simulationCoverage *SimulationCoverageTracker
	rng                *rand.Rand
}

func sampleSurfaceEventPayload(
	pending *PendingAction,
	action *model_state.Action,
	deps surfaceEventSamplingDeps,
) (map[string]object.Object, error) {
	if pending.Event == nil {
		return map[string]object.Object{}, nil
	}
	if action != nil && actions.ActionHasParameterSimulation(*action) {
		bindings := actions.BuildSimulationBindings(deps.bindingsBuilder, deps.catalog.ClassNameMap(), pending.Instance)
		var used map[identity.Key]bool
		if deps.simulationCoverage != nil {
			used = deps.simulationCoverage.UsedSimulationParams
		}
		return actions.SampleSurfaceEventPayload(*pending.Event, action, actions.SurfaceEventSamplingContext{
			Binder:               deps.paramGen.Binder,
			Sampler:              deps.paramGen.Sampler,
			Bindings:             bindings,
			UsedSimulationParams: used,
			RNG:                  deps.rng,
		})
	}
	return sampleEventPayload(pending.Event, action, deps.paramGen, deps.rng)
}

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

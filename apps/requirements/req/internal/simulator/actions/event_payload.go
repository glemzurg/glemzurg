package actions

import (
	"math/rand"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// SampleEventPayload builds the parameter map carried by an event.
// Names that match action parameters are sampled with type and requires awareness;
// other event names receive an untyped value (empty set or random string).
func SampleEventPayload(
	event model_state.Event,
	action *model_state.Action,
	binder *ParameterBinder,
	sampler *ParameterSampler,
	rng *rand.Rand,
) (map[string]object.Object, error) {
	matched := matchActionParametersByEventNames(event.ParameterNames, action)

	var sampled map[string]object.Object
	var err error
	if action != nil && len(matched) > 0 && sampler != nil {
		owner := ParameterOwnerFromAction(*action)
		if owner.NeedsRequiresAwareSampling(matched) {
			sampled, err = sampler.SampleParameters(owner, matched, rng)
		} else {
			sampled = binder.GenerateRandomParameters(matched, rng)
		}
	} else {
		sampled = binder.GenerateRandomParameters(matched, rng)
	}
	if err != nil {
		return nil, err
	}

	result := remapSampledToEventNames(event.ParameterNames, matched, sampled)
	for _, name := range event.ParameterNames {
		if _, ok := result[name]; ok {
			continue
		}
		result[name] = binder.GenerateEventOnlyParameterValue(rng)
	}
	return result, nil
}

// GenerateEventOnlyParameterValue samples a value for an event name with no action/query type.
func (b *ParameterBinder) GenerateEventOnlyParameterValue(rng *rand.Rand) object.Object {
	return randomString(rng)
}

func matchActionParametersByEventNames(
	eventNames []string,
	action *model_state.Action,
) []model_state.Parameter {
	if action == nil || len(action.Parameters) == 0 || len(eventNames) == 0 {
		return nil
	}

	byNorm := make(map[string]model_state.Parameter, len(action.Parameters))
	for _, param := range action.Parameters {
		byNorm[identity.NormalizeSubKey(param.Name)] = param
	}

	var matched []model_state.Parameter
	seen := make(map[string]bool, len(eventNames))
	for _, name := range eventNames {
		norm := identity.NormalizeSubKey(name)
		if seen[norm] {
			continue
		}
		param, ok := byNorm[norm]
		if !ok {
			continue
		}
		seen[norm] = true
		matched = append(matched, param)
	}
	return matched
}

func remapSampledToEventNames(
	eventNames []string,
	matched []model_state.Parameter,
	sampled map[string]object.Object,
) map[string]object.Object {
	if len(matched) == 0 {
		if sampled == nil {
			return make(map[string]object.Object)
		}
		return sampled
	}

	normToEventName := make(map[string]string, len(eventNames))
	for _, name := range eventNames {
		normToEventName[identity.NormalizeSubKey(name)] = name
	}

	result := make(map[string]object.Object, len(matched))
	for _, param := range matched {
		eventName, ok := normToEventName[identity.NormalizeSubKey(param.Name)]
		if !ok {
			continue
		}
		if val, ok := sampled[param.Name]; ok {
			result[eventName] = val
		}
	}
	return result
}

package convert

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// associationClassEndpointParamNames is the TLA identifier pair for implicit
// host-end parameters on association-class _new. Distinct ends use each class's
// TLA name; the same class on both ends uses From/To prefixes.
func associationClassEndpointParamNames(fromClassName, toClassName string) (fromParam, toParam string) {
	from := model_class.ClassTLAName(fromClassName)
	to := model_class.ClassTLAName(toClassName)
	if from == to {
		return "From" + from, "To" + to
	}
	return from, to
}

// injectAssociationClassEndpointParams adds the implicit host-endpoint parameters
// to association-class creation events and their Initialize (creation) actions.
//
// Authors do not declare these in parser_human / parser_ai. Injection runs before
// lowering so well-formed logic may reference them (e.g. Partner.HasCustomers).
func injectAssociationClassEndpointParams(model *core.Model) error {
	if model == nil {
		return nil
	}
	classes := allClassesByKey(model)
	for _, assoc := range model.GetClassAssociations() {
		if assoc.AssociationClassKey == nil {
			continue
		}
		acClass, ok := classes[*assoc.AssociationClassKey]
		if !ok {
			continue
		}
		fromClass, okFrom := classes[assoc.FromClassKey]
		toClass, okTo := classes[assoc.ToClassKey]
		if !okFrom || !okTo {
			continue
		}
		fromParam, toParam := associationClassEndpointParamNames(fromClass.Name, toClass.Name)
		if err := injectACCreationEndpoints(&acClass, fromParam, toParam, fromClass, toClass); err != nil {
			return fmt.Errorf("association class %q: %w", acClass.Name, err)
		}
		// Write back into model tree.
		if err := writeClassBack(model, acClass); err != nil {
			return err
		}
	}
	return nil
}

func allClassesByKey(model *core.Model) map[identity.Key]model_class.Class {
	out := make(map[identity.Key]model_class.Class)
	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			for k, c := range subdomain.Classes {
				out[k] = c
			}
		}
	}
	return out
}

func injectACCreationEndpoints(
	ac *model_class.Class,
	fromParam, toParam string,
	fromClass, toClass model_class.Class,
) error {
	creationEventKeys := creationEventKeysOnClass(*ac)
	if len(creationEventKeys) == 0 {
		return nil
	}

	events := ac.Events
	if events == nil {
		events = make(map[identity.Key]model_state.Event)
	}
	actions := ac.Actions
	if actions == nil {
		actions = make(map[identity.Key]model_state.Action)
	}

	for ek := range creationEventKeys {
		ev, ok := events[ek]
		if !ok {
			continue
		}
		ev.ParameterNames = prependEndpointParamNames(ev.ParameterNames, fromParam, toParam)
		events[ek] = ev
	}

	// Creation transition actions need matching parameters for lower + surface sampling.
	creationActionKeys := creationActionKeysOnClass(*ac)
	for ak := range creationActionKeys {
		act, ok := actions[ak]
		if !ok {
			continue
		}
		params, err := prependEndpointActionParams(act, fromParam, toParam, fromClass, toClass)
		if err != nil {
			return err
		}
		act.Parameters = params
		actions[ak] = act
	}

	ac.SetEvents(events)
	ac.SetActions(actions)
	return nil
}

func creationEventKeysOnClass(class model_class.Class) map[identity.Key]struct{} {
	out := make(map[identity.Key]struct{})
	for _, t := range class.Transitions {
		if t.FromStateKey == nil {
			out[t.EventKey] = struct{}{}
		}
	}
	return out
}

func creationActionKeysOnClass(class model_class.Class) map[identity.Key]struct{} {
	out := make(map[identity.Key]struct{})
	for _, t := range class.Transitions {
		if t.FromStateKey == nil && t.ActionKey != nil {
			out[*t.ActionKey] = struct{}{}
		}
	}
	return out
}

func prependEndpointParamNames(names []string, fromParam, toParam string) []string {
	// Already injected or authored in correct leading positions.
	if len(names) >= 2 && names[0] == fromParam && names[1] == toParam {
		return names
	}
	// Drop accidental authoring of endpoint names anywhere, then put them first.
	rest := make([]string, 0, len(names))
	for _, n := range names {
		if n == fromParam || n == toParam {
			continue
		}
		rest = append(rest, n)
	}
	return append([]string{fromParam, toParam}, rest...)
}

func prependEndpointActionParams(
	act model_state.Action,
	fromParam, toParam string,
	fromClass, toClass model_class.Class,
) ([]model_state.Parameter, error) {
	// Already has both endpoint params (by name).
	if hasParamNamed(act.Parameters, fromParam) && hasParamNamed(act.Parameters, toParam) {
		// Reorder so endpoints are first when both exist.
		return reorderEndpointParamsFirst(act.Parameters, fromParam, toParam), nil
	}

	fromP, err := newObjectEndpointParam(act.Key, fromParam, fromClass)
	if err != nil {
		return nil, err
	}
	toP, err := newObjectEndpointParam(act.Key, toParam, toClass)
	if err != nil {
		return nil, err
	}

	rest := make([]model_state.Parameter, 0, len(act.Parameters))
	for _, p := range act.Parameters {
		if p.Name == fromParam || p.Name == toParam {
			continue
		}
		rest = append(rest, p)
	}
	return append([]model_state.Parameter{fromP, toP}, rest...), nil
}

func hasParamNamed(params []model_state.Parameter, name string) bool {
	for _, p := range params {
		if p.Name == name {
			return true
		}
	}
	return false
}

func reorderEndpointParamsFirst(params []model_state.Parameter, fromParam, toParam string) []model_state.Parameter {
	var fromP, toP *model_state.Parameter
	rest := make([]model_state.Parameter, 0, len(params))
	for i := range params {
		switch params[i].Name {
		case fromParam:
			p := params[i]
			fromP = &p
		case toParam:
			p := params[i]
			toP = &p
		default:
			rest = append(rest, params[i])
		}
	}
	out := make([]model_state.Parameter, 0, len(params))
	if fromP != nil {
		out = append(out, *fromP)
	}
	if toP != nil {
		out = append(out, *toP)
	}
	return append(out, rest...)
}

func newObjectEndpointParam(actionKey identity.Key, name string, endpointClass model_class.Class) (model_state.Parameter, error) {
	// Rules use the class sub-key (same as authored "object of partner").
	rules := "object of " + endpointClass.Key.SubKey
	return model_state.NewParameter(actionKey, name, rules, false)
}

func writeClassBack(model *core.Model, class model_class.Class) error {
	for dKey, domain := range model.Domains {
		for sKey, subdomain := range domain.Subdomains {
			if _, ok := subdomain.Classes[class.Key]; !ok {
				continue
			}
			subdomain.Classes[class.Key] = class
			domain.Subdomains[sKey] = subdomain
			model.Domains[dKey] = domain
			return nil
		}
	}
	return fmt.Errorf("class %s not found in model tree", class.Key.String())
}

package engine

import (
	"fmt"
	"sort"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_data_type"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/actions"
)

// validateSimulationModel rejects models with no simulatable classes and checks
// parsed action requires the parameter sampler cannot satisfy.
func validateSimulationModel(model *core.Model) error {
	if err := validateAtLeastOneSimulatableClass(model); err != nil {
		return err
	}
	if err := validateEventActionParameters(model); err != nil {
		return err
	}
	if err := validateRequiresSamplingSupport(model); err != nil {
		return err
	}
	return validateReferenceDataTypeInvariants(model)
}

func validateAtLeastOneSimulatableClass(model *core.Model) error {
	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			for _, class := range subdomain.Classes {
				if len(class.States) > 0 {
					return nil
				}
			}
		}
	}
	return fmt.Errorf("simulation requires at least one class with a state machine in scope")
}

type eventActionBindingKey struct {
	classKey  string
	eventKey  string
	actionKey string
}

func validateEventActionParameters(model *core.Model) error {
	seen := make(map[eventActionBindingKey]bool)
	var messages []string

	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			for _, class := range subdomain.Classes {
				if len(class.States) == 0 {
					continue
				}
				messages = append(messages, collectEventActionParameterMismatches(class, seen)...)
			}
		}
	}

	if len(messages) == 0 {
		return nil
	}
	sort.Strings(messages)
	return fmt.Errorf("%s", strings.Join(messages, "; "))
}

func collectEventActionParameterMismatches(class model_class.Class, seen map[eventActionBindingKey]bool) []string {
	var messages []string

	for _, transition := range class.Transitions {
		if transition.ActionKey == nil {
			continue
		}

		key := eventActionBindingKey{
			classKey:  class.Key.String(),
			eventKey:  transition.EventKey.String(),
			actionKey: transition.ActionKey.String(),
		}
		if seen[key] {
			continue
		}
		seen[key] = true

		event, eventOK := class.Events[transition.EventKey]
		if !eventOK {
			continue
		}
		action, actionOK := class.Actions[*transition.ActionKey]
		if !actionOK {
			continue
		}

		missing := actionParametersMissingFromEvent(event, action)
		if len(missing) == 0 {
			continue
		}

		messages = append(messages, formatActionParametersMissingFromEvent(class.Name, event.Name, action.Name, missing))
	}

	return messages
}

func actionParametersMissingFromEvent(event model_state.Event, action model_state.Action) []string {
	onEvent := make(map[string]bool, len(event.ParameterNames))
	for _, name := range event.ParameterNames {
		onEvent[identity.NormalizeSubKey(name)] = true
	}

	var missing []string
	for _, param := range action.Parameters {
		if onEvent[identity.NormalizeSubKey(param.Name)] {
			continue
		}
		missing = append(missing, param.Name)
	}
	sort.Strings(missing)
	return missing
}

func formatActionParametersMissingFromEvent(className, eventName, actionName string, missing []string) string {
	switch len(missing) {
	case 1:
		return fmt.Sprintf(
			`class %q event %q action %q: action parameter %q is not declared on the event`,
			className, eventName, actionName, missing[0],
		)
	default:
		return fmt.Sprintf(
			`class %q event %q action %q: action parameters %s are not declared on the event`,
			className, eventName, actionName, strings.Join(missing, ", "),
		)
	}
}

func validateClassOwnerRequiresSampling(class model_class.Class) error {
	for _, action := range class.Actions {
		if err := actions.ValidateOwnerRequiresSamplingSupport(class.Name, actions.ParameterOwnerFromAction(action)); err != nil {
			return err
		}
	}
	for _, query := range class.Queries {
		if err := actions.ValidateOwnerRequiresSamplingSupport(class.Name, actions.ParameterOwnerFromQuery(query)); err != nil {
			return err
		}
	}
	return nil
}

func validateRequiresSamplingSupport(model *core.Model) error {
	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			for _, class := range subdomain.Classes {
				if len(class.States) == 0 {
					continue
				}
				if err := validateClassOwnerRequiresSampling(class); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// validateReferenceDataTypeInvariants rejects models where a reference data type lacks a
// formal constraint the simulator can enforce on attributes. Action and query parameters
// use parameter invariants, owner requires, or implicit reference assessments at sampling time.
func validateReferenceDataTypeInvariants(model *core.Model) error {
	var messages []string

	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			for _, class := range subdomain.Classes {
				if len(class.States) == 0 {
					continue
				}
				messages = append(messages, collectReferenceAttributeInvariantGaps(class)...)
			}
		}
	}

	if len(messages) == 0 {
		return nil
	}
	sort.Strings(messages)
	return fmt.Errorf("%s", strings.Join(messages, "; "))
}

func collectReferenceAttributeInvariantGaps(class model_class.Class) []string {
	var messages []string

	for _, attr := range class.Attributes {
		if !model_data_type.ContainsReferenceConstraint(attr.DataType) {
			continue
		}
		if hasParsedLogic(attr.Invariants) {
			continue
		}
		messages = append(messages, fmt.Sprintf(
			`class %q attribute %q: reference data type has no invariant`,
			class.Name, attr.Name,
		))
	}

	return messages
}

func hasParsedLogic(logics []model_logic.Logic) bool {
	for _, logic := range logics {
		if logic.Spec.Expression != nil {
			return true
		}
	}
	return false
}

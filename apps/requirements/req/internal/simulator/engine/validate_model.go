package engine

import (
	"fmt"
	"sort"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/actions"
)

// validateSimulationModel rejects models where any in-scope class lacks a state machine
// or defines parsed action requires the parameter sampler cannot satisfy.
func validateSimulationModel(model *core.Model) error {
	if err := validateClassesHaveStates(model); err != nil {
		return err
	}
	return validateRequiresSamplingSupport(model)
}

func validateClassesHaveStates(model *core.Model) error {
	var missing []string
	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			for _, class := range subdomain.Classes {
				if len(class.States) == 0 {
					missing = append(missing, class.Name)
				}
			}
		}
	}
	if len(missing) == 0 {
		return nil
	}
	sort.Strings(missing)
	return fmt.Errorf(
		"simulation requires every class to have a state machine; missing states: %s",
		strings.Join(missing, ", "),
	)
}

func validateRequiresSamplingSupport(model *core.Model) error {
	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			for _, class := range subdomain.Classes {
				if len(class.States) == 0 {
					continue
				}
				for _, action := range class.Actions {
					if err := actions.ValidateActionRequiresSamplingSupport(class.Name, action); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

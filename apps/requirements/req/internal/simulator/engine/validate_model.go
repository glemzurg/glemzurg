package engine

import (
	"fmt"
	"sort"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
)

// validateSimulationModel rejects models where any in-scope class lacks a state machine.
func validateSimulationModel(model *core.Model) error {
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

package invariants

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
)

// AssociationInstancePairChecker validates that each association has at most one
// link per from/to instance pair.
type AssociationInstancePairChecker struct {
	associations []model_class.Association
}

// NewAssociationInstancePairChecker builds instance-pair uniqueness metadata from the model.
func NewAssociationInstancePairChecker(model *core.Model) *AssociationInstancePairChecker {
	checker := &AssociationInstancePairChecker{}

	classes := make(map[identity.Key]model_class.Class)
	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			for _, class := range subdomain.Classes {
				classes[class.Key] = class
			}
		}
	}

	for _, assoc := range model.GetClassAssociations() {
		if _, ok := classes[assoc.FromClassKey]; !ok {
			continue
		}
		if _, ok := classes[assoc.ToClassKey]; !ok {
			continue
		}
		checker.associations = append(checker.associations, assoc)
	}

	return checker
}

// CheckState validates instance-pair uniqueness across all associations.
func (c *AssociationInstancePairChecker) CheckState(simState *state.SimulationState) ViolationErrors {
	var violations ViolationErrors
	for _, assoc := range c.associations {
		violations = append(violations, c.checkAssociation(simState, assoc)...)
	}
	return violations
}

func (c *AssociationInstancePairChecker) checkAssociation(
	simState *state.SimulationState,
	assoc model_class.Association,
) ViolationErrors {
	links := collectAssociationLinks(simState, assoc)
	if len(links) == 0 {
		return nil
	}

	counts := make(map[associationLinkEndpoints]int)
	var violations ViolationErrors
	for _, link := range links {
		counts[link]++
		if counts[link] > 1 {
			violations = append(violations, NewAssociationDuplicateLinkViolation(AssociationDuplicateLinkViolationParams{
				AssociationName: assoc.Name,
				FromInstanceID:  link.fromID,
				ToInstanceID:    link.toID,
				ActualCount:     counts[link],
			}))
		}
	}
	return violations
}

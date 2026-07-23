package invariants

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/instance"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/schema"
)

// AssociationInstancePairChecker validates that each association has at most one
// link per from/to instance pair.
type AssociationInstancePairChecker struct {
	associations []model_class.Association
}

// NewAssociationInstancePairChecker builds instance-pair uniqueness metadata from schema.
func NewAssociationInstancePairChecker(sch *schema.Schema) *AssociationInstancePairChecker {
	checker := &AssociationInstancePairChecker{}

	sch.ForEachAssociation(func(assoc model_class.Association) {
		if !sch.IsClassInScope(assoc.FromClassKey) || !sch.IsClassInScope(assoc.ToClassKey) {
			return
		}
		checker.associations = append(checker.associations, assoc)
	})

	return checker
}

// CheckState validates instance-pair uniqueness across all associations.
func (c *AssociationInstancePairChecker) CheckState(simState *instance.State) ViolationErrors {
	var violations ViolationErrors
	for _, assoc := range c.associations {
		violations = append(violations, c.checkAssociation(simState, assoc)...)
	}
	return violations
}

func (c *AssociationInstancePairChecker) checkAssociation(
	simState *instance.State,
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

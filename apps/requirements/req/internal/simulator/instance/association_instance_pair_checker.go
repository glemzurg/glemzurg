package instance

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/schema"
)

// AssociationInstancePairChecker validates that each association has at most one
// link per from/to instance pair. Associations are loaded from schema at check time.
type AssociationInstancePairChecker struct {
	sch *schema.Schema
}

// NewAssociationInstancePairChecker binds an instance-pair checker to schema.
func NewAssociationInstancePairChecker(sch *schema.Schema) *AssociationInstancePairChecker {
	return &AssociationInstancePairChecker{sch: sch}
}

// CheckState validates instance-pair uniqueness across all associations.
func (c *AssociationInstancePairChecker) CheckState(simState *State) ViolationErrors {
	var violations ViolationErrors
	if c == nil || c.sch == nil {
		return violations
	}
	for _, view := range c.sch.ScopedAssociations() {
		violations = append(violations, c.checkAssociation(simState, view.Association)...)
	}
	return violations
}

func (c *AssociationInstancePairChecker) checkAssociation(
	simState *State,
	assoc model_class.Association,
) ViolationErrors {
	return checkAssociationInstancePairs(assoc, collectAssociationLinks(simState, assoc))
}

// checkAssociationInstancePairs reports when the same from/to pair appears more than once.
func checkAssociationInstancePairs(
	assoc model_class.Association,
	links []associationLinkEndpoints,
) ViolationErrors {
	if len(links) == 0 {
		return nil
	}

	counts := make(map[associationLinkEndpoints]int)
	var violations ViolationErrors
	for _, link := range links {
		counts[link]++
		if counts[link] > 1 {
			violations = append(violations, newAssociationDuplicateLinkViolation(AssociationDuplicateLinkViolationParams{
				AssociationName: assoc.Name,
				FromInstanceID:  link.fromID,
				ToInstanceID:    link.toID,
				ActualCount:     counts[link],
			}))
		}
	}
	return violations
}

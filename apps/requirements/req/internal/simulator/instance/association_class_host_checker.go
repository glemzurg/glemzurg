package instance

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/schema"
)

// AssociationClassHostChecker reports plain binary host links on associations that
// require an in-scope association class. When the AC is out of scope, the host is
// allowed to degrade to plain endpoint links and this check is silent.
type AssociationClassHostChecker struct {
	sch *schema.Schema
}

// NewAssociationClassHostChecker binds the checker to schema.
func NewAssociationClassHostChecker(sch *schema.Schema) *AssociationClassHostChecker {
	return &AssociationClassHostChecker{sch: sch}
}

// CheckState finds plain LinkTable edges for AC-hosted associations whose AC is in scope.
func (c *AssociationClassHostChecker) CheckState(simState *State) ViolationErrors {
	if c == nil || c.sch == nil || simState == nil {
		return nil
	}
	var violations ViolationErrors
	for _, view := range c.sch.ScopedAssociations() {
		assoc := view.Association
		if assoc.AssociationClassKey == nil {
			continue
		}
		acKey := *assoc.AssociationClassKey
		// Only when the association class participates in the run; otherwise plain links are intentional.
		if !c.sch.IsClassInScope(acKey) {
			continue
		}
		simState.ForEachBinaryLinkOfAssociation(assoc.Key, func(fromID, toID ID) {
			violations = append(violations, newAssociationClassMissingViolation(AssociationClassMissingViolationParams{
				AssociationName:     assoc.Name,
				AssociationClassKey: acKey,
				FromInstanceID:      fromID,
				ToInstanceID:        toID,
			}))
		})
	}
	return violations
}

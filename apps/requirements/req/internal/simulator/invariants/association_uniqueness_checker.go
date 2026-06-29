package invariants

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
)

// AssociationUniquenessChecker validates per-pair link caps declared on associations.
type AssociationUniquenessChecker struct {
	associations []model_class.Association
}

// NewAssociationUniquenessChecker builds association uniqueness metadata from the model.
func NewAssociationUniquenessChecker(model *core.Model) *AssociationUniquenessChecker {
	checker := &AssociationUniquenessChecker{}

	classes := make(map[identity.Key]model_class.Class)
	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			for _, class := range subdomain.Classes {
				classes[class.Key] = class
			}
		}
	}

	for _, assoc := range model.GetClassAssociations() {
		if assoc.Uniqueness.LowerBound == 0 && assoc.Uniqueness.HigherBound == 0 {
			continue
		}
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

// CheckState validates all association uniqueness constraints.
func (c *AssociationUniquenessChecker) CheckState(simState *state.SimulationState) ViolationErrors {
	var violations ViolationErrors
	for _, assoc := range c.associations {
		violations = append(violations, c.checkAssociation(simState, assoc)...)
	}
	return violations
}

func (c *AssociationUniquenessChecker) checkAssociation(
	simState *state.SimulationState,
	assoc model_class.Association,
) ViolationErrors {
	if assoc.AssociationClassKey != nil {
		return c.checkAssociationClassPairs(simState, assoc)
	}
	return c.checkDirectAssociationPairs(simState, assoc)
}

func (c *AssociationUniquenessChecker) checkAssociationClassPairs(
	simState *state.SimulationState,
	assoc model_class.Association,
) ViolationErrors {
	seen := make(map[string]bool)
	var violations ViolationErrors
	for _, link := range simState.AssociationLinks().AllLinks() {
		if link.HostAssocKey != assoc.Key {
			continue
		}
		violations = append(violations, c.violationForPair(simState, assoc, seen, link.FromEndpointID, link.ToEndpointID)...)
	}
	return violations
}

func (c *AssociationUniquenessChecker) checkDirectAssociationPairs(
	simState *state.SimulationState,
	assoc model_class.Association,
) ViolationErrors {
	seen := make(map[string]bool)
	var violations ViolationErrors
	assocKey := evaluator.AssociationKey(assoc.Key.String())
	for _, inst := range simState.AllInstances() {
		if inst.ClassKey != assoc.FromClassKey {
			continue
		}
		for _, link := range simState.Links().GetAllForward(evaluator.ObjectID(inst.ID)) {
			if link.AssociationKey != assocKey {
				continue
			}
			violations = append(violations, c.violationForPair(simState, assoc, seen, inst.ID, state.InstanceID(link.ToID))...)
		}
	}
	return violations
}

func (c *AssociationUniquenessChecker) violationForPair(
	simState *state.SimulationState,
	assoc model_class.Association,
	seen map[string]bool,
	fromID, toID state.InstanceID,
) ViolationErrors {
	pairKey := fmt.Sprintf("%d:%d", fromID, toID)
	if seen[pairKey] {
		return nil
	}
	seen[pairKey] = true

	count := simState.CountActivePairLinks(assoc, fromID, toID)
	msg := checkUniquenessUpperBound(count, assoc.Uniqueness)
	if msg == "" {
		return nil
	}
	return ViolationErrors{NewAssociationUniquenessViolation(AssociationUniquenessViolationParams{
		AssociationName: assoc.Name,
		FromInstanceID:  fromID,
		ToInstanceID:    toID,
		ActualCount:     count,
		RequiredMin:     0,
		RequiredMax:     assoc.Uniqueness.HigherBound,
		Message:         msg,
	})}
}

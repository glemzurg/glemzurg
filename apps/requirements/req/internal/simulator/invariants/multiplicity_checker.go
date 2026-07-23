package invariants

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/instance"
)

// associationBinding holds one association edge as seen from a participating class.
type associationBinding struct {
	association  model_class.Association
	fromClassKey identity.Key
	toClassKey   identity.Key
}

// MultiplicityChecker validates association multiplicity constraints as implicit invariants.
type MultiplicityChecker struct {
	classAssocs map[identity.Key][]associationBinding
}

// NewMultiplicityChecker builds association multiplicity metadata from the model.
func NewMultiplicityChecker(model *core.Model) *MultiplicityChecker {
	checker := &MultiplicityChecker{
		classAssocs: make(map[identity.Key][]associationBinding),
	}

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
		binding := associationBinding{
			association:  assoc,
			fromClassKey: assoc.FromClassKey,
			toClassKey:   assoc.ToClassKey,
		}
		checker.classAssocs[assoc.FromClassKey] = append(checker.classAssocs[assoc.FromClassKey], binding)
		if assoc.FromClassKey != assoc.ToClassKey {
			checker.classAssocs[assoc.ToClassKey] = append(checker.classAssocs[assoc.ToClassKey], binding)
		}
	}

	return checker
}

// CheckState validates all association multiplicities across every live instance.
func (c *MultiplicityChecker) CheckState(simState *instance.State) ViolationErrors {
	var violations ViolationErrors
	simState.ForEachInstance(func(inst *instance.Instance) {
		violations = append(violations, c.CheckInstance(inst, simState)...)
	})
	return violations
}

// CheckInstance validates all multiplicity constraints for a single instance.
func (c *MultiplicityChecker) CheckInstance(
	instance *instance.Instance,
	simState *instance.State,
) ViolationErrors {
	if instance == nil {
		return nil
	}

	assocs := c.classAssocs[instance.ClassKey]
	if len(assocs) == 0 {
		return nil
	}

	var violations ViolationErrors

	for _, binding := range assocs {
		if binding.fromClassKey == instance.ClassKey {
			count := c.countActiveForwardLinks(instance.ID, binding, simState)
			if msg := checkMultiplicityBounds(count, binding.association.ToMultiplicity.LowerBound, binding.association.ToMultiplicity.HigherBound); msg != "" {
				violations = append(violations, NewMultiplicityViolation(MultiplicityViolationParams{
					InstanceID:      instance.ID,
					ClassKey:        instance.ClassKey,
					AssociationName: binding.association.Name,
					Direction:       "forward",
					ActualCount:     count,
					RequiredMin:     binding.association.ToMultiplicity.LowerBound,
					RequiredMax:     binding.association.ToMultiplicity.HigherBound,
					Message:         msg,
				}))
			}
		}

		if binding.toClassKey == instance.ClassKey {
			count := c.countActiveReverseLinks(instance.ID, binding, simState)
			if msg := checkMultiplicityBounds(count, binding.association.FromMultiplicity.LowerBound, binding.association.FromMultiplicity.HigherBound); msg != "" {
				violations = append(violations, NewMultiplicityViolation(MultiplicityViolationParams{
					InstanceID:      instance.ID,
					ClassKey:        instance.ClassKey,
					AssociationName: binding.association.Name,
					Direction:       "reverse",
					ActualCount:     count,
					RequiredMin:     binding.association.FromMultiplicity.LowerBound,
					RequiredMax:     binding.association.FromMultiplicity.HigherBound,
					Message:         msg,
				}))
			}
		}
	}

	return violations
}

func (c *MultiplicityChecker) countActiveForwardLinks(
	fromID instance.ID,
	binding associationBinding,
	simState *instance.State,
) int {
	if binding.association.AssociationClassKey != nil {
		return c.countActiveAssociationLinksFrom(fromID, binding.association.Key, simState)
	}
	linked := simState.GetLinkedForward(fromID, binding.association.Key)
	return c.countActiveLinkedInstances(linked, simState)
}

func (c *MultiplicityChecker) countActiveReverseLinks(
	toID instance.ID,
	binding associationBinding,
	simState *instance.State,
) int {
	if binding.association.AssociationClassKey != nil {
		return c.countActiveAssociationLinksTo(toID, binding.association.Key, simState)
	}
	linked := simState.GetLinkedReverse(toID, binding.association.Key)
	return c.countActiveLinkedInstances(linked, simState)
}

func (c *MultiplicityChecker) countActiveAssociationLinksFrom(
	fromID instance.ID,
	hostAssocKey identity.Key,
	simState *instance.State,
) int {
	links := simState.AssociationLinksFromEndpoint(hostAssocKey, fromID)
	count := 0
	for _, link := range links {
		linkInst := simState.GetInstance(link.LinkInstanceID)
		if linkInst != nil {
			count++
		}
	}
	return count
}

func (c *MultiplicityChecker) countActiveAssociationLinksTo(
	toID instance.ID,
	hostAssocKey identity.Key,
	simState *instance.State,
) int {
	links := simState.AssociationLinksToEndpoint(hostAssocKey, toID)
	count := 0
	for _, link := range links {
		linkInst := simState.GetInstance(link.LinkInstanceID)
		if linkInst != nil {
			count++
		}
	}
	return count
}

func (c *MultiplicityChecker) countActiveLinkedInstances(
	linked []instance.ID,
	simState *instance.State,
) int {
	count := 0
	for _, id := range linked {
		inst := simState.GetInstance(id)
		if inst != nil {
			count++
		}
	}
	return count
}

func checkMultiplicityBounds(count int, lowerBound, upperBound uint) string {
	if lowerBound > 0 && uint(count) < lowerBound { //nolint:gosec // count is a link count from a small in-memory graph, no overflow risk
		return fmt.Sprintf("expected at least %d links, got %d", lowerBound, count)
	}
	if upperBound > 0 && uint(count) > upperBound { //nolint:gosec // count is a link count from a small in-memory graph, no overflow risk
		return fmt.Sprintf("expected at most %d links, got %d", upperBound, count)
	}
	return ""
}

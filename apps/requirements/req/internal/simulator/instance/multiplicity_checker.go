package instance

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/schema"
)

// associationBinding holds one association edge as seen from a participating class.
type associationBinding struct {
	association  model_class.Association
	fromClassKey identity.Key
	toClassKey   identity.Key
}

// MultiplicityChecker validates association multiplicity constraints as implicit invariants.
// Association edges are loaded per class from schema when that instance is checked.
type MultiplicityChecker struct {
	sch *schema.Schema
}

// NewMultiplicityChecker binds a multiplicity checker to schema.
func NewMultiplicityChecker(sch *schema.Schema) *MultiplicityChecker {
	return &MultiplicityChecker{sch: sch}
}

// CheckState validates all association multiplicities across every live instance.
func (c *MultiplicityChecker) CheckState(simState *State) ViolationErrors {
	var violations ViolationErrors
	simState.ForEachInstance(func(inst *Instance) {
		violations = append(violations, c.CheckInstance(inst, simState)...)
	})
	return violations
}

// CheckInstance validates all multiplicity constraints for a single instance.
func (c *MultiplicityChecker) CheckInstance(
	instance *Instance,
	simState *State,
) ViolationErrors {
	if c == nil || c.sch == nil || instance == nil {
		return nil
	}

	views, inScope, err := c.sch.AssociationsForClass(instance.ClassKey)
	if err != nil || !inScope || len(views) == 0 {
		return nil
	}

	var violations ViolationErrors

	for _, view := range views {
		binding := associationBinding{
			association:  view.Association,
			fromClassKey: view.FromClassKey,
			toClassKey:   view.ToClassKey,
		}
		if binding.fromClassKey == instance.ClassKey {
			count := c.countActiveForwardLinks(instance.ID, binding, simState)
			if msg := checkMultiplicityBounds(count, binding.association.ToMultiplicity.LowerBound, binding.association.ToMultiplicity.HigherBound); msg != "" {
				violations = append(violations, newMultiplicityViolation(MultiplicityViolationParams{
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
				violations = append(violations, newMultiplicityViolation(MultiplicityViolationParams{
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
	fromID ID,
	binding associationBinding,
	simState *State,
) int {
	if binding.association.AssociationClassKey != nil {
		return c.countActiveAssociationLinksFrom(fromID, binding.association.Key, simState)
	}
	linked := simState.GetLinkedForward(fromID, binding.association.Key)
	return c.countActiveLinkedInstances(linked, simState)
}

func (c *MultiplicityChecker) countActiveReverseLinks(
	toID ID,
	binding associationBinding,
	simState *State,
) int {
	if binding.association.AssociationClassKey != nil {
		return c.countActiveAssociationLinksTo(toID, binding.association.Key, simState)
	}
	linked := simState.GetLinkedReverse(toID, binding.association.Key)
	return c.countActiveLinkedInstances(linked, simState)
}

func (c *MultiplicityChecker) countActiveAssociationLinksFrom(
	fromID ID,
	hostAssocKey identity.Key,
	simState *State,
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
	toID ID,
	hostAssocKey identity.Key,
	simState *State,
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
	linked []ID,
	simState *State,
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

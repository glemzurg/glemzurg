package engine

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
)

// MultiplicityViolation records that an association multiplicity constraint
// was not satisfied for a particular instance.
type MultiplicityViolation struct {
	AssociationKey  identity.Key
	AssociationName string
	InstanceID      state.InstanceID
	ClassKey        identity.Key
	Direction       string // "forward" or "reverse"
	ActualCount     int
	RequiredMin     uint
	RequiredMax     uint // 0 = unbounded
	Message         string
}

// MultiplicityChecker validates association multiplicity constraints.
type MultiplicityChecker struct {
	catalog *ClassCatalog
}

// NewMultiplicityChecker creates a new multiplicity checker.
func NewMultiplicityChecker(catalog *ClassCatalog) *MultiplicityChecker {
	return &MultiplicityChecker{
		catalog: catalog,
	}
}

// CheckInstance validates all multiplicity constraints for a single instance.
func (c *MultiplicityChecker) CheckInstance(
	instance *state.ClassInstance,
	simState *state.SimulationState,
) []MultiplicityViolation {
	if instance == nil {
		return nil
	}

	assocs := c.catalog.GetAssociationsForClass(instance.ClassKey)
	if len(assocs) == 0 {
		return nil
	}

	var violations []MultiplicityViolation

	for _, ai := range assocs {
		if ai.FromClassKey == instance.ClassKey {
			// This instance is the "from" side — check ToMultiplicity.
			count := c.countActiveForwardLinks(instance.ID, ai, simState)

			v := checkBounds(count, ai.Association.ToMultiplicity.LowerBound, ai.Association.ToMultiplicity.HigherBound)
			if v != "" {
				violations = append(violations, MultiplicityViolation{
					AssociationKey:  ai.Association.Key,
					AssociationName: ai.Association.Name,
					InstanceID:      instance.ID,
					ClassKey:        instance.ClassKey,
					Direction:       "forward",
					ActualCount:     count,
					RequiredMin:     ai.Association.ToMultiplicity.LowerBound,
					RequiredMax:     ai.Association.ToMultiplicity.HigherBound,
					Message:         v,
				})
			}
		}

		if ai.ToClassKey == instance.ClassKey {
			// This instance is the "to" side — check FromMultiplicity.
			count := c.countActiveReverseLinks(instance.ID, ai, simState)

			v := checkBounds(count, ai.Association.FromMultiplicity.LowerBound, ai.Association.FromMultiplicity.HigherBound)
			if v != "" {
				violations = append(violations, MultiplicityViolation{
					AssociationKey:  ai.Association.Key,
					AssociationName: ai.Association.Name,
					InstanceID:      instance.ID,
					ClassKey:        instance.ClassKey,
					Direction:       "reverse",
					ActualCount:     count,
					RequiredMin:     ai.Association.FromMultiplicity.LowerBound,
					RequiredMax:     ai.Association.FromMultiplicity.HigherBound,
					Message:         v,
				})
			}
		}
	}

	return violations
}

func (c *MultiplicityChecker) countActiveForwardLinks(
	fromID state.InstanceID,
	ai AssociationInfo,
	simState *state.SimulationState,
) int {
	linked := simState.GetLinkedForward(fromID, ai.Association.Key)
	return c.countActiveLinkedInstances(linked, ai.ToClassKey, simState)
}

func (c *MultiplicityChecker) countActiveReverseLinks(
	toID state.InstanceID,
	ai AssociationInfo,
	simState *state.SimulationState,
) int {
	linked := simState.GetLinkedReverse(toID, ai.Association.Key)
	return c.countActiveLinkedInstances(linked, ai.FromClassKey, simState)
}

func (c *MultiplicityChecker) countActiveLinkedInstances(
	linked []state.InstanceID,
	_ identity.Key,
	simState *state.SimulationState,
) int {
	count := 0
	for _, id := range linked {
		inst := simState.GetInstance(id)
		if inst == nil {
			continue
		}
		if !IsActiveAssociationClassInstance(c.catalog, inst.ClassKey, getInstanceStateName(inst)) {
			continue
		}
		count++
	}
	return count
}

// CheckState validates all multiplicity constraints across all instances.
func (c *MultiplicityChecker) CheckState(
	simState *state.SimulationState,
) []MultiplicityViolation {
	var violations []MultiplicityViolation

	for _, instance := range simState.AllInstances() {
		violations = append(violations, c.CheckInstance(instance, simState)...)
	}

	return violations
}

// checkBounds checks if a count satisfies lower/upper bounds.
// Returns empty string if satisfied, otherwise a violation message.
func checkBounds(count int, lowerBound, upperBound uint) string {
	if lowerBound > 0 && uint(count) < lowerBound { //nolint:gosec // count is a link count from a small in-memory graph, no overflow risk
		return fmt.Sprintf("expected at least %d links, got %d", lowerBound, count)
	}
	if upperBound > 0 && uint(count) > upperBound { //nolint:gosec // count is a link count from a small in-memory graph, no overflow risk
		return fmt.Sprintf("expected at most %d links, got %d", upperBound, count)
	}
	return ""
}

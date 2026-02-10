package engine

import (
	"sort"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/invariants"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
)

// LivenessChecker performs post-simulation coverage analysis.
// It verifies that every in-scope class was instantiated, every
// non-derived attribute was written, and every in-scope association
// had at least one link created.
type LivenessChecker struct {
	catalog *ClassCatalog
}

// NewLivenessChecker creates a new liveness checker.
func NewLivenessChecker(catalog *ClassCatalog) *LivenessChecker {
	return &LivenessChecker{catalog: catalog}
}

// Check performs all liveness checks against a completed simulation result.
func (lc *LivenessChecker) Check(result *SimulationResult) invariants.ViolationList {
	var violations invariants.ViolationList
	violations = append(violations, lc.checkClassInstantiation(result)...)
	violations = append(violations, lc.checkAttributeWriteCoverage(result)...)
	violations = append(violations, lc.checkAssociationCoverage(result)...)
	return violations
}

// checkClassInstantiation verifies every simulatable class had at least
// one instance created during the simulation.
func (lc *LivenessChecker) checkClassInstantiation(result *SimulationResult) invariants.ViolationList {
	instantiated := make(map[identity.Key]bool)
	collectInstantiatedClasses(result.Steps, instantiated)

	var violations invariants.ViolationList
	for _, classInfo := range lc.catalog.AllSimulatableClasses() {
		if !instantiated[classInfo.ClassKey] {
			violations = append(violations, invariants.NewLivenessClassNotInstantiatedViolation(
				classInfo.ClassKey,
				classInfo.Class.Name,
			))
		}
	}
	return violations
}

// collectInstantiatedClasses walks steps (including cascaded) and records
// all class keys that had creation steps.
func collectInstantiatedClasses(steps []*SimulationStep, out map[identity.Key]bool) {
	for _, step := range steps {
		if step.Kind == StepKindCreation {
			out[step.ClassKey] = true
		}
		if len(step.CascadedSteps) > 0 {
			collectInstantiatedClasses(step.CascadedSteps, out)
		}
	}
}

// checkAttributeWriteCoverage verifies every non-derived attribute of each
// in-scope class was written at least once during the simulation.
func (lc *LivenessChecker) checkAttributeWriteCoverage(result *SimulationResult) invariants.ViolationList {
	written := make(map[identity.Key]map[string]bool)
	collectWrittenAttributes(result.Steps, written)

	var violations invariants.ViolationList
	for _, classInfo := range lc.catalog.AllSimulatableClasses() {
		classWritten := written[classInfo.ClassKey]

		// Collect non-derived attribute names, sorted for deterministic output.
		var attrNames []string
		for _, attr := range classInfo.Class.Attributes {
			if attr.DerivationPolicy != nil {
				continue // Derived attributes are computed, not written.
			}
			attrNames = append(attrNames, attr.Name)
		}
		sort.Strings(attrNames)

		for _, attrName := range attrNames {
			if classWritten == nil || !classWritten[attrName] {
				violations = append(violations, invariants.NewLivenessAttributeNotWrittenViolation(
					classInfo.ClassKey,
					classInfo.Class.Name,
					attrName,
				))
			}
		}
	}
	return violations
}

// collectWrittenAttributes walks steps and records all (classKey, attrName)
// pairs that were written via primed assignments.
func collectWrittenAttributes(steps []*SimulationStep, out map[identity.Key]map[string]bool) {
	for _, step := range steps {
		// Check transition result.
		if step.TransitionResult != nil && step.TransitionResult.ActionResult != nil {
			recordPrimedWrites(step.ClassKey, step.TransitionResult.ActionResult.PrimedAssignments, out)
		}
		// Check do action result.
		if step.DoActionResult != nil {
			recordPrimedWrites(step.ClassKey, step.DoActionResult.PrimedAssignments, out)
		}
		// Recurse into cascaded steps.
		if len(step.CascadedSteps) > 0 {
			collectWrittenAttributes(step.CascadedSteps, out)
		}
	}
}

// recordPrimedWrites records attribute names from primed assignments for a class.
func recordPrimedWrites(classKey identity.Key, assignments map[state.InstanceID]map[string]object.Object, out map[identity.Key]map[string]bool) {
	for _, fields := range assignments {
		for fieldName := range fields {
			if out[classKey] == nil {
				out[classKey] = make(map[string]bool)
			}
			out[classKey][fieldName] = true
		}
	}
}

// checkAssociationCoverage verifies every in-scope association had at
// least one link created during the simulation.
func (lc *LivenessChecker) checkAssociationCoverage(result *SimulationResult) invariants.ViolationList {
	if result.FinalState == nil {
		return nil
	}

	// Get all association keys that have at least one link in the final state.
	linkedAssocs := result.FinalState.Links().AllAssociationKeys()

	var violations invariants.ViolationList
	for _, assocInfo := range lc.catalog.AllAssociations() {
		assocKeyStr := evaluator.AssociationKey(assocInfo.Association.Key.String())
		if !linkedAssocs[assocKeyStr] {
			violations = append(violations, invariants.NewLivenessAssociationNotLinkedViolation(
				assocInfo.Association.Key,
				assocInfo.Association.Name,
				assocInfo.FromClassKey,
				assocInfo.ToClassKey,
			))
		}
	}
	return violations
}

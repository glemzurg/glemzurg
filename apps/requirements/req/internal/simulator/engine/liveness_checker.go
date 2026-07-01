package engine

import (
	"sort"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/invariants"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
)

// LivenessChecker performs post-simulation coverage analysis across the whole
// simulation surface. Violations highlight gaps in subdomain logic — classes,
// associations, events, queries, and actions that the run never exercised.
type LivenessChecker struct {
	catalog *ClassCatalog
}

// NewLivenessChecker creates a new liveness checker.
func NewLivenessChecker(catalog *ClassCatalog) *LivenessChecker {
	return &LivenessChecker{catalog: catalog}
}

// Check performs all liveness checks against a completed simulation result.
func (lc *LivenessChecker) Check(result *SimulationResult) invariants.ViolationErrors {
	var violations invariants.ViolationErrors
	violations = append(violations, lc.checkClassInstantiation(result)...)
	violations = append(violations, lc.checkAttributeWriteCoverage(result)...)
	violations = append(violations, lc.checkAssociationCoverage(result)...)
	violations = append(violations, lc.checkEventCoverage(result)...)
	violations = append(violations, lc.checkQueryCoverage(result)...)
	violations = append(violations, lc.checkDerivedAttributeReadCoverage(result)...)
	violations = append(violations, lc.checkActionCoverage(result)...)
	violations = append(violations, lc.checkParameterSimulationCoverage(result)...)
	return violations
}

// checkClassInstantiation verifies every in-scope class had at least one instance
// created during the simulation.
func (lc *LivenessChecker) checkClassInstantiation(result *SimulationResult) invariants.ViolationErrors {
	instantiated := make(map[identity.Key]bool)
	collectInstantiatedClasses(result.Steps, instantiated)

	var violations invariants.ViolationErrors
	for _, classInfo := range lc.catalog.AllScopedClasses() {
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
func (lc *LivenessChecker) checkAttributeWriteCoverage(result *SimulationResult) invariants.ViolationErrors {
	written := make(map[identity.Key]map[string]bool)
	collectWrittenAttributes(result.Steps, written)

	var violations invariants.ViolationErrors
	for _, classInfo := range lc.catalog.AllScopedClasses() {
		classWritten := written[classInfo.ClassKey]

		type attrCoverage struct {
			subKey string
			name   string
		}
		var attrs []attrCoverage
		for _, attr := range classInfo.Class.Attributes {
			if attr.DerivationPolicy != nil {
				continue
			}
			attrs = append(attrs, attrCoverage{subKey: attr.Key.SubKey, name: attr.Name})
		}
		sort.Slice(attrs, func(i, j int) bool { return attrs[i].name < attrs[j].name })

		for _, attr := range attrs {
			if classWritten == nil || !classWritten[attr.subKey] {
				violations = append(violations, invariants.NewLivenessAttributeNotWrittenViolation(
					classInfo.ClassKey,
					classInfo.Class.Name,
					attr.name,
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
		if step.TransitionResult != nil && step.TransitionResult.ActionResult != nil {
			recordPrimedWrites(step.ClassKey, step.TransitionResult.ActionResult.PrimedAssignments, out)
		}
		if step.DoActionResult != nil {
			recordPrimedWrites(step.ClassKey, step.DoActionResult.PrimedAssignments, out)
		}
		if len(step.CascadedSteps) > 0 {
			collectWrittenAttributes(step.CascadedSteps, out)
		}
	}
}

// recordPrimedWrites records attribute subKeys from primed assignments for a class.
func recordPrimedWrites(classKey identity.Key, assignments map[state.InstanceID]map[string]object.Object, out map[identity.Key]map[string]bool) {
	for _, fields := range assignments {
		for fieldName := range fields {
			if out[classKey] == nil {
				out[classKey] = make(map[string]bool)
			}
			out[classKey][identity.NormalizeSubKey(fieldName)] = true
		}
	}
}

// checkAssociationCoverage verifies every in-scope association had at least
// one link created during the simulation.
func (lc *LivenessChecker) checkAssociationCoverage(result *SimulationResult) invariants.ViolationErrors {
	if result.FinalState == nil {
		return nil
	}

	linkedAssocs := result.FinalState.Links().AllAssociationKeys()
	for hostKey := range result.FinalState.AssociationLinks().AllHostAssociationKeys() {
		linkedAssocs[hostKey] = true
	}

	var violations invariants.ViolationErrors
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

type simulationCoverage struct {
	events       map[identity.Key]bool
	queries      map[identity.Key]bool
	derivedAttrs map[identity.Key]bool
	actions      map[identity.Key]bool
}

func collectSimulationCoverage(steps []*SimulationStep, catalog *ClassCatalog, out *simulationCoverage) {
	for _, step := range steps {
		if step.EventName != "" {
			out.events[step.EventKey] = true
		}
		if step.QueryName != "" {
			out.queries[step.QueryKey] = true
		}
		if step.DerivedAttributeName != "" {
			out.derivedAttrs[step.DerivedAttributeKey] = true
		}
		for _, actionKey := range step.ExecutedActionKeys {
			out.actions[actionKey] = true
		}
		recordTransitionActionCoverage(step, catalog, out)
		if len(step.CascadedSteps) > 0 {
			collectSimulationCoverage(step.CascadedSteps, catalog, out)
		}
	}
}

func recordTransitionActionCoverage(step *SimulationStep, catalog *ClassCatalog, out *simulationCoverage) {
	if step.TransitionResult == nil || step.TransitionResult.ActionResult == nil {
		return
	}
	classInfo := catalog.GetClassInfo(step.ClassKey)
	if classInfo == nil {
		return
	}
	transition, ok := classInfo.Class.Transitions[step.TransitionResult.TransitionKey]
	if !ok || transition.ActionKey == nil {
		return
	}
	out.actions[*transition.ActionKey] = true
}

func newSimulationCoverage() *simulationCoverage {
	return &simulationCoverage{
		events:       make(map[identity.Key]bool),
		queries:      make(map[identity.Key]bool),
		derivedAttrs: make(map[identity.Key]bool),
		actions:      make(map[identity.Key]bool),
	}
}

func (lc *LivenessChecker) checkEventCoverage(result *SimulationResult) invariants.ViolationErrors {
	coverage := newSimulationCoverage()
	collectSimulationCoverage(result.Steps, lc.catalog, coverage)

	var violations invariants.ViolationErrors
	for _, classInfo := range lc.catalog.AllScopedClasses() {
		events := sortedClassEvents(classInfo)
		if len(events) == 0 {
			continue
		}
		for _, event := range events {
			if !coverage.events[event.Key] {
				violations = append(violations, invariants.NewLivenessEventNotSentViolation(
					classInfo.ClassKey,
					classInfo.Class.Name,
					event.Name,
				))
			}
		}
	}
	return violations
}

func (lc *LivenessChecker) checkQueryCoverage(result *SimulationResult) invariants.ViolationErrors {
	coverage := newSimulationCoverage()
	collectSimulationCoverage(result.Steps, lc.catalog, coverage)

	var violations invariants.ViolationErrors
	for _, classInfo := range lc.catalog.AllScopedClasses() {
		queries := sortedClassQueries(classInfo)
		if len(queries) == 0 {
			continue
		}
		for _, query := range queries {
			if !coverage.queries[query.Key] {
				violations = append(violations, invariants.NewLivenessQueryNotRunViolation(
					classInfo.ClassKey,
					classInfo.Class.Name,
					query.Name,
				))
			}
		}
	}
	return violations
}

func (lc *LivenessChecker) checkDerivedAttributeReadCoverage(result *SimulationResult) invariants.ViolationErrors {
	coverage := newSimulationCoverage()
	collectSimulationCoverage(result.Steps, lc.catalog, coverage)

	var violations invariants.ViolationErrors
	for _, classInfo := range lc.catalog.AllScopedClasses() {
		derivedAttrs := sortedExternalDerivedAttributes(lc.catalog, classInfo)
		if len(derivedAttrs) == 0 {
			continue
		}
		for _, attr := range derivedAttrs {
			if !coverage.derivedAttrs[attr.Key] {
				violations = append(violations, invariants.NewLivenessAttributeNotReadViolation(
					classInfo.ClassKey,
					classInfo.Class.Name,
					attr.Name,
				))
			}
		}
	}
	return violations
}

func (lc *LivenessChecker) checkParameterSimulationCoverage(result *SimulationResult) invariants.ViolationErrors {
	used := map[identity.Key]bool{}
	if result.SimulationCoverage != nil {
		used = result.SimulationCoverage.UsedSimulationParams
	}

	var violations invariants.ViolationErrors
	for _, classInfo := range lc.catalog.AllScopedClasses() {
		actions := sortedClassActions(classInfo)
		for _, action := range actions {
			for _, param := range action.Parameters {
				if param.Simulation == nil || param.Simulation.Specification == nil {
					continue
				}
				if used[param.Key] {
					continue
				}
				violations = append(violations, invariants.NewLivenessParameterSimulationNotUsedViolation(
					classInfo.ClassKey,
					classInfo.Class.Name,
					action.Name,
					param.Name,
				))
			}
		}
	}
	return violations
}

func (lc *LivenessChecker) checkActionCoverage(result *SimulationResult) invariants.ViolationErrors {
	coverage := newSimulationCoverage()
	collectSimulationCoverage(result.Steps, lc.catalog, coverage)

	var violations invariants.ViolationErrors
	for _, classInfo := range lc.catalog.AllScopedClasses() {
		actions := sortedClassActions(classInfo)
		if len(actions) == 0 {
			continue
		}
		for _, action := range actions {
			if !coverage.actions[action.Key] {
				violations = append(violations, invariants.NewLivenessActionNotExecutedViolation(
					classInfo.ClassKey,
					classInfo.Class.Name,
					action.Name,
				))
			}
		}
	}
	return violations
}

func sortedClassEvents(classInfo *ClassInfo) []model_state.Event {
	events := make([]model_state.Event, 0, len(classInfo.Class.Events))
	for _, event := range classInfo.Class.Events {
		events = append(events, event)
	}
	sort.Slice(events, func(i, j int) bool { return events[i].Name < events[j].Name })
	return events
}

func sortedClassQueries(classInfo *ClassInfo) []model_state.Query {
	queries := make([]model_state.Query, 0, len(classInfo.Class.Queries))
	for _, query := range classInfo.Class.Queries {
		queries = append(queries, query)
	}
	sort.Slice(queries, func(i, j int) bool { return queries[i].Name < queries[j].Name })
	return queries
}

func sortedExternalDerivedAttributes(catalog *ClassCatalog, classInfo *ClassInfo) []model_class.Attribute {
	attrs := catalog.ExternalDerivedAttributes(classInfo.ClassKey)
	sort.Slice(attrs, func(i, j int) bool { return attrs[i].Name < attrs[j].Name })
	return attrs
}

func sortedClassActions(classInfo *ClassInfo) []model_state.Action {
	actions := make([]model_state.Action, 0, len(classInfo.Class.Actions))
	for _, action := range classInfo.Class.Actions {
		actions = append(actions, action)
	}
	sort.Slice(actions, func(i, j int) bool { return actions[i].Name < actions[j].Name })
	return actions
}

package invariants

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
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
	acInactive  map[identity.Key]map[string]bool
}

// NewMultiplicityChecker builds association multiplicity metadata from the model.
func NewMultiplicityChecker(model *core.Model) *MultiplicityChecker {
	checker := &MultiplicityChecker{
		classAssocs: make(map[identity.Key][]associationBinding),
		acInactive:  make(map[identity.Key]map[string]bool),
	}

	classes := make(map[identity.Key]model_class.Class)
	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			for _, class := range subdomain.Classes {
				classes[class.Key] = class
			}
		}
	}

	allAssocs := model.GetClassAssociations()
	for _, assoc := range allAssocs {
		if model_class.IsReverseInvariantOnlyAssociation(allAssocs, assoc) {
			continue
		}
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
		if assoc.AssociationClassKey != nil {
			if acClass, ok := classes[*assoc.AssociationClassKey]; ok {
				checker.acInactive[*assoc.AssociationClassKey] = inactiveAssociationClassStates(acClass)
			}
		}
	}

	return checker
}

// CheckState validates all association multiplicities across every live instance.
func (c *MultiplicityChecker) CheckState(simState *state.SimulationState) ViolationErrors {
	var violations ViolationErrors
	for _, instance := range simState.AllInstances() {
		violations = append(violations, c.CheckInstance(instance, simState)...)
	}
	return violations
}

// CheckInstance validates all multiplicity constraints for a single instance.
func (c *MultiplicityChecker) CheckInstance(
	instance *state.ClassInstance,
	simState *state.SimulationState,
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
	fromID state.InstanceID,
	binding associationBinding,
	simState *state.SimulationState,
) int {
	if binding.association.AssociationClassKey != nil {
		return c.countActiveAssociationLinksFrom(fromID, binding.association.Key, simState)
	}
	linked := simState.GetLinkedForward(fromID, binding.association.Key)
	return c.countActiveLinkedInstances(linked, simState)
}

func (c *MultiplicityChecker) countActiveReverseLinks(
	toID state.InstanceID,
	binding associationBinding,
	simState *state.SimulationState,
) int {
	if binding.association.AssociationClassKey != nil {
		return c.countActiveAssociationLinksTo(toID, binding.association.Key, simState)
	}
	linked := simState.GetLinkedReverse(toID, binding.association.Key)
	return c.countActiveLinkedInstances(linked, simState)
}

func (c *MultiplicityChecker) countActiveAssociationLinksFrom(
	fromID state.InstanceID,
	hostAssocKey identity.Key,
	simState *state.SimulationState,
) int {
	links := simState.AssociationLinksFromEndpoint(hostAssocKey, fromID)
	count := 0
	for _, link := range links {
		linkInst := simState.GetInstance(link.LinkInstanceID)
		if linkInst == nil {
			continue
		}
		if c.isActiveAssociationClassInstance(linkInst.ClassKey, instanceStateName(linkInst)) {
			count++
		}
	}
	return count
}

func (c *MultiplicityChecker) countActiveAssociationLinksTo(
	toID state.InstanceID,
	hostAssocKey identity.Key,
	simState *state.SimulationState,
) int {
	links := simState.AssociationLinksToEndpoint(hostAssocKey, toID)
	count := 0
	for _, link := range links {
		linkInst := simState.GetInstance(link.LinkInstanceID)
		if linkInst == nil {
			continue
		}
		if c.isActiveAssociationClassInstance(linkInst.ClassKey, instanceStateName(linkInst)) {
			count++
		}
	}
	return count
}

func (c *MultiplicityChecker) countActiveLinkedInstances(
	linked []state.InstanceID,
	simState *state.SimulationState,
) int {
	count := 0
	for _, id := range linked {
		inst := simState.GetInstance(id)
		if inst == nil {
			continue
		}
		if !c.isActiveAssociationClassInstance(inst.ClassKey, instanceStateName(inst)) {
			continue
		}
		count++
	}
	return count
}

func (c *MultiplicityChecker) isActiveAssociationClassInstance(classKey identity.Key, stateName string) bool {
	inactive, isAC := c.acInactive[classKey]
	if !isAC || len(inactive) == 0 {
		return true
	}
	return !inactive[stateName]
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

func instanceStateName(instance *state.ClassInstance) string {
	if instance == nil {
		return ""
	}
	stateAttr := instance.GetAttribute("_state")
	if stateAttr == nil {
		return ""
	}
	if strObj, ok := stateAttr.(*object.String); ok {
		return strObj.Value()
	}
	return ""
}

// inactiveAssociationClassStates marks AC states that cannot reach any creation target state.
func inactiveAssociationClassStates(class model_class.Class) map[string]bool {
	creationStates := creationTargetStateNames(class)
	inactive := make(map[string]bool)
	for _, state := range class.States {
		if !stateCanReachCreation(class, state.Name, creationStates) {
			inactive[state.Name] = true
		}
	}
	return inactive
}

func stateCanReachCreation(class model_class.Class, startName string, creationStates map[string]bool) bool {
	if creationStates[startName] {
		return true
	}
	reachable := statesReachableFrom(class, map[string]bool{startName: true})
	for name := range creationStates {
		if reachable[name] {
			return true
		}
	}
	return false
}

func creationTargetStateNames(class model_class.Class) map[string]bool {
	targets := make(map[string]bool)
	for _, t := range class.Transitions {
		if t.FromStateKey != nil || t.ToStateKey == nil {
			continue
		}
		if name := stateKeyToNameInClass(*t.ToStateKey, class); name != "" {
			targets[name] = true
		}
	}
	return targets
}

func statesReachableFrom(class model_class.Class, seeds map[string]bool) map[string]bool {
	reachable := make(map[string]bool)
	queue := make([]string, 0, len(seeds))
	for name := range seeds {
		reachable[name] = true
		queue = append(queue, name)
	}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		currentKey := stateNameToKeyInClass(current, class)
		if currentKey == nil {
			continue
		}
		for _, t := range class.Transitions {
			if t.FromStateKey == nil || *t.FromStateKey != *currentKey || t.ToStateKey == nil {
				continue
			}
			nextName := stateKeyToNameInClass(*t.ToStateKey, class)
			if nextName == "" || reachable[nextName] {
				continue
			}
			reachable[nextName] = true
			queue = append(queue, nextName)
		}
	}
	return reachable
}

func stateKeyToNameInClass(stateKey identity.Key, class model_class.Class) string {
	if s, ok := class.States[stateKey]; ok {
		return s.Name
	}
	return ""
}

func stateNameToKeyInClass(stateName string, class model_class.Class) *identity.Key {
	for _, s := range class.States {
		if s.Name == stateName {
			key := s.Key
			return &key
		}
	}
	return nil
}

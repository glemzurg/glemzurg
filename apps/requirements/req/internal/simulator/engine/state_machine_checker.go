package engine

import (
	"sort"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/invariants"
)

// StateMachineChecker reports structural defects in scoped class state machines.
type StateMachineChecker struct {
	catalog *ClassCatalog
}

// NewStateMachineChecker creates a checker for state machine completeness rules.
func NewStateMachineChecker(catalog *ClassCatalog) *StateMachineChecker {
	return &StateMachineChecker{catalog: catalog}
}

// Check returns violations for in-scope classes whose state machine omits the system _new event.
func (smc *StateMachineChecker) Check() invariants.ViolationErrors {
	var violations invariants.ViolationErrors

	classes := smc.catalog.AllScopedClasses()
	sort.Slice(classes, func(i, j int) bool {
		return classes[i].Class.Name < classes[j].Class.Name
	})

	for _, classInfo := range classes {
		if !classHasStateMachine(classInfo.Class) || classStateMachineHasNewEvent(classInfo.Class) {
			continue
		}
		violations = append(violations, invariants.NewStateMachineIncompleteViolation(
			classInfo.ClassKey,
			classInfo.Class.Name,
		))
	}

	return violations
}

// classHasStateMachine reports whether the class declares any states or transitions.
func classHasStateMachine(class model_class.Class) bool {
	return len(class.States) > 0 || len(class.Transitions) > 0
}

// classStateMachineHasNewEvent reports whether the class defines the system _new event.
func classStateMachineHasNewEvent(class model_class.Class) bool {
	for _, event := range class.Events {
		if model_state.IsSystemCreationEvent(event.Name) {
			return true
		}
	}
	return false
}

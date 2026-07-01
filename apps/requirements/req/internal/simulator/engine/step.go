package engine

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/actions"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/invariants"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
)

// StepKind describes the type of simulation step.
type StepKind int

const (
	// StepKindCreation is a new instance created (from initial state).
	StepKindCreation StepKind = iota
	// StepKindNormal is a normal state transition.
	StepKindNormal
	// StepKindDestroy is an instance destroyed (to final state).
	StepKindDestroy
)

// String returns a human-readable name for the step kind.
func (k StepKind) String() string {
	switch k {
	case StepKindCreation:
		return "creation"
	case StepKindNormal:
		return "normal"
	case StepKindDestroy:
		return "destroy"
	default:
		return "unknown"
	}
}

// SimulationStep records one atomic unit of simulation work.
type SimulationStep struct {
	// StepNumber is the ordinal position in the simulation (1-based).
	StepNumber int

	// Kind is the type of step (creation, normal, destroy).
	Kind StepKind

	// ClassKey is the class being acted upon.
	ClassKey identity.Key

	// ClassName is the human-readable name of the class.
	ClassName string

	// EventKey is the event that triggered this step.
	EventKey identity.Key

	// EventName is the human-readable name of the event.
	EventName string

	// InstanceID is the instance that was acted upon (assigned after creation).
	InstanceID state.InstanceID

	// FromState is the state name before the transition (empty for creation).
	FromState string

	// ToState is the state name after the transition (empty for destroy).
	ToState string

	// Parameters are the event parameters that were passed.
	Parameters map[string]object.Object

	// TransitionResult is the detailed result from the action executor.
	TransitionResult *actions.TransitionResult

	// DoActionResult is the result from a "do" action execution (nil for transition steps).
	DoActionResult *actions.ActionResult

	// QueryResult is the result from a query execution (nil for non-query steps).
	QueryResult *actions.QueryResult

	// QueryKey is the query invoked (for query steps).
	QueryKey identity.Key

	// QueryName is the human-readable query name (for query steps).
	QueryName string

	// DerivedAttributeKey is the derived attribute read (for derived-read steps).
	DerivedAttributeKey identity.Key

	// DerivedAttributeName is the human-readable derived attribute name.
	DerivedAttributeName string

	// DerivedReadValue is the computed value from a derived attribute read step.
	DerivedReadValue object.Object

	// ExecutedActionKeys records actions run during this step (transition, do, entry, exit).
	ExecutedActionKeys []identity.Key

	// CascadedSteps holds child steps from creation chain cascading.
	CascadedSteps []*SimulationStep

	// Violations contains any invariant violations detected during this step.
	Violations invariants.ViolationErrors
}

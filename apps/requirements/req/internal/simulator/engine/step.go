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
	// StepKindDeletion is an instance deleted (to final state).
	StepKindDeletion
)

// String returns a human-readable name for the step kind.
func (k StepKind) String() string {
	switch k {
	case StepKindCreation:
		return "creation"
	case StepKindNormal:
		return "normal"
	case StepKindDeletion:
		return "deletion"
	default:
		return "unknown"
	}
}

// SimulationStep records one atomic unit of simulation work.
type SimulationStep struct {
	// StepNumber is the ordinal position in the simulation (1-based).
	StepNumber int

	// Kind is the type of step (creation, normal, deletion).
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

	// ToState is the state name after the transition (empty for deletion).
	ToState string

	// Parameters are the event parameters that were passed.
	Parameters map[string]object.Object

	// TransitionResult is the detailed result from the action executor.
	TransitionResult *actions.TransitionResult

	// DoActionResult is the result from a "do" action execution (nil for transition steps).
	DoActionResult *actions.ActionResult

	// CascadedSteps holds child steps from creation chain cascading.
	CascadedSteps []*SimulationStep

	// Violations contains any invariant violations detected during this step.
	Violations invariants.ViolationList
}

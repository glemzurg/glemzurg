package trace

import (
	"encoding/json"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/actions"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/engine"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/invariants"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
	"github.com/stretchr/testify/suite"
)

func TestTraceSuite(t *testing.T) {
	suite.Run(t, new(TraceSuite))
}

type TraceSuite struct {
	suite.Suite
}

func mustKey(s string) identity.Key {
	k, err := identity.ParseKey(s)
	if err != nil {
		panic(err)
	}
	return k
}

func (s *TraceSuite) TestEmptyResult() {
	result := &engine.SimulationResult{
		StepsTaken:        0,
		TerminationReason: "deadlock",
	}

	tr := FromResult(result)

	s.Equal(0, tr.StepsTaken)
	s.Equal("deadlock", tr.TerminationReason)
	s.Empty(tr.Steps)
	s.Nil(tr.FinalState)
}

func (s *TraceSuite) TestCreationStep() {
	classKey := mustKey("domain/d/subdomain/s/class/order")
	eventKey := mustKey("domain/d/subdomain/s/class/order/event/create")

	result := &engine.SimulationResult{
		StepsTaken:        1,
		TerminationReason: "max_steps",
		Steps: []*engine.SimulationStep{
			{
				StepNumber: 1,
				Kind:       engine.StepKindCreation,
				ClassKey:   classKey,
				ClassName:  "Order",
				EventKey:   eventKey,
				EventName:  "create",
				InstanceID: 1,
				ToState:    "Open",
			},
		},
	}

	tr := FromResult(result)

	s.Len(tr.Steps, 1)
	step := tr.Steps[0]
	s.Equal(1, step.StepNumber)
	s.Equal("creation", step.Kind)
	s.Equal("Order", step.ClassName)
	s.Equal(classKey.String(), step.ClassKey)
	s.Equal("create", step.EventName)
	s.Equal(uint64(1), step.InstanceID)
	s.Empty(step.FromState)
	s.Equal("Open", step.ToState)
}

func (s *TraceSuite) TestNormalTransitionStep() {
	classKey := mustKey("domain/d/subdomain/s/class/order")
	eventKey := mustKey("domain/d/subdomain/s/class/order/event/close")

	result := &engine.SimulationResult{
		StepsTaken:        1,
		TerminationReason: "max_steps",
		Steps: []*engine.SimulationStep{
			{
				StepNumber: 2,
				Kind:       engine.StepKindNormal,
				ClassKey:   classKey,
				ClassName:  "Order",
				EventKey:   eventKey,
				EventName:  "close",
				InstanceID: 1,
				FromState:  "Open",
				ToState:    "Closed",
				TransitionResult: &actions.TransitionResult{
					InstanceID: 1,
					FromState:  "Open",
					ToState:    "Closed",
					ActionResult: &actions.ActionResult{
						InstanceID: 1,
						PrimedAssignments: map[state.InstanceID]map[string]object.Object{
							1: {
								"amount": object.NewInteger(42),
							},
						},
					},
				},
			},
		},
	}

	tr := FromResult(result)

	s.Len(tr.Steps, 1)
	step := tr.Steps[0]
	s.Equal("normal", step.Kind)
	s.Equal("Open", step.FromState)
	s.Equal("Closed", step.ToState)
	s.Require().NotNil(step.Assignments)
	s.Equal("42", step.Assignments["amount"])
}

func (s *TraceSuite) TestDoActionStep() {
	classKey := mustKey("domain/d/subdomain/s/class/order")

	result := &engine.SimulationResult{
		StepsTaken:        1,
		TerminationReason: "max_steps",
		Steps: []*engine.SimulationStep{
			{
				StepNumber: 3,
				Kind:       engine.StepKindNormal,
				ClassKey:   classKey,
				ClassName:  "Order",
				InstanceID: 1,
				DoActionResult: &actions.ActionResult{
					InstanceID: 1,
					PrimedAssignments: map[state.InstanceID]map[string]object.Object{
						1: {
							"counter": object.NewInteger(5),
						},
					},
				},
			},
		},
	}

	tr := FromResult(result)

	s.Len(tr.Steps, 1)
	step := tr.Steps[0]
	s.Require().NotNil(step.Assignments)
	s.Equal("5", step.Assignments["counter"])
}

func (s *TraceSuite) TestCascadedSteps() {
	orderKey := mustKey("domain/d/subdomain/s/class/order")
	itemKey := mustKey("domain/d/subdomain/s/class/item")

	result := &engine.SimulationResult{
		StepsTaken:        1,
		TerminationReason: "max_steps",
		Steps: []*engine.SimulationStep{
			{
				StepNumber: 1,
				Kind:       engine.StepKindCreation,
				ClassKey:   orderKey,
				ClassName:  "Order",
				InstanceID: 1,
				ToState:    "Open",
				CascadedSteps: []*engine.SimulationStep{
					{
						StepNumber: 0,
						Kind:       engine.StepKindCreation,
						ClassKey:   itemKey,
						ClassName:  "Item",
						InstanceID: 2,
						ToState:    "Active",
					},
				},
			},
		},
	}

	tr := FromResult(result)

	s.Len(tr.Steps, 1)
	s.Len(tr.Steps[0].CascadedSteps, 1)
	cascaded := tr.Steps[0].CascadedSteps[0]
	s.Equal("creation", cascaded.Kind)
	s.Equal("Item", cascaded.ClassName)
	s.Equal(uint64(2), cascaded.InstanceID)
}

func (s *TraceSuite) TestStepWithViolations() {
	classKey := mustKey("domain/d/subdomain/s/class/order")

	result := &engine.SimulationResult{
		StepsTaken:        1,
		TerminationReason: "violation",
		Steps: []*engine.SimulationStep{
			{
				StepNumber: 1,
				Kind:       engine.StepKindNormal,
				ClassKey:   classKey,
				ClassName:  "Order",
				InstanceID: 1,
				Violations: invariants.ViolationErrors{
					{
						Type:    invariants.ViolationTypeModelInvariant,
						Message: "invariant failed: x > 0",
					},
				},
			},
		},
	}

	tr := FromResult(result)

	s.Len(tr.Steps, 1)
	s.Len(tr.Steps[0].Violations, 1)
	s.Equal("invariant failed: x > 0", tr.Steps[0].Violations[0])
}

func (s *TraceSuite) TestStepWithParameters() {
	classKey := mustKey("domain/d/subdomain/s/class/order")

	result := &engine.SimulationResult{
		StepsTaken:        1,
		TerminationReason: "max_steps",
		Steps: []*engine.SimulationStep{
			{
				StepNumber: 1,
				Kind:       engine.StepKindCreation,
				ClassKey:   classKey,
				ClassName:  "Order",
				InstanceID: 1,
				ToState:    "Open",
				Parameters: map[string]object.Object{
					"qty": object.NewInteger(10),
				},
			},
		},
	}

	tr := FromResult(result)

	s.Require().NotNil(tr.Steps[0].Parameters)
	s.Equal("10", tr.Steps[0].Parameters["qty"])
}

func (s *TraceSuite) TestFinalState() {
	simState := state.NewSimulationState()
	classKey := mustKey("domain/d/subdomain/s/class/order")

	attrs := object.NewRecord()
	attrs.Set("amount", object.NewInteger(100))
	simState.CreateInstance(classKey, attrs)

	result := &engine.SimulationResult{
		StepsTaken:        1,
		TerminationReason: "max_steps",
		FinalState:        simState,
	}

	tr := FromResult(result)

	s.Require().NotNil(tr.FinalState)
	s.Equal(1, tr.FinalState.InstanceCount)
	s.Require().Len(tr.FinalState.Instances, 1)
	inst := tr.FinalState.Instances[0]
	s.Equal(uint64(1), inst.InstanceID)
	s.Equal(classKey.String(), inst.ClassKey)
	s.Equal("100", inst.Attributes["amount"])
}

func (s *TraceSuite) TestAssociationClassMaterializationStep() {
	classKey := mustKey("domain/d/subdomain/s/class/link_def")
	fromClassKey := mustKey("domain/d/subdomain/s/class/partner")
	toClassKey := mustKey("domain/d/subdomain/s/class/jurisdiction")
	hostAssocKey := mustKey("domain/d/subdomain/s/cassociation/class/partner/class/jurisdiction/configures")

	result := &engine.SimulationResult{
		StepsTaken:        1,
		TerminationReason: "max_steps",
		Steps: []*engine.SimulationStep{
			{
				StepNumber: 1,
				Kind:       engine.StepKindCreation,
				ClassKey:   classKey,
				ClassName:  "LinkDef",
				EventName:  "Add",
				InstanceID: 3,
				ToState:    "Active",
				TransitionResult: &actions.TransitionResult{
					InstanceID:  3,
					WasCreation: true,
					AssociationMaterialization: &actions.AssociationMaterialization{
						HostAssociationName: "Configures",
						HostAssociationKey:  hostAssocKey,
						FromClassName:       "Partner",
						FromClassKey:        fromClassKey,
						ToClassName:         "Jurisdiction",
						ToClassKey:          toClassKey,
						FromInstanceID:      1,
						ToInstanceID:        2,
					},
				},
			},
		},
	}

	tr := FromResult(result)
	s.Require().Len(tr.Steps, 1)
	mat := tr.Steps[0].AssociationMaterialization
	s.Require().NotNil(mat)
	s.Equal("Configures", mat.AssociationName)
	s.Equal(hostAssocKey.String(), mat.AssociationKey)
	s.Equal("Partner", mat.FromClassName)
	s.Equal(fromClassKey.String(), mat.FromClassKey)
	s.Equal("Jurisdiction", mat.ToClassName)
	s.Equal(toClassKey.String(), mat.ToClassKey)
	s.Equal(uint64(1), mat.FromInstanceID)
	s.Equal(uint64(2), mat.ToInstanceID)

	text := tr.FormatText()
	s.Contains(text, "materializes: Configures (Partner#1 -> Jurisdiction#2)")
}

func (s *TraceSuite) TestFormatTextOutput() {
	classKey := mustKey("domain/d/subdomain/s/class/order")

	result := &engine.SimulationResult{
		StepsTaken:        1,
		TerminationReason: "max_steps",
		Steps: []*engine.SimulationStep{
			{
				StepNumber: 1,
				Kind:       engine.StepKindCreation,
				ClassKey:   classKey,
				ClassName:  "Order",
				InstanceID: 1,
				ToState:    "Open",
				EventName:  "create",
			},
		},
	}

	tr := FromResult(result)
	text := tr.FormatText()

	s.Contains(text, "Simulation: 1 steps, terminated: max_steps")
	s.Contains(text, "CREATE Order#1 -> Open")
	s.Contains(text, "(event: create)")
}

func (s *TraceSuite) TestFormatJSONRoundTrip() {
	classKey := mustKey("domain/d/subdomain/s/class/order")

	result := &engine.SimulationResult{
		StepsTaken:        2,
		TerminationReason: "max_steps",
		Steps: []*engine.SimulationStep{
			{
				StepNumber: 1,
				Kind:       engine.StepKindCreation,
				ClassKey:   classKey,
				ClassName:  "Order",
				InstanceID: 1,
				ToState:    "Open",
			},
		},
	}

	tr := FromResult(result)
	data, err := tr.FormatJSON()
	s.Require().NoError(err)

	var decoded SimulationTrace
	err = json.Unmarshal(data, &decoded)
	s.Require().NoError(err)

	s.Equal(2, decoded.StepsTaken)
	s.Equal("max_steps", decoded.TerminationReason)
	s.Len(decoded.Steps, 1)
	s.Equal("creation", decoded.Steps[0].Kind)
	s.Equal("Order", decoded.Steps[0].ClassName)
}

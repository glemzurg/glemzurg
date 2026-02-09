package engine

import (
	"math/rand"
	"testing"

	"github.com/glemzurg/go-tlaplus/internal/identity"
	"github.com/glemzurg/go-tlaplus/internal/req_model/model_class"
	"github.com/glemzurg/go-tlaplus/internal/req_model/model_state"
	"github.com/glemzurg/go-tlaplus/internal/simulator/object"
	"github.com/glemzurg/go-tlaplus/internal/simulator/state"
	"github.com/stretchr/testify/suite"
)

type ActionSelectorSuite struct {
	suite.Suite
}

func TestActionSelectorSuite(t *testing.T) {
	suite.Run(t, new(ActionSelectorSuite))
}

func (s *ActionSelectorSuite) TestCreationEligibleWhenNoInstancesExist() {
	orderClass, orderKey := testOrderClass()
	model := testModel(classEntry(orderClass, orderKey))

	catalog := NewClassCatalog(model)
	rng := rand.New(rand.NewSource(42))
	selector := NewActionSelector(catalog, rng)

	simState := state.NewSimulationState()

	action, err := selector.SelectAction(simState)
	s.NoError(err)
	s.NotNil(action)
	s.True(action.IsCreation)
	s.Nil(action.Instance)
}

func (s *ActionSelectorSuite) TestNormalEventsEligibleForExistingInstances() {
	orderClass, orderKey := testOrderClass()
	model := testModel(classEntry(orderClass, orderKey))

	catalog := NewClassCatalog(model)
	rng := rand.New(rand.NewSource(42))
	selector := NewActionSelector(catalog, rng)

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("_state", object.NewString("Open"))
	attrs.Set("amount", object.NewInteger(0))
	simState.CreateInstance(orderKey, attrs)

	// With an instance in "Open" state, should get at least creation + close event.
	// Run multiple selections to verify both types exist.
	foundCreation := false
	foundNormal := false
	for i := 0; i < 50; i++ {
		action, err := selector.SelectAction(simState)
		s.NoError(err)
		if action.IsCreation {
			foundCreation = true
		} else if !action.IsDo {
			foundNormal = true
		}
		if foundCreation && foundNormal {
			break
		}
	}
	s.True(foundCreation, "should find creation events")
	s.True(foundNormal, "should find normal events")
}

func (s *ActionSelectorSuite) TestDeadlockWhenNoActionsEligible() {
	// A class with no creation transitions and no existing instances → deadlock.
	classKey := mustKey("domain/d/subdomain/s/class/stuck")
	stateActiveKey := mustKey("domain/d/subdomain/s/class/stuck/state/active")
	eventUpdateKey := mustKey("domain/d/subdomain/s/class/stuck/event/update")
	transUpdateKey := mustKey("domain/d/subdomain/s/class/stuck/transition/update")

	class := model_class.Class{
		Key:        classKey,
		Name:       "Stuck",
		Attributes: map[identity.Key]model_class.Attribute{},
		States: map[identity.Key]model_state.State{
			stateActiveKey: {Key: stateActiveKey, Name: "Active"},
		},
		Events: map[identity.Key]model_state.Event{
			eventUpdateKey: {Key: eventUpdateKey, Name: "update"},
		},
		Guards:  map[identity.Key]model_state.Guard{},
		Actions: map[identity.Key]model_state.Action{},
		Queries: map[identity.Key]model_state.Query{},
		Transitions: map[identity.Key]model_state.Transition{
			transUpdateKey: {
				Key:          transUpdateKey,
				FromStateKey: &stateActiveKey,
				EventKey:     eventUpdateKey,
				ToStateKey:   &stateActiveKey,
			},
		},
	}

	model := testModel(classEntry(class, classKey))
	catalog := NewClassCatalog(model)
	rng := rand.New(rand.NewSource(42))
	selector := NewActionSelector(catalog, rng)

	// No creation transitions and no instances → deadlock.
	simState := state.NewSimulationState()
	_, err := selector.SelectAction(simState)
	s.Error(err)
	s.Contains(err.Error(), "deadlock")
}

func (s *ActionSelectorSuite) TestDoActionsEligibleAsEvents() {
	classKey := mustKey("domain/d/subdomain/s/class/counter")
	stateActiveKey := mustKey("domain/d/subdomain/s/class/counter/state/active")
	eventCreateKey := mustKey("domain/d/subdomain/s/class/counter/event/create")
	actionDoKey := mustKey("domain/d/subdomain/s/class/counter/action/do_count")
	stateActionKey := mustKey("domain/d/subdomain/s/class/counter/state/active/saction/do/do_count")
	transCreateKey := mustKey("domain/d/subdomain/s/class/counter/transition/create")

	class := model_class.Class{
		Key:        classKey,
		Name:       "Counter",
		Attributes: map[identity.Key]model_class.Attribute{},
		States: map[identity.Key]model_state.State{
			stateActiveKey: {
				Key:  stateActiveKey,
				Name: "Active",
				Actions: []model_state.StateAction{
					{Key: stateActionKey, ActionKey: actionDoKey, When: "do"},
				},
			},
		},
		Events: map[identity.Key]model_state.Event{
			eventCreateKey: {Key: eventCreateKey, Name: "create"},
		},
		Guards: map[identity.Key]model_state.Guard{},
		Actions: map[identity.Key]model_state.Action{
			actionDoKey: {
				Key:  actionDoKey,
				Name: "DoCount",
				TlaGuarantees: []string{
					"self.count' = self.count + 1",
				},
			},
		},
		Queries: map[identity.Key]model_state.Query{},
		Transitions: map[identity.Key]model_state.Transition{
			transCreateKey: {
				Key:          transCreateKey,
				FromStateKey: nil,
				EventKey:     eventCreateKey,
				ToStateKey:   &stateActiveKey,
			},
		},
	}

	model := testModel(classEntry(class, classKey))
	catalog := NewClassCatalog(model)
	rng := rand.New(rand.NewSource(42))
	selector := NewActionSelector(catalog, rng)

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("_state", object.NewString("Active"))
	attrs.Set("count", object.NewInteger(0))
	simState.CreateInstance(classKey, attrs)

	// Should find "do" actions as eligible.
	foundDo := false
	for i := 0; i < 50; i++ {
		action, err := selector.SelectAction(simState)
		s.NoError(err)
		if action.IsDo {
			foundDo = true
			s.Equal("DoCount", action.DoAction.Name)
			s.NotNil(action.Instance)
			break
		}
	}
	s.True(foundDo, "should find do actions as eligible events")
}

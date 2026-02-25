package engine

import (
	"math/rand"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
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

	eventUpdate := helper.Must(model_state.NewEvent(eventUpdateKey, "update", "", nil))

	class := helper.Must(model_class.NewClass(classKey, "Stuck", "", nil, nil, nil, ""))
	class.Attributes = map[identity.Key]model_class.Attribute{}
	class.States = map[identity.Key]model_state.State{
		stateActiveKey: {Key: stateActiveKey, Name: "Active"},
	}
	class.Events = map[identity.Key]model_state.Event{
		eventUpdateKey: eventUpdate,
	}
	class.Guards = map[identity.Key]model_state.Guard{}
	class.Actions = map[identity.Key]model_state.Action{}
	class.Queries = map[identity.Key]model_state.Query{}
	class.Transitions = map[identity.Key]model_state.Transition{
		transUpdateKey: {
			Key:          transUpdateKey,
			FromStateKey: &stateActiveKey,
			EventKey:     eventUpdateKey,
			ToStateKey:   &stateActiveKey,
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

	eventCreate := helper.Must(model_state.NewEvent(eventCreateKey, "create", "", nil))

	guaranteeKey := helper.Must(identity.NewActionGuaranteeKey(actionDoKey, "0"))
	guaranteeLogic := helper.Must(model_logic.NewLogic(guaranteeKey, model_logic.LogicTypeStateChange, "Postcondition.", model_logic.NotationTLAPlus, "self.count' = self.count + 1"))
	actionDo := helper.Must(model_state.NewAction(actionDoKey, "DoCount", "", nil, []model_logic.Logic{guaranteeLogic}, nil, nil))

	class := helper.Must(model_class.NewClass(classKey, "Counter", "", nil, nil, nil, ""))
	class.Attributes = map[identity.Key]model_class.Attribute{}
	class.States = map[identity.Key]model_state.State{
		stateActiveKey: {
			Key:  stateActiveKey,
			Name: "Active",
			Actions: []model_state.StateAction{
				{Key: stateActionKey, ActionKey: actionDoKey, When: "do"},
			},
		},
	}
	class.Events = map[identity.Key]model_state.Event{
		eventCreateKey: eventCreate,
	}
	class.Guards = map[identity.Key]model_state.Guard{}
	class.Actions = map[identity.Key]model_state.Action{
		actionDoKey: actionDo,
	}
	class.Queries = map[identity.Key]model_state.Query{}
	class.Transitions = map[identity.Key]model_state.Transition{
		transCreateKey: {
			Key:          transCreateKey,
			FromStateKey: nil,
			EventKey:     eventCreateKey,
			ToStateKey:   &stateActiveKey,
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

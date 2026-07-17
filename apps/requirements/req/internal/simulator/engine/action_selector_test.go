package engine

import (
	"math/rand"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_data_type"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
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
	rng := rand.New(rand.NewSource(42)) //nolint:gosec // deterministic seed for reproducible tests //nolint:gosec // deterministic seed for reproducible tests
	selector := NewActionSelector(catalog, nil, nil, nil, rng)

	simState := state.NewSimulationState()

	action, err := selector.SelectAction(simState)
	s.Require().NoError(err)
	s.NotNil(action)
	s.True(action.IsCreation)
	s.Nil(action.Instance)
}

func (s *ActionSelectorSuite) TestNormalEventsEligibleForExistingInstances() {
	orderClass, orderKey := testOrderClass()
	model := testModel(classEntry(orderClass, orderKey))

	catalog := NewClassCatalog(model)
	rng := rand.New(rand.NewSource(42)) //nolint:gosec // deterministic seed for reproducible tests //nolint:gosec // deterministic seed for reproducible tests
	selector := NewActionSelector(catalog, nil, nil, nil, rng)

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("_state", object.NewString("Open"))
	attrs.Set("amount", object.NewInteger(0))
	simState.CreateInstance(orderKey, attrs)

	// With an instance in "Open" state, should get at least creation + close event.
	// Run multiple selections to verify both types exist.
	foundCreation := false
	foundNormal := false
	for range 50 {
		action, err := selector.SelectAction(simState)
		s.Require().NoError(err)
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

	eventUpdate := model_state.NewEvent(eventUpdateKey, "update", "", nil)

	stateActive := model_state.NewState(stateActiveKey, "Active", "", "")

	transUpdate := model_state.NewTransition(transUpdateKey, eventUpdateKey, model_state.TransitionStateKeys{FromStateKey: &stateActiveKey, ToStateKey: &stateActiveKey}, model_state.TransitionLogicKeys{GuardKey: nil, ActionKey: nil}, "")

	class := model_class.NewClass(classKey, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: "Stuck", Details: "", UnfinishedNotes: "", UmlComment: ""})
	class.SetAttributes(nil)
	class.SetStates(map[identity.Key]model_state.State{
		stateActiveKey: stateActive,
	})
	class.SetEvents(map[identity.Key]model_state.Event{
		eventUpdateKey: eventUpdate,
	})
	class.SetGuards(map[identity.Key]model_state.Guard{})
	class.SetActions(map[identity.Key]model_state.Action{})
	class.SetQueries(map[identity.Key]model_state.Query{})
	class.SetTransitions(map[identity.Key]model_state.Transition{
		transUpdateKey: transUpdate,
	})

	model := testModel(classEntry(class, classKey))
	catalog := NewClassCatalog(model)
	rng := rand.New(rand.NewSource(42)) //nolint:gosec // deterministic seed for reproducible tests
	selector := NewActionSelector(catalog, nil, nil, nil, rng)

	// No creation transitions and no instances → deadlock.
	simState := state.NewSimulationState()
	_, err := selector.SelectAction(simState)
	s.Require().Error(err)
	s.Contains(err.Error(), "deadlock")
}

func (s *ActionSelectorSuite) TestCreationBlockedUntilObjectParamClassHasInstance() {
	// Owner._new(Peer) requires an in-scope Peer instance before Owner creation is eligible.
	ownerClass, ownerKey, peerClass, peerKey := ownerWithObjectParamPeer()
	model := testModel(classEntry(ownerClass, ownerKey), classEntry(peerClass, peerKey))
	catalog := NewClassCatalog(model)
	selector := NewActionSelector(catalog, nil, state.NewBindingsBuilder(state.NewSimulationState()), nil, rand.New(rand.NewSource(1))) //nolint:gosec

	simState := state.NewSimulationState()

	// Only Peer creation should be eligible (Owner needs a Peer instance).
	for range 20 {
		action, err := selector.SelectAction(simState)
		s.Require().NoError(err)
		s.True(action.IsCreation)
		s.Equal(peerKey, action.Class.ClassKey, "owner creation must wait for peer instances")
	}

	// After a Peer exists, Owner creation becomes eligible.
	peerAttrs := object.NewRecord()
	peerAttrs.Set("_state", object.NewString("Active"))
	simState.CreateInstance(peerKey, peerAttrs)

	foundOwner := false
	for range 40 {
		action, err := selector.SelectAction(simState)
		s.Require().NoError(err)
		if action.IsCreation && action.Class.ClassKey == ownerKey {
			foundOwner = true
			break
		}
	}
	s.True(foundOwner, "owner creation should become eligible once peer exists")
}

func (s *ActionSelectorSuite) TestCreationAllowedWhenObjectParamClassOutOfScope() {
	// Catalog has only Owner; Peer is registered as OOS extent — Owner _new always eligible.
	ownerClass, ownerKey, peerClass, peerKey := ownerWithObjectParamPeer()
	full := testModel(classEntry(ownerClass, ownerKey), classEntry(peerClass, peerKey))
	active := testModel(classEntry(ownerClass, ownerKey))
	catalog := NewClassCatalog(active)
	catalog.RegisterOutOfScopeMetadata(full)
	selector := NewActionSelector(catalog, nil, nil, nil, rand.New(rand.NewSource(1))) //nolint:gosec

	simState := state.NewSimulationState()
	action, err := selector.SelectAction(simState)
	s.Require().NoError(err)
	s.True(action.IsCreation)
	s.Equal(ownerKey, action.Class.ClassKey)
}

// ownerWithObjectParamPeer builds Owner with _new → Initialize(Peer object of peer).
func ownerWithObjectParamPeer() (ownerClass model_class.Class, ownerKey identity.Key, peerClass model_class.Class, peerKey identity.Key) {
	peerClass, peerKey = simpleCreateClass("peer", "Peer")
	ownerKey = mustKey("domain/d/subdomain/s/class/owner")
	stateActive := mustKey("domain/d/subdomain/s/class/owner/state/active")
	eventNew := mustKey("domain/d/subdomain/s/class/owner/event/_new")
	actionInit := mustKey("domain/d/subdomain/s/class/owner/action/initialize")
	transNew := mustKey("domain/d/subdomain/s/class/owner/transition/create")

	param, err := model_state.NewParameter(actionInit, "Peer", "object of peer", false)
	if err != nil {
		panic(err)
	}
	objKey := "peer"
	param.DataType = &model_data_type.DataType{
		CollectionType: model_data_type.COLLECTION_TYPE_ATOMIC,
		Atomic: &model_data_type.Atomic{
			ConstraintType: model_data_type.CONSTRAINT_TYPE_OBJECT,
			ObjectClassKey: &objKey,
		},
	}
	action := model_state.NewAction(
		actionInit,
		model_state.ActionDetails{Name: "Initialize", Details: ""},
		nil, nil, nil,
		[]model_state.Parameter{param},
	)
	event := model_state.NewEvent(eventNew, "_new", "", []string{"Peer"})
	ownerClass = model_class.NewClass(ownerKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Owner", Details: "", UnfinishedNotes: "", UmlComment: ""})
	ownerClass.SetStates(map[identity.Key]model_state.State{
		stateActive: model_state.NewState(stateActive, "Active", "", ""),
	})
	ownerClass.SetEvents(map[identity.Key]model_state.Event{eventNew: event})
	ownerClass.SetActions(map[identity.Key]model_state.Action{actionInit: action})
	ownerClass.SetTransitions(map[identity.Key]model_state.Transition{
		transNew: model_state.NewTransition(
			transNew, eventNew,
			model_state.TransitionStateKeys{FromStateKey: nil, ToStateKey: &stateActive},
			model_state.TransitionLogicKeys{GuardKey: nil, ActionKey: &actionInit},
			"",
		),
	})
	return ownerClass, ownerKey, peerClass, peerKey
}

func (s *ActionSelectorSuite) TestDoActionsEligibleOnExistingInstances() {
	classKey := mustKey("domain/d/subdomain/s/class/counter")
	stateActiveKey := mustKey("domain/d/subdomain/s/class/counter/state/active")
	eventCreateKey := mustKey("domain/d/subdomain/s/class/counter/event/create")
	actionDoKey := mustKey("domain/d/subdomain/s/class/counter/action/do_count")
	stateActionKey := mustKey("domain/d/subdomain/s/class/counter/state/active/saction/do/do_count")
	transCreateKey := mustKey("domain/d/subdomain/s/class/counter/transition/create")

	eventCreate := model_state.NewEvent(eventCreateKey, "create", "", nil)

	guaranteeKey := helper.Must(identity.NewActionGuaranteeKey(actionDoKey, "0"))
	guaranteeLogic := model_logic.NewLogic(guaranteeKey, model_logic.LogicTypeStateChange, "Postcondition.", "count", counterSpec(), nil)
	actionDo := model_state.NewAction(actionDoKey, model_state.ActionDetails{Name: "DoCount", Details: ""}, nil, []model_logic.Logic{guaranteeLogic}, nil, nil)

	stateActionDo := model_state.NewStateAction(stateActionKey, actionDoKey, "do")

	stateActive := model_state.NewState(stateActiveKey, "Active", "", "")
	stateActive.SetActions([]model_state.StateAction{stateActionDo})

	transCreate := model_state.NewTransition(transCreateKey, eventCreateKey, model_state.TransitionStateKeys{FromStateKey: nil, ToStateKey: &stateActiveKey}, model_state.TransitionLogicKeys{GuardKey: nil, ActionKey: nil}, "")

	class := model_class.NewClass(classKey, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: "Counter", Details: "", UnfinishedNotes: "", UmlComment: ""})
	class.SetAttributes(nil)
	class.SetStates(map[identity.Key]model_state.State{
		stateActiveKey: stateActive,
	})
	class.SetEvents(map[identity.Key]model_state.Event{
		eventCreateKey: eventCreate,
	})
	class.SetGuards(map[identity.Key]model_state.Guard{})
	class.SetActions(map[identity.Key]model_state.Action{
		actionDoKey: actionDo,
	})
	class.SetQueries(map[identity.Key]model_state.Query{})
	class.SetTransitions(map[identity.Key]model_state.Transition{
		transCreateKey: transCreate,
	})

	model := testModel(classEntry(class, classKey))
	catalog := NewClassCatalog(model)
	rng := rand.New(rand.NewSource(42)) //nolint:gosec // deterministic seed for reproducible tests
	selector := NewActionSelector(catalog, nil, nil, nil, rng)

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("_state", object.NewString("Active"))
	attrs.Set("count", object.NewInteger(0))
	simState.CreateInstance(classKey, attrs)

	foundDo := false
	for range 50 {
		action, err := selector.SelectAction(simState)
		s.Require().NoError(err)
		if action.IsDo {
			foundDo = true
			s.Equal("DoCount", action.DoAction.Name)
			s.NotNil(action.Instance)
			break
		}
	}
	s.True(foundDo, "do actions are surface-level on existing instances")
}

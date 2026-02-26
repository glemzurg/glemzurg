package engine

import (
	"math/rand"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/actions"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
	"github.com/stretchr/testify/suite"
)

type CreationChainSuite struct {
	suite.Suite
}

func TestCreationChainSuite(t *testing.T) {
	suite.Run(t, new(CreationChainSuite))
}

type testChainModel struct {
	model    *req_model.Model
	orderKey identity.Key
	itemKey  identity.Key
	assocKey identity.Key
}

// buildChainTestComponents builds all components needed for creation chain tests.
func buildChainTestComponents(
	tcm *testChainModel,
) (*CreationChainHandler, *state.SimulationState, *actions.ActionExecutor) {
	simState := state.NewSimulationState()
	bb := state.NewBindingsBuilder(simState)
	ge := actions.NewGuardEvaluator(bb)
	rng := rand.New(rand.NewSource(42))
	ae := actions.NewActionExecutor(bb, nil, nil, nil, ge, rng)
	pb := actions.NewParameterBinder()
	sae := NewStateActionExecutor(ae)

	catalog := NewClassCatalog(tcm.model)
	handler := NewCreationChainHandler(catalog, ae, sae, pb, rng)

	return handler, simState, ae
}

// buildOrderItemModel creates a model with Order and Item classes,
// where Order has a mandatory outbound association to Item.
func buildOrderItemModel(mandatory bool) *testChainModel {
	orderClass, orderKey := testOrderClass()
	itemClass, itemKey := testItemClass()

	assocKey := testAssocKey(orderKey, itemKey, "OrderItem")

	var toMultStr string
	if mandatory {
		toMultStr = "1..many"
	} else {
		toMultStr = "any"
	}
	fromMult := helper.Must(model_class.NewMultiplicity("1"))
	toMult := helper.Must(model_class.NewMultiplicity(toMultStr))
	assoc := helper.Must(model_class.NewAssociation(assocKey, "OrderItem", "", orderKey, fromMult, itemKey, toMult, nil, ""))

	m := testModel(classEntry(orderClass, orderKey), classEntry(itemClass, itemKey))
	m.ClassAssociations = map[identity.Key]model_class.Association{
		assocKey: assoc,
	}

	return &testChainModel{
		model:    m,
		orderKey: orderKey,
		itemKey:  itemKey,
		assocKey: assocKey,
	}
}

func (s *CreationChainSuite) TestNoMandatoryAssociationsReturnsEmpty() {
	tcm := buildOrderItemModel(false) // not mandatory
	handler, simState, ae := buildChainTestComponents(tcm)

	// Create an Order instance first.
	orderClass, _ := testOrderClass()
	eventCreateKey := mustKey("domain/d/subdomain/s/class/order/event/create")
	event := orderClass.Events[eventCreateKey]

	result, err := ae.ExecuteTransition(orderClass, event, nil, nil, nil, nil)
	s.Require().NoError(err)

	// Handle creation chain — nothing should cascade.
	steps, violations, err := handler.HandleCreationChain(result.InstanceID, simState, 0)
	s.NoError(err)
	s.Empty(steps)
	s.Empty(violations)
}

func (s *CreationChainSuite) TestMandatoryAssociationCreatesLinkedInstance() {
	tcm := buildOrderItemModel(true) // mandatory
	handler, simState, ae := buildChainTestComponents(tcm)

	// Create an Order instance.
	orderClass, _ := testOrderClass()
	eventCreateKey := mustKey("domain/d/subdomain/s/class/order/event/create")
	event := orderClass.Events[eventCreateKey]

	result, err := ae.ExecuteTransition(orderClass, event, nil, nil, nil, nil)
	s.Require().NoError(err)
	s.True(result.WasCreation)

	// Handle creation chain — should cascade and create an Item.
	steps, _, err := handler.HandleCreationChain(result.InstanceID, simState, 0)
	s.NoError(err)
	s.Len(steps, 1)
	s.Equal("Item", steps[0].ClassName)
	s.Equal(StepKindCreation, steps[0].Kind)

	// Should now have 2 instances total: 1 Order + 1 Item.
	s.Equal(2, simState.InstanceCount())

	// The Item should be linked to the Order.
	links := simState.GetLinkedForward(result.InstanceID, tcm.assocKey)
	s.Len(links, 1)
}

func (s *CreationChainSuite) TestCascadeDepthLimitReturnsError() {
	tcm := buildOrderItemModel(true)
	handler, simState, ae := buildChainTestComponents(tcm)

	// Create an Order instance.
	orderClass, _ := testOrderClass()
	eventCreateKey := mustKey("domain/d/subdomain/s/class/order/event/create")
	event := orderClass.Events[eventCreateKey]

	result, err := ae.ExecuteTransition(orderClass, event, nil, nil, nil, nil)
	s.Require().NoError(err)

	// Simulate exceeding depth limit.
	_, _, err = handler.HandleCreationChain(result.InstanceID, simState, maxCascadeDepth+1)
	s.Error(err)
	s.Contains(err.Error(), "max depth")
}

func (s *CreationChainSuite) TestMissingCreationTransitionReturnsError() {
	// Build a model where Item has no creation transition.
	orderClass, orderKey := testOrderClass()

	// Item with state but NO creation transition.
	itemKey := mustKey("domain/d/subdomain/s/class/item")
	stateActiveKey := mustKey("domain/d/subdomain/s/class/item/state/active")
	eventUpdateKey := mustKey("domain/d/subdomain/s/class/item/event/update")
	transUpdateKey := mustKey("domain/d/subdomain/s/class/item/transition/update")

	eventUpdate := helper.Must(model_state.NewEvent(eventUpdateKey, "update", "", nil))
	stateActive := helper.Must(model_state.NewState(stateActiveKey, "Active", "", ""))
	transUpdate := helper.Must(model_state.NewTransition(transUpdateKey, &stateActiveKey, eventUpdateKey, nil, nil, &stateActiveKey, ""))

	itemClass := helper.Must(model_class.NewClass(itemKey, "Item", "", nil, nil, nil, ""))
	itemClass.SetAttributes(map[identity.Key]model_class.Attribute{})
	itemClass.SetStates(map[identity.Key]model_state.State{
		stateActiveKey: stateActive,
	})
	itemClass.SetEvents(map[identity.Key]model_state.Event{
		eventUpdateKey: eventUpdate,
	})
	itemClass.SetGuards(map[identity.Key]model_state.Guard{})
	itemClass.SetActions(map[identity.Key]model_state.Action{})
	itemClass.SetQueries(map[identity.Key]model_state.Query{})
	itemClass.SetTransitions(map[identity.Key]model_state.Transition{
		transUpdateKey: transUpdate,
	})

	assocKey := testAssocKey(orderKey, itemKey, "OrderItem")
	fromMult := helper.Must(model_class.NewMultiplicity("1"))
	toMult := helper.Must(model_class.NewMultiplicity("1..many"))
	assoc := helper.Must(model_class.NewAssociation(assocKey, "OrderItem", "", orderKey, fromMult, itemKey, toMult, nil, ""))

	model := testModel(classEntry(orderClass, orderKey), classEntry(itemClass, itemKey))
	model.ClassAssociations = map[identity.Key]model_class.Association{
		assocKey: assoc,
	}

	simState := state.NewSimulationState()
	bb := state.NewBindingsBuilder(simState)
	ge := actions.NewGuardEvaluator(bb)
	rng := rand.New(rand.NewSource(42))
	ae := actions.NewActionExecutor(bb, nil, nil, nil, ge, rng)
	pb := actions.NewParameterBinder()
	sae := NewStateActionExecutor(ae)
	catalog := NewClassCatalog(model)
	handler := NewCreationChainHandler(catalog, ae, sae, pb, rng)

	// Create an Order.
	eventCreateKey := mustKey("domain/d/subdomain/s/class/order/event/create")
	event := orderClass.Events[eventCreateKey]
	result, err := ae.ExecuteTransition(orderClass, event, nil, nil, nil, nil)
	s.Require().NoError(err)

	// Handle chain — should fail because Item has no creation transition.
	_, _, err = handler.HandleCreationChain(result.InstanceID, simState, 0)
	s.Error(err)
	s.Contains(err.Error(), "no creation transition")
}

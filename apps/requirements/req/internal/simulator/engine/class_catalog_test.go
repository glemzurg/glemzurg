package engine

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"
	"github.com/stretchr/testify/suite"
)

type ClassCatalogSuite struct {
	suite.Suite
}

func TestClassCatalogSuite(t *testing.T) {
	suite.Run(t, new(ClassCatalogSuite))
}

// ========================================================================
// Tests
// ========================================================================

func (s *ClassCatalogSuite) TestCatalogFromModelWithOneClass() {
	orderClass, orderKey := testOrderClass()
	model := testModel(classEntry(orderClass, orderKey))

	catalog := NewClassCatalog(model)

	info := catalog.GetClassInfo(orderKey)
	s.NotNil(info)
	s.Equal("Order", info.Class.Name)
	s.True(info.HasStates)

	// Should have one creation event (create)
	s.Len(info.CreationEvents, 1)
	s.Equal("create", info.CreationEvents[0].Name)
}

func (s *ClassCatalogSuite) TestCatalogWithMultipleClasses() {
	orderClass, orderKey := testOrderClass()
	itemClass, itemKey := testItemClass()
	model := testModel(classEntry(orderClass, orderKey), classEntry(itemClass, itemKey))

	catalog := NewClassCatalog(model)

	all := catalog.AllSimulatableClasses()
	s.Len(all, 2)

	s.NotNil(catalog.GetClassInfo(orderKey))
	s.NotNil(catalog.GetClassInfo(itemKey))
}

func (s *ClassCatalogSuite) TestClassWithNoStatesExcluded() {
	classKey := mustKey("domain/d/subdomain/s/class/simple")

	simpleClass := helper.Must(model_class.NewClass(classKey, "Simple", "", nil, nil, nil, ""))
	simpleClass.SetAttributes(map[identity.Key]model_class.Attribute{})
	simpleClass.SetStates(map[identity.Key]model_state.State{})
	simpleClass.SetEvents(map[identity.Key]model_state.Event{})
	simpleClass.SetGuards(map[identity.Key]model_state.Guard{})
	simpleClass.SetActions(map[identity.Key]model_state.Action{})
	simpleClass.SetQueries(map[identity.Key]model_state.Query{})
	simpleClass.SetTransitions(map[identity.Key]model_state.Transition{})

	model := testModel(classEntry(simpleClass, classKey))

	catalog := NewClassCatalog(model)

	s.Nil(catalog.GetClassInfo(classKey))
	s.Empty(catalog.AllSimulatableClasses())
}

func (s *ClassCatalogSuite) TestStateEventsIndexedCorrectly() {
	orderClass, orderKey := testOrderClass()
	model := testModel(classEntry(orderClass, orderKey))

	catalog := NewClassCatalog(model)
	info := catalog.GetClassInfo(orderKey)

	// From "Open" state, "close" event should be eligible.
	openEvents := info.StateEvents["Open"]
	s.Len(openEvents, 1)
	s.Equal("close", openEvents[0].Event.Name)
	s.Len(openEvents[0].Transitions, 1)

	// From "Closed" state, no events should be eligible.
	closedEvents := info.StateEvents["Closed"]
	s.Empty(closedEvents)
}

func (s *ClassCatalogSuite) TestMandatoryAssociationsDetected() {
	orderClass, orderKey := testOrderClass()
	itemClass, itemKey := testItemClass()

	assocKey := testAssocKey(orderKey, itemKey, "OrderItem")
	fromMult := helper.Must(model_class.NewMultiplicity("1"))
	toMult := helper.Must(model_class.NewMultiplicity("1..many"))
	assoc := helper.Must(model_class.NewAssociation(assocKey, "OrderItem", "", orderKey, fromMult, itemKey, toMult, nil, ""))

	model := testModel(classEntry(orderClass, orderKey), classEntry(itemClass, itemKey))
	model.ClassAssociations = map[identity.Key]model_class.Association{
		assocKey: assoc,
	}

	catalog := NewClassCatalog(model)

	mandatory := catalog.GetMandatoryOutboundAssociations(orderKey)
	s.Len(mandatory, 1)
	s.Equal(itemKey, mandatory[0].ToClassKey)
	s.True(mandatory[0].MandatoryTo)
	s.Equal(uint(1), mandatory[0].MinTo)
}

func (s *ClassCatalogSuite) TestDoActionsRecorded() {
	classKey := mustKey("domain/d/subdomain/s/class/counter")
	stateActiveKey := mustKey("domain/d/subdomain/s/class/counter/state/active")
	eventCreateKey := mustKey("domain/d/subdomain/s/class/counter/event/create")
	actionDoKey := mustKey("domain/d/subdomain/s/class/counter/action/do_count")
	stateActionKey := mustKey("domain/d/subdomain/s/class/counter/state/active/saction/do/do_count")
	transCreateKey := mustKey("domain/d/subdomain/s/class/counter/transition/create")

	eventCreate := helper.Must(model_state.NewEvent(eventCreateKey, "create", "", nil))

	guaranteeKey := helper.Must(identity.NewActionGuaranteeKey(actionDoKey, "0"))
	guaranteeLogic := helper.Must(model_logic.NewLogic(guaranteeKey, model_logic.LogicTypeStateChange, "Postcondition.", "count", model_logic.NotationTLAPlus, "self.count + 1"))
	actionDo := helper.Must(model_state.NewAction(actionDoKey, "DoCount", "", nil, []model_logic.Logic{guaranteeLogic}, nil, nil))

	stateActionDo := helper.Must(model_state.NewStateAction(stateActionKey, actionDoKey, "do"))

	stateActive := helper.Must(model_state.NewState(stateActiveKey, "Active", "", ""))
	stateActive.SetActions([]model_state.StateAction{stateActionDo})

	transCreate := helper.Must(model_state.NewTransition(transCreateKey, nil, eventCreateKey, nil, nil, &stateActiveKey, ""))

	class := helper.Must(model_class.NewClass(classKey, "Counter", "", nil, nil, nil, ""))
	class.SetAttributes(map[identity.Key]model_class.Attribute{})
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

	info := catalog.GetClassInfo(classKey)
	s.NotNil(info)

	doActions := info.DoActions["Active"]
	s.Len(doActions, 1)
	s.Equal("DoCount", doActions[0].Name)
}

func (s *ClassCatalogSuite) TestExternalCreationEventsNoAssociation() {
	orderClass, orderKey := testOrderClass()
	model := testModel(classEntry(orderClass, orderKey))

	catalog := NewClassCatalog(model)

	// Without associations, creation events are external.
	ext := catalog.ExternalCreationEvents(orderKey)
	s.Len(ext, 1)
	s.Equal("create", ext[0].Name)
}

func (s *ClassCatalogSuite) TestExternalCreationEventsWithMandatoryAssociation() {
	orderClass, orderKey := testOrderClass()
	itemClass, itemKey := testItemClass()

	assocKey := testAssocKey(orderKey, itemKey, "OrderItem")
	fromMult := helper.Must(model_class.NewMultiplicity("1"))
	toMult := helper.Must(model_class.NewMultiplicity("1..many"))
	assoc := helper.Must(model_class.NewAssociation(assocKey, "OrderItem", "", orderKey, fromMult, itemKey, toMult, nil, ""))

	model := testModel(classEntry(orderClass, orderKey), classEntry(itemClass, itemKey))
	model.ClassAssociations = map[identity.Key]model_class.Association{
		assocKey: assoc,
	}

	catalog := NewClassCatalog(model)

	// Item creation is driven by Order → Item is NOT external.
	ext := catalog.ExternalCreationEvents(itemKey)
	s.Empty(ext)

	// Order creation is still external (nothing creates Order).
	ext = catalog.ExternalCreationEvents(orderKey)
	s.Len(ext, 1)
}

func (s *ClassCatalogSuite) TestGetCreationEvent() {
	orderClass, orderKey := testOrderClass()
	model := testModel(classEntry(orderClass, orderKey))

	catalog := NewClassCatalog(model)

	ev, found := catalog.GetCreationEvent(orderKey)
	s.True(found)
	s.Equal("create", ev.Name)

	// Non-existent class
	_, found = catalog.GetCreationEvent(mustKey("domain/d/subdomain/s/class/nope"))
	s.False(found)
}

// ========================================================================
// SentBy / CalledBy filtering tests
// ========================================================================

func (s *ClassCatalogSuite) TestExternalStateEvents_NoSentBy() {
	// With no SentBy data, all events are external.
	orderClass, orderKey := testOrderClass()
	model := testModel(classEntry(orderClass, orderKey))

	catalog := NewClassCatalog(model)

	ext := catalog.ExternalStateEvents(orderKey, "Open")
	s.Len(ext, 1)
	s.Equal("close", ext[0].Event.Name)
}

func (s *ClassCatalogSuite) TestExternalStateEvents_SentByInScope() {
	// Event with SentBy pointing to an in-scope class → internal (filtered out).
	orderClass, orderKey := testOrderClass()
	itemClass, itemKey := testItemClass()
	model := testModel(classEntry(orderClass, orderKey), classEntry(itemClass, itemKey))

	catalog := NewClassCatalog(model)

	// Mark the "close" event as sent by Item (in-scope).
	closeEventKey := mustKey("domain/d/subdomain/s/class/order/event/close")
	catalog.SetEventSentBy(closeEventKey, []identity.Key{itemKey})

	ext := catalog.ExternalStateEvents(orderKey, "Open")
	s.Empty(ext, "close event should be internal because sender Item is in scope")
}

func (s *ClassCatalogSuite) TestExternalStateEvents_SentByOutOfScope() {
	// Event with SentBy pointing to an out-of-scope class → still external.
	orderClass, orderKey := testOrderClass()
	model := testModel(classEntry(orderClass, orderKey))

	catalog := NewClassCatalog(model)

	// Mark the "close" event as sent by a class NOT in the catalog.
	closeEventKey := mustKey("domain/d/subdomain/s/class/order/event/close")
	outsideClassKey := mustKey("domain/d/subdomain/s/class/external_system")
	catalog.SetEventSentBy(closeEventKey, []identity.Key{outsideClassKey})

	ext := catalog.ExternalStateEvents(orderKey, "Open")
	s.Len(ext, 1, "close event should be external because sender is not in scope")
	s.Equal("close", ext[0].Event.Name)
}

func (s *ClassCatalogSuite) TestExternalDoActions_NoCalledBy() {
	// With no CalledBy data, all do-actions are external.
	classKey := mustKey("domain/d/subdomain/s/class/counter")
	stateActiveKey := mustKey("domain/d/subdomain/s/class/counter/state/active")
	eventCreateKey := mustKey("domain/d/subdomain/s/class/counter/event/create")
	actionDoKey := mustKey("domain/d/subdomain/s/class/counter/action/do_count")
	stateActionKey := mustKey("domain/d/subdomain/s/class/counter/state/active/saction/do/do_count")
	transCreateKey := mustKey("domain/d/subdomain/s/class/counter/transition/create")

	eventCreate := helper.Must(model_state.NewEvent(eventCreateKey, "create", "", nil))

	guaranteeKey := helper.Must(identity.NewActionGuaranteeKey(actionDoKey, "0"))
	guaranteeLogic := helper.Must(model_logic.NewLogic(guaranteeKey, model_logic.LogicTypeStateChange, "Postcondition.", "count", model_logic.NotationTLAPlus, "self.count + 1"))
	actionDo := helper.Must(model_state.NewAction(actionDoKey, "DoCount", "", nil, []model_logic.Logic{guaranteeLogic}, nil, nil))

	stateActionDo := helper.Must(model_state.NewStateAction(stateActionKey, actionDoKey, "do"))

	stateActive := helper.Must(model_state.NewState(stateActiveKey, "Active", "", ""))
	stateActive.SetActions([]model_state.StateAction{stateActionDo})

	transCreate := helper.Must(model_state.NewTransition(transCreateKey, nil, eventCreateKey, nil, nil, &stateActiveKey, ""))

	class := helper.Must(model_class.NewClass(classKey, "Counter", "", nil, nil, nil, ""))
	class.SetAttributes(map[identity.Key]model_class.Attribute{})
	class.SetStates(map[identity.Key]model_state.State{
		stateActiveKey: stateActive,
	})
	class.SetEvents(map[identity.Key]model_state.Event{eventCreateKey: eventCreate})
	class.SetGuards(map[identity.Key]model_state.Guard{})
	class.SetActions(map[identity.Key]model_state.Action{actionDoKey: actionDo})
	class.SetQueries(map[identity.Key]model_state.Query{})
	class.SetTransitions(map[identity.Key]model_state.Transition{
		transCreateKey: transCreate,
	})

	model := testModel(classEntry(class, classKey))
	catalog := NewClassCatalog(model)

	ext := catalog.ExternalDoActions(classKey, "Active")
	s.Len(ext, 1)
	s.Equal("DoCount", ext[0].Name)
}

func (s *ClassCatalogSuite) TestExternalDoActions_CalledByInScope() {
	// Action with CalledBy pointing to an in-scope class → internal (filtered out).
	classKey := mustKey("domain/d/subdomain/s/class/counter")
	stateActiveKey := mustKey("domain/d/subdomain/s/class/counter/state/active")
	eventCreateKey := mustKey("domain/d/subdomain/s/class/counter/event/create")
	actionDoKey := mustKey("domain/d/subdomain/s/class/counter/action/do_count")
	stateActionKey := mustKey("domain/d/subdomain/s/class/counter/state/active/saction/do/do_count")
	transCreateKey := mustKey("domain/d/subdomain/s/class/counter/transition/create")

	eventCreate := helper.Must(model_state.NewEvent(eventCreateKey, "create", "", nil))

	guaranteeKey := helper.Must(identity.NewActionGuaranteeKey(actionDoKey, "0"))
	guaranteeLogic := helper.Must(model_logic.NewLogic(guaranteeKey, model_logic.LogicTypeStateChange, "Postcondition.", "count", model_logic.NotationTLAPlus, "self.count + 1"))
	actionDo := helper.Must(model_state.NewAction(actionDoKey, "DoCount", "", nil, []model_logic.Logic{guaranteeLogic}, nil, nil))

	stateActionDo := helper.Must(model_state.NewStateAction(stateActionKey, actionDoKey, "do"))

	stateActive := helper.Must(model_state.NewState(stateActiveKey, "Active", "", ""))
	stateActive.SetActions([]model_state.StateAction{stateActionDo})

	transCreate := helper.Must(model_state.NewTransition(transCreateKey, nil, eventCreateKey, nil, nil, &stateActiveKey, ""))

	orderClass, orderKey := testOrderClass()

	class := helper.Must(model_class.NewClass(classKey, "Counter", "", nil, nil, nil, ""))
	class.SetAttributes(map[identity.Key]model_class.Attribute{})
	class.SetStates(map[identity.Key]model_state.State{
		stateActiveKey: stateActive,
	})
	class.SetEvents(map[identity.Key]model_state.Event{eventCreateKey: eventCreate})
	class.SetGuards(map[identity.Key]model_state.Guard{})
	class.SetActions(map[identity.Key]model_state.Action{actionDoKey: actionDo})
	class.SetQueries(map[identity.Key]model_state.Query{})
	class.SetTransitions(map[identity.Key]model_state.Transition{
		transCreateKey: transCreate,
	})

	model := testModel(classEntry(class, classKey), classEntry(orderClass, orderKey))
	catalog := NewClassCatalog(model)

	// Mark the action as called by Order (in-scope).
	catalog.SetActionCalledBy(actionDoKey, []identity.Key{orderKey})

	ext := catalog.ExternalDoActions(classKey, "Active")
	s.Empty(ext, "do action should be internal because caller Order is in scope")
}

func (s *ClassCatalogSuite) TestCallerDataExport() {
	orderClass, orderKey := testOrderClass()
	itemClass, itemKey := testItemClass()
	model := testModel(classEntry(orderClass, orderKey), classEntry(itemClass, itemKey))

	catalog := NewClassCatalog(model)

	eventKey := mustKey("domain/d/subdomain/s/class/order/event/close")
	actionKey := mustKey("domain/d/subdomain/s/class/order/action/do_close")
	queryKey := mustKey("domain/d/subdomain/s/class/order/query/get_total")

	catalog.SetEventSentBy(eventKey, []identity.Key{itemKey})
	catalog.SetActionCalledBy(actionKey, []identity.Key{itemKey})
	catalog.SetQueryCalledBy(queryKey, []identity.Key{orderKey})

	cd := catalog.CallerData()
	s.NotNil(cd)
	s.Equal([]identity.Key{itemKey}, cd.EventSentBy[eventKey])
	s.Equal([]identity.Key{itemKey}, cd.ActionCalledBy[actionKey])
	s.Equal([]identity.Key{orderKey}, cd.QueryCalledBy[queryKey])
}

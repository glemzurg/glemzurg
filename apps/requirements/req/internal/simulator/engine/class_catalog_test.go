package engine

import (
	"testing"

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
	simpleClass := model_class.Class{
		Key:         classKey,
		Name:        "Simple",
		Attributes:  map[identity.Key]model_class.Attribute{},
		States:      map[identity.Key]model_state.State{},
		Events:      map[identity.Key]model_state.Event{},
		Guards:      map[identity.Key]model_state.Guard{},
		Actions:     map[identity.Key]model_state.Action{},
		Queries:     map[identity.Key]model_state.Query{},
		Transitions: map[identity.Key]model_state.Transition{},
	}
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
	model := testModel(classEntry(orderClass, orderKey), classEntry(itemClass, itemKey))
	model.ClassAssociations = map[identity.Key]model_class.Association{
		assocKey: {
			Key:              assocKey,
			Name:             "OrderItem",
			FromClassKey:     orderKey,
			ToClassKey:       itemKey,
			FromMultiplicity: model_class.Multiplicity{LowerBound: 1, HigherBound: 1},
			ToMultiplicity:   model_class.Multiplicity{LowerBound: 1, HigherBound: 0}, // 1..*
		},
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
				Guarantees: []model_logic.Logic{
					{Key: "guar_1", Description: "Postcondition.", Notation: model_logic.NotationTLAPlus, Specification: "self.count' = self.count + 1"},
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
	model := testModel(classEntry(orderClass, orderKey), classEntry(itemClass, itemKey))
	model.ClassAssociations = map[identity.Key]model_class.Association{
		assocKey: {
			Key:              assocKey,
			Name:             "OrderItem",
			FromClassKey:     orderKey,
			ToClassKey:       itemKey,
			FromMultiplicity: model_class.Multiplicity{LowerBound: 1, HigherBound: 1},
			ToMultiplicity:   model_class.Multiplicity{LowerBound: 1, HigherBound: 0}, // 1..*
		},
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

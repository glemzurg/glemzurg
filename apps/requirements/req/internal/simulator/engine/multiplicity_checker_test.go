package engine

import (
	"testing"

	"github.com/glemzurg/go-tlaplus/internal/identity"
	"github.com/glemzurg/go-tlaplus/internal/req_model/model_class"
	"github.com/glemzurg/go-tlaplus/internal/simulator/object"
	"github.com/glemzurg/go-tlaplus/internal/simulator/state"
	"github.com/stretchr/testify/suite"
)

type MultiplicityCheckerSuite struct {
	suite.Suite
}

func TestMultiplicityCheckerSuite(t *testing.T) {
	suite.Run(t, new(MultiplicityCheckerSuite))
}

func (s *MultiplicityCheckerSuite) TestValidMultiplicities() {
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
			ToMultiplicity:   model_class.Multiplicity{LowerBound: 1, HigherBound: 3},
		},
	}

	catalog := NewClassCatalog(model)
	checker := NewMultiplicityChecker(catalog)

	simState := state.NewSimulationState()
	order := simState.CreateInstance(orderKey, object.NewRecord())
	item := simState.CreateInstance(itemKey, object.NewRecord())
	simState.AddLink(assocKey, order.ID, item.ID)

	// Order has 1 forward link (within 1..3) → valid.
	// Item has 1 reverse link (within 1..1) → valid.
	violations := checker.CheckState(simState)
	s.Empty(violations)
}

func (s *MultiplicityCheckerSuite) TestLowerBoundViolation() {
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
			ToMultiplicity:   model_class.Multiplicity{LowerBound: 2, HigherBound: 0}, // At least 2 items
		},
	}

	catalog := NewClassCatalog(model)
	checker := NewMultiplicityChecker(catalog)

	simState := state.NewSimulationState()
	order := simState.CreateInstance(orderKey, object.NewRecord())
	item := simState.CreateInstance(itemKey, object.NewRecord())
	simState.AddLink(assocKey, order.ID, item.ID)

	// Order has 1 forward link but needs at least 2 → violation.
	violations := checker.CheckInstance(order, simState)
	s.Len(violations, 1)
	s.Contains(violations[0].Message, "at least 2")
	s.Equal("forward", violations[0].Direction)
}

func (s *MultiplicityCheckerSuite) TestUpperBoundViolation() {
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
			FromMultiplicity: model_class.Multiplicity{LowerBound: 0, HigherBound: 0}, // any
			ToMultiplicity:   model_class.Multiplicity{LowerBound: 0, HigherBound: 1}, // at most 1
		},
	}

	catalog := NewClassCatalog(model)
	checker := NewMultiplicityChecker(catalog)

	simState := state.NewSimulationState()
	order := simState.CreateInstance(orderKey, object.NewRecord())
	item1 := simState.CreateInstance(itemKey, object.NewRecord())
	item2 := simState.CreateInstance(itemKey, object.NewRecord())
	simState.AddLink(assocKey, order.ID, item1.ID)
	simState.AddLink(assocKey, order.ID, item2.ID)

	// Order has 2 forward links but max is 1 → violation.
	violations := checker.CheckInstance(order, simState)
	s.Len(violations, 1)
	s.Contains(violations[0].Message, "at most 1")
}

func (s *MultiplicityCheckerSuite) TestOptionalAssociationNeverViolated() {
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
			FromMultiplicity: model_class.Multiplicity{LowerBound: 0, HigherBound: 0}, // any
			ToMultiplicity:   model_class.Multiplicity{LowerBound: 0, HigherBound: 0}, // any
		},
	}

	catalog := NewClassCatalog(model)
	checker := NewMultiplicityChecker(catalog)

	simState := state.NewSimulationState()
	simState.CreateInstance(orderKey, object.NewRecord())
	simState.CreateInstance(itemKey, object.NewRecord())

	// No links, but both multiplicities are 0..* → no violation.
	violations := checker.CheckState(simState)
	s.Empty(violations)
}

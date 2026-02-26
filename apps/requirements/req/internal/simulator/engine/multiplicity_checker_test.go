package engine

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
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
	fromMult := helper.Must(model_class.NewMultiplicity("1"))
	toMult := helper.Must(model_class.NewMultiplicity("1..3"))
	assoc := helper.Must(model_class.NewAssociation(assocKey, "OrderItem", "", orderKey, fromMult, itemKey, toMult, nil, ""))

	model := testModel(classEntry(orderClass, orderKey), classEntry(itemClass, itemKey))
	model.ClassAssociations = map[identity.Key]model_class.Association{
		assocKey: assoc,
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
	fromMult := helper.Must(model_class.NewMultiplicity("1"))
	toMult := helper.Must(model_class.NewMultiplicity("2..many"))
	assoc := helper.Must(model_class.NewAssociation(assocKey, "OrderItem", "", orderKey, fromMult, itemKey, toMult, nil, ""))

	model := testModel(classEntry(orderClass, orderKey), classEntry(itemClass, itemKey))
	model.ClassAssociations = map[identity.Key]model_class.Association{
		assocKey: assoc,
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
	fromMult := helper.Must(model_class.NewMultiplicity("any"))
	toMult := helper.Must(model_class.NewMultiplicity("0..1"))
	assoc := helper.Must(model_class.NewAssociation(assocKey, "OrderItem", "", orderKey, fromMult, itemKey, toMult, nil, ""))

	model := testModel(classEntry(orderClass, orderKey), classEntry(itemClass, itemKey))
	model.ClassAssociations = map[identity.Key]model_class.Association{
		assocKey: assoc,
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
	fromMult := helper.Must(model_class.NewMultiplicity("any"))
	toMult := helper.Must(model_class.NewMultiplicity("any"))
	assoc := helper.Must(model_class.NewAssociation(assocKey, "OrderItem", "", orderKey, fromMult, itemKey, toMult, nil, ""))

	model := testModel(classEntry(orderClass, orderKey), classEntry(itemClass, itemKey))
	model.ClassAssociations = map[identity.Key]model_class.Association{
		assocKey: assoc,
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

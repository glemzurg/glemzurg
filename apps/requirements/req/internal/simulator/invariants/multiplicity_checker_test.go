package invariants

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
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
	orderClass, orderKey := multiplicityTestOrderClass()
	itemClass, itemKey := multiplicityTestItemClass()

	assocKey := multiplicityTestAssocKey(orderKey, itemKey)
	fromMult := helper.Must(model_class.NewMultiplicity("1"))
	toMult := helper.Must(model_class.NewMultiplicity("1..3"))
	assoc := model_class.NewAssociation(assocKey, model_class.AssociationDetails{Name: "OrderItem", Details: ""}, model_class.AssociationEnd{ClassKey: orderKey, Multiplicity: fromMult}, model_class.AssociationEnd{ClassKey: itemKey, Multiplicity: toMult}, model_class.AssociationOptions{AssociationClassKey: nil, UmlComment: ""})

	model := multiplicityTestModel(classEntry(orderClass, orderKey), classEntry(itemClass, itemKey))
	model.ClassAssociations = map[identity.Key]model_class.Association{
		assocKey: assoc,
	}

	checker := NewMultiplicityChecker(model)

	simState := state.NewSimulationState()
	order := simState.CreateInstance(orderKey, object.NewRecord())
	item := simState.CreateInstance(itemKey, object.NewRecord())
	s.Require().NoError(simState.AddLink(assocKey, order.ID, item.ID))

	violations := checker.CheckState(simState)
	s.Empty(violations)
}

func (s *MultiplicityCheckerSuite) TestLowerBoundViolation() {
	orderClass, orderKey := multiplicityTestOrderClass()
	itemClass, itemKey := multiplicityTestItemClass()

	assocKey := multiplicityTestAssocKey(orderKey, itemKey)
	fromMult := helper.Must(model_class.NewMultiplicity("1"))
	toMult := helper.Must(model_class.NewMultiplicity("2..many"))
	assoc := model_class.NewAssociation(assocKey, model_class.AssociationDetails{Name: "OrderItem", Details: ""}, model_class.AssociationEnd{ClassKey: orderKey, Multiplicity: fromMult}, model_class.AssociationEnd{ClassKey: itemKey, Multiplicity: toMult}, model_class.AssociationOptions{AssociationClassKey: nil, UmlComment: ""})

	model := multiplicityTestModel(classEntry(orderClass, orderKey), classEntry(itemClass, itemKey))
	model.ClassAssociations = map[identity.Key]model_class.Association{
		assocKey: assoc,
	}

	checker := NewMultiplicityChecker(model)

	simState := state.NewSimulationState()
	order := simState.CreateInstance(orderKey, object.NewRecord())
	item := simState.CreateInstance(itemKey, object.NewRecord())
	s.Require().NoError(simState.AddLink(assocKey, order.ID, item.ID))

	violations := checker.CheckInstance(order, simState)
	s.Len(violations, 1)
	s.Contains(violations[0].Message, "at least 2")
}

func (s *MultiplicityCheckerSuite) TestUpperBoundViolation() {
	orderClass, orderKey := multiplicityTestOrderClass()
	itemClass, itemKey := multiplicityTestItemClass()

	assocKey := multiplicityTestAssocKey(orderKey, itemKey)
	fromMult := helper.Must(model_class.NewMultiplicity("any"))
	toMult := helper.Must(model_class.NewMultiplicity("0..1"))
	assoc := model_class.NewAssociation(assocKey, model_class.AssociationDetails{Name: "OrderItem", Details: ""}, model_class.AssociationEnd{ClassKey: orderKey, Multiplicity: fromMult}, model_class.AssociationEnd{ClassKey: itemKey, Multiplicity: toMult}, model_class.AssociationOptions{AssociationClassKey: nil, UmlComment: ""})

	model := multiplicityTestModel(classEntry(orderClass, orderKey), classEntry(itemClass, itemKey))
	model.ClassAssociations = map[identity.Key]model_class.Association{
		assocKey: assoc,
	}

	checker := NewMultiplicityChecker(model)

	simState := state.NewSimulationState()
	order := simState.CreateInstance(orderKey, object.NewRecord())
	item1 := simState.CreateInstance(itemKey, object.NewRecord())
	item2 := simState.CreateInstance(itemKey, object.NewRecord())
	s.Require().NoError(simState.AddLink(assocKey, order.ID, item1.ID))
	s.Require().NoError(simState.AddLink(assocKey, order.ID, item2.ID))

	violations := checker.CheckInstance(order, simState)
	s.Len(violations, 1)
	s.Contains(violations[0].Message, "at most 1")
}

func (s *MultiplicityCheckerSuite) TestOptionalAssociationNeverViolated() {
	orderClass, orderKey := multiplicityTestOrderClass()
	itemClass, itemKey := multiplicityTestItemClass()

	assocKey := multiplicityTestAssocKey(orderKey, itemKey)
	fromMult := helper.Must(model_class.NewMultiplicity("any"))
	toMult := helper.Must(model_class.NewMultiplicity("any"))
	assoc := model_class.NewAssociation(assocKey, model_class.AssociationDetails{Name: "OrderItem", Details: ""}, model_class.AssociationEnd{ClassKey: orderKey, Multiplicity: fromMult}, model_class.AssociationEnd{ClassKey: itemKey, Multiplicity: toMult}, model_class.AssociationOptions{AssociationClassKey: nil, UmlComment: ""})

	model := multiplicityTestModel(classEntry(orderClass, orderKey), classEntry(itemClass, itemKey))
	model.ClassAssociations = map[identity.Key]model_class.Association{
		assocKey: assoc,
	}

	checker := NewMultiplicityChecker(model)

	simState := state.NewSimulationState()
	simState.CreateInstance(orderKey, object.NewRecord())
	simState.CreateInstance(itemKey, object.NewRecord())

	violations := checker.CheckState(simState)
	s.Empty(violations)
}

func multiplicityTestOrderClass() (model_class.Class, identity.Key) {
	classKey := multiplicityMustKey("domain/d/subdomain/s/class/order")
	stateOpenKey := multiplicityMustKey("domain/d/subdomain/s/class/order/state/open")
	eventCreateKey := multiplicityMustKey("domain/d/subdomain/s/class/order/event/create")
	transCreateKey := multiplicityMustKey("domain/d/subdomain/s/class/order/transition/create")

	eventCreate := model_state.NewEvent(eventCreateKey, "create", "", nil)
	stateOpen := model_state.NewState(stateOpenKey, "Open", "", "")
	transCreate := model_state.NewTransition(transCreateKey, eventCreateKey, model_state.TransitionStateKeys{FromStateKey: nil, ToStateKey: &stateOpenKey}, model_state.TransitionLogicKeys{GuardKey: nil, ActionKey: nil}, "")

	class := model_class.NewClass(classKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Order"})
	class.SetStates(map[identity.Key]model_state.State{stateOpenKey: stateOpen})
	class.SetEvents(map[identity.Key]model_state.Event{eventCreateKey: eventCreate})
	class.SetTransitions(map[identity.Key]model_state.Transition{transCreateKey: transCreate})
	return class, classKey
}

func multiplicityTestItemClass() (model_class.Class, identity.Key) {
	classKey := multiplicityMustKey("domain/d/subdomain/s/class/item")
	stateActiveKey := multiplicityMustKey("domain/d/subdomain/s/class/item/state/active")
	eventCreateKey := multiplicityMustKey("domain/d/subdomain/s/class/item/event/create")
	transCreateKey := multiplicityMustKey("domain/d/subdomain/s/class/item/transition/create")

	eventCreate := model_state.NewEvent(eventCreateKey, "create", "", nil)
	stateActive := model_state.NewState(stateActiveKey, "Active", "", "")
	transCreate := model_state.NewTransition(transCreateKey, eventCreateKey, model_state.TransitionStateKeys{FromStateKey: nil, ToStateKey: &stateActiveKey}, model_state.TransitionLogicKeys{GuardKey: nil, ActionKey: nil}, "")

	class := model_class.NewClass(classKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Item"})
	class.SetStates(map[identity.Key]model_state.State{stateActiveKey: stateActive})
	class.SetEvents(map[identity.Key]model_state.Event{eventCreateKey: eventCreate})
	class.SetTransitions(map[identity.Key]model_state.Transition{transCreateKey: transCreate})
	return class, classKey
}

func multiplicityMustKey(s string) identity.Key {
	k, err := identity.ParseKey(s)
	if err != nil {
		panic(err)
	}
	return k
}

func multiplicityTestModel(classes ...struct {
	class model_class.Class
	key   identity.Key
}) *core.Model {
	subdomainKey := multiplicityMustKey("domain/d/subdomain/s")
	domainKey := multiplicityMustKey("domain/d")

	classMap := make(map[identity.Key]model_class.Class)
	for _, c := range classes {
		classMap[c.key] = c.class
	}

	subdomain := model_domain.NewSubdomain(subdomainKey, "S", "", "", "")
	subdomain.Classes = classMap

	domain := model_domain.NewDomain(domainKey, "D", "", "", false, "")
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{
		subdomainKey: subdomain,
	}

	model := core.NewModel("test", core.ModelDetails{Name: "Test", Details: ""}, "", nil, nil, nil)
	model.Domains = map[identity.Key]model_domain.Domain{
		domainKey: domain,
	}
	return &model
}

func classEntry(class model_class.Class, key identity.Key) struct {
	class model_class.Class
	key   identity.Key
} {
	return struct {
		class model_class.Class
		key   identity.Key
	}{class, key}
}

func multiplicityTestAssocKey(fromClassKey, toClassKey identity.Key) identity.Key {
	parentKey := multiplicityMustKey("domain/d/subdomain/s")
	k, err := identity.NewClassAssociationKey(parentKey, fromClassKey, toClassKey, "OrderItem")
	if err != nil {
		panic(err)
	}
	return k
}

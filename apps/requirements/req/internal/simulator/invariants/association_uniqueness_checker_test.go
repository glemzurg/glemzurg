package invariants

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
	"github.com/stretchr/testify/suite"
)

type AssociationUniquenessCheckerSuite struct {
	suite.Suite
}

func TestAssociationUniquenessCheckerSuite(t *testing.T) {
	suite.Run(t, new(AssociationUniquenessCheckerSuite))
}

func (s *AssociationUniquenessCheckerSuite) buildModel(uniqueness string) (*core.Model, identity.Key, identity.Key, identity.Key, identity.Key) {
	fromClass, fromKey := multiplicityTestOrderClass()
	toClass, toKey := multiplicityTestItemClass()
	acClass, acKey := associationUniquenessTestLinkClass()

	assocKey := multiplicityTestAssocKey(fromKey, toKey)
	fromMult := helper.Must(model_class.NewMultiplicity("any"))
	toMult := helper.Must(model_class.NewMultiplicity("any"))
	uniq := helper.Must(model_class.NewMultiplicity(uniqueness))

	assoc := model_class.NewAssociation(
		assocKey,
		model_class.AssociationDetails{Name: "OrderItem", Details: ""},
		model_class.AssociationEnd{ClassKey: fromKey, Multiplicity: fromMult},
		model_class.AssociationEnd{ClassKey: toKey, Multiplicity: toMult},
		uniq, model_class.AssociationOptions{AssociationClassKey: &acKey, UmlComment: ""},
	)

	model := multiplicityTestModel(
		classEntry(fromClass, fromKey),
		classEntry(toClass, toKey),
		classEntry(acClass, acKey),
	)
	domainKey := multiplicityMustKey("domain/d")
	subdomainKey := multiplicityMustKey("domain/d/subdomain/s")
	domain := model.Domains[domainKey]
	subdomain := domain.Subdomains[subdomainKey]
	subdomain.ClassAssociations = map[identity.Key]model_class.Association{
		assocKey: assoc,
	}
	domain.Subdomains[subdomainKey] = subdomain
	model.Domains[domainKey] = domain
	return model, assocKey, fromKey, toKey, acKey
}

func (s *AssociationUniquenessCheckerSuite) TestWithinCapNoViolation() {
	model, assocKey, fromKey, toKey, acKey := s.buildModel("0..1")
	checker := NewAssociationUniquenessChecker(model)

	simState := state.NewSimulationState()
	fromInst := simState.CreateInstance(fromKey, object.NewRecord())
	toInst := simState.CreateInstance(toKey, object.NewRecord())
	linkInst := simState.CreateInstance(acKey, object.NewRecord())
	simState.AddAssociationLink(assocKey, fromInst.ID, toInst.ID, linkInst.ID)

	violations := checker.CheckState(simState)
	s.Empty(violations)
}

func (s *AssociationUniquenessCheckerSuite) TestNoLinksDoesNotViolate() {
	model, _, _, _, _ := s.buildModel("1")
	checker := NewAssociationUniquenessChecker(model)

	simState := state.NewSimulationState()
	simState.CreateInstance(multiplicityMustKey("domain/d/subdomain/s/class/order"), object.NewRecord())
	simState.CreateInstance(multiplicityMustKey("domain/d/subdomain/s/class/item"), object.NewRecord())

	violations := checker.CheckState(simState)
	s.Empty(violations)
}

func (s *AssociationUniquenessCheckerSuite) TestExceedsCapReportsViolation() {
	model, assocKey, fromKey, toKey, acKey := s.buildModel("0..1")
	checker := NewAssociationUniquenessChecker(model)

	simState := state.NewSimulationState()
	fromInst := simState.CreateInstance(fromKey, object.NewRecord())
	toInst := simState.CreateInstance(toKey, object.NewRecord())
	link1 := simState.CreateInstance(acKey, object.NewRecord())
	link2 := simState.CreateInstance(acKey, object.NewRecord())
	simState.AddAssociationLink(assocKey, fromInst.ID, toInst.ID, link1.ID)
	simState.AddAssociationLink(assocKey, fromInst.ID, toInst.ID, link2.ID)

	violations := checker.CheckState(simState)
	s.Require().Len(violations, 1)
	s.Equal(ViolationTypeAssociationUniqueness, violations[0].Type)
}

func associationUniquenessTestLinkClass() (model_class.Class, identity.Key) {
	classKey := multiplicityMustKey("domain/d/subdomain/s/class/link")
	stateActiveKey := multiplicityMustKey("domain/d/subdomain/s/class/link/state/active")
	eventCreateKey := multiplicityMustKey("domain/d/subdomain/s/class/link/event/create")
	transCreateKey := multiplicityMustKey("domain/d/subdomain/s/class/link/transition/create")

	eventCreate := model_state.NewEvent(eventCreateKey, "create", "", nil)
	stateActive := model_state.NewState(stateActiveKey, "Active", "", "")
	transCreate := model_state.NewTransition(transCreateKey, eventCreateKey, model_state.TransitionStateKeys{FromStateKey: nil, ToStateKey: &stateActiveKey}, model_state.TransitionLogicKeys{GuardKey: nil, ActionKey: nil}, "")

	class := model_class.NewClass(classKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Link"})
	class.SetStates(map[identity.Key]model_state.State{stateActiveKey: stateActive})
	class.SetEvents(map[identity.Key]model_state.Event{eventCreateKey: eventCreate})
	class.SetTransitions(map[identity.Key]model_state.Transition{transCreateKey: transCreate})
	return class, classKey
}

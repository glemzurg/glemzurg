package instance

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/schema"
	"github.com/stretchr/testify/suite"
)

type AssociationClassHostCheckerSuite struct {
	suite.Suite
}

func TestAssociationClassHostCheckerSuite(t *testing.T) {
	suite.Run(t, new(AssociationClassHostCheckerSuite))
}

func (s *AssociationClassHostCheckerSuite) TestPlainLinkWithInScopeACReportsViolation() {
	model, assocKey, fromKey, toKey, acKey := s.buildACHostedModel()
	sch := schema.New(model, schema.RunScopeAll())
	checker := NewAssociationClassHostChecker(sch)

	simState := NewState(sch)
	from := simState.CreateInstance(fromKey, object.NewRecord())
	to := simState.CreateInstance(toKey, object.NewRecord())
	// Model/engine mistake: plain host link while AC is in scope.
	s.Require().NoError(simState.AddLink(assocKey, from.GetID(), to.GetID()))

	violations := checker.CheckState(simState)
	s.Require().Len(violations, 1)
	s.Equal(ViolationTypeAssociationClassMissing, violations[0].Type)
	s.Contains(violations[0].Message, "association class missing")
	s.Equal(acKey, violations[0].ClassKey)
	s.Equal(from.GetID(), violations[0].InstanceID)
}

func (s *AssociationClassHostCheckerSuite) TestAssociationClassRowNoViolation() {
	model, assocKey, fromKey, toKey, acKey := s.buildACHostedModel()
	sch := schema.New(model, schema.RunScopeAll())
	checker := NewAssociationClassHostChecker(sch)

	simState := NewState(sch)
	from := simState.CreateInstance(fromKey, object.NewRecord())
	to := simState.CreateInstance(toKey, object.NewRecord())
	ac := simState.CreateInstance(acKey, object.NewRecord())
	s.Require().NoError(simState.AddAssociationLink(assocKey, from.GetID(), to.GetID(), ac.GetID()))

	violations := checker.CheckState(simState)
	s.Empty(violations)
}

func (s *AssociationClassHostCheckerSuite) TestPlainLinkWhenACOutOfScopeNoViolation() {
	model, assocKey, fromKey, toKey, acKey := s.buildACHostedModel()
	// Endpoints only — AC deliberately out of run scope (host degrades to plain links).
	sch := schema.New(model, schema.NewRunScope([]identity.Key{fromKey, toKey}))
	s.False(sch.IsClassInScope(acKey))
	checker := NewAssociationClassHostChecker(sch)

	simState := NewState(sch)
	from := simState.CreateInstance(fromKey, object.NewRecord())
	to := simState.CreateInstance(toKey, object.NewRecord())
	s.Require().NoError(simState.AddLink(assocKey, from.GetID(), to.GetID()))

	violations := checker.CheckState(simState)
	s.Empty(violations)
}

func (s *AssociationClassHostCheckerSuite) buildACHostedModel() (
	model *core.Model,
	assocKey, fromKey, toKey, acKey identity.Key,
) {
	fromClass, fromKey := multiplicityTestOrderClass()
	toClass, toKey := multiplicityTestItemClass()
	acKey = multiplicityMustKey("domain/d/subdomain/s/class/link_def")
	acClass := s.linkDefClass(acKey)

	assocKey = multiplicityTestAssocKey(fromKey, toKey)
	assoc := model_class.NewAssociation(
		assocKey,
		model_class.AssociationDetails{Name: "Configures", Details: ""},
		model_class.AssociationEnd{ClassKey: fromKey, Multiplicity: helper.Must(model_class.NewMultiplicity("any"))},
		model_class.AssociationEnd{ClassKey: toKey, Multiplicity: helper.Must(model_class.NewMultiplicity("any"))},
		model_class.AssociationOptions{AssociationClassKey: &acKey},
	)

	model = multiplicityTestModel(
		classEntry(fromClass, fromKey),
		classEntry(toClass, toKey),
		classEntry(acClass, acKey),
	)
	model.ClassAssociations = map[identity.Key]model_class.Association{assocKey: assoc}
	return model, assocKey, fromKey, toKey, acKey
}

func (s *AssociationClassHostCheckerSuite) linkDefClass(classKey identity.Key) model_class.Class {
	stateKey := multiplicityMustKey(classKey.String() + "/state/active")
	eventKey := multiplicityMustKey(classKey.String() + "/event/new")
	transKey := multiplicityMustKey(classKey.String() + "/transition/create")
	state := model_state.NewState(stateKey, "Active", "", "")
	event := model_state.NewEvent(eventKey, "_new", "", nil)
	trans := model_state.NewTransition(
		transKey, eventKey,
		model_state.TransitionStateKeys{FromStateKey: nil, ToStateKey: &stateKey},
		model_state.TransitionLogicKeys{},
		"",
	)
	class := model_class.NewClass(classKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Link Def"})
	class.SetStates(map[identity.Key]model_state.State{stateKey: state})
	class.SetEvents(map[identity.Key]model_state.Event{eventKey: event})
	class.SetTransitions(map[identity.Key]model_state.Transition{transKey: trans})
	return class
}

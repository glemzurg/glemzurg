package engine

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/suite"
)

func TestValidateModelSuite(t *testing.T) {
	suite.Run(t, new(ValidateModelSuite))
}

type ValidateModelSuite struct {
	suite.Suite
}

func (s *ValidateModelSuite) jurisdictionClassMissingEventParam() (model_class.Class, identity.Key) {
	classKey := mustKey("domain/finance/subdomain/wallet/class/jurisdiction")
	stateActiveKey := mustKey("domain/finance/subdomain/wallet/class/jurisdiction/state/active")
	eventAddKey := mustKey("domain/finance/subdomain/wallet/class/jurisdiction/event/add")
	actionAddKey := mustKey("domain/finance/subdomain/wallet/class/jurisdiction/action/add")
	transAddKey := mustKey("domain/finance/subdomain/wallet/class/jurisdiction/transition/add")

	eventAdd := model_state.NewEvent(eventAddKey, "Add", "", []string{
		"Name",
		"CountryCode",
		"StateCode",
	})
	actionAdd := model_state.NewAction(actionAddKey, model_state.ActionDetails{Name: "Add", Details: ""}, nil, nil, nil, []model_state.Parameter{
		helper.Must(model_state.NewParameter(actionAddKey, "Name", "unconstrained", false)),
		helper.Must(model_state.NewParameter(actionAddKey, "SocialOnly", "enum of TRUE, FALSE", false)),
		helper.Must(model_state.NewParameter(actionAddKey, "CountryCode", "ref", true)),
		helper.Must(model_state.NewParameter(actionAddKey, "StateCode", "ref", true)),
	})
	stateActive := model_state.NewState(stateActiveKey, "Active", "", "")
	transAdd := model_state.NewTransition(transAddKey, eventAddKey, model_state.TransitionStateKeys{FromStateKey: nil, ToStateKey: &stateActiveKey}, model_state.TransitionLogicKeys{GuardKey: nil, ActionKey: &actionAddKey}, "")

	class := model_class.NewClass(classKey, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: "Jurisdiction", Details: "", UnfinishedNotes: "", UmlComment: ""})
	class.SetAttributes(nil)
	class.SetStates(map[identity.Key]model_state.State{
		stateActiveKey: stateActive,
	})
	class.SetEvents(map[identity.Key]model_state.Event{
		eventAddKey: eventAdd,
	})
	class.SetGuards(map[identity.Key]model_state.Guard{})
	class.SetActions(map[identity.Key]model_state.Action{
		actionAddKey: actionAdd,
	})
	class.SetQueries(map[identity.Key]model_state.Query{})
	class.SetTransitions(map[identity.Key]model_state.Transition{
		transAddKey: transAdd,
	})

	return class, classKey
}

func (s *ValidateModelSuite) TestValidateEventActionParametersReportsMissingActionParam() {
	class, classKey := s.jurisdictionClassMissingEventParam()
	model := testModel(classEntry(class, classKey))

	err := validateSimulationModel(model)
	s.Require().Error(err)
	s.Contains(err.Error(), `class "Jurisdiction" event "Add" action "Add"`)
	s.Contains(err.Error(), `action parameter "SocialOnly" is not declared on the event`)
}

func (s *ValidateModelSuite) TestNewSimulationEngineFailsWithClearEventActionParameterError() {
	class, classKey := s.jurisdictionClassMissingEventParam()
	model := testModel(classEntry(class, classKey))

	_, err := NewSimulationEngine(model, SimulationConfig{MaxSteps: 1, RandomSeed: 42})
	s.Require().Error(err)
	s.Contains(err.Error(), `action parameter "SocialOnly" is not declared on the event`)
}

func (s *ValidateModelSuite) jurisdictionClassReferenceParamsWithoutInvariant() (model_class.Class, identity.Key) {
	class, classKey := s.jurisdictionClassMissingEventParam()
	actionAddKey := mustKey("domain/finance/subdomain/wallet/class/jurisdiction/action/add")

	action := class.Actions[actionAddKey]
	action.Parameters = []model_state.Parameter{
		helper.Must(model_state.NewParameter(actionAddKey, "Name", "unconstrained", false)),
		helper.Must(model_state.NewParameter(actionAddKey, "CountryCode", "ref of ISO 3166-1 two-letter codes", true)),
		helper.Must(model_state.NewParameter(actionAddKey, "StateCode", "ref of ISO 3166-2 subdivision codes", true)),
	}
	class.Actions[actionAddKey] = action

	eventAddKey := mustKey("domain/finance/subdomain/wallet/class/jurisdiction/event/add")
	event := class.Events[eventAddKey]
	event.ParameterNames = []string{"Name", "CountryCode", "StateCode"}
	class.Events[eventAddKey] = event

	return class, classKey
}

func (s *ValidateModelSuite) TestValidateReferenceDataTypeInvariantsReportsMissingActionParameterInvariant() {
	class, classKey := s.jurisdictionClassReferenceParamsWithoutInvariant()
	model := testModel(classEntry(class, classKey))

	err := validateSimulationModel(model)
	s.Require().Error(err)
	s.Contains(err.Error(), `class "Jurisdiction" action "Add" parameter "CountryCode": reference data type has no invariant`)
	s.Contains(err.Error(), `class "Jurisdiction" action "Add" parameter "StateCode": reference data type has no invariant`)
}

func (s *ValidateModelSuite) TestValidateReferenceDataTypeInvariantsAcceptsActionRequire() {
	class, classKey := s.jurisdictionClassReferenceParamsWithoutInvariant()
	actionAddKey := mustKey("domain/finance/subdomain/wallet/class/jurisdiction/action/add")

	requireLogic := model_logic.NewLogic(
		helper.Must(identity.NewActionRequireKey(actionAddKey, "0")),
		model_logic.LogicTypeAssessment,
		"Placeholder precondition for reference parameters.",
		"",
		parsedSpec("TRUE"),
		nil,
	)
	action := class.Actions[actionAddKey]
	action.Requires = []model_logic.Logic{requireLogic}
	class.Actions[actionAddKey] = action

	model := testModel(classEntry(class, classKey))
	s.Require().NoError(validateSimulationModel(model))
}

func (s *ValidateModelSuite) TestValidateReferenceDataTypeInvariantsReportsMissingAttributeInvariant() {
	classKey := mustKey("domain/finance/subdomain/wallet/class/jurisdiction")
	stateActiveKey := mustKey("domain/finance/subdomain/wallet/class/jurisdiction/state/active")
	eventCreateKey := mustKey("domain/finance/subdomain/wallet/class/jurisdiction/event/create")
	transCreateKey := mustKey("domain/finance/subdomain/wallet/class/jurisdiction/transition/create")
	attrCountryCodeKey := mustKey("domain/finance/subdomain/wallet/class/jurisdiction/attribute/country_code")

	attrCountryCode := helper.Must(model_class.NewAttribute(
		attrCountryCodeKey,
		model_class.AttributeDetails{Name: "Country Code", Details: ""},
		"ref of ISO 3166-1 two-letter codes",
		nil,
		false,
		model_class.AttributeAnnotations{},
	))

	eventCreate := model_state.NewEvent(eventCreateKey, "create", "", nil)
	stateActive := model_state.NewState(stateActiveKey, "Active", "", "")
	transCreate := model_state.NewTransition(transCreateKey, eventCreateKey, model_state.TransitionStateKeys{FromStateKey: nil, ToStateKey: &stateActiveKey}, model_state.TransitionLogicKeys{GuardKey: nil, ActionKey: nil}, "")

	class := model_class.NewClass(classKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Jurisdiction"})
	class.SetAttributes([]model_class.Attribute{attrCountryCode})
	class.SetStates(map[identity.Key]model_state.State{stateActiveKey: stateActive})
	class.SetEvents(map[identity.Key]model_state.Event{eventCreateKey: eventCreate})
	class.SetGuards(map[identity.Key]model_state.Guard{})
	class.SetActions(map[identity.Key]model_state.Action{})
	class.SetQueries(map[identity.Key]model_state.Query{})
	class.SetTransitions(map[identity.Key]model_state.Transition{transCreateKey: transCreate})

	model := testModel(classEntry(class, classKey))

	err := validateSimulationModel(model)
	s.Require().Error(err)
	s.Contains(err.Error(), `class "Jurisdiction" attribute "Country Code": reference data type has no invariant`)
}

package engine

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
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
	actionAdd := model_state.NewAction(
		actionAddKey,
		"Add",
		"",
		nil,
		nil,
		nil,
		[]model_state.Parameter{
			helper.Must(model_state.NewParameter(actionAddKey, "Name", "unconstrained", false)),
			helper.Must(model_state.NewParameter(actionAddKey, "SocialOnly", "enum of TRUE, FALSE", false)),
			helper.Must(model_state.NewParameter(actionAddKey, "CountryCode", "ref", true)),
			helper.Must(model_state.NewParameter(actionAddKey, "StateCode", "ref", true)),
		},
	)
	stateActive := model_state.NewState(stateActiveKey, "Active", "", "")
	transAdd := model_state.NewTransition(transAddKey, nil, eventAddKey, nil, &actionAddKey, &stateActiveKey, "")

	class := model_class.NewClass(classKey, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: "Jurisdiction", Details: "", UnfinishedNotes: "", UmlComment: ""})
	class.SetAttributes(map[identity.Key]model_class.Attribute{})
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

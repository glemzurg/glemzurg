package actions

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/stretchr/testify/suite"
)

func TestSimulationParameterSuite(t *testing.T) {
	suite.Run(t, new(SimulationParameterSuite))
}

type SimulationParameterSuite struct {
	suite.Suite
}

func (s *SimulationParameterSuite) actionWithSimulation() model_state.Action {
	classKey := helper.Must(identity.NewClassKey(
		helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("finance")), "wallet")),
		"transaction",
	))
	actionKey := helper.Must(identity.NewActionKey(classKey, "initialize"))
	paramKey := helper.Must(identity.NewParameterKey(actionKey, "amounts"))
	reqKey := helper.Must(identity.NewParameterSimulationRequireKey(paramKey, "0"))
	specKey := helper.Must(identity.NewParameterSimulationSpecKey(paramKey))

	reqLogic := model_logic.NewLogic(
		reqKey,
		model_logic.LogicTypeAssessment,
		"",
		"",
		logic_spec.ExpressionSpec{
			Notation:      model_logic.NotationTLAPlus,
			Specification: tlaLiteralTrue,
			Expression:    &me.BoolLiteral{Value: true},
		},
		nil,
	)
	specLogic := model_logic.NewLogic(
		specKey,
		model_logic.LogicTypeValue,
		"",
		"",
		logic_spec.ExpressionSpec{
			Notation:      model_logic.NotationTLAPlus,
			Specification: "{}",
			Expression:    &me.SetLiteral{Elements: nil},
		},
		nil,
	)
	param := helper.Must(model_state.NewParameter(actionKey, "Amounts", "unordered of unconstrained", false))
	param.SetSimulation(&model_state.ParameterSimulation{
		Requires:      []model_logic.Logic{reqLogic},
		Specification: &specLogic,
	})
	return model_state.NewAction(
		actionKey,
		model_state.ActionDetails{Name: "Initialize", Details: ""},
		nil,
		nil,
		nil,
		[]model_state.Parameter{param},
	)
}

func (s *SimulationParameterSuite) TestActionHasParameterSimulation() {
	action := s.actionWithSimulation()
	s.True(ActionHasParameterSimulation(action))

	plain := model_state.NewAction(
		action.Key,
		model_state.ActionDetails{Name: "Initialize", Details: ""},
		nil,
		nil,
		nil,
		nil,
	)
	s.False(ActionHasParameterSimulation(plain))
}

func (s *SimulationParameterSuite) TestActionSimulationRequiresMet() {
	action := s.actionWithSimulation()
	ok, err := ActionSimulationRequiresMet(action, evaluator.NewBindings())
	s.Require().NoError(err)
	s.True(ok)
}

func (s *SimulationParameterSuite) TestEvaluateSimulationSpecification() {
	action := s.actionWithSimulation()
	value, err := EvaluateSimulationSpecification(action.Parameters[0], evaluator.NewBindings())
	s.Require().NoError(err)
	s.NotNil(value)
}

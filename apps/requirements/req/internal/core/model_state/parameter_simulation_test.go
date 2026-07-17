package model_state

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

type ParameterSimulationTestSuite struct {
	suite.Suite
}

func TestParameterSimulationSuite(t *testing.T) {
	suite.Run(t, new(ParameterSimulationTestSuite))
}

func (s *ParameterSimulationTestSuite) actionParamKey() identity.Key {
	actionKey := helper.Must(identity.NewActionKey(testClassKey(), "initialize"))
	return helper.Must(identity.NewParameterKey(actionKey, "amounts"))
}

func testClassKey() identity.Key {
	return helper.Must(identity.NewClassKey(
		helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("domain_a")), "subdomain_a")),
		"transaction",
	))
}

func (s *ParameterSimulationTestSuite) TestValidateSimulationRules() {
	paramKey := s.actionParamKey()
	reqKey0 := helper.Must(identity.NewParameterSimulationRequireKey(paramKey, "0"))
	specKey0 := helper.Must(identity.NewParameterSimulationSpecKey(paramKey, "0"))
	reqKey1 := helper.Must(identity.NewParameterSimulationRequireKey(paramKey, "1"))
	specKey1 := helper.Must(identity.NewParameterSimulationSpecKey(paramKey, "1"))

	spec0 := model_logic.NewLogic(specKey0, model_logic.LogicTypeValue, "single", "", logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "{}"}, nil)
	spec1 := model_logic.NewLogic(specKey1, model_logic.LogicTypeValue, "multi", "", logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "{a}"}, nil)
	sim := &ParameterSimulation{
		Details: "sample amounts",
		Rules: []ParameterSimulationRule{
			{
				Details: "one account",
				Requires: []model_logic.Logic{
					model_logic.NewLogic(reqKey0, model_logic.LogicTypeAssessment, "pool non-empty", "", logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "TRUE"}, nil),
				},
				Specification: &spec0,
			},
			{
				Details: "three accounts",
				Requires: []model_logic.Logic{
					model_logic.NewLogic(reqKey1, model_logic.LogicTypeAssessment, "pool size", "", logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "TRUE"}, nil),
				},
				Specification: &spec1,
			},
		},
	}
	ctx := coreerr.NewContext("test", "")
	s.Require().NoError(sim.Validate(ctx, paramKey))
	s.True(sim.HasSimulation())
	s.True(sim.Rules[0].HasSpecification())
}

func (s *ParameterSimulationTestSuite) TestValidateRejectsWrongRequireType() {
	paramKey := s.actionParamKey()
	reqKey := helper.Must(identity.NewParameterSimulationRequireKey(paramKey, "0"))
	sim := &ParameterSimulation{
		Rules: []ParameterSimulationRule{
			{
				Requires: []model_logic.Logic{
					model_logic.NewLogic(reqKey, model_logic.LogicTypeValue, "", "", logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus}, nil),
				},
			},
		},
	}
	ctx := coreerr.NewContext("test", "")
	err := sim.Validate(ctx, paramKey)
	s.Require().Error(err)
	s.Contains(err.Error(), "assessment")
}

func (s *ParameterSimulationTestSuite) TestSpecKeyRequiresIntegerRuleIndex() {
	paramKey := s.actionParamKey()
	_, err := identity.NewParameterSimulationSpecKey(paramKey, "spec")
	s.Require().Error(err)
	_, err = identity.NewParameterSimulationSpecKey(paramKey, "0")
	s.Require().NoError(err)
}

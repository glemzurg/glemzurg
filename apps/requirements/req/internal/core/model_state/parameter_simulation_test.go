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

func (s *ParameterSimulationTestSuite) TestValidateSimulationRequiresAndSpec() {
	paramKey := s.actionParamKey()
	reqKey := helper.Must(identity.NewParameterSimulationRequireKey(paramKey, "0"))
	specKey := helper.Must(identity.NewParameterSimulationSpecKey(paramKey))

	spec := model_logic.NewLogic(specKey, model_logic.LogicTypeValue, "amounts value", "", logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "{}"}, nil)
	sim := &ParameterSimulation{
		Details: "sample amounts",
		Requires: []model_logic.Logic{
			model_logic.NewLogic(reqKey, model_logic.LogicTypeAssessment, "pool size", "", logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "TRUE"}, nil),
		},
		Specification: &spec,
	}
	ctx := coreerr.NewContext("test", "")
	s.Require().NoError(sim.Validate(ctx, paramKey))
}

func (s *ParameterSimulationTestSuite) TestValidateRejectsWrongRequireType() {
	paramKey := s.actionParamKey()
	reqKey := helper.Must(identity.NewParameterSimulationRequireKey(paramKey, "0"))
	sim := &ParameterSimulation{
		Requires: []model_logic.Logic{
			model_logic.NewLogic(reqKey, model_logic.LogicTypeValue, "", "", logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus}, nil),
		},
	}
	ctx := coreerr.NewContext("test", "")
	err := sim.Validate(ctx, paramKey)
	s.Require().Error(err)
	s.Contains(err.Error(), "assessment")
}

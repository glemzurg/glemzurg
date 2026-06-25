package actions

import (
	"math/rand"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/convert"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

func (s *ParameterSamplerSuite) currencyAddActionWithISORequire() model_state.Action {
	classKey := mustKey("domain/finance/wallet/class/currency")
	actionKey := helper.Must(identity.NewActionKey(classKey, "add"))
	requireLogic := model_logic.NewLogic(
		helper.Must(identity.NewActionRequireKey(actionKey, "0")),
		model_logic.LogicTypeAssessment,
		"Valid ISO 4217 code when ISO is provided.",
		"",
		currencyRequireSpec(),
		nil,
	)
	return model_state.NewAction(actionKey, model_state.ActionDetails{Name: "Add", Details: ""}, []model_logic.Logic{requireLogic}, nil, nil, []model_state.Parameter{
		helper.Must(model_state.NewParameter(actionKey, "Type", "enum of SOCIAL, REAL", false)),
		helper.Must(model_state.NewParameter(actionKey, "ISO", "ref of valid ISO 4217 codes", true)),
	})
}

func (s *ParameterSamplerSuite) TestSampleImplicitEnumParameterWithoutExplicitRequire() {
	action := s.currencyAddActionWithISORequire()
	params := []model_state.Parameter{
		helper.Must(model_state.NewParameter(action.Key, "Type", "enum of SOCIAL, REAL", false)),
	}

	binder := NewParameterBinder()
	sampler := NewParameterSampler(binder, nil)
	rng := rand.New(rand.NewSource(11)) //nolint:gosec // deterministic test seed

	result, err := sampler.SampleFromRequires(params, &action, rng)
	s.Require().NoError(err)

	value, ok := result["Type"].(*object.String)
	s.Require().True(ok)
	s.Contains([]string{"SOCIAL", "REAL"}, value.Value())
}

func (s *ParameterSamplerSuite) TestSampleQueryImplicitEnumParameter() {
	classKey := mustKey("domain/finance/wallet/class/currency")
	queryKey := helper.Must(identity.NewQueryKey(classKey, "filter"))
	query := model_state.NewQuery(queryKey, "Filter", "", nil, nil, []model_state.Parameter{
		helper.Must(model_state.NewParameter(queryKey, "Type", "enum of SOCIAL, REAL", false)),
	})

	binder := NewParameterBinder()
	sampler := NewParameterSampler(binder, nil)
	rng := rand.New(rand.NewSource(13)) //nolint:gosec // deterministic test seed

	result, err := sampler.SampleQueryFromRequires(query, rng)
	s.Require().NoError(err)

	value, ok := result["Type"].(*object.String)
	s.Require().True(ok)
	s.Contains([]string{"SOCIAL", "REAL"}, value.Value())
}

func (s *ParameterSamplerSuite) TestEffectiveRequiresSkipsExplicitEnumRequire() {
	classKey := mustKey("domain/finance/wallet/class/jurisdiction")
	actionKey := helper.Must(identity.NewActionKey(classKey, "add"))
	ctx := &convert.LowerContext{
		ClassKey:   classKey,
		Parameters: map[string]bool{"SocialOnly": true},
	}
	pf := convert.NewExpressionParseFunc(ctx)
	explicit := []model_logic.Logic{
		model_logic.NewLogic(
			helper.Must(identity.NewActionRequireKey(actionKey, "0")),
			model_logic.LogicTypeAssessment,
			"Social only flag must be valid.",
			"",
			helper.Must(logic_spec.NewExpressionSpec("tla_plus", `SocialOnly \in {"TRUE", "FALSE"}`, pf)),
			nil,
		),
	}
	params := []model_state.Parameter{
		helper.Must(model_state.NewParameter(actionKey, "SocialOnly", "enum of TRUE, FALSE", false)),
	}

	owner := ParameterOwner{Key: actionKey, Kind: logicOwnerKindAction, Parameters: params, Requires: explicit}
	effective, err := owner.EffectiveRequiresFor(params)
	s.Require().NoError(err)
	s.Len(effective, 1)
}

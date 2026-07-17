package actions

import (
	"math/rand"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/convert"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

func bareISOInvariantSpec() logic_spec.ExpressionSpec {
	classKey := mustKey("domain/finance/wallet/class/currency")
	iso4217CodesKey := helper.Must(identity.NewNamedSetKey("iso4217codes"))
	ctx := &convert.LowerContext{
		ClassKey: classKey,
		Parameters: map[string]bool{
			"ISO": true,
		},
		NamedSets: map[string]identity.Key{
			"_Iso4217Codes": iso4217CodesKey,
		},
	}
	pf := convert.NewExpressionParseFunc(ctx)
	return helper.Must(logic_spec.NewExpressionSpec("tla_plus", `ISO \in _Iso4217Codes`, pf))
}

func accountIDInvariantSpec() logic_spec.ExpressionSpec {
	classKey := mustKey("domain/finance/wallet/class/account")
	ctx := &convert.LowerContext{
		ClassKey: classKey,
		Parameters: map[string]bool{
			"accountId": true,
		},
	}
	pf := convert.NewExpressionParseFunc(ctx)
	return helper.Must(logic_spec.NewExpressionSpec("tla_plus", "accountId > 0", pf))
}

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

func (s *ParameterSamplerSuite) TestSampleParameterInvariantMembershipFromNamedSet() {
	classKey := mustKey("domain/finance/wallet/class/currency")
	actionKey := helper.Must(identity.NewActionKey(classKey, "add"))
	isoParamKey := helper.Must(identity.NewParameterKey(actionKey, "iso"))
	isoParam := helper.Must(model_state.NewParameter(actionKey, "ISO", "ref of valid ISO 4217 codes", true))
	isoParam.SetInvariants([]model_logic.Logic{
		model_logic.NewLogic(
			helper.Must(identity.NewParameterInvariantKey(isoParamKey, "0")),
			model_logic.LogicTypeAssessment,
			"Valid ISO 4217 code when ISO is provided.",
			"",
			currencyRequireSpec(),
			nil,
		),
	})
	action := model_state.NewAction(actionKey, model_state.ActionDetails{Name: "Add", Details: ""}, nil, nil, nil, []model_state.Parameter{isoParam})
	params := []model_state.Parameter{isoParam}

	binder := NewParameterBinder()
	sampler := NewParameterSampler(binder, s.iso4217NamedSet())

	for seed := range 200 {
		result, err := sampler.SampleFromRequires(params, &action, rand.New(rand.NewSource(int64(seed)))) //nolint:gosec // deterministic test seed
		s.Require().NoError(err)
		if object.IsNull(result["ISO"]) {
			continue
		}
		s.Contains([]string{"USD", "GBP", "CAD", "EUR"}, result["ISO"].(*object.String).Value())
	}
}

func (s *ParameterSamplerSuite) TestEffectiveRequiresExcludesParameterInvariant() {
	classKey := mustKey("domain/finance/wallet/class/currency")
	actionKey := helper.Must(identity.NewActionKey(classKey, "add"))
	isoParamKey := helper.Must(identity.NewParameterKey(actionKey, "iso"))
	isoParam := helper.Must(model_state.NewParameter(actionKey, "ISO", "ref of valid ISO 4217 codes", true))
	isoParam.SetInvariants([]model_logic.Logic{
		model_logic.NewLogic(
			helper.Must(identity.NewParameterInvariantKey(isoParamKey, "0")),
			model_logic.LogicTypeAssessment,
			"Valid ISO 4217 code when ISO is provided.",
			"",
			currencyRequireSpec(),
			nil,
		),
	})
	owner := ParameterOwnerFromAction(model_state.NewAction(actionKey, model_state.ActionDetails{Name: "Add", Details: ""}, nil, nil, nil, []model_state.Parameter{isoParam}))

	effective, err := owner.EffectiveRequiresFor([]model_state.Parameter{isoParam})
	s.Require().NoError(err)
	s.Require().Len(effective, 1)
	s.Contains(effective[0].Key.String(), "implicit_ref_0")
}

func (s *ParameterSamplerSuite) TestSamplingLogicsForIncludesParameterInvariant() {
	classKey := mustKey("domain/finance/wallet/class/currency")
	actionKey := helper.Must(identity.NewActionKey(classKey, "add"))
	isoParamKey := helper.Must(identity.NewParameterKey(actionKey, "iso"))
	isoParam := helper.Must(model_state.NewParameter(actionKey, "ISO", "ref of valid ISO 4217 codes", true))
	isoParam.SetInvariants([]model_logic.Logic{
		model_logic.NewLogic(
			helper.Must(identity.NewParameterInvariantKey(isoParamKey, "0")),
			model_logic.LogicTypeAssessment,
			"Valid ISO 4217 code when ISO is provided.",
			"",
			currencyRequireSpec(),
			nil,
		),
	})
	owner := ParameterOwnerFromAction(model_state.NewAction(actionKey, model_state.ActionDetails{Name: "Add", Details: ""}, nil, nil, nil, []model_state.Parameter{isoParam}))

	sampling, err := owner.SamplingLogicsFor([]model_state.Parameter{isoParam})
	s.Require().NoError(err)
	s.Require().Len(sampling, 2)
	s.Equal(isoParam.Invariants[0].Key, sampling[1].Key)
}

func (s *ParameterSamplerSuite) TestSamplingLogicsWrapsNullableParameterInvariantWithoutAuthorGuard() {
	classKey := mustKey("domain/finance/wallet/class/currency")
	actionKey := helper.Must(identity.NewActionKey(classKey, "add"))
	isoParamKey := helper.Must(identity.NewParameterKey(actionKey, "iso"))
	isoParam := helper.Must(model_state.NewParameter(actionKey, "ISO", "ref of valid ISO 4217 codes", true))
	isoParam.SetInvariants([]model_logic.Logic{
		model_logic.NewLogic(
			helper.Must(identity.NewParameterInvariantKey(isoParamKey, "0")),
			model_logic.LogicTypeAssessment,
			"Valid ISO 4217 code when ISO is provided.",
			"",
			bareISOInvariantSpec(),
			nil,
		),
	})
	owner := ParameterOwnerFromAction(model_state.NewAction(actionKey, model_state.ActionDetails{Name: "Add", Details: ""}, nil, nil, nil, []model_state.Parameter{isoParam}))

	sampling, err := owner.SamplingLogicsFor([]model_state.Parameter{isoParam})
	s.Require().NoError(err)
	s.Require().GreaterOrEqual(len(sampling), 1)

	var wrapped *model_logic.Logic
	for i := range sampling {
		if sampling[i].Key == isoParam.Invariants[0].Key {
			wrapped = &sampling[i]
			break
		}
	}
	s.Require().NotNil(wrapped)
	s.Contains(wrapped.Spec.Specification, "_GZ!WhenNotNull(ISO,")
}

func (s *ParameterSamplerSuite) TestAssessParameterInvariantSkipsWhenNullableAndUnset() {
	classKey := mustKey("domain/finance/wallet/class/account")
	actionKey := helper.Must(identity.NewActionKey(classKey, "open"))
	accountIDParamKey := helper.Must(identity.NewParameterKey(actionKey, "accountid"))
	accountIDParam := helper.Must(model_state.NewParameter(actionKey, "accountId", "positive account id", true))
	accountIDParam.SetInvariants([]model_logic.Logic{
		model_logic.NewLogic(
			helper.Must(identity.NewParameterInvariantKey(accountIDParamKey, "0")),
			model_logic.LogicTypeAssessment,
			"Account id must be positive when provided.",
			"",
			accountIDInvariantSpec(),
			nil,
		),
	})
	owner := ParameterOwnerFromAction(model_state.NewAction(actionKey, model_state.ActionDetails{Name: "Open", Details: ""}, nil, nil, nil, []model_state.Parameter{accountIDParam}))

	bindings := evaluator.NewBindings()
	failures, err := owner.AssessParameterInvariants([]model_state.Parameter{accountIDParam}, bindings)
	s.Require().NoError(err)
	s.Empty(failures)
}

func (s *ParameterSamplerSuite) TestAssessParameterInvariantSkipsParameterEquality() {
	classKey := mustKey("domain/finance/wallet/class/currency")
	actionKey := helper.Must(identity.NewActionKey(classKey, "add"))
	isoParamKey := helper.Must(identity.NewParameterKey(actionKey, "iso"))
	isoParam := helper.Must(model_state.NewParameter(actionKey, "ISO", "ref of valid ISO 4217 codes", true))
	abbrParam := helper.Must(model_state.NewParameter(actionKey, "Abbr", "abbreviation", false))
	ctx := &convert.LowerContext{
		ClassKey:   classKey,
		Parameters: map[string]bool{"ISO": true, "Abbr": true},
	}
	pf := convert.NewExpressionParseFunc(ctx)
	isoParam.SetInvariants([]model_logic.Logic{
		model_logic.NewLogic(
			helper.Must(identity.NewParameterInvariantKey(isoParamKey, "1")),
			model_logic.LogicTypeAssessment,
			"ISO must match Abbr when provided.",
			"",
			helper.Must(logic_spec.NewExpressionSpec("tla_plus", "ISO = Abbr", pf)),
			nil,
		),
	})
	owner := ParameterOwnerFromAction(model_state.NewAction(actionKey, model_state.ActionDetails{Name: "Add", Details: ""}, nil, nil, nil, []model_state.Parameter{isoParam, abbrParam}))

	bindings := evaluator.NewBindings()
	bindings.Set("ISO", object.NewString("USD"), evaluator.NamespaceLocal)
	bindings.Set("Abbr", object.NewString("NS"), evaluator.NamespaceLocal)

	failures, err := owner.AssessParameterInvariants([]model_state.Parameter{isoParam, abbrParam}, bindings)
	s.Require().NoError(err)
	s.Empty(failures)
}

func (s *ParameterSamplerSuite) TestAssessParameterInvariantFailsWhenSetAndViolated() {
	classKey := mustKey("domain/finance/wallet/class/account")
	actionKey := helper.Must(identity.NewActionKey(classKey, "open"))
	accountIDParamKey := helper.Must(identity.NewParameterKey(actionKey, "accountid"))
	accountIDParam := helper.Must(model_state.NewParameter(actionKey, "accountId", "positive account id", true))
	accountIDParam.SetInvariants([]model_logic.Logic{
		model_logic.NewLogic(
			helper.Must(identity.NewParameterInvariantKey(accountIDParamKey, "0")),
			model_logic.LogicTypeAssessment,
			"Account id must be positive when provided.",
			"",
			accountIDInvariantSpec(),
			nil,
		),
	})
	owner := ParameterOwnerFromAction(model_state.NewAction(actionKey, model_state.ActionDetails{Name: "Open", Details: ""}, nil, nil, nil, []model_state.Parameter{accountIDParam}))

	bindings := evaluator.NewBindings()
	bindings.Set("accountId", object.NewInteger(0), evaluator.NamespaceLocal)

	failures, err := owner.AssessParameterInvariants([]model_state.Parameter{accountIDParam}, bindings)
	s.Require().NoError(err)
	s.Require().Len(failures, 1)
}

func (s *ParameterSamplerSuite) TestEffectiveRequiresSynthesizesImplicitReferenceRequire() {
	classKey := mustKey("domain/finance/wallet/class/currency")
	actionKey := helper.Must(identity.NewActionKey(classKey, "add"))
	isoParam := helper.Must(model_state.NewParameter(actionKey, "ISO", "ref of valid ISO 4217 codes", true))
	owner := ParameterOwnerFromAction(model_state.NewAction(actionKey, model_state.ActionDetails{Name: "Add", Details: ""}, nil, nil, nil, []model_state.Parameter{isoParam}))

	effective, err := owner.EffectiveRequiresFor([]model_state.Parameter{isoParam})
	s.Require().NoError(err)
	s.Require().Len(effective, 1)
	s.Contains(effective[0].Key.String(), "implicit_ref_0")
	s.Equal("TRUE", effective[0].Spec.Specification)
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

package actions

import (
	"math/rand"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/convert"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/stretchr/testify/suite"
)

type ParameterSamplerSuite struct {
	suite.Suite
}

func TestParameterSamplerSuite(t *testing.T) {
	suite.Run(t, new(ParameterSamplerSuite))
}

func jurisdictionRequireSpec(tla string) logic_spec.ExpressionSpec {
	classKey := mustKey("domain/finance/wallet/class/jurisdiction")
	jurisdictionCodesKey := helper.Must(identity.NewNamedSetKey("jurisdictioncodes"))
	ctx := &convert.LowerContext{
		ClassKey: classKey,
		Parameters: map[string]bool{
			"CountryCode": true,
			"StateCode":   true,
		},
		NamedSets: map[string]identity.Key{
			"_JurisdictionCodes": jurisdictionCodesKey,
		},
	}
	pf := convert.NewExpressionParseFunc(ctx)
	return helper.Must(logic_spec.NewExpressionSpec("tla_plus", tla, pf))
}

func (s *ParameterSamplerSuite) jurisdictionNamedSet() map[string]object.Object {
	jurisdictionCodes := object.NewSetFromElements([]object.Object{
		object.NewTupleFromElements([]object.Object{object.NewString("US"), object.NewString("CA")}),
		object.NewTupleFromElements([]object.Object{object.NewString("US"), object.NewString("NY")}),
		object.NewTupleFromElements([]object.Object{object.NewString("GB"), object.NewString("")}),
	})
	return map[string]object.Object{
		"jurisdictioncodes": jurisdictionCodes,
	}
}

func (s *ParameterSamplerSuite) jurisdictionAction() model_state.Action {
	classKey := mustKey("domain/finance/wallet/class/jurisdiction")
	actionKey := helper.Must(identity.NewActionKey(classKey, "add"))
	requireLogic := model_logic.NewLogic(
		helper.Must(identity.NewActionRequireKey(actionKey, "0")),
		model_logic.LogicTypeAssessment,
		"Valid jurisdiction pair when country is provided.",
		"",
		jurisdictionRequireSpec(`IF CountryCode = NULL THEN StateCode = NULL ELSE <<CountryCode, StateCode>> \in _JurisdictionCodes`),
		nil,
	)
	return model_state.NewAction(actionKey, "Add", "", []model_logic.Logic{requireLogic}, nil, nil, nil)
}

func (s *ParameterSamplerSuite) jurisdictionParams() []model_state.Parameter {
	classKey := mustKey("domain/finance/wallet/class/jurisdiction")
	actionKey := helper.Must(identity.NewActionKey(classKey, "add"))
	return []model_state.Parameter{
		helper.Must(model_state.NewParameter(actionKey, "CountryCode", "ref of ISO 3166-1 two-letter codes", false)),
		helper.Must(model_state.NewParameter(actionKey, "StateCode", "ref of ISO 3166-2 subdivision codes", false)),
	}
}

func (s *ParameterSamplerSuite) TestExtractNullableElseTupleConstraint() {
	constraints := extractParameterConstraints([]model_logic.Logic{
		model_logic.NewLogic(
			helper.Must(identity.NewActionRequireKey(helper.Must(identity.NewActionKey(mustKey("domain/finance/wallet/class/jurisdiction"), "add")), "0")),
			model_logic.LogicTypeAssessment,
			"Valid jurisdiction pair when country is provided.",
			"",
			jurisdictionRequireSpec(`IF CountryCode = NULL THEN StateCode = NULL ELSE <<CountryCode, StateCode>> \in _JurisdictionCodes`),
			nil,
		),
	})
	s.Require().NotNil(constraints.nullableElseTuple)
	s.Equal("CountryCode", constraints.nullableElseTuple.paramNames[0])
	s.Equal("jurisdictioncodes", constraints.nullableElseTuple.setSubKey)
}

func (s *ParameterSamplerSuite) TestExtractLoweredNullableElseTupleConstraint() {
	constraints := extractParameterConstraints([]model_logic.Logic{
		model_logic.NewLogic(
			helper.Must(identity.NewActionRequireKey(helper.Must(identity.NewActionKey(mustKey("domain/finance/wallet/class/jurisdiction"), "add")), "0")),
			model_logic.LogicTypeAssessment,
			"Valid jurisdiction pair when country is provided.",
			"",
			jurisdictionRequireSpec(`IF CountryCode = {} THEN StateCode = {} ELSE ⟨CountryCode, StateCode⟩ ∈ _JurisdictionCodes`),
			nil,
		),
	})
	s.Require().NotNil(constraints.nullableElseTuple)
}

func (s *ParameterSamplerSuite) TestPickRandomTupleFromNamedSet() {
	for seed := range 5 {
		tuple, ok := pickRandomTuple("jurisdictioncodes", s.jurisdictionNamedSet(), rand.New(rand.NewSource(int64(seed)))) //nolint:gosec // deterministic test seed
		s.True(ok, "seed %d", seed)
		s.NotNil(tuple, "seed %d", seed)
	}
}

func (s *ParameterSamplerSuite) TestApplyNullableElseTupleSetsCoupledValues() {
	constraint := &nullableElseTupleConstraint{
		conditionParam: "CountryCode",
		thenParam:      "StateCode",
		paramNames:     []string{"CountryCode", "StateCode"},
		setSubKey:      "jurisdictioncodes",
	}

	for seed := range 30 {
		result := map[string]object.Object{
			"CountryCode": object.NewString("junk"),
			"StateCode":   object.NewString("junk"),
		}
		applyNullableElseTuple(result, constraint, rand.New(rand.NewSource(int64(seed))), s.jurisdictionNamedSet()) //nolint:gosec // deterministic test seed
		s.assertJurisdictionPair(result["CountryCode"], result["StateCode"], int64(seed))
	}
}

func (s *ParameterSamplerSuite) assertJurisdictionPair(country, state object.Object, seed int64) {
	if object.IsNull(country) {
		s.True(object.IsNull(state), "seed %d", seed)
		return
	}
	countryStr, ok := country.(*object.String)
	s.Require().True(ok, "seed %d", seed)
	switch countryStr.Value() {
	case "GB":
		s.True(object.IsNull(state), "seed %d", seed)
	case "US":
		stateStr, ok := state.(*object.String)
		s.Require().True(ok, "seed %d", seed)
		s.Contains([]string{"CA", "NY"}, stateStr.Value(), "seed %d", seed)
	default:
		s.Failf("unexpected country", "country %q seed %d", countryStr.Value(), seed)
	}
}

func (s *ParameterSamplerSuite) TestSampleNullableElseTupleFromNamedSet() {
	binder := NewParameterBinder()
	sampler := NewParameterSampler(binder, s.jurisdictionNamedSet())
	action := s.jurisdictionAction()
	params := s.jurisdictionParams()

	for seed := range 30 {
		result, err := sampler.SampleFromRequires(params, &action, rand.New(rand.NewSource(int64(seed)))) //nolint:gosec // deterministic test seed
		s.Require().NoError(err)
		s.assertJurisdictionPair(result["CountryCode"], result["StateCode"], int64(seed))
	}
}

func (s *ParameterSamplerSuite) TestSampleEnumConstraint() {
	classKey := mustKey("domain/finance/wallet/class/jurisdiction")
	actionKey := helper.Must(identity.NewActionKey(classKey, "add"))
	ctx := &convert.LowerContext{
		ClassKey: classKey,
		Parameters: map[string]bool{
			"SocialOnly": true,
		},
	}
	pf := convert.NewExpressionParseFunc(ctx)
	requireLogic := model_logic.NewLogic(
		helper.Must(identity.NewActionRequireKey(actionKey, "0")),
		model_logic.LogicTypeAssessment,
		"Social only flag must be valid.",
		"",
		helper.Must(logic_spec.NewExpressionSpec("tla_plus", `SocialOnly \in {"TRUE", "FALSE"}`, pf)),
		nil,
	)
	action := model_state.NewAction(actionKey, "Add", "", []model_logic.Logic{requireLogic}, nil, nil, nil)
	params := []model_state.Parameter{
		helper.Must(model_state.NewParameter(actionKey, "SocialOnly", "enum of TRUE, FALSE", false)),
	}

	binder := NewParameterBinder()
	sampler := NewParameterSampler(binder, nil)
	rng := rand.New(rand.NewSource(7)) //nolint:gosec // deterministic test seed

	result, err := sampler.SampleFromRequires(params, &action, rng)
	s.Require().NoError(err)
	value, ok := result["SocialOnly"].(*object.String)
	s.Require().True(ok)
	s.Contains([]string{"TRUE", "FALSE"}, value.Value())
}

func (s *ParameterSamplerSuite) TestSampleFallsBackWithoutRequires() {
	classKey := mustKey("domain/finance/wallet/class/jurisdiction")
	actionKey := helper.Must(identity.NewActionKey(classKey, "add"))
	params := []model_state.Parameter{
		helper.Must(model_state.NewParameter(actionKey, "Name", "unconstrained", false)),
	}

	binder := NewParameterBinder()
	sampler := NewParameterSampler(binder, nil)
	rng := rand.New(rand.NewSource(1)) //nolint:gosec // deterministic test seed

	result, err := sampler.SampleFromRequires(params, nil, rng)
	s.Require().NoError(err)
	s.NotNil(result["Name"])
}

func (s *ParameterSamplerSuite) TestBareParameterReferenceRequireIsSupported() {
	classKey := mustKey("domain/finance/wallet/class/jurisdiction")
	actionKey := helper.Must(identity.NewActionKey(classKey, "add"))
	ctx := &convert.LowerContext{
		ClassKey: classKey,
		Parameters: map[string]bool{
			"Name": true,
		},
	}
	pf := convert.NewExpressionParseFunc(ctx)
	requireLogic := model_logic.NewLogic(
		helper.Must(identity.NewActionRequireKey(actionKey, "0")),
		model_logic.LogicTypeAssessment,
		"Name is provided.",
		"",
		helper.Must(logic_spec.NewExpressionSpec("tla_plus", "Name", pf)),
		nil,
	)
	action := model_state.NewAction(
		actionKey,
		"Add",
		"",
		[]model_logic.Logic{requireLogic},
		nil,
		nil,
		[]model_state.Parameter{
			helper.Must(model_state.NewParameter(actionKey, "Name", "display name", false)),
		},
	)

	err := ValidateActionRequiresSamplingSupport("Jurisdiction", action)
	s.NoError(err)
}

func (s *ParameterSamplerSuite) TestUnsupportedCompareRequireReturnsSpecificError() {
	classKey := mustKey("domain/d/subdomain/s/class/order")
	actionKey := helper.Must(identity.NewActionKey(classKey, "set_amount"))
	spec := orderSpecWithParams("Amount > 0", []string{"Amount"})
	requireLogic := model_logic.NewLogic(
		helper.Must(identity.NewActionRequireKey(actionKey, "0")),
		model_logic.LogicTypeAssessment,
		"Amount must be positive.",
		"",
		spec,
		nil,
	)
	action := model_state.NewAction(
		actionKey,
		"SetAmount",
		"",
		[]model_logic.Logic{requireLogic},
		nil,
		nil,
		[]model_state.Parameter{
			helper.Must(model_state.NewParameter(actionKey, "Amount", "positive amount", false)),
		},
	)

	err := ValidateActionRequiresSamplingSupport("Order", action)
	s.Require().Error(err)

	var unsupported *UnsupportedRequiresSamplingError
	s.Require().ErrorAs(err, &unsupported)
	s.Equal("Order", unsupported.ClassName)
	s.Equal("SetAmount", unsupported.ActionName)
	s.Contains(unsupported.Error(), "cannot derive random parameter values")
}

func (s *ParameterSamplerSuite) TestSampleFromRequiresReturnsUnsupportedError() {
	classKey := mustKey("domain/d/subdomain/s/class/order")
	actionKey := helper.Must(identity.NewActionKey(classKey, "set_amount"))
	spec := orderSpecWithParams("Amount > 0", []string{"Amount"})
	action := model_state.NewAction(
		actionKey,
		"SetAmount",
		"",
		[]model_logic.Logic{
			model_logic.NewLogic(
				helper.Must(identity.NewActionRequireKey(actionKey, "0")),
				model_logic.LogicTypeAssessment,
				"Amount must be positive.",
				"",
				spec,
				nil,
			),
		},
		nil,
		nil,
		[]model_state.Parameter{
			helper.Must(model_state.NewParameter(actionKey, "Amount", "positive amount", false)),
		},
	)
	params := []model_state.Parameter{
		helper.Must(model_state.NewParameter(actionKey, "Amount", "positive amount", false)),
	}

	binder := NewParameterBinder()
	sampler := NewParameterSampler(binder, nil)
	_, err := sampler.SampleFromRequires(params, &action, rand.New(rand.NewSource(1))) //nolint:gosec // deterministic test seed
	s.Require().Error(err)
	s.ErrorAs(err, new(*UnsupportedRequiresSamplingError))
}

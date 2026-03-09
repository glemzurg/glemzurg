package convert

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

type LowerModelTestSuite struct {
	suite.Suite
}

func TestLowerModelSuite(t *testing.T) {
	suite.Run(t, new(LowerModelTestSuite))
}

// mustKey is a test helper that panics on error.
func mustKey(key identity.Key, err error) identity.Key {
	if err != nil {
		panic(err)
	}
	return key
}

// mustSpec creates a TLA+ expression spec for testing.
func mustSpec(spec string) logic_spec.ExpressionSpec {
	s, err := logic_spec.NewExpressionSpec("tla_plus", spec, nil)
	if err != nil {
		panic(err)
	}
	return s
}

// buildTestModel creates a minimal model that exercises all LowerModel paths.
func buildTestModel() *core.Model {
	// Keys.
	domainKey := mustKey(identity.NewDomainKey("d"))
	subKey := mustKey(identity.NewSubdomainKey(domainKey, "s"))
	classKey := mustKey(identity.NewClassKey(subKey, "Account"))
	attrKey := mustKey(identity.NewAttributeKey(classKey, "balance"))
	actionKey := mustKey(identity.NewActionKey(classKey, "Deposit"))
	queryKey := mustKey(identity.NewQueryKey(classKey, "GetBalance"))
	guardKey := mustKey(identity.NewGuardKey(classKey, "HasFunds"))

	invariantKey := mustKey(identity.NewInvariantKey("0"))
	classInvKey := mustKey(identity.NewClassInvariantKey(classKey, "0"))
	attrInvKey := mustKey(identity.NewAttributeInvariantKey(attrKey, "0"))
	attrDerivKey := mustKey(identity.NewAttributeDerivationKey(attrKey, "0"))
	globalFuncKey := mustKey(identity.NewGlobalFunctionKey("_Max"))
	namedSetKey := mustKey(identity.NewNamedSetKey("valid_statuses"))

	actionReqKey := mustKey(identity.NewActionRequireKey(actionKey, "0"))
	actionGuarKey := mustKey(identity.NewActionGuaranteeKey(actionKey, "0"))
	actionSafeKey := mustKey(identity.NewActionSafetyKey(actionKey, "0"))
	queryReqKey := mustKey(identity.NewQueryRequireKey(queryKey, "0"))
	queryGuarKey := mustKey(identity.NewQueryGuaranteeKey(queryKey, "0"))

	// Build Logic objects.
	modelInvariant := model_logic.Logic{
		Key:         invariantKey,
		Type:        model_logic.LogicTypeAssessment,
		Description: "model invariant",
		Spec:        mustSpec("TRUE"),
	}

	classInvariant := model_logic.Logic{
		Key:         classInvKey,
		Type:        model_logic.LogicTypeAssessment,
		Description: "class invariant",
		Spec:        mustSpec("balance ≥ 0"),
	}

	attrInvariant := model_logic.Logic{
		Key:         attrInvKey,
		Type:        model_logic.LogicTypeAssessment,
		Description: "attr invariant",
		Spec:        mustSpec("balance ≥ 0"),
	}

	derivationPolicy := model_logic.Logic{
		Key:         attrDerivKey,
		Type:        model_logic.LogicTypeValue,
		Description: "derivation",
		Spec:        mustSpec("0"),
	}

	guardLogic := model_logic.Logic{
		Key:         guardKey,
		Type:        model_logic.LogicTypeAssessment,
		Description: "guard logic",
		Spec:        mustSpec("balance > 0"),
	}

	actionRequire := model_logic.Logic{
		Key:         actionReqKey,
		Type:        model_logic.LogicTypeAssessment,
		Description: "require",
		Spec:        mustSpec("amount > 0"),
	}

	actionGuarantee := model_logic.Logic{
		Key:         actionGuarKey,
		Type:        model_logic.LogicTypeStateChange,
		Description: "guarantee",
		Target:      "balance",
		Spec:        mustSpec("balance + amount"),
	}

	actionSafety := model_logic.Logic{
		Key:         actionSafeKey,
		Type:        model_logic.LogicTypeSafetyRule,
		Description: "safety",
		Spec:        mustSpec("balance' ≥ 0"),
	}

	queryRequire := model_logic.Logic{
		Key:         queryReqKey,
		Type:        model_logic.LogicTypeAssessment,
		Description: "query require",
		Spec:        mustSpec("TRUE"),
	}

	queryGuarantee := model_logic.Logic{
		Key:         queryGuarKey,
		Type:        model_logic.LogicTypeQuery,
		Description: "query guarantee",
		Target:      "result",
		Spec:        mustSpec("balance"),
	}

	globalFuncLogic := model_logic.Logic{
		Key:         globalFuncKey,
		Type:        model_logic.LogicTypeValue,
		Description: "max function",
		Spec:        mustSpec("IF x > y THEN x ELSE y"),
	}

	// Build model.
	model := &core.Model{
		Key:        "test",
		Name:       "Test",
		Invariants: []model_logic.Logic{modelInvariant},
		GlobalFunctions: map[identity.Key]model_logic.GlobalFunction{
			globalFuncKey: {
				Key:        globalFuncKey,
				Name:       "_Max",
				Parameters: []string{"x", "y"},
				Logic:      globalFuncLogic,
			},
		},
		NamedSets: map[identity.Key]model_logic.NamedSet{
			namedSetKey: {
				Key:  namedSetKey,
				Name: "valid_statuses",
				Spec: mustSpec(`{"active", "inactive"}`),
			},
		},
		Domains: map[identity.Key]model_domain.Domain{
			domainKey: {
				Key:  domainKey,
				Name: "d",
				Subdomains: map[identity.Key]model_domain.Subdomain{
					subKey: {
						Key:  subKey,
						Name: "s",
						Classes: map[identity.Key]model_class.Class{
							classKey: {
								Key:        classKey,
								Name:       "Account",
								Invariants: []model_logic.Logic{classInvariant},
								Attributes: map[identity.Key]model_class.Attribute{
									attrKey: {
										Key:              attrKey,
										Name:             "balance",
										DerivationPolicy: &derivationPolicy,
										Invariants:       []model_logic.Logic{attrInvariant},
									},
								},
								Guards: map[identity.Key]model_state.Guard{
									guardKey: {
										Key:   guardKey,
										Name:  "HasFunds",
										Logic: guardLogic,
									},
								},
								Actions: map[identity.Key]model_state.Action{
									actionKey: {
										Key:  actionKey,
										Name: "Deposit",
										Parameters: []model_state.Parameter{
											{Name: "amount", DataTypeRules: "Int"},
										},
										Requires:    []model_logic.Logic{actionRequire},
										Guarantees:  []model_logic.Logic{actionGuarantee},
										SafetyRules: []model_logic.Logic{actionSafety},
									},
								},
								Queries: map[identity.Key]model_state.Query{
									queryKey: {
										Key:        queryKey,
										Name:       "GetBalance",
										Requires:   []model_logic.Logic{queryRequire},
										Guarantees: []model_logic.Logic{queryGuarantee},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return model
}

func (s *LowerModelTestSuite) TestLowerModelSuccess() {
	model := buildTestModel()

	err := LowerModel(model)
	s.Require().NoError(err)

	// Verify model invariant was lowered.
	s.NotNil(model.Invariants[0].Spec.Expression)
	_, isBool := model.Invariants[0].Spec.Expression.(*logic_expression.BoolLiteral)
	s.True(isBool, "model invariant should be BoolLiteral")

	// Verify global function was lowered.
	for _, gf := range model.GlobalFunctions {
		s.NotNil(gf.Logic.Spec.Expression, "global function should be lowered")
		_, isITE := gf.Logic.Spec.Expression.(*logic_expression.IfThenElse)
		s.True(isITE, "global function should be IfThenElse")
	}

	// Verify named set was lowered.
	for _, ns := range model.NamedSets {
		s.NotNil(ns.Spec.Expression, "named set should be lowered")
		_, isSet := ns.Spec.Expression.(*logic_expression.SetLiteral)
		s.True(isSet, "named set should be SetLiteral")
	}

	// Navigate to the class.
	var class model_class.Class
	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			for _, c := range subdomain.Classes {
				class = c
			}
		}
	}

	// Verify class invariant was lowered.
	s.NotNil(class.Invariants[0].Spec.Expression, "class invariant should be lowered")

	// Verify attribute derivation policy was lowered.
	for _, attr := range class.Attributes {
		s.NotNil(attr.DerivationPolicy.Spec.Expression, "derivation policy should be lowered")
		s.NotNil(attr.Invariants[0].Spec.Expression, "attribute invariant should be lowered")
	}

	// Verify guard was lowered.
	for _, guard := range class.Guards {
		s.NotNil(guard.Logic.Spec.Expression, "guard should be lowered")
	}

	// Verify action requires/guarantees/safety were lowered.
	for _, action := range class.Actions {
		s.NotNil(action.Requires[0].Spec.Expression, "action require should be lowered")
		s.NotNil(action.Guarantees[0].Spec.Expression, "action guarantee should be lowered")
		s.NotNil(action.SafetyRules[0].Spec.Expression, "action safety rule should be lowered")
	}

	// Verify query requires/guarantees were lowered.
	for _, query := range class.Queries {
		s.NotNil(query.Requires[0].Spec.Expression, "query require should be lowered")
		s.NotNil(query.Guarantees[0].Spec.Expression, "query guarantee should be lowered")
	}
}

func (s *LowerModelTestSuite) TestLowerModelSkipsAlreadyLowered() {
	model := buildTestModel()

	// Lower once.
	err := LowerModel(model)
	s.Require().NoError(err)

	// Capture the expression pointer.
	expr := model.Invariants[0].Spec.Expression

	// Lower again — should skip already-lowered specs.
	err = LowerModel(model)
	s.Require().NoError(err)

	// The expression should be the same pointer (not re-lowered).
	s.Equal(expr, model.Invariants[0].Spec.Expression)
}

func (s *LowerModelTestSuite) TestLowerModelSkipsEmptySpec() {
	model := buildTestModel()

	// Set a spec with empty specification — should be skipped.
	model.Invariants[0].Spec.Specification = ""

	err := LowerModel(model)
	s.Require().NoError(err)

	// Expression should remain nil.
	s.Nil(model.Invariants[0].Spec.Expression)
}

func (s *LowerModelTestSuite) TestLowerModelActionParameterScope() {
	model := buildTestModel()

	err := LowerModel(model)
	s.Require().NoError(err)

	// The action require "amount > 0" should resolve 'amount' as a LocalVar.
	for _, action := range getClassFromModel(model).Actions {
		require := action.Requires[0].Spec.Expression
		cmp, ok := require.(*logic_expression.Compare)
		s.True(ok)
		lv, ok := cmp.Left.(*logic_expression.LocalVar)
		s.True(ok)
		s.Equal("amount", lv.Name)
	}
}

func (s *LowerModelTestSuite) TestLowerModelAttributeResolves() {
	model := buildTestModel()

	err := LowerModel(model)
	s.Require().NoError(err)

	// The class invariant "balance ≥ 0" should resolve 'balance' as an AttributeRef.
	class := getClassFromModel(model)
	inv := class.Invariants[0].Spec.Expression
	cmp, ok := inv.(*logic_expression.Compare)
	s.True(ok)
	_, isAttr := cmp.Left.(*logic_expression.AttributeRef)
	s.True(isAttr, "balance should resolve to AttributeRef")
}

func (s *LowerModelTestSuite) TestLowerModelAllExpressionsValidate() {
	model := buildTestModel()

	err := LowerModel(model)
	s.Require().NoError(err)

	// Validate every lowered expression.
	for _, inv := range model.Invariants {
		if inv.Spec.Expression != nil {
			s.Require().NoError(inv.Spec.Expression.Validate())
		}
	}
	for _, gf := range model.GlobalFunctions {
		if gf.Logic.Spec.Expression != nil {
			s.Require().NoError(gf.Logic.Spec.Expression.Validate())
		}
	}
	for _, ns := range model.NamedSets {
		if ns.Spec.Expression != nil {
			s.Require().NoError(ns.Spec.Expression.Validate())
		}
	}

	class := getClassFromModel(model)
	for _, inv := range class.Invariants {
		if inv.Spec.Expression != nil {
			s.Require().NoError(inv.Spec.Expression.Validate())
		}
	}
	for _, attr := range class.Attributes {
		if attr.DerivationPolicy != nil && attr.DerivationPolicy.Spec.Expression != nil {
			s.Require().NoError(attr.DerivationPolicy.Spec.Expression.Validate())
		}
		for _, inv := range attr.Invariants {
			if inv.Spec.Expression != nil {
				s.Require().NoError(inv.Spec.Expression.Validate())
			}
		}
	}
	for _, guard := range class.Guards {
		if guard.Logic.Spec.Expression != nil {
			s.Require().NoError(guard.Logic.Spec.Expression.Validate())
		}
	}
	for _, action := range class.Actions {
		for _, r := range action.Requires {
			if r.Spec.Expression != nil {
				s.Require().NoError(r.Spec.Expression.Validate())
			}
		}
		for _, g := range action.Guarantees {
			if g.Spec.Expression != nil {
				s.Require().NoError(g.Spec.Expression.Validate())
			}
		}
		for _, sr := range action.SafetyRules {
			if sr.Spec.Expression != nil {
				s.Require().NoError(sr.Spec.Expression.Validate())
			}
		}
	}
	for _, query := range class.Queries {
		for _, r := range query.Requires {
			if r.Spec.Expression != nil {
				s.Require().NoError(r.Spec.Expression.Validate())
			}
		}
		for _, g := range query.Guarantees {
			if g.Spec.Expression != nil {
				s.Require().NoError(g.Spec.Expression.Validate())
			}
		}
	}
}

// getClassFromModel navigates the model tree to get the single class.
func getClassFromModel(model *core.Model) model_class.Class {
	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			for _, class := range subdomain.Classes {
				return class
			}
		}
	}
	panic("no class found in model")
}

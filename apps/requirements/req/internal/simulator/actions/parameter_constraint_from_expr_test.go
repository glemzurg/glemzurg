package actions

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/convert"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/stretchr/testify/require"
)

func TestExtractNullableElseBooleanConstantConstraint(t *testing.T) {
	classKey := mustKey("domain/finance/wallet/class/jurisdiction")
	actionKey := helper.Must(identity.NewActionKey(classKey, "update"))
	ctx := &convert.LowerContext{
		ClassKey: classKey,
		Parameters: map[string]bool{
			"JurisdictionCode": true,
			"SocialOnly":       true,
		},
	}
	pf := convert.NewExpressionParseFunc(ctx)
	logic := model_logic.NewLogic(
		helper.Must(identity.NewActionRequireKey(actionKey, "social_only")),
		model_logic.LogicTypeAssessment,
		"If no jurisdiction code then social only.",
		"",
		helper.Must(logic_spec.NewExpressionSpec(
			"tla_plus",
			`IF JurisdictionCode = NULL THEN SocialOnly = TRUE ELSE TRUE`,
			pf,
		)),
		nil,
	)

	constraints := extractParameterConstraints([]model_logic.Logic{logic})
	require.NotNil(t, constraints.nullableElseBooleanConstant)
	require.Equal(t, "JurisdictionCode", constraints.nullableElseBooleanConstant.driverParam)
	require.Equal(t, "SocialOnly", constraints.nullableElseBooleanConstant.followerParam)
	require.True(t, constraints.nullableElseBooleanConstant.value)
}

func TestApplyNullableElseBooleanConstantSetsSocialOnlyWhenCodeNull(t *testing.T) {
	result := map[string]object.Object{
		"JurisdictionCode": object.NewSet(),
		"SocialOnly":       object.NewBoolean(false),
	}
	applyNullableElseBooleanConstant(result, &nullableElseBooleanConstantConstraint{
		driverParam:   "JurisdictionCode",
		followerParam: "SocialOnly",
		value:         true,
	})

	boolVal, ok := result["SocialOnly"].(*object.Boolean)
	require.True(t, ok)
	require.True(t, boolVal.Value())
}

func TestExtractParamInNamedSetMinusPeerField(t *testing.T) {
	classKey := mustKey("domain/finance/wallet/class/jurisdiction")
	actionKey := helper.Must(identity.NewActionKey(classKey, "add"))
	jurisdictionCodesKey := helper.Must(identity.NewNamedSetKey("jurisdictioncodes"))
	ctx := &convert.LowerContext{
		ClassKey: classKey,
		Parameters: map[string]bool{
			"JurisdictionCode": true,
		},
		ClassNames: map[string]identity.Key{
			"Jurisdiction": classKey,
		},
		NamedSets: map[string]identity.Key{
			"_JurisdictionCodes": jurisdictionCodesKey,
		},
	}
	pf := convert.NewExpressionParseFunc(ctx)
	logic := model_logic.NewLogic(
		helper.Must(identity.NewActionRequireKey(actionKey, "0")),
		model_logic.LogicTypeAssessment,
		"Unused allowed jurisdiction code.",
		"",
		helper.Must(logic_spec.NewExpressionSpec(
			"tla_plus",
			`JurisdictionCode \in (_JurisdictionCodes \ { j.jurisdiction_code : j \in Jurisdiction })`,
			pf,
		)),
		nil,
	)

	constraints := extractParameterConstraints([]model_logic.Logic{logic})
	require.NotNil(t, constraints.paramInNamedSetMinusPeerField)
	require.Equal(t, "JurisdictionCode", constraints.paramInNamedSetMinusPeerField.paramName)
	require.Equal(t, "jurisdictioncodes", constraints.paramInNamedSetMinusPeerField.setSubKey)
	require.Equal(t, "jurisdiction_code", constraints.paramInNamedSetMinusPeerField.fieldSubKey)
	require.Equal(t, classKey, constraints.paramInNamedSetMinusPeerField.classKey)
	require.True(t, membershipSupportsParamSampling(logic.Spec.Expression.(*me.Membership)))
}

func TestExtractPlainParamInNamedSet(t *testing.T) {
	classKey := mustKey("domain/finance/wallet/class/currency")
	actionKey := helper.Must(identity.NewActionKey(classKey, "add"))
	isoKey := helper.Must(identity.NewNamedSetKey("iso4217codes"))
	ctx := &convert.LowerContext{
		ClassKey:   classKey,
		Parameters: map[string]bool{"ISO": true},
		NamedSets:  map[string]identity.Key{"_Iso4217Codes": isoKey},
	}
	pf := convert.NewExpressionParseFunc(ctx)
	logic := model_logic.NewLogic(
		helper.Must(identity.NewActionRequireKey(actionKey, "0")),
		model_logic.LogicTypeAssessment,
		"ISO from set.",
		"",
		helper.Must(logic_spec.NewExpressionSpec("tla_plus", `ISO \in _Iso4217Codes`, pf)),
		nil,
	)

	constraints := extractParameterConstraints([]model_logic.Logic{logic})
	require.NotNil(t, constraints.paramInNamedSet)
	require.Equal(t, "ISO", constraints.paramInNamedSet.paramName)
	require.Equal(t, "iso4217codes", constraints.paramInNamedSet.setSubKey)
}

package actions

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
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

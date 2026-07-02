package actions

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/stretchr/testify/require"
)

func booleanEnumParameter(actionKey identity.Key, name string) model_state.Parameter {
	param := helper.Must(model_state.NewParameter(actionKey, name, "enum of TRUE, FALSE", false))
	boolTypeSpec := helper.Must(logic_spec.NewTypeSpec(model_logic.NotationTLAPlus, "BOOLEAN", nil))
	param.DataType.TypeSpec = &boolTypeSpec
	return param
}

func TestCoerceValueForDataTypeStoresBooleanEnumAsBoolean(t *testing.T) {
	classKey := mustKey("domain/finance/wallet/class/jurisdiction")
	actionKey := helper.Must(identity.NewActionKey(classKey, "add"))
	param := booleanEnumParameter(actionKey, "SocialOnly")

	coerced := CoerceValueForDataType(param.DataType, object.NewString("TRUE"))
	boolVal, ok := coerced.(*object.Boolean)
	require.True(t, ok)
	require.True(t, boolVal.Value())

	coerced = CoerceValueForDataType(param.DataType, object.NewString("FALSE"))
	boolVal, ok = coerced.(*object.Boolean)
	require.True(t, ok)
	require.False(t, boolVal.Value())
}

func TestEnumMembershipSpecificationUsesBooleanSetForBooleanTypeSpec(t *testing.T) {
	classKey := mustKey("domain/finance/wallet/class/jurisdiction")
	actionKey := helper.Must(identity.NewActionKey(classKey, "add"))
	param := booleanEnumParameter(actionKey, "SocialOnly")

	spec := enumMembershipSpecification(
		param.Name,
		param.DataType,
		[]string{"TRUE", "FALSE"},
		false,
	)
	require.Equal(t, `SocialOnly \in BOOLEAN`, spec)
}

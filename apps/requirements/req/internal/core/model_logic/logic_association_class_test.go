package model_logic

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/require"
)

func actionGuaranteeKey(t *testing.T) identity.Key {
	t.Helper()
	return helper.Must(identity.NewActionGuaranteeKey(
		helper.Must(identity.NewActionKey(
			helper.Must(identity.NewClassKey(
				helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("d")), "s")),
				"c",
			)),
			"a",
		)),
		"0",
	))
}

func TestValidateAssociationClassReifyValid(t *testing.T) {
	tests := []struct {
		name string
		spec string
	}{
		{name: "singleton new", spec: "_new(Amount)"},
		{name: "set-map new", spec: `{ _new(r.amount) : r \in Amounts }`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			logic := NewLogic(
				actionGuaranteeKey(t),
				LogicTypeStateChange,
				"Create AC rows",
				"AccountBalanceChange",
				logic_spec.ExpressionSpec{Notation: NotationTLAPlus, Specification: tc.spec},
				nil,
			)
			logic.SetEndpointSelectorSpec(logic_spec.ExpressionSpec{Notation: NotationTLAPlus, Specification: "r.account"})
			require.NoError(t, logic.Validate(coreerr.NewContext("test", "")))
			require.True(t, IsAssociationClassReify(logic))
		})
	}
}

func TestValidateEndpointSelectorRequiresActionGuarantee(t *testing.T) {
	queryKey := helper.Must(identity.NewQueryKey(
		helper.Must(identity.NewClassKey(
			helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("d")), "s")),
			"c",
		)),
		"q",
	))
	guarKey := helper.Must(identity.NewQueryGuaranteeKey(queryKey, "0"))
	logic := NewLogic(
		guarKey,
		LogicTypeStateChange,
		"Bad",
		"AccountBalanceChange",
		logic_spec.ExpressionSpec{Notation: NotationTLAPlus, Specification: "_new(Amount)"},
		nil,
	)
	logic.SetEndpointSelectorSpec(logic_spec.ExpressionSpec{Notation: NotationTLAPlus, Specification: "acct"})
	err := logic.Validate(coreerr.NewContext("test", ""))
	require.Error(t, err)
	require.Contains(t, err.Error(), "action guarantees")
}

func TestValidatePlainStateChangeWithoutEndpointSelector(t *testing.T) {
	logic := NewLogic(
		actionGuaranteeKey(t),
		LogicTypeStateChange,
		"Plain attr",
		"timestamp",
		logic_spec.ExpressionSpec{Notation: NotationTLAPlus, Specification: "Timestamp"},
		nil,
	)
	require.NoError(t, logic.Validate(coreerr.NewContext("test", "")))
	require.False(t, IsAssociationClassReify(logic))
}

package model_class

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/require"
)

func TestComputedSimpleActionGuaranteeDescription(t *testing.T) {
	classKey := helper.Must(identity.NewClassKey(
		helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("finance")), "wallet")),
		"jurisdiction",
	))
	nameKey := helper.Must(identity.NewAttributeKey(classKey, "name"))
	socialKey := helper.Must(identity.NewAttributeKey(classKey, "social_only"))

	attributes := map[identity.Key]Attribute{
		nameKey: {
			Key:  nameKey,
			Name: "Display Name",
		},
		socialKey: {
			Key:  socialKey,
			Name: "Is Social Only",
		},
	}

	guaranteeKey := helper.Must(identity.NewActionGuaranteeKey(
		helper.Must(identity.NewActionKey(classKey, "add")),
		"0",
	))

	tests := []struct {
		name      string
		guarantee model_logic.Logic
		wantDesc  string
		wantOK    bool
	}{
		{
			name: "simple assignment",
			guarantee: model_logic.Logic{
				Key:    guaranteeKey,
				Type:   model_logic.LogicTypeStateChange,
				Target: "name",
				Spec:   logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "Name"},
			},
			wantDesc: "Set Display Name",
			wantOK:   true,
		},
		{
			name: "second attribute",
			guarantee: model_logic.Logic{
				Key:    guaranteeKey,
				Type:   model_logic.LogicTypeStateChange,
				Target: "social_only",
				Spec:   logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "SocialOnly"},
			},
			wantDesc: "Set Is Social Only",
			wantOK:   true,
		},
		{
			name: "multi word spec",
			guarantee: model_logic.Logic{
				Key:    guaranteeKey,
				Type:   model_logic.LogicTypeStateChange,
				Target: "name",
				Spec:   logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "self.name' = Name"},
			},
			wantOK: false,
		},
		{
			name: "unknown target",
			guarantee: model_logic.Logic{
				Key:    guaranteeKey,
				Type:   model_logic.LogicTypeStateChange,
				Target: "missing",
				Spec:   logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "Name"},
			},
			wantOK: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			desc, ok := ComputedSimpleActionGuaranteeDescription(tc.guarantee, attributes)
			require.Equal(t, tc.wantOK, ok)
			if ok {
				require.Equal(t, tc.wantDesc, desc)
			}
		})
	}
}

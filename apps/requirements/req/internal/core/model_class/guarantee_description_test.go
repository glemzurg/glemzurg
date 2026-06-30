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

	attributes := []Attribute{
		{Key: nameKey, Name: "Display Name"},
		{Key: socialKey, Name: "Is Social Only"},
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

func TestComputedAssociationDestroyGuaranteeDescription(t *testing.T) {
	subdomainKey := helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("d")), "s"))
	fromKey := helper.Must(identity.NewClassKey(subdomainKey, "wallet"))
	toKey := helper.Must(identity.NewClassKey(subdomainKey, "behavior"))
	actionKey := helper.Must(identity.NewActionKey(fromKey, "remove"))
	guaranteeKey := helper.Must(identity.NewActionGuaranteeKey(actionKey, "0"))

	assoc := NewAssociation(
		helper.Must(identity.NewClassAssociationKey(subdomainKey, fromKey, toKey, "applies_social")),
		AssociationDetails{Name: "Applies Social Currency Logic", Details: ""},
		AssociationEnd{ClassKey: fromKey, Multiplicity: helper.Must(NewMultiplicity("1"))},
		AssociationEnd{ClassKey: toKey, Multiplicity: helper.Must(NewMultiplicity("0..1"))},
		Multiplicity{},
		AssociationOptions{},
	)

	guarantee := model_logic.Logic{
		Key:    guaranteeKey,
		Type:   model_logic.LogicTypeDelete,
		Target: "AppliesSocialCurrencyLogic",
		Spec: logic_spec.ExpressionSpec{
			Notation:      model_logic.NotationTLAPlus,
			Specification: `{ b \in AppliesSocialCurrencyLogic : TRUE }`,
		},
	}

	desc, ok := ComputedAssociationDestroyGuaranteeDescription(guarantee, map[identity.Key]Association{assoc.Key: assoc})
	require.True(t, ok)
	require.Equal(t, "Set Applies Social Currency Logic", desc)
}

func TestComputedAssociationSetAddGuaranteeDescription(t *testing.T) {
	subdomainKey := helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("d")), "s"))
	fromKey := helper.Must(identity.NewClassKey(subdomainKey, "container"))
	toKey := helper.Must(identity.NewClassKey(subdomainKey, "part"))
	actionKey := helper.Must(identity.NewActionKey(fromKey, "split"))
	guaranteeKey := helper.Must(identity.NewActionGuaranteeKey(actionKey, "0"))

	assoc := NewAssociation(
		helper.Must(identity.NewClassAssociationKey(subdomainKey, fromKey, toKey, "is_subdivided_into")),
		AssociationDetails{Name: "Is Subdivided Into", Details: ""},
		AssociationEnd{ClassKey: fromKey, Multiplicity: helper.Must(NewMultiplicity("1"))},
		AssociationEnd{ClassKey: toKey, Multiplicity: helper.Must(NewMultiplicity("any"))},
		Multiplicity{},
		AssociationOptions{},
	)

	guarantee := model_logic.Logic{
		Key:    guaranteeKey,
		Type:   model_logic.LogicTypeStateChange,
		Target: "IsSubdividedInto",
		Spec: logic_spec.ExpressionSpec{
			Notation:      model_logic.NotationTLAPlus,
			Specification: `IsSubdividedInto \union {_new(PartId)}`,
		},
	}

	desc, ok := ComputedAssociationSetAddGuaranteeDescription(guarantee, map[identity.Key]Association{assoc.Key: assoc})
	require.True(t, ok)
	require.Equal(t, "Add to Is Subdivided Into", desc)
}

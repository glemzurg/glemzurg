package model_class_test

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/require"
)

func TestMatchAssociationDeleteGuarantee(t *testing.T) {
	assocKey := helper.Must(identity.NewClassAssociationKey(
		helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("d")), "s")),
		helper.Must(identity.NewClassKey(helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("d")), "s")), "from")),
		helper.Must(identity.NewClassKey(helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("d")), "s")), "to")),
		"assoc",
	))
	eventKey := helper.Must(identity.NewEventKey(
		helper.Must(identity.NewClassKey(helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("d")), "s")), "to")),
		"_delete",
	))

	logic := model_logic.NewLogic(
		identity.Key{},
		model_logic.LogicTypeDelete,
		"Remove peers",
		"AssocField",
		logic_spec.ExpressionSpec{
			Expression: &me.SetFilter{
				Variable:  "b",
				Set:       &me.AssociationRef{AssociationKey: assocKey},
				Predicate: &me.BoolLiteral{Value: true},
			},
		},
		nil,
	)
	logic.SetDeleteEventSpec(logic_spec.ExpressionSpec{
		Expression: &me.EventCall{
			EventKey: eventKey,
			Args:     []me.Expression{&me.LocalVar{Name: "b"}},
		},
	})

	assocRef, selection, eventCall, ok := model_class.MatchAssociationDeleteGuarantee(logic)
	require.True(t, ok)
	require.Equal(t, assocKey, assocRef.AssociationKey)
	require.Equal(t, "b", selection.Variable)
	require.Equal(t, eventKey, eventCall.EventKey)

	eventKeyOut, ok := model_class.AssociationDeleteEventKey(logic)
	require.True(t, ok)
	require.Equal(t, eventKey, eventKeyOut)
	require.False(t, model_class.DeleteGuaranteeHasInlineStateChange(logic))
}

func TestMatchAssociationDeleteGuaranteeInlineStateChange(t *testing.T) {
	assocKey := helper.Must(identity.NewClassAssociationKey(
		helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("d")), "s")),
		helper.Must(identity.NewClassKey(helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("d")), "s")), "from")),
		helper.Must(identity.NewClassKey(helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("d")), "s")), "to")),
		"assoc",
	))
	eventKey := helper.Must(identity.NewEventKey(
		helper.Must(identity.NewClassKey(helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("d")), "s")), "to")),
		"_delete",
	))
	selection := &me.SetFilter{
		Variable:  "b",
		Set:       &me.AssociationRef{AssociationKey: assocKey},
		Predicate: &me.BoolLiteral{Value: true},
	}
	logic := model_logic.NewLogic(
		identity.Key{},
		model_logic.LogicTypeDelete,
		"Remove peers",
		"AssocField",
		logic_spec.ExpressionSpec{
			Expression: &me.SetOp{
				Op:    me.SetDifference,
				Left:  &me.AssociationRef{AssociationKey: assocKey},
				Right: selection,
			},
		},
		nil,
	)
	logic.SetDeleteEventSpec(logic_spec.ExpressionSpec{
		Expression: &me.EventCall{
			EventKey: eventKey,
			Args:     []me.Expression{&me.LocalVar{Name: "item"}},
		},
	})

	assocRef, matchedSelection, eventCall, ok := model_class.MatchAssociationDeleteGuarantee(logic)
	require.True(t, ok)
	require.True(t, model_class.DeleteGuaranteeHasInlineStateChange(logic))
	require.Equal(t, assocKey, assocRef.AssociationKey)
	require.Equal(t, "b", matchedSelection.Variable)
	require.Equal(t, eventKey, eventCall.EventKey)
}

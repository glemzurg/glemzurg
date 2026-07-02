package model_class_test

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/require"
)

func TestMatchAssociationSetAddExpr(t *testing.T) {
	assocKey := helper.Must(identity.NewClassAssociationKey(
		helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("d")), "s")),
		helper.Must(identity.NewClassKey(helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("d")), "s")), "from")),
		helper.Must(identity.NewClassKey(helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("d")), "s")), "to")),
		"assoc",
	))
	eventKey := helper.Must(identity.NewEventKey(
		helper.Must(identity.NewClassKey(helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("d")), "s")), "from")),
		"_new",
	))

	expr := &me.SetOp{
		Op:   me.SetUnion,
		Left: &me.AssociationRef{AssociationKey: assocKey},
		Right: &me.SetLiteral{
			Elements: []me.Expression{
				&me.EventCall{
					EventKey: eventKey,
					Args: []me.Expression{
						&me.LocalVar{Name: "MinimumBalance"},
						&me.LocalVar{Name: "TopoffBalance"},
					},
				},
			},
		},
	}

	assocRef, eventCall, ok := model_class.MatchAssociationSetAddExpr(expr)
	require.True(t, ok)
	require.Equal(t, assocKey, assocRef.AssociationKey)
	require.Equal(t, eventKey, eventCall.EventKey)
	require.Len(t, eventCall.Args, 2)
}

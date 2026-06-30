package convert_test

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/convert"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/parser"
	"github.com/stretchr/testify/require"
)

func TestAssociationDestroyGuaranteeSelectionTLALower(t *testing.T) {
	ctx := associationSetMapDeleteFixture()
	selectionSpec := `{ b \in AppliesSocialCurrencyLogic : TRUE }`

	selectionAST, err := parser.ParseExpression(selectionSpec)
	require.NoError(t, err)
	selectionLowered, err := convert.Lower(selectionAST, ctx)
	require.NoError(t, err)

	selection, ok := selectionLowered.(*me.SetFilter)
	require.True(t, ok)
	require.Equal(t, "b", selection.Variable)

	eventDeleteKey := selection.Set.(*me.AssociationRef).AssociationKey
	logic := model_logic.NewLogic(
		identity.Key{},
		model_logic.LogicTypeDestroy,
		"Remove peers",
		"AppliesSocialCurrencyLogic",
		logic_spec.ExpressionSpec{Expression: selectionLowered},
		nil,
	)
	logic.SetDestroyEventSpec(logic_spec.ExpressionSpec{
		Expression: &me.EventCall{
			EventKey: identity.Key{SubKey: model_state.EventNameDestroy},
			Args:     []me.Expression{&me.LocalVar{Name: "b"}},
		},
	})

	assocRef, matchedSelection, eventCall, ok := model_class.MatchAssociationDestroyGuarantee(logic)
	require.True(t, ok)
	require.Equal(t, eventDeleteKey, assocRef.AssociationKey)
	require.Equal(t, "b", matchedSelection.Variable)
	require.Equal(t, model_state.EventNameDestroy, eventCall.EventKey.SubKey)
}

func TestAssociationDestroyGuaranteeInlineDifferenceTLALower(t *testing.T) {
	ctx := associationSetMapDeleteFixture()
	spec := `AppliesSocialCurrencyLogic \ { b \in AppliesSocialCurrencyLogic : TRUE }`

	astExpr, err := parser.ParseExpression(spec)
	require.NoError(t, err)
	lowered, err := convert.Lower(astExpr, ctx)
	require.NoError(t, err)

	logic := model_logic.NewLogic(
		identity.Key{},
		model_logic.LogicTypeDestroy,
		"Remove peers",
		"AppliesSocialCurrencyLogic",
		logic_spec.ExpressionSpec{Expression: lowered},
		nil,
	)
	logic.SetDestroyEventSpec(logic_spec.ExpressionSpec{
		Expression: &me.EventCall{
			EventKey: identity.Key{SubKey: model_state.EventNameDestroy},
			Args:     []me.Expression{&me.LocalVar{Name: "item"}},
		},
	})

	_, matchedSelection, eventCall, ok := model_class.MatchAssociationDestroyGuarantee(logic)
	require.True(t, ok)
	require.True(t, model_class.DestroyGuaranteeHasInlineStateChange(logic))
	require.Equal(t, "b", matchedSelection.Variable)
	require.Equal(t, model_state.EventNameDestroy, eventCall.EventKey.SubKey)
	require.Equal(t, "item", eventCall.Args[0].(*me.LocalVar).Name)
}

package convert_test

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/convert"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/parser"
	"github.com/stretchr/testify/require"
)

func TestAssociationSetMapGuaranteeTLARoundTrip(t *testing.T) {
	ctx := associationSetMapFixture()
	spec := `{Update(r) : r \in AppliesSocialCurrencyLogic}`

	astExpr, err := parser.ParseExpression(spec)
	require.NoError(t, err)
	_, ok := astExpr.(*ast.SetMap)
	require.True(t, ok, "expected SetMap AST, got %T", astExpr)

	lowered, err := convert.Lower(astExpr, ctx)
	require.NoError(t, err)

	setMap, ok := lowered.(*me.SetMap)
	require.True(t, ok)
	_, eventCall, ok := model_class.MatchAssociationSetMapExpr(setMap)
	require.True(t, ok)
	require.NotEmpty(t, eventCall.EventKey)

	raised, err := convert.Raise(lowered, raiseContextForAssociationSetMap(ctx))
	require.NoError(t, err)

	printed := ast.Print(raised)
	require.Contains(t, printed, "Update")
	require.Contains(t, printed, "AppliesSocialCurrencyLogic")
	require.Contains(t, printed, "∈")
}

func associationSetMapFixture() *convert.LowerContext {
	subdomainKey := helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("d")), "s"))
	fromKey := helper.Must(identity.NewClassKey(subdomainKey, "currency_wallet_definition"))
	toKey := helper.Must(identity.NewClassKey(subdomainKey, "social_currency_behavior"))
	assocKey := helper.Must(identity.NewClassAssociationKey(subdomainKey, fromKey, toKey, "applies_social"))
	eventUpdateKey := helper.Must(identity.NewEventKey(toKey, "Update"))

	assoc := model_class.NewAssociation(
		assocKey,
		model_class.AssociationDetails{Name: "Applies Social Currency Logic", Details: ""},
		model_class.AssociationEnd{ClassKey: fromKey, Multiplicity: helper.Must(model_class.NewMultiplicity("1"))},
		model_class.AssociationEnd{ClassKey: toKey, Multiplicity: helper.Must(model_class.NewMultiplicity("0..1"))},
		model_class.Multiplicity{},
		model_class.AssociationOptions{},
	)

	eventNewFromKey := helper.Must(identity.NewEventKey(fromKey, "_new"))
	fromClass := model_class.NewClass(fromKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Currency Wallet Definition"})
	fromClass.SetEvents(map[identity.Key]model_state.Event{
		eventNewFromKey: model_state.NewEvent(eventNewFromKey, "_new", "", []string{"MinimumBalance", "TopoffBalance"}),
	})
	peerClass := model_class.NewClass(toKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Social Currency Behavior"})
	peerClass.SetEvents(map[identity.Key]model_state.Event{
		eventUpdateKey: model_state.NewEvent(eventUpdateKey, "Update", "", []string{"MinimumBalance", "TopoffBalance"}),
	})

	classes := map[identity.Key]model_class.Class{fromKey: fromClass, toKey: peerClass}
	associations := map[identity.Key]model_class.Association{assocKey: assoc}
	ctx := &convert.LowerContext{
		ClassKey:         fromKey,
		AssociationNames: convert.BuildOutgoingAssociationFieldNameMap(fromKey, associations),
		SystemEventNames: convert.BuildSystemEventNameMap(&fromClass),
		PeerEventNames:   convert.BuildPeerEventNameMap(fromKey, associations, classes),
		Parameters:       map[string]bool{"MinimumBalance": true, "TopoffBalance": true},
	}
	return ctx
}

func raiseContextForAssociationSetMap(ctx *convert.LowerContext) *convert.RaiseContext {
	return &convert.RaiseContext{
		AssociationNames: invertKeyStringMap(ctx.AssociationNames),
		PeerEventNames:   invertKeyStringMap(ctx.PeerEventNames),
	}
}

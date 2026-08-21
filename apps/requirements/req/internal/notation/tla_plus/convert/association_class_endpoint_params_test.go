package convert

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/require"
)

func TestAssociationClassEndpointParamNames(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		from, to string
		wantFrom string
		wantTo   string
	}{
		{
			name:     "distinct class display names",
			from:     "Partner",
			to:       "Jurisdiction",
			wantFrom: "Partner",
			wantTo:   "Jurisdiction",
		},
		{
			name:     "spaces stripped to TLA form",
			from:     "Jurisdictional Wallet Definition",
			to:       "Currency",
			wantFrom: "JurisdictionalWalletDefinition",
			wantTo:   "Currency",
		},
		{
			name:     "same class both ends",
			from:     "Account",
			to:       "Account",
			wantFrom: "FromAccount",
			wantTo:   "ToAccount",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			gotFrom, gotTo := associationClassEndpointParamNames(tc.from, tc.to)
			require.Equal(t, tc.wantFrom, gotFrom)
			require.Equal(t, tc.wantTo, gotTo)
		})
	}
}

func TestInjectAssociationClassEndpointParams(t *testing.T) {
	t.Parallel()

	parse := func(s string) identity.Key {
		k, err := identity.ParseKey(s)
		require.NoError(t, err)
		return k
	}

	domainKey := parse("domain/d")
	subKey := parse("domain/d/subdomain/s")
	partnerKey := parse("domain/d/subdomain/s/class/partner")
	jurisdictionKey := parse("domain/d/subdomain/s/class/jurisdiction")
	linkDefKey := parse("domain/d/subdomain/s/class/link_def")
	hostAssocKey := parse("domain/d/subdomain/s/cassociation/class/partner/class/jurisdiction/configures")

	partner := model_class.NewClass(partnerKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Partner"})
	jurisdiction := model_class.NewClass(jurisdictionKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Jurisdiction"})
	linkDef := model_class.NewClass(linkDefKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "LinkDef"})

	stateActive := parse("domain/d/subdomain/s/class/link_def/state/active")
	eventAdd := parse("domain/d/subdomain/s/class/link_def/event/add")
	actionInit := parse("domain/d/subdomain/s/class/link_def/action/initialize")
	transAdd := parse("domain/d/subdomain/s/class/link_def/transition/add")

	ev := model_state.NewEvent(eventAdd, "Add", "", nil)
	act := model_state.NewAction(actionInit, model_state.ActionDetails{Name: "Initialize", Details: ""}, nil, nil, nil, nil)
	linkDef.SetStates(map[identity.Key]model_state.State{
		stateActive: model_state.NewState(stateActive, "Active", "", ""),
	})
	linkDef.SetEvents(map[identity.Key]model_state.Event{eventAdd: ev})
	linkDef.SetActions(map[identity.Key]model_state.Action{actionInit: act})
	linkDef.SetTransitions(map[identity.Key]model_state.Transition{
		transAdd: model_state.NewTransition(
			transAdd, eventAdd,
			model_state.TransitionStateKeys{FromStateKey: nil, ToStateKey: &stateActive},
			model_state.TransitionLogicKeys{GuardKey: nil, ActionKey: &actionInit},
			"",
		),
	})

	fromMult := helper.Must(model_class.NewMultiplicity("1"))
	toMult := helper.Must(model_class.NewMultiplicity("any"))
	hostAssoc := model_class.NewAssociation(
		hostAssocKey,
		model_class.AssociationDetails{Name: "Configures", Details: ""},
		model_class.AssociationEnd{ClassKey: partnerKey, Multiplicity: fromMult},
		model_class.AssociationEnd{ClassKey: jurisdictionKey, Multiplicity: toMult},
		model_class.AssociationOptions{AssociationClassKey: &linkDefKey, UmlComment: ""},
	)

	sub := model_domain.NewSubdomain(subKey, "S", "", "", "")
	sub.Classes = map[identity.Key]model_class.Class{
		partnerKey: partner, jurisdictionKey: jurisdiction, linkDefKey: linkDef,
	}
	domain := model_domain.NewDomain(domainKey, "D", "", "", false, "")
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{subKey: sub}

	model := core.NewModel("m", core.ModelDetails{Name: "M", Details: ""}, "", nil, nil, nil)
	model.Domains = map[identity.Key]model_domain.Domain{domainKey: domain}
	require.NoError(t, model.SetClassAssociations(map[identity.Key]model_class.Association{hostAssocKey: hostAssoc}))

	require.NoError(t, injectAssociationClassEndpointParams(&model))

	got := model.Domains[domainKey].Subdomains[subKey].Classes[linkDefKey]
	gotEv := got.Events[eventAdd]
	require.Equal(t, []string{"Partner", "Jurisdiction"}, gotEv.ParameterNames)

	gotAct := got.Actions[actionInit]
	require.Len(t, gotAct.Parameters, 2)
	require.Equal(t, "Partner", gotAct.Parameters[0].Name)
	require.Equal(t, "Jurisdiction", gotAct.Parameters[1].Name)
}

package generate

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/generate/req_flat"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/require"
)

func TestGenerateClassMarkdownUsesScopedDisplayNames(t *testing.T) {
	t.Parallel()

	backofficeDomain := helper.Must(identity.NewDomainKey("backoffice"))
	backofficeDefault := helper.Must(identity.NewSubdomainKey(backofficeDomain, "default"))
	adminKey := helper.Must(identity.NewClassKey(backofficeDefault, "administrator"))
	adminClass := model_class.NewClass(adminKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Administrator", Details: "Configures leaderboards.", UnfinishedNotes: "", UmlComment: ""})

	platformDomain := helper.Must(identity.NewDomainKey("platform"))
	platformLeaderboards := helper.Must(identity.NewSubdomainKey(platformDomain, "leaderboards"))
	resolverKey := helper.Must(identity.NewClassKey(platformLeaderboards, "resolver"))
	resolverClass := model_class.NewClass(resolverKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Resolver", Details: "Leaderboard rules.", UnfinishedNotes: "", UmlComment: ""})

	assocKey := helper.Must(identity.NewClassAssociationKey(identity.Key{}, adminKey, resolverKey, "configures"))
	assoc := model_class.NewAssociation(
		assocKey,
		model_class.AssociationDetails{Name: "Configures", Details: ""},
		model_class.AssociationEnd{ClassKey: adminKey, Multiplicity: helper.Must(model_class.NewMultiplicity("1"))},
		model_class.AssociationEnd{ClassKey: resolverKey, Multiplicity: helper.Must(model_class.NewMultiplicity("any"))},
		model_class.AssociationOptions{},
	)

	model := core.Model{
		Key:  "evenplay",
		Name: "Evenplay",
		Domains: map[identity.Key]model_domain.Domain{
			backofficeDomain: {
				Key:  backofficeDomain,
				Name: "Backoffice",
				Subdomains: map[identity.Key]model_domain.Subdomain{
					backofficeDefault: {
						Key:     backofficeDefault,
						Name:    "Default",
						Classes: map[identity.Key]model_class.Class{adminKey: adminClass},
					},
				},
			},
			platformDomain: {
				Key:  platformDomain,
				Name: "Platform",
				Subdomains: map[identity.Key]model_domain.Subdomain{
					platformLeaderboards: {
						Key:     platformLeaderboards,
						Name:    "Leaderboards",
						Classes: map[identity.Key]model_class.Class{resolverKey: resolverClass},
					},
				},
			},
		},
		ClassAssociations: map[identity.Key]model_class.Association{assocKey: assoc},
	}

	reqs := req_flat.NewRequirements(model)
	_, diagramClasses, associations := reqs.RegardingClasses([]model_class.Class{adminClass})
	require.Len(t, associations, 1)

	md, err := generateClassMdContents(reqs, adminClass, diagramClasses, "", "")
	require.NoError(t, err)
	require.Contains(t, md, "Platform::Leaderboards::Resolver")
	require.NotContains(t, md, "[Resolver](")
}

func TestGenerateClassMarkdownOmitsDefaultSubdomainInScopedNames(t *testing.T) {
	t.Parallel()

	backofficeDomain := helper.Must(identity.NewDomainKey("backoffice"))
	backofficeDefault := helper.Must(identity.NewSubdomainKey(backofficeDomain, "default"))
	adminKey := helper.Must(identity.NewClassKey(backofficeDefault, "administrator"))
	adminClass := model_class.NewClass(adminKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Administrator", Details: "Configures leaderboards."})

	platformDomain := helper.Must(identity.NewDomainKey("platform"))
	platformLeaderboards := helper.Must(identity.NewSubdomainKey(platformDomain, "leaderboards"))
	resolverKey := helper.Must(identity.NewClassKey(platformLeaderboards, "resolver"))
	resolverClass := model_class.NewClass(resolverKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Resolver", Details: "Leaderboard rules."})

	assocKey := helper.Must(identity.NewClassAssociationKey(identity.Key{}, resolverKey, adminKey, "configured_by"))
	assoc := model_class.NewAssociation(
		assocKey,
		model_class.AssociationDetails{Name: "Configured By", Details: ""},
		model_class.AssociationEnd{ClassKey: resolverKey, Multiplicity: helper.Must(model_class.NewMultiplicity("1"))},
		model_class.AssociationEnd{ClassKey: adminKey, Multiplicity: helper.Must(model_class.NewMultiplicity("any"))},
		model_class.AssociationOptions{},
	)

	model := core.Model{
		Key:  "evenplay",
		Name: "Evenplay",
		Domains: map[identity.Key]model_domain.Domain{
			backofficeDomain: {
				Key:  backofficeDomain,
				Name: "Backoffice",
				Subdomains: map[identity.Key]model_domain.Subdomain{
					backofficeDefault: {
						Key:     backofficeDefault,
						Name:    "Default",
						Classes: map[identity.Key]model_class.Class{adminKey: adminClass},
					},
				},
			},
			platformDomain: {
				Key:  platformDomain,
				Name: "Platform",
				Subdomains: map[identity.Key]model_domain.Subdomain{
					platformLeaderboards: {
						Key:     platformLeaderboards,
						Name:    "Leaderboards",
						Classes: map[identity.Key]model_class.Class{resolverKey: resolverClass},
					},
				},
			},
		},
		ClassAssociations: map[identity.Key]model_class.Association{assocKey: assoc},
	}

	reqs := req_flat.NewRequirements(model)
	_, diagramClasses, _ := reqs.RegardingClasses([]model_class.Class{resolverClass})

	md, err := generateClassMdContents(reqs, resolverClass, diagramClasses, "", "")
	require.NoError(t, err)
	require.Contains(t, md, "Backoffice::Administrator")
	require.NotContains(t, md, "Backoffice::Default::Administrator")
}

func TestGenerateSubdomainMarkdownUsesScopedGeneralizationNames(t *testing.T) {
	t.Parallel()

	financeDomain := helper.Must(identity.NewDomainKey("finance"))
	walletSub := helper.Must(identity.NewSubdomainKey(financeDomain, "wallet"))
	opsSub := helper.Must(identity.NewSubdomainKey(financeDomain, "operations"))

	partnerKey := helper.Must(identity.NewClassKey(walletSub, "partner"))
	playerKey := helper.Must(identity.NewClassKey(opsSub, "player"))
	genKey := helper.Must(identity.NewGeneralizationKey(walletSub, "account_holder"))
	gen := model_class.NewGeneralization(genKey, model_class.GeneralizationDetails{Name: "Account Holder", Details: ""}, "", model_class.GeneralizationTraits{IsComplete: true, IsStatic: true}, "")
	partnerClass := model_class.NewClass(partnerKey, model_class.ClassLinks{SuperclassOfKey: &genKey}, model_class.ClassDetails{Name: "Partner", Details: "", UnfinishedNotes: "", UmlComment: ""})
	playerClass := model_class.NewClass(playerKey, model_class.ClassLinks{SubclassOfKey: &genKey}, model_class.ClassDetails{Name: "Player", Details: "", UnfinishedNotes: "", UmlComment: ""})

	domain := model_domain.Domain{
		Key:  financeDomain,
		Name: "Finance",
		Subdomains: map[identity.Key]model_domain.Subdomain{
			walletSub: {
				Key:             walletSub,
				Name:            "Wallet",
				Classes:         map[identity.Key]model_class.Class{partnerKey: partnerClass},
				Generalizations: map[identity.Key]model_class.Generalization{genKey: gen},
			},
			opsSub: {
				Key:     opsSub,
				Name:    "Operations",
				Classes: map[identity.Key]model_class.Class{playerKey: playerClass},
			},
		},
	}
	model := core.Model{
		Key:     "finance",
		Name:    "Finance",
		Domains: map[identity.Key]model_domain.Domain{financeDomain: domain},
	}

	reqs := req_flat.NewRequirements(model)
	md, err := generateSubdomainMdContents(reqs, model, domain, domain.Subdomains[walletSub], "", "")
	require.NoError(t, err)
	require.Contains(t, md, "Operations::Player")
}

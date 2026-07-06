package generate

import (
	"strings"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateClassesMermaidUsesNamespacesForCrossScopeClasses(t *testing.T) {
	t.Parallel()

	backofficeDomain := helper.Must(identity.NewDomainKey("backoffice"))
	backofficeDefault := helper.Must(identity.NewSubdomainKey(backofficeDomain, "default"))
	adminKey := helper.Must(identity.NewClassKey(backofficeDefault, "administrator"))
	adminClass := model_class.NewClass(adminKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Administrator", Details: "Configures leaderboards."})

	platformDomain := helper.Must(identity.NewDomainKey("platform"))
	platformLeaderboards := helper.Must(identity.NewSubdomainKey(platformDomain, "leaderboards"))
	resolverKey := helper.Must(identity.NewClassKey(platformLeaderboards, "resolver"))
	resolverClass := model_class.NewClass(resolverKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Resolver", Details: "Leaderboard rules."})

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

	writer := newCollectWriter()
	require.NoError(t, GenerateMdToWriter(model, writer, nil))

	adminFile := convertKeyToFilename("class", adminKey.String(), "", ".md")
	adminBody := string(writer.md[adminFile])
	resolverNode := nodeIDFor("class", resolverKey)

	assert.Contains(t, adminBody, "namespace Platform.Leaderboards {")
	assert.Contains(t, adminBody, resolverNode)
	assert.NotContains(t, adminBody, "namespace Backoffice")

	resolverFile := convertKeyToFilename("class", resolverKey.String(), "", ".md")
	resolverBody := string(writer.md[resolverFile])
	adminNode := nodeIDFor("class", adminKey)

	assert.Contains(t, resolverBody, "namespace Backoffice {")
	assert.Contains(t, resolverBody, adminNode)
	assert.NotContains(t, resolverBody, "namespace Backoffice.Default")
}

func TestGenerateClassesMermaidActorStereotypeOutsideNamespace(t *testing.T) {
	t.Parallel()

	actorKey := helper.Must(identity.NewActorKey("administrator_actor"))
	backofficeDomain := helper.Must(identity.NewDomainKey("backoffice"))
	backofficeDefault := helper.Must(identity.NewSubdomainKey(backofficeDomain, "default"))
	adminKey := helper.Must(identity.NewClassKey(backofficeDefault, "administrator"))
	adminClass := model_class.NewClass(
		adminKey,
		model_class.ClassLinks{ActorKey: &actorKey},
		model_class.ClassDetails{Name: "Administrator", Details: "Configures leaderboards."},
	)

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
		Actors: map[identity.Key]model_actor.Actor{
			actorKey: model_actor.NewActor(actorKey, "person", model_actor.GeneralizationRefs{}, model_actor.ActorDetails{Name: "Administrator", Details: ""}),
		},
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

	writer := newCollectWriter()
	require.NoError(t, GenerateMdToWriter(model, writer, nil))

	resolverFile := convertKeyToFilename("class", resolverKey.String(), "", ".md")
	body := string(writer.md[resolverFile])
	adminNode := nodeIDFor("class", adminKey)
	actorStereotype := "<<actor>> " + adminNode

	namespaceStart := strings.Index(body, "namespace Backoffice {")
	require.Positive(t, namespaceStart)
	suffix := body[namespaceStart:]
	stereotypeIdx := strings.Index(suffix, actorStereotype)
	require.Positive(t, stereotypeIdx)
	beforeStereotype := suffix[:stereotypeIdx]
	assert.Contains(t, beforeStereotype, "}", "namespace block should close before actor stereotype")
	assert.NotContains(t, beforeStereotype, "<<actor>>", "actor stereotype must not appear inside namespace block")
}

func TestGenerateClassesMermaidUsesSubdomainNamespaceWithinDomain(t *testing.T) {
	t.Parallel()

	financeDomain := helper.Must(identity.NewDomainKey("finance"))
	walletSub := helper.Must(identity.NewSubdomainKey(financeDomain, "wallet"))
	opsSub := helper.Must(identity.NewSubdomainKey(financeDomain, "operations"))

	partnerKey := helper.Must(identity.NewClassKey(walletSub, "partner"))
	playerKey := helper.Must(identity.NewClassKey(opsSub, "player"))
	one := helper.Must(model_class.NewMultiplicity("1"))
	assocKey := helper.Must(identity.NewClassAssociationKey(financeDomain, partnerKey, playerKey, "serves"))
	assoc := model_class.NewAssociation(
		assocKey,
		model_class.AssociationDetails{Name: "Serves", Details: ""},
		model_class.AssociationEnd{ClassKey: partnerKey, Multiplicity: one},
		model_class.AssociationEnd{ClassKey: playerKey, Multiplicity: one},
		model_class.AssociationOptions{},
	)

	model := core.Model{
		Key:  "finance",
		Name: "Finance",
		Domains: map[identity.Key]model_domain.Domain{
			financeDomain: {
				Key:  financeDomain,
				Name: "Finance",
				Subdomains: map[identity.Key]model_domain.Subdomain{
					walletSub: {
						Key:  walletSub,
						Name: "Wallet",
						Classes: map[identity.Key]model_class.Class{
							partnerKey: model_class.NewClass(partnerKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Partner", Details: ""}),
						},
					},
					opsSub: {
						Key:  opsSub,
						Name: "Operations",
						Classes: map[identity.Key]model_class.Class{
							playerKey: model_class.NewClass(playerKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Player", Details: ""}),
						},
					},
				},
			},
		},
		ClassAssociations: map[identity.Key]model_class.Association{assocKey: assoc},
	}

	writer := newCollectWriter()
	require.NoError(t, GenerateMdToWriter(model, writer, nil))

	walletFile := convertKeyToFilename("subdomain", walletSub.String(), "", ".md")
	walletBody := string(writer.md[walletFile])
	playerNode := nodeIDFor("class", playerKey)

	assert.Contains(t, walletBody, "namespace Operations {")
	assert.Contains(t, walletBody, playerNode)
	assert.NotContains(t, walletBody, "namespace Finance.Operations")
}

func TestMermaidNamespaceSegmentStripsNonAlphanumerics(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "DX", mermaidNamespaceSegment("D X"))
	assert.Equal(t, "Scope", mermaidNamespaceSegment("!!!"))
}

func TestMermaidNamespacePathFromSegments(t *testing.T) {
	t.Parallel()

	assert.Empty(t, mermaidNamespacePathFromSegments(nil))
	assert.Equal(t, "Platform.Leaderboards", mermaidNamespacePathFromSegments([]string{"Platform", "Leaderboards"}))
}

func TestGenerateClassesMermaidSameSubdomainHasNoNamespaces(t *testing.T) {
	t.Parallel()

	domainKey := helper.Must(identity.NewDomainKey("d"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "s"))
	alphaKey := helper.Must(identity.NewClassKey(subdomainKey, "alpha"))
	betaKey := helper.Must(identity.NewClassKey(subdomainKey, "beta"))
	one := helper.Must(model_class.NewMultiplicity("1"))
	assocKey := helper.Must(identity.NewClassAssociationKey(subdomainKey, alphaKey, betaKey, "links"))
	assoc := model_class.NewAssociation(
		assocKey,
		model_class.AssociationDetails{Name: "links", Details: ""},
		model_class.AssociationEnd{ClassKey: alphaKey, Multiplicity: one},
		model_class.AssociationEnd{ClassKey: betaKey, Multiplicity: one},
		model_class.AssociationOptions{},
	)

	model := core.Model{
		Key:  "local_only",
		Name: "Test",
		Domains: map[identity.Key]model_domain.Domain{
			domainKey: {
				Key: domainKey,
				Subdomains: map[identity.Key]model_domain.Subdomain{
					subdomainKey: {
						Key: subdomainKey,
						Classes: map[identity.Key]model_class.Class{
							alphaKey: model_class.NewClass(alphaKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Alpha", Details: ""}),
							betaKey:  model_class.NewClass(betaKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Beta", Details: ""}),
						},
						ClassAssociations: map[identity.Key]model_class.Association{assocKey: assoc},
					},
				},
			},
		},
	}

	writer := newCollectWriter()
	require.NoError(t, GenerateMdToWriter(model, writer, nil))

	alphaFile := convertKeyToFilename("class", alphaKey.String(), "", ".md")
	body := string(writer.md[alphaFile])
	diagram := body[strings.Index(body, "```mermaid"):]
	assert.NotContains(t, diagram, "namespace ")
}

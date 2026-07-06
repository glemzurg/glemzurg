package generate

import (
	"strings"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateDomainMdListsExternalDiagramClasses(t *testing.T) {
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

	domainFile := convertKeyToFilename("domain", backofficeDomain.String(), "", ".md")
	got := string(writer.md[domainFile])

	classesIdx := strings.Index(got, "## Classes")
	require.Positive(t, classesIdx)
	classesSection := got[classesIdx:]

	assert.Contains(t, classesSection, "- **[Administrator]")
	assert.Contains(t, classesSection, "Platform::Leaderboards::Resolver")
	assert.NotContains(t, classesSection, "- **[Resolver]")

	adminBullet := strings.Index(classesSection, "- **[Administrator]")
	resolverBullet := strings.Index(classesSection, "Platform::Leaderboards::Resolver")
	require.Positive(t, adminBullet)
	require.Positive(t, resolverBullet)
	assert.Less(t, adminBullet, resolverBullet, "in-domain classes should precede external diagram classes")
}

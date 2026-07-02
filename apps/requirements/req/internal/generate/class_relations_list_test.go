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

func TestGenerateClassMdListsDiagramClassesBeneathRelations(t *testing.T) {
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

	alpha := model_class.NewClass(alphaKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Alpha", Details: "First class."})
	beta := model_class.NewClass(betaKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Beta", Details: "Second class."})

	model := core.Model{
		Key:  "class_relations_list",
		Name: "Test",
		Domains: map[identity.Key]model_domain.Domain{
			domainKey: {
				Key: domainKey,
				Subdomains: map[identity.Key]model_domain.Subdomain{
					subdomainKey: {
						Key: subdomainKey,
						Classes: map[identity.Key]model_class.Class{
							alphaKey: alpha,
							betaKey:  beta,
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
	got := string(writer.md[alphaFile])

	relationsIdx := strings.Index(got, "## Relations")
	require.Positive(t, relationsIdx)
	relationsSection := got[relationsIdx:]

	assert.Contains(t, relationsSection, "The classes in this diagram.")
	assert.Contains(t, relationsSection, "- **[Alpha]("+convertKeyToFilename("class", alphaKey.String(), "", ".md")+").** First class.")
	assert.Contains(t, relationsSection, "- **[Beta]("+convertKeyToFilename("class", betaKey.String(), "", ".md")+").** Second class.")

	alphaBullet := strings.Index(relationsSection, "- **[Alpha]")
	betaBullet := strings.Index(relationsSection, "- **[Beta]")
	require.Positive(t, alphaBullet)
	require.Positive(t, betaBullet)
	assert.Less(t, alphaBullet, betaBullet, "diagram class list should be sorted by name like subdomain pages")
}

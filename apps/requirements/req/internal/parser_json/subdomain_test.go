package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_use_case"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubdomainInOutRoundTrip(t *testing.T) {
	domainKey, err := identity.NewDomainKey("domain1")
	require.NoError(t, err)
	subdomainKey, err := identity.NewSubdomainKey(domainKey, "sub1")
	require.NoError(t, err)
	genKey, err := identity.NewGeneralizationKey(subdomainKey, "gen1")
	require.NoError(t, err)
	classKey, err := identity.NewClassKey(subdomainKey, "class1")
	require.NoError(t, err)
	useCaseKey, err := identity.NewUseCaseKey(subdomainKey, "uc1")
	require.NoError(t, err)
	assocKey, err := identity.NewClassAssociationKey(subdomainKey, classKey, classKey)
	require.NoError(t, err)

	original := model_domain.Subdomain{
		Key:        subdomainKey,
		Name:       "Sub1",
		Details:    "Details",
		UmlComment: "comment",
		Generalizations: map[identity.Key]model_class.Generalization{
			genKey: {
				Key:  genKey,
				Name: "Gen1",
			},
		},
		Classes: map[identity.Key]model_class.Class{
			classKey: {
				Key:  classKey,
				Name: "Class1",
			},
		},
		UseCases: map[identity.Key]model_use_case.UseCase{
			useCaseKey: {
				Key:   useCaseKey,
				Name:  "UC1",
				Level: "sea",
			},
		},
		ClassAssociations: map[identity.Key]model_class.Association{
			assocKey: {
				Key:          assocKey,
				Name:         "Assoc1",
				FromClassKey: classKey,
				ToClassKey:   classKey,
			},
		},
	}

	inOut := FromRequirementsSubdomain(original)
	back, err := inOut.ToRequirements()
	require.NoError(t, err)
	assert.Equal(t, original, back)
}

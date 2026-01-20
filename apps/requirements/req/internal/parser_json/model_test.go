package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModelInOutRoundTrip(t *testing.T) {
	actorKey, err := identity.NewActorKey("actor1")
	require.NoError(t, err)
	domain1Key, err := identity.NewDomainKey("domain1")
	require.NoError(t, err)
	domain2Key, err := identity.NewDomainKey("domain2")
	require.NoError(t, err)
	daKey, err := identity.NewDomainAssociationKey(domain1Key, domain2Key)
	require.NoError(t, err)
	subdomainKey, err := identity.NewSubdomainKey(domain1Key, "sub1")
	require.NoError(t, err)
	class1Key, err := identity.NewClassKey(subdomainKey, "class1")
	require.NoError(t, err)
	// Use domain2 for class2 so they're in different domains (for model-level association)
	subdomain2Key, err := identity.NewSubdomainKey(domain2Key, "sub2")
	require.NoError(t, err)
	class2Key, err := identity.NewClassKey(subdomain2Key, "class2")
	require.NoError(t, err)
	// Model-level class associations span domains, so they have no parent
	assocKey, err := identity.NewClassAssociationKey(identity.Key{}, class1Key, class2Key)
	require.NoError(t, err)

	original := req_model.Model{
		Key:     "model1",
		Name:    "Test Model",
		Details: "Details",

		Actors: map[identity.Key]model_actor.Actor{
			actorKey: {Key: actorKey, Name: "User", Type: "person", UmlComment: "comment"},
		},
		Domains: map[identity.Key]model_domain.Domain{
			domain1Key: {Key: domain1Key, Name: "Domain1", Realized: true, UmlComment: "comment"},
		},
		DomainAssociations: map[identity.Key]model_domain.Association{
			daKey: {Key: daKey, ProblemDomainKey: domain1Key, SolutionDomainKey: domain2Key, UmlComment: "comment"},
		},
		ClassAssociations: map[identity.Key]model_class.Association{
			assocKey: {Key: assocKey, Name: "Assoc1", FromClassKey: class1Key, ToClassKey: class2Key, UmlComment: "comment"},
		},
	}

	inOut := FromRequirementsModel(original)
	back, err := inOut.ToRequirements()
	require.NoError(t, err)
	assert.Equal(t, original, back)
}

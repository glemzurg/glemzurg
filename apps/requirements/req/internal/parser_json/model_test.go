package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/stretchr/testify/assert"
)

func TestModelInOutRoundTrip(t *testing.T) {
	original := requirements.Requirements{
		Model: requirements.Model{
			Key:     "model1",
			Name:    "Test Model",
			Details: "Details",
		},
		Actors: []requirements.Actor{
			{Key: "actor1", Name: "User", Type: "person", UmlComment: "comment"},
		},
		Domains: []requirements.Domain{
			{Key: "domain1", Name: "Domain1", Realized: true, UmlComment: "comment"},
		},
		Subdomains: map[string][]requirements.Subdomain{
			"domain1": {{Key: "sub1", Name: "Sub1", UmlComment: "comment"}},
		},
		DomainAssociations: []requirements.DomainAssociation{
			{Key: "da1", ProblemDomainKey: "domain1", SolutionDomainKey: "domain2", UmlComment: "comment"},
		},
		Associations: []requirements.Association{
			{Key: "assoc1", Name: "Assoc1", FromClassKey: "class1", ToClassKey: "class2", UmlComment: "comment"},
		},
	}

	inOut := FromRequirementsModel(original)
	back := inOut.ToRequirements()

	// Check model
	assert.Equal(t, original.Model, back.Model)

	// Check actors
	assert.Len(t, back.Actors, len(original.Actors))

	// Check domains
	assert.Len(t, back.Domains, len(original.Domains))

	// Check subdomains
	assert.Len(t, back.Subdomains, len(original.Subdomains))

	// Check domain associations
	assert.Len(t, back.DomainAssociations, len(original.DomainAssociations))

	// Check associations
	assert.Len(t, back.Associations, len(original.Associations))
}

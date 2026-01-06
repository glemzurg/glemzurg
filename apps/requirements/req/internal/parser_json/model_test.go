package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/stretchr/testify/assert"
)

func TestModelInOutRoundTrip(t *testing.T) {
	original := requirements.Model{
		Key:     "model1",
		Name:    "Test Model",
		Details: "Details",

		Actors: []requirements.Actor{
			{Key: "actor1", Name: "User", Type: "person", UmlComment: "comment"},
		},
		Domains: []requirements.Domain{
			{Key: "domain1", Name: "Domain1", Realized: true, UmlComment: "comment"},
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
	assert.Equal(t, original, back)
}

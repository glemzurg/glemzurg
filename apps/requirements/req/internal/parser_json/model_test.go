package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model"
	"github.com/stretchr/testify/assert"
)

func TestModelInOutRoundTrip(t *testing.T) {
	original := model.Model{
		Key:     "model1",
		Name:    "Test Model",
		Details: "Details",

		Actors: []actor.Actor{
			{Key: "actor1", Name: "User", Type: "person", UmlComment: "comment"},
		},
		Domains: []domain.Domain{
			{Key: "domain1", Name: "Domain1", Realized: true, UmlComment: "comment"},
		},
		DomainAssociations: []domain.DomainAssociation{
			{Key: "da1", ProblemDomainKey: "domain1", SolutionDomainKey: "domain2", UmlComment: "comment"},
		},
		Associations: []class.Association{
			{Key: "assoc1", Name: "Assoc1", FromClassKey: "class1", ToClassKey: "class2", UmlComment: "comment"},
		},
	}

	inOut := FromRequirementsModel(original)
	back := inOut.ToRequirements()
	assert.Equal(t, original, back)
}

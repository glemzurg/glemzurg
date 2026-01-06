package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/stretchr/testify/assert"
)

func TestSubdomainInOutRoundTrip(t *testing.T) {
	original := requirements.Subdomain{
		Key:        "sub1",
		Name:       "Sub1",
		Details:    "Details",
		UmlComment: "comment",
		Generalizations: []requirements.Generalization{
			{
				Key: "gen1",
			},
		},
		Classes: []requirements.Class{
			{
				Key: "class1",
			},
		},
		UseCases: []requirements.UseCase{
			{
				Key: "uc1",
			},
		},
		Associations: []requirements.Association{
			{
				Key: "assoc1",
			},
		},
	}

	inOut := FromRequirementsSubdomain(original)
	back := inOut.ToRequirements()
	assert.Equal(t, original, back)
}

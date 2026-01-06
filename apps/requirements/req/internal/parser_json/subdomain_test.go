package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/use_case"
	"github.com/stretchr/testify/assert"
)

func TestSubdomainInOutRoundTrip(t *testing.T) {
	original := domain.Subdomain{
		Key:        "sub1",
		Name:       "Sub1",
		Details:    "Details",
		UmlComment: "comment",
		Generalizations: []class.Generalization{
			{
				Key: "gen1",
			},
		},
		Classes: []class.Class{
			{
				Key: "class1",
			},
		},
		UseCases: []use_case.UseCase{
			{
				Key: "uc1",
			},
		},
		Associations: []class.Association{
			{
				Key: "assoc1",
			},
		},
	}

	inOut := FromRequirementsSubdomain(original)
	back := inOut.ToRequirements()
	assert.Equal(t, original, back)
}

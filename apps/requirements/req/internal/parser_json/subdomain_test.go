package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_use_case"
	"github.com/stretchr/testify/assert"
)

func TestSubdomainInOutRoundTrip(t *testing.T) {
	original := model_domain.Subdomain{
		Key:        "sub1",
		Name:       "Sub1",
		Details:    "Details",
		UmlComment: "comment",
		Generalizations: []model_class.Generalization{
			{
				Key: "gen1",
			},
		},
		Classes: []model_class.Class{
			{
				Key: "class1",
			},
		},
		UseCases: []model_use_case.UseCase{
			{
				Key: "uc1",
			},
		},
		Associations: []model_class.Association{
			{
				Key: "assoc1",
			},
		},
	}

	inOut := FromRequirementsSubdomain(original)
	back := inOut.ToRequirements()
	assert.Equal(t, original, back)
}

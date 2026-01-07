package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_domain"
	"github.com/stretchr/testify/assert"
)

func TestDomainAssociationInOutRoundTrip(t *testing.T) {
	original := model_domain.DomainAssociation{
		Key:               "da1",
		ProblemDomainKey:  "domain1",
		SolutionDomainKey: "domain2",
		UmlComment:        "comment",
	}

	inOut := FromRequirementsDomainAssociation(original)
	back := inOut.ToRequirements()
	assert.Equal(t, original, back)
}

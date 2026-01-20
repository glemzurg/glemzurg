package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDomainAssociationInOutRoundTrip(t *testing.T) {
	domain1Key, err := identity.NewDomainKey("domain1")
	require.NoError(t, err)
	domain2Key, err := identity.NewDomainKey("domain2")
	require.NoError(t, err)
	daKey, err := identity.NewDomainAssociationKey(domain1Key, domain2Key)
	require.NoError(t, err)

	original := model_domain.Association{
		Key:               daKey,
		ProblemDomainKey:  domain1Key,
		SolutionDomainKey: domain2Key,
		UmlComment:        "comment",
	}

	inOut := FromRequirementsDomainAssociation(original)
	back, err := inOut.ToRequirements()
	require.NoError(t, err)
	assert.Equal(t, original, back)
}

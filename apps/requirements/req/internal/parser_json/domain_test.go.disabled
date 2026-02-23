package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDomainInOutRoundTrip(t *testing.T) {
	key, err := identity.NewDomainKey("domain1")
	require.NoError(t, err)

	original := model_domain.Domain{
		Key:        key,
		Name:       "Domain1",
		Details:    "Details",
		Realized:   true,
		UmlComment: "comment",
	}

	inOut := FromRequirementsDomain(original)
	back, err := inOut.ToRequirements()
	require.NoError(t, err)
	assert.Equal(t, original, back)
}

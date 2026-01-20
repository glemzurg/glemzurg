package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGeneralizationInOutRoundTrip(t *testing.T) {
	domainKey, err := identity.NewDomainKey("domain1")
	require.NoError(t, err)
	subdomainKey, err := identity.NewSubdomainKey(domainKey, "sub1")
	require.NoError(t, err)
	genKey, err := identity.NewGeneralizationKey(subdomainKey, "gen1")
	require.NoError(t, err)

	original := model_class.Generalization{
		Key:        genKey,
		Name:       "Gen1",
		Details:    "Details",
		IsComplete: true,
		IsStatic:   false,
		UmlComment: "comment",
	}

	inOut := FromRequirementsGeneralization(original)
	back, err := inOut.ToRequirements()
	require.NoError(t, err)
	assert.Equal(t, original, back)
}

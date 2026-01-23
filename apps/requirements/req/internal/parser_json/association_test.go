package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssociationInOutRoundTrip(t *testing.T) {
	domainKey, err := identity.NewDomainKey("domain1")
	require.NoError(t, err)
	subdomainKey, err := identity.NewSubdomainKey(domainKey, "sub1")
	require.NoError(t, err)
	class1Key, err := identity.NewClassKey(subdomainKey, "class1")
	require.NoError(t, err)
	class2Key, err := identity.NewClassKey(subdomainKey, "class2")
	require.NoError(t, err)
	aclassKey, err := identity.NewClassKey(subdomainKey, "aclass")
	require.NoError(t, err)
	assocKey, err := identity.NewClassAssociationKey(subdomainKey, class1Key, class2Key)
	require.NoError(t, err)

	original := model_class.Association{
		Key:                 assocKey,
		Name:                "Assoc1",
		Details:             "Details",
		FromClassKey:        class1Key,
		FromMultiplicity:    model_class.Multiplicity{LowerBound: 1, HigherBound: 1},
		ToClassKey:          class2Key,
		ToMultiplicity:      model_class.Multiplicity{LowerBound: 0, HigherBound: 5},
		AssociationClassKey: &aclassKey,
		UmlComment:          "comment",
	}

	inOut := FromRequirementsAssociation(original)
	back, err := inOut.ToRequirements()
	require.NoError(t, err)
	assert.Equal(t, original, back)
}

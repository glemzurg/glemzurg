package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_data_type"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAttributeInOutRoundTrip(t *testing.T) {
	domainKey, err := identity.NewDomainKey("domain1")
	require.NoError(t, err)
	subdomainKey, err := identity.NewSubdomainKey(domainKey, "sub1")
	require.NoError(t, err)
	classKey, err := identity.NewClassKey(subdomainKey, "class1")
	require.NoError(t, err)
	attrKey, err := identity.NewAttributeKey(classKey, "attr1")
	require.NoError(t, err)

	original := model_class.Attribute{
		Key:              attrKey,
		Name:             "Name",
		Details:          "Details",
		DataTypeRules:    "string",
		DerivationPolicy: "derived",
		Nullable:         true,
		UmlComment:       "comment",
		IndexNums:        []uint{1},
		DataType: &model_data_type.DataType{
			Key: "dt1",
		},
	}

	inOut := FromRequirementsAttribute(original)
	back, err := inOut.ToRequirements()
	require.NoError(t, err)
	assert.Equal(t, original, back)
}

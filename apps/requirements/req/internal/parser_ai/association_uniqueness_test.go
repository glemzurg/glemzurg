package parser_ai

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/require"
)

func TestConvertUniquenessRoundTrip(t *testing.T) {
	subdomainKey := helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("finance")), "wallet"))
	partnerClass := helper.Must(identity.NewClassKey(subdomainKey, "partner"))
	jurisdictionClass := helper.Must(identity.NewClassKey(subdomainKey, "jurisdiction"))
	fromAttr := helper.Must(identity.NewAttributeKey(partnerClass, "partner_code"))
	toAttr := helper.Must(identity.NewAttributeKey(jurisdictionClass, "jurisdiction_code"))
	uniqueness := model_class.NewAssociationUniqueness([]identity.Key{fromAttr}, []identity.Key{toAttr})

	input := convertUniquenessFromModel(&uniqueness)
	require.NotNil(t, input)
	require.Equal(t, []string{"partner_code"}, input.FromAttributes)
	require.Equal(t, []string{"jurisdiction_code"}, input.ToAttributes)

	result, err := convertInputUniqueness(input, partnerClass, jurisdictionClass, "test.assoc.json")
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, uniqueness, *result)
}

func TestConvertUniquenessAbsentIsNilPointer(t *testing.T) {
	result, err := convertInputUniqueness(nil, identity.Key{}, identity.Key{}, "test.assoc.json")
	require.NoError(t, err)
	require.Nil(t, result)
}

func TestConvertUniquenessFromModelNilPointer(t *testing.T) {
	require.Nil(t, convertUniquenessFromModel(nil))
}

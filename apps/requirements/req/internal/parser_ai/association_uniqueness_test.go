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
	uniqueness := []model_class.AssociationUniqueness{
		model_class.NewAssociationUniqueness([]identity.Key{fromAttr}, []identity.Key{toAttr}),
	}

	input := convertUniquenessFromModel(uniqueness)
	require.Len(t, input, 1)
	require.Equal(t, []string{"partner_code"}, input[0].FromAttributes)
	require.Equal(t, []string{"jurisdiction_code"}, input[0].ToAttributes)

	result, err := convertInputUniqueness(input, partnerClass, jurisdictionClass, "test.assoc.json")
	require.NoError(t, err)
	require.Equal(t, uniqueness, result)
}

func TestConvertUniquenessMultipleRoundTrip(t *testing.T) {
	subdomainKey := helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("process")), "family"))
	familyClass := helper.Must(identity.NewClassKey(subdomainKey, "family"))
	phaseClass := helper.Must(identity.NewClassKey(subdomainKey, "phase"))
	nameAttr := helper.Must(identity.NewAttributeKey(phaseClass, "name"))
	numAttr := helper.Must(identity.NewAttributeKey(phaseClass, "num"))
	uniqueness := []model_class.AssociationUniqueness{
		model_class.NewAssociationUniqueness(nil, []identity.Key{nameAttr}),
		model_class.NewAssociationUniqueness(nil, []identity.Key{numAttr}),
	}

	input := convertUniquenessFromModel(uniqueness)
	require.Len(t, input, 2)
	require.Equal(t, []string{"name"}, input[0].ToAttributes)
	require.Equal(t, []string{"num"}, input[1].ToAttributes)

	result, err := convertInputUniqueness(input, familyClass, phaseClass, "test.assoc.json")
	require.NoError(t, err)
	require.Equal(t, uniqueness, result)
}

func TestConvertUniquenessAbsentIsNil(t *testing.T) {
	result, err := convertInputUniqueness(nil, identity.Key{}, identity.Key{}, "test.assoc.json")
	require.NoError(t, err)
	require.Nil(t, result)
}

func TestConvertUniquenessFromModelEmpty(t *testing.T) {
	require.Nil(t, convertUniquenessFromModel(nil))
	require.Nil(t, convertUniquenessFromModel([]model_class.AssociationUniqueness{}))
}

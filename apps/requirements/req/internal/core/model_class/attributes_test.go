package model_class

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAttributesByKey(t *testing.T) {
	classKey := helper.Must(identity.NewClassKey(
		helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("domain_a")), "sub_a")),
		"class_a",
	))
	firstKey := helper.Must(identity.NewAttributeKey(classKey, "first"))
	secondKey := helper.Must(identity.NewAttributeKey(classKey, "second"))

	attrs := []Attribute{
		{Key: firstKey, Name: "First"},
		{Key: secondKey, Name: "Second"},
	}

	byKey := AttributesByKey(attrs)
	assert.Len(t, byKey, 2)
	assert.Equal(t, "First", byKey[firstKey].Name)
	assert.Equal(t, "Second", byKey[secondKey].Name)
}

func TestAttributeBySubKey(t *testing.T) {
	classKey := helper.Must(identity.NewClassKey(
		helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("domain_a")), "sub_a")),
		"class_a",
	))
	attrKey := helper.Must(identity.NewAttributeKey(classKey, "abbr"))
	attrs := []Attribute{{Key: attrKey, Name: "Abbr"}}

	got, ok := AttributeBySubKey(attrs, "abbr")
	require.True(t, ok)
	assert.Equal(t, "Abbr", got.Name)

	_, ok = AttributeBySubKey(attrs, "missing")
	assert.False(t, ok)
}

func TestValidateUniqueAttributeKeys(t *testing.T) {
	classKey := helper.Must(identity.NewClassKey(
		helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("domain_a")), "sub_a")),
		"class_a",
	))
	attrKey := helper.Must(identity.NewAttributeKey(classKey, "total"))

	ctx := coreerr.NewContext("class", classKey.String())
	err := validateUniqueAttributeKeys(ctx, []Attribute{
		{Key: attrKey, Name: "Total"},
		{Key: attrKey, Name: "Total Again"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate attribute key")
}

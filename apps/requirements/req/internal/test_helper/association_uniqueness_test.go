package test_helper

import (
	"maps"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetTestModelAssociationUniqueness(t *testing.T) {
	model := GetTestModel()

	var fromOnly, toOnly, bothSides, withUniqueness int
	for _, assoc := range model.GetClassAssociations() {
		if assoc.Uniqueness == nil {
			continue
		}
		withUniqueness++
		fromCount := len(assoc.Uniqueness.FromAttributeKeys)
		toCount := len(assoc.Uniqueness.ToAttributeKeys)
		switch {
		case fromCount > 0 && toCount > 0:
			bothSides++
		case fromCount > 0:
			fromOnly++
		case toCount > 0:
			toOnly++
		default:
			t.Fatalf("association %q has empty uniqueness tuple", assoc.Name)
		}
	}

	assert.Equal(t, 3, withUniqueness, "test model should exercise uniqueness on three associations")
	assert.Equal(t, 1, fromOnly, "expected one from-only uniqueness association")
	assert.Equal(t, 1, toOnly, "expected one to-only uniqueness association")
	assert.Equal(t, 1, bothSides, "expected one both-side uniqueness association")

	classes := allClassesFromModel(model)

	toOnlyAssoc, ok := findAssociationByName(model, "order belongs to customer")
	require.True(t, ok)
	require.NotNil(t, toOnlyAssoc.Uniqueness)
	require.Empty(t, toOnlyAssoc.Uniqueness.FromAttributeKeys)
	require.Len(t, toOnlyAssoc.Uniqueness.ToAttributeKeys, 1)
	assert.True(t, classHasAttributeKey(classes[toOnlyAssoc.ToClassKey], toOnlyAssoc.Uniqueness.ToAttributeKeys[0]))

	fromOnlyAssoc, ok := findAssociationByName(model, "product stored on shelf")
	require.True(t, ok)
	require.NotNil(t, fromOnlyAssoc.Uniqueness)
	require.Len(t, fromOnlyAssoc.Uniqueness.FromAttributeKeys, 1)
	require.Empty(t, fromOnlyAssoc.Uniqueness.ToAttributeKeys)
	assert.True(t, classHasAttributeKey(classes[fromOnlyAssoc.FromClassKey], fromOnlyAssoc.Uniqueness.FromAttributeKeys[0]))

	bothAssoc, ok := findAssociationByName(model, "order has shipment")
	require.True(t, ok)
	require.NotNil(t, bothAssoc.Uniqueness)
	require.Len(t, bothAssoc.Uniqueness.FromAttributeKeys, 1)
	require.Len(t, bothAssoc.Uniqueness.ToAttributeKeys, 1)
	assert.True(t, classHasAttributeKey(classes[bothAssoc.FromClassKey], bothAssoc.Uniqueness.FromAttributeKeys[0]))
	assert.True(t, classHasAttributeKey(classes[bothAssoc.ToClassKey], bothAssoc.Uniqueness.ToAttributeKeys[0]))
}

func allClassesFromModel(model core.Model) map[identity.Key]model_class.Class {
	classes := make(map[identity.Key]model_class.Class)
	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			maps.Copy(classes, subdomain.Classes)
		}
	}
	return classes
}

func findAssociationByName(model core.Model, name string) (model_class.Association, bool) {
	for _, assoc := range model.GetClassAssociations() {
		if assoc.Name == name {
			return assoc, true
		}
	}
	return model_class.Association{}, false
}

func classHasAttributeKey(class model_class.Class, attrKey identity.Key) bool {
	for _, attr := range class.Attributes {
		if attr.Key == attrKey {
			return true
		}
	}
	return false
}

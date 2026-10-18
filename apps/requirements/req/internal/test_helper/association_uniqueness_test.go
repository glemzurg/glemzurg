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

	var fromOnly, toOnly, bothSides, withUniqueness, multiConstraint int
	for _, assoc := range model.GetClassAssociations() {
		if len(assoc.Uniqueness) == 0 {
			continue
		}
		withUniqueness++
		if len(assoc.Uniqueness) > 1 {
			multiConstraint++
		}
		for i, uniqueness := range assoc.Uniqueness {
			fromCount := len(uniqueness.FromAttributeKeys)
			toCount := len(uniqueness.ToAttributeKeys)
			switch {
			case fromCount > 0 && toCount > 0:
				bothSides++
			case fromCount > 0:
				fromOnly++
			case toCount > 0:
				toOnly++
			default:
				t.Fatalf("association %q uniqueness[%d] has empty uniqueness tuple", assoc.Name, i)
			}
		}
	}

	assert.Equal(t, 3, withUniqueness, "test model should exercise uniqueness on three associations")
	assert.Equal(t, 1, multiConstraint, "test model should exercise two uniqueness constraints on one association")
	assert.Equal(t, 2, fromOnly, "expected two from-only uniqueness constraints")
	assert.Equal(t, 1, toOnly, "expected one to-only uniqueness constraint")
	assert.Equal(t, 1, bothSides, "expected one both-side uniqueness constraint")

	classes := allClassesFromModel(model)

	multiAssoc, ok := findAssociationByName(model, "order belongs to customer")
	require.True(t, ok)
	require.Len(t, multiAssoc.Uniqueness, 2)
	require.Empty(t, multiAssoc.Uniqueness[0].FromAttributeKeys)
	require.Len(t, multiAssoc.Uniqueness[0].ToAttributeKeys, 1)
	assert.True(t, classHasAttributeKey(classes[multiAssoc.ToClassKey], multiAssoc.Uniqueness[0].ToAttributeKeys[0]))
	require.Len(t, multiAssoc.Uniqueness[1].FromAttributeKeys, 1)
	require.Empty(t, multiAssoc.Uniqueness[1].ToAttributeKeys)
	assert.True(t, classHasAttributeKey(classes[multiAssoc.FromClassKey], multiAssoc.Uniqueness[1].FromAttributeKeys[0]))

	fromOnlyAssoc, ok := findAssociationByName(model, "product stored on shelf")
	require.True(t, ok)
	require.Len(t, fromOnlyAssoc.Uniqueness, 1)
	require.Len(t, fromOnlyAssoc.Uniqueness[0].FromAttributeKeys, 1)
	require.Empty(t, fromOnlyAssoc.Uniqueness[0].ToAttributeKeys)
	assert.True(t, classHasAttributeKey(classes[fromOnlyAssoc.FromClassKey], fromOnlyAssoc.Uniqueness[0].FromAttributeKeys[0]))

	bothAssoc, ok := findAssociationByName(model, "order has shipment")
	require.True(t, ok)
	require.Len(t, bothAssoc.Uniqueness, 1)
	require.Len(t, bothAssoc.Uniqueness[0].FromAttributeKeys, 1)
	require.Len(t, bothAssoc.Uniqueness[0].ToAttributeKeys, 1)
	assert.True(t, classHasAttributeKey(classes[bothAssoc.FromClassKey], bothAssoc.Uniqueness[0].FromAttributeKeys[0]))
	assert.True(t, classHasAttributeKey(classes[bothAssoc.ToClassKey], bothAssoc.Uniqueness[0].ToAttributeKeys[0]))
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

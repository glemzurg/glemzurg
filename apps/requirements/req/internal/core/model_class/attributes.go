package model_class

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// attributesByKey indexes attributes by key for lookup. Slice order is not represented.
func attributesByKey(attributes []Attribute) map[identity.Key]Attribute {
	if len(attributes) == 0 {
		return nil
	}
	byKey := make(map[identity.Key]Attribute, len(attributes))
	for _, attr := range attributes {
		byKey[attr.Key] = attr
	}
	return byKey
}

// attributeBySubKey returns the attribute whose key sub-key matches, using the same
// normalization as identity key construction.
func attributeBySubKey(attributes []Attribute, subKey string) (Attribute, bool) {
	normalized := identity.NormalizeSubKey(subKey)
	for _, attr := range attributes {
		if attr.Key.SubKey == normalized {
			return attr, true
		}
	}
	return Attribute{}, false
}

func validateUniqueAttributeKeys(ctx *coreerr.ValidationContext, attributes []Attribute) error {
	seen := make(map[identity.Key]bool, len(attributes))
	for i, attr := range attributes {
		if seen[attr.Key] {
			attrCtx := ctx.Child("attribute", fmt.Sprintf("%d", i))
			return coreerr.NewWithValues(attrCtx, coreerr.AttrDuplicateKey, fmt.Sprintf("duplicate attribute key %q", attr.Key.String()), "Key", attr.Key.String(), "")
		}
		seen[attr.Key] = true
	}
	return nil
}

package view_helper

import (
	"sort"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
)

// Sorting constants for attributes.
const _superHighIndexNumForSort = 100000

// SortAttributes sorts a slice of attributes for display.
// Sort order:
// 1. By index number (attributes with indexes come first, sorted by first index)
// 2. Non-derived attributes before derived attributes
// 3. By name alphabetically
func SortAttributes(attributes []model_class.Attribute) {
	sort.Slice(attributes, func(i, j int) bool {
		// First, if one has an index and another doesn't use the index.
		// And if they both have indexes sort by the indexes.
		iIndexNum, jIndexNum := uint(_superHighIndexNumForSort), uint(_superHighIndexNumForSort)
		if len(attributes[i].IndexNums) > 0 {
			iIndexNum = attributes[i].IndexNums[0]
		}
		if len(attributes[j].IndexNums) > 0 {
			jIndexNum = attributes[j].IndexNums[0]
		}
		if iIndexNum != jIndexNum {
			return iIndexNum < jIndexNum
		}

		// Non-derived attributes before derived attributes.
		iDerived := attributes[i].DerivationPolicy
		jDerived := attributes[j].DerivationPolicy
		switch {
		case iDerived == nil && jDerived != nil:
			return true // i is first.
		case jDerived == nil && iDerived != nil:
			return false // j is first.
		}

		// Then order by name.
		return attributes[i].Name < attributes[j].Name
	})
}

// GetAttributesSorted converts a map of attributes to a sorted slice.
// This is a convenience function for templates that need sorted attributes.
func GetAttributesSorted(attributes map[identity.Key]model_class.Attribute) []model_class.Attribute {
	result := make([]model_class.Attribute, 0, len(attributes))
	for _, attr := range attributes {
		result = append(result, attr)
	}
	SortAttributes(result)
	return result
}

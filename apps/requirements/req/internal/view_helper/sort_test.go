package view_helper

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_logic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestSortSuite(t *testing.T) {
	suite.Run(t, new(SortSuite))
}

type SortSuite struct {
	suite.Suite
}

// TestSortAttributes tests the sorting logic for attributes.
func (suite *SortSuite) TestSortAttributes() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))

	// Create test attributes with different characteristics.
	attrWithIndex0 := model_class.Attribute{
		Key:       helper.Must(identity.NewAttributeKey(classKey, "indexed0")),
		Name:      "Indexed0",
		IndexNums: []uint{0},
	}
	attrWithIndex1 := model_class.Attribute{
		Key:       helper.Must(identity.NewAttributeKey(classKey, "indexed1")),
		Name:      "Indexed1",
		IndexNums: []uint{1},
	}
	attrWithIndex2 := model_class.Attribute{
		Key:       helper.Must(identity.NewAttributeKey(classKey, "indexed2")),
		Name:      "Indexed2",
		IndexNums: []uint{2},
	}
	attrNoIndexA := model_class.Attribute{
		Key:  helper.Must(identity.NewAttributeKey(classKey, "no_index_a")),
		Name: "NoIndexA",
	}
	attrNoIndexB := model_class.Attribute{
		Key:  helper.Must(identity.NewAttributeKey(classKey, "no_index_b")),
		Name: "NoIndexB",
	}
	attrDerived := model_class.Attribute{
		Key:              helper.Must(identity.NewAttributeKey(classKey, "derived")),
		Name:             "Derived",
		DerivationPolicy: &model_logic.Logic{Description: "some derivation", Notation: "tla_plus"},
	}
	attrDerivedWithIndex := model_class.Attribute{
		Key:              helper.Must(identity.NewAttributeKey(classKey, "derived_indexed")),
		Name:             "DerivedIndexed",
		IndexNums:        []uint{3},
		DerivationPolicy: &model_logic.Logic{Description: "some derivation", Notation: "tla_plus"},
	}

	tests := []struct {
		name     string
		input    []model_class.Attribute
		expected []string // Expected order of attribute names
	}{
		{
			name:     "sort by index number",
			input:    []model_class.Attribute{attrWithIndex2, attrWithIndex0, attrWithIndex1},
			expected: []string{"Indexed0", "Indexed1", "Indexed2"},
		},
		{
			name:     "indexed attributes before non-indexed",
			input:    []model_class.Attribute{attrNoIndexA, attrWithIndex0, attrNoIndexB},
			expected: []string{"Indexed0", "NoIndexA", "NoIndexB"},
		},
		{
			name:     "non-derived before derived (same index level)",
			input:    []model_class.Attribute{attrDerived, attrNoIndexA},
			expected: []string{"NoIndexA", "Derived"},
		},
		{
			name:     "alphabetical order for same characteristics",
			input:    []model_class.Attribute{attrNoIndexB, attrNoIndexA},
			expected: []string{"NoIndexA", "NoIndexB"},
		},
		{
			name:     "complex sort: index > derived > alphabetical",
			input:    []model_class.Attribute{attrDerived, attrNoIndexB, attrWithIndex1, attrNoIndexA, attrWithIndex0, attrDerivedWithIndex},
			expected: []string{"Indexed0", "Indexed1", "DerivedIndexed", "NoIndexA", "NoIndexB", "Derived"},
		},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			// Make a copy to avoid modifying the original.
			attrs := make([]model_class.Attribute, len(tt.input))
			copy(attrs, tt.input)

			SortAttributes(attrs)

			actualNames := make([]string, len(attrs))
			for i, attr := range attrs {
				actualNames[i] = attr.Name
			}
			assert.Equal(t, tt.expected, actualNames)
		})
	}
}

// TestGetAttributesSorted tests the convenience function that converts a map to sorted slice.
func (suite *SortSuite) TestGetAttributesSorted() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))

	attr1Key := helper.Must(identity.NewAttributeKey(classKey, "attr1"))
	attr2Key := helper.Must(identity.NewAttributeKey(classKey, "attr2"))
	attr3Key := helper.Must(identity.NewAttributeKey(classKey, "attr3"))

	attrs := map[identity.Key]model_class.Attribute{
		attr3Key: {Key: attr3Key, Name: "Zebra"},
		attr1Key: {Key: attr1Key, Name: "Apple", IndexNums: []uint{0}},
		attr2Key: {Key: attr2Key, Name: "Banana"},
	}

	sorted := GetAttributesSorted(attrs)

	assert.Len(suite.T(), sorted, 3)
	// Indexed first, then alphabetical.
	assert.Equal(suite.T(), "Apple", sorted[0].Name)
	assert.Equal(suite.T(), "Banana", sorted[1].Name)
	assert.Equal(suite.T(), "Zebra", sorted[2].Name)
}

package core_test

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/test_helper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetTestModelExploresAssociationUniqueness ensures the shared fixture validates
// association uniqueness through the full model tree, not only unit-level checks.
func TestGetTestModelExploresAssociationUniqueness(t *testing.T) {
	model := test_helper.GetTestModel()
	require.NoError(t, model.Validate())

	var fromOnly, toOnly, bothSides int
	for _, assoc := range model.GetClassAssociations() {
		for i, uniqueness := range assoc.Uniqueness {
			fromCount := len(uniqueness.FromAttributeKeys)
			toCount := len(uniqueness.ToAttributeKeys)
			require.Positive(t, fromCount+toCount, "association %q uniqueness[%d] must list at least one uniqueness attribute", assoc.Name, i)
			switch {
			case fromCount > 0 && toCount > 0:
				bothSides++
			case fromCount > 0:
				fromOnly++
			case toCount > 0:
				toOnly++
			}
		}
	}

	assert.Positive(t, fromOnly)
	assert.Positive(t, toOnly)
	assert.Positive(t, bothSides)
}

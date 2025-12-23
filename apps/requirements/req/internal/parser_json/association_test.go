package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/stretchr/testify/assert"
)

func TestAssociationInOutRoundTrip(t *testing.T) {
	original := requirements.Association{
		Key:                 "assoc1",
		Name:                "Assoc1",
		Details:             "Details",
		FromClassKey:        "class1",
		FromMultiplicity:    requirements.Multiplicity{LowerBound: 1, HigherBound: 1},
		ToClassKey:          "class2",
		ToMultiplicity:      requirements.Multiplicity{LowerBound: 0, HigherBound: 5},
		AssociationClassKey: "aclass",
		UmlComment:          "comment",
	}

	inOut := FromRequirementsAssociation(original)
	back := inOut.ToRequirements()

	assert.Equal(t, original.Key, back.Key)
	assert.Equal(t, original.Name, back.Name)
	assert.Equal(t, original.Details, back.Details)
	assert.Equal(t, original.FromClassKey, back.FromClassKey)
	assert.Equal(t, original.ToClassKey, back.ToClassKey)
	assert.Equal(t, original.AssociationClassKey, back.AssociationClassKey)
	assert.Equal(t, original.UmlComment, back.UmlComment)
	assert.Equal(t, original.FromMultiplicity, back.FromMultiplicity)
	assert.Equal(t, original.ToMultiplicity, back.ToMultiplicity)
}

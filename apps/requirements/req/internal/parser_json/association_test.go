package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_class"
	"github.com/stretchr/testify/assert"
)

func TestAssociationInOutRoundTrip(t *testing.T) {
	original := model_class.Association{
		Key:                 "assoc1",
		Name:                "Assoc1",
		Details:             "Details",
		FromClassKey:        "class1",
		FromMultiplicity:    model_class.Multiplicity{LowerBound: 1, HigherBound: 1},
		ToClassKey:          "class2",
		ToMultiplicity:      model_class.Multiplicity{LowerBound: 0, HigherBound: 5},
		AssociationClassKey: "aclass",
		UmlComment:          "comment",
	}

	inOut := FromRequirementsAssociation(original)
	back := inOut.ToRequirements()
	assert.Equal(t, original, back)
}

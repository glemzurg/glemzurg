package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
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

	if back.Key != original.Key || back.Name != original.Name || back.Details != original.Details ||
		back.FromClassKey != original.FromClassKey || back.ToClassKey != original.ToClassKey ||
		back.AssociationClassKey != original.AssociationClassKey || back.UmlComment != original.UmlComment ||
		back.FromMultiplicity != original.FromMultiplicity || back.ToMultiplicity != original.ToMultiplicity {
		t.Errorf("Round trip failed: got %+v, want %+v", back, original)
	}
}

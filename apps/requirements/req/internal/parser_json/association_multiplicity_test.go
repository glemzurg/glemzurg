package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
)

func TestMultiplicityInOutRoundTrip(t *testing.T) {
	original := requirements.Multiplicity{
		LowerBound:  1,
		HigherBound: 5,
	}

	inOut := FromRequirementsMultiplicity(original)
	back := inOut.ToRequirements()

	if back != original {
		t.Errorf("Round trip failed: got %+v, want %+v", back, original)
	}
}

package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/stretchr/testify/assert"
)

func TestMultiplicityInOutRoundTrip(t *testing.T) {
	original := requirements.Multiplicity{
		LowerBound:  1,
		HigherBound: 5,
	}

	inOut := FromRequirementsMultiplicity(original)
	back := inOut.ToRequirements()

	assert.Equal(t, original.LowerBound, back.LowerBound)
	assert.Equal(t, original.HigherBound, back.HigherBound)
}

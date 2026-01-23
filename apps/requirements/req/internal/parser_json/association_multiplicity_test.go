package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/stretchr/testify/assert"
)

func TestMultiplicityInOutRoundTrip(t *testing.T) {
	original := model_class.Multiplicity{
		LowerBound:  1,
		HigherBound: 5,
	}

	inOut := FromRequirementsMultiplicity(original)
	back := inOut.ToRequirements()
	assert.Equal(t, original, back)
}

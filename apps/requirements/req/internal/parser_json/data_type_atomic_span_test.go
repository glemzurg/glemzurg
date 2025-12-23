package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/data_type"
	"github.com/stretchr/testify/assert"
)

func TestAtomicSpanInOutRoundTrip(t *testing.T) {
	lowerVal := 1
	lowerDenom := 2
	higherVal := 10
	higherDenom := 20
	original := data_type.AtomicSpan{
		LowerType:         "closed",
		LowerValue:        &lowerVal,
		LowerDenominator:  &lowerDenom,
		HigherType:        "open",
		HigherValue:       &higherVal,
		HigherDenominator: &higherDenom,
		Units:             "kg",
		Precision:         0.1,
	}

	inOut := FromRequirementsAtomicSpan(original)
	back := inOut.ToRequirements()
	assert.Equal(t, original, back)
}

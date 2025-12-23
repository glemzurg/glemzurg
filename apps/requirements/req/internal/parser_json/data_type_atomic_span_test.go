package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/data_type"
	"github.com/stretchr/testify/assert"
)

func TestAtomicSpanInOutRoundTrip(t *testing.T) {
	lowerVal := 1
	higherVal := 10
	original := data_type.AtomicSpan{
		LowerType:         "closed",
		LowerValue:        &lowerVal,
		LowerDenominator:  nil,
		HigherType:        "open",
		HigherValue:       &higherVal,
		HigherDenominator: nil,
		Units:             "kg",
		Precision:         0.1,
	}

	inOut := FromRequirementsAtomicSpan(original)
	back := inOut.ToRequirements()

	assert.Equal(t, original.LowerType, back.LowerType)
	assert.NotNil(t, back.LowerValue)
	assert.Equal(t, *original.LowerValue, *back.LowerValue)
	assert.Nil(t, back.LowerDenominator)
	assert.Equal(t, original.HigherType, back.HigherType)
	assert.NotNil(t, back.HigherValue)
	assert.Equal(t, *original.HigherValue, *back.HigherValue)
	assert.Nil(t, back.HigherDenominator)
	assert.Equal(t, original.Units, back.Units)
	assert.Equal(t, original.Precision, back.Precision)
}

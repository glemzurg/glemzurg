package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/data_type"
)

func TestAtomicSpanInOutRoundTrip(t *testing.T) {
	lowerVal := 1
	higherVal := 10
	original := data_type.AtomicSpan{
		LowerType:        "closed",
		LowerValue:       &lowerVal,
		LowerDenominator: nil,
		HigherType:        "open",
		HigherValue:       &higherVal,
		HigherDenominator: nil,
		Units:             "kg",
		Precision:         0.1,
	}

	inOut := FromRequirementsAtomicSpan(original)
	back := inOut.ToRequirements()

	if back.LowerType != original.LowerType || (back.LowerValue == nil && original.LowerValue != nil) || (back.LowerValue != nil && *back.LowerValue != *original.LowerValue) ||
		back.HigherType != original.HigherType || (back.HigherValue == nil && original.HigherValue != nil) || (back.HigherValue != nil && *back.HigherValue != *original.HigherValue) ||
		back.Units != original.Units || back.Precision != original.Precision {
		t.Errorf("Round trip failed: got %+v, want %+v", back, original)
	}
}
package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_data_type"
	"github.com/stretchr/testify/assert"
)

func TestAtomicSpanInOutRoundTrip(t *testing.T) {

	original := model_data_type.AtomicSpan{
		LowerType:         "closed",
		LowerValue:        t_IntPtr(1),
		LowerDenominator:  t_IntPtr(2),
		HigherType:        "open",
		HigherValue:       t_IntPtr(10),
		HigherDenominator: t_IntPtr(20),
		Units:             "kg",
		Precision:         0.1,
	}

	inOut := FromRequirementsAtomicSpan(original)
	back := inOut.ToRequirements()
	assert.Equal(t, original, back)
}

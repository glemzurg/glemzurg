package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/data_type"
	"github.com/stretchr/testify/assert"
)

func TestAtomicEnumInOutRoundTrip(t *testing.T) {
	original := data_type.AtomicEnum{
		Value:     "value1",
		SortOrder: 1,
	}

	inOut := FromRequirementsAtomicEnum(original)
	back := inOut.ToRequirements()
	assert.Equal(t, original, back)
}

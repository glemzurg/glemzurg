package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/stretchr/testify/assert"
)

func TestStateActionInOutRoundTrip(t *testing.T) {
	originalReq := requirements.StateAction{
		Key:       "state_action1",
		ActionKey: "action1",
		When:      "entry",
		StateKey:  "state1", // This won't be preserved in round trip
	}

	// Convert to InOut
	inOut := FromRequirementsStateAction(originalReq)

	// Convert back to requirements
	convertedBack := inOut.ToRequirements()

	// Check fields
	assert.Equal(t, originalReq.Key, convertedBack.Key)
	assert.Equal(t, originalReq.ActionKey, convertedBack.ActionKey)
	assert.Equal(t, originalReq.When, convertedBack.When)

	// Note: StateKey is not stored in JSON, so it's empty in convertedBack
}

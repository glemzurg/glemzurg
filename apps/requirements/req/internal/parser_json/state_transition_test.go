package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/stretchr/testify/assert"
)

func TestTransitionInOutRoundTrip(t *testing.T) {
	originalReq := requirements.Transition{
		Key:          "transition1",
		FromStateKey: "state1",
		EventKey:     "event1",
		GuardKey:     "guard1",
		ActionKey:    "action1",
		ToStateKey:   "state2",
		UmlComment:   "Transition comment",
	}

	// Convert to InOut
	inOut := FromRequirementsTransition(originalReq)

	// Convert back to requirements
	convertedBack := inOut.ToRequirements()

	// Check fields
	assert.Equal(t, originalReq.Key, convertedBack.Key)
	assert.Equal(t, originalReq.FromStateKey, convertedBack.FromStateKey)
	assert.Equal(t, originalReq.EventKey, convertedBack.EventKey)
	assert.Equal(t, originalReq.GuardKey, convertedBack.GuardKey)
	assert.Equal(t, originalReq.ActionKey, convertedBack.ActionKey)
	assert.Equal(t, originalReq.ToStateKey, convertedBack.ToStateKey)
	assert.Equal(t, originalReq.UmlComment, convertedBack.UmlComment)
}

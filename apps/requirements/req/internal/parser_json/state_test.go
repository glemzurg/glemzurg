package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/stretchr/testify/assert"
)

func TestStateInOutRoundTrip(t *testing.T) {
	originalReq := requirements.State{
		Key:        "state1",
		Name:       "Initial State",
		Details:    "The starting state",
		UmlComment: "State comment",
		Actions: []requirements.StateAction{
			{
				Key:       "action1",
				ActionKey: "entry_action",
				When:      "entry",
			},
			{
				Key:       "action2",
				ActionKey: "exit_action",
				When:      "exit",
			},
		},
	}

	// Convert to InOut
	inOut := FromRequirementsState(originalReq)

	// Convert back to requirements
	convertedBack := inOut.ToRequirements()

	// Check fields
	assert.Equal(t, originalReq.Key, convertedBack.Key)
	assert.Equal(t, originalReq.Name, convertedBack.Name)
	assert.Equal(t, originalReq.Details, convertedBack.Details)
	assert.Equal(t, originalReq.UmlComment, convertedBack.UmlComment)
	assert.Len(t, convertedBack.Actions, len(originalReq.Actions))

	for i, action := range originalReq.Actions {
		assert.Equal(t, action.Key, convertedBack.Actions[i].Key)
		assert.Equal(t, action.ActionKey, convertedBack.Actions[i].ActionKey)
		assert.Equal(t, action.When, convertedBack.Actions[i].When)
	}
}

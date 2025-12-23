package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
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
	if convertedBack.Key != originalReq.Key ||
		convertedBack.Name != originalReq.Name ||
		convertedBack.Details != originalReq.Details ||
		convertedBack.UmlComment != originalReq.UmlComment ||
		len(convertedBack.Actions) != len(originalReq.Actions) {
		t.Errorf("Round trip failed: got %+v, want %+v", convertedBack, originalReq)
	}

	for i, action := range originalReq.Actions {
		if convertedBack.Actions[i].Key != action.Key ||
			convertedBack.Actions[i].ActionKey != action.ActionKey ||
			convertedBack.Actions[i].When != action.When {
			t.Errorf("Action[%d] mismatch: got %+v, want %+v", i, convertedBack.Actions[i], action)
		}
	}
}

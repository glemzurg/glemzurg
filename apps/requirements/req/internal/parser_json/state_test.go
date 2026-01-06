package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/stretchr/testify/assert"
)

func TestStateInOutRoundTrip(t *testing.T) {
	original := requirements.State{
		Key:        "state1",
		Name:       "Initial State",
		Details:    "The starting state",
		UmlComment: "State comment",
		Actions: []requirements.StateAction{
			{
				Key: "action1",
			},
			{
				Key: "action2",
			},
		},
	}

	inOut := FromRequirementsState(original)
	back := inOut.ToRequirements()
	assert.Equal(t, original, back)
}

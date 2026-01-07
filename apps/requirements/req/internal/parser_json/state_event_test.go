package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_state"
	"github.com/stretchr/testify/assert"
)

func TestEventInOutRoundTrip(t *testing.T) {
	original := model_state.Event{
		Key:     "event1",
		Name:    "Login Event",
		Details: "User attempts to log in",
		Parameters: []model_state.EventParameter{
			{
				Name: "username",
			},
			{
				Name: "password",
			},
		},
	}

	inOut := FromRequirementsEvent(original)
	back := inOut.ToRequirements()
	assert.Equal(t, original, back)
}

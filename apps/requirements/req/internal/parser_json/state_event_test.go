package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/stretchr/testify/assert"
)

func TestEventInOutRoundTrip(t *testing.T) {
	originalReq := requirements.Event{
		Key:     "event1",
		Name:    "Login Event",
		Details: "User attempts to log in",
		Parameters: []requirements.EventParameter{
			{
				Name:   "username",
				Source: "user_input",
			},
			{
				Name:   "password",
				Source: "user_input",
			},
		},
	}

	// Convert to InOut
	inOut := FromRequirementsEvent(originalReq)

	// Convert back to requirements
	convertedBack := inOut.ToRequirements()

	// Check fields
	assert.Equal(t, originalReq.Key, convertedBack.Key)
	assert.Equal(t, originalReq.Name, convertedBack.Name)
	assert.Equal(t, originalReq.Details, convertedBack.Details)
	assert.Len(t, convertedBack.Parameters, len(originalReq.Parameters))

	for i, param := range originalReq.Parameters {
		assert.Equal(t, param.Name, convertedBack.Parameters[i].Name)
		assert.Equal(t, param.Source, convertedBack.Parameters[i].Source)
	}
}

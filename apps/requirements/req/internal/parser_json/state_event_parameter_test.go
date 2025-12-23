package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/stretchr/testify/assert"
)

func TestEventParameterInOutRoundTrip(t *testing.T) {
	originalReq := requirements.EventParameter{
		Name:   "username",
		Source: "user_input",
	}

	// Convert to InOut
	inOut := FromRequirementsEventParameter(originalReq)

	// Convert back to requirements
	convertedBack := inOut.ToRequirements()

	// Check fields
	assert.Equal(t, originalReq.Name, convertedBack.Name)
	assert.Equal(t, originalReq.Source, convertedBack.Source)
}

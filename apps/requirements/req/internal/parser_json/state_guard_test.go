package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/stretchr/testify/assert"
)

func TestGuardInOutRoundTrip(t *testing.T) {
	originalReq := requirements.Guard{
		Key:     "guard1",
		Name:    "Authenticated",
		Details: "User must be authenticated",
	}

	// Convert to InOut
	inOut := FromRequirementsGuard(originalReq)

	// Convert back to requirements
	convertedBack := inOut.ToRequirements()

	// Check fields
	assert.Equal(t, originalReq.Key, convertedBack.Key)
	assert.Equal(t, originalReq.Name, convertedBack.Name)
	assert.Equal(t, originalReq.Details, convertedBack.Details)
}

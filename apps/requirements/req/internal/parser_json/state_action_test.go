package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/stretchr/testify/assert"
)

func TestActionInOutRoundTrip(t *testing.T) {
	originalReq := requirements.Action{
		Key:        "action1",
		Name:       "Login Action",
		Details:    "User logs in",
		Requires:   []string{"user_authenticated"},
		Guarantees: []string{"session_created"},
	}

	// Convert to InOut
	inOut := FromRequirementsAction(originalReq)

	// Convert back to requirements
	convertedBack := inOut.ToRequirements()

	// Check fields
	assert.Equal(t, originalReq.Key, convertedBack.Key)
	assert.Equal(t, originalReq.Name, convertedBack.Name)
	assert.Equal(t, originalReq.Details, convertedBack.Details)
	assert.Len(t, convertedBack.Requires, len(originalReq.Requires))
	assert.Len(t, convertedBack.Guarantees, len(originalReq.Guarantees))

	for i, req := range originalReq.Requires {
		assert.Equal(t, req, convertedBack.Requires[i])
	}

	for i, gua := range originalReq.Guarantees {
		assert.Equal(t, gua, convertedBack.Guarantees[i])
	}
}

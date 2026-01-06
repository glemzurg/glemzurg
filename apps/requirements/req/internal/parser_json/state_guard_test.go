package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/stretchr/testify/assert"
)

func TestGuardInOutRoundTrip(t *testing.T) {
	original := requirements.Guard{
		Key:     "guard1",
		Name:    "Authenticated",
		Details: "User must be authenticated",
	}

	inOut := FromRequirementsGuard(original)
	back := inOut.ToRequirements()
	assert.Equal(t, original, back)
}

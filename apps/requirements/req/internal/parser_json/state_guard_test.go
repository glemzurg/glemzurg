package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/state"
	"github.com/stretchr/testify/assert"
)

func TestGuardInOutRoundTrip(t *testing.T) {
	original := state.Guard{
		Key:     "guard1",
		Name:    "Authenticated",
		Details: "User must be authenticated",
	}

	inOut := FromRequirementsGuard(original)
	back := inOut.ToRequirements()
	assert.Equal(t, original, back)
}

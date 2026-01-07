package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_state"
	"github.com/stretchr/testify/assert"
)

func TestActionInOutRoundTrip(t *testing.T) {
	original := model_state.Action{
		Key:        "action1",
		Name:       "Login Action",
		Details:    "User logs in",
		Requires:   []string{"user_authenticated"},
		Guarantees: []string{"session_created"},
	}

	inOut := FromRequirementsAction(original)
	back := inOut.ToRequirements()
	assert.Equal(t, original, back)
}

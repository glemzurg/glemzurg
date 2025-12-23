package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/stretchr/testify/assert"
)

func TestStateActionInOutRoundTrip(t *testing.T) {
	original := requirements.StateAction{
		Key:       "state_action1",
		ActionKey: "action1",
		When:      "entry",
	}

	inOut := FromRequirementsStateAction(original)
	back := inOut.ToRequirements()
	assert.Equal(t, original, back)
}

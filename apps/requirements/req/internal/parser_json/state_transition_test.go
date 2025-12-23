package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/stretchr/testify/assert"
)

func TestTransitionInOutRoundTrip(t *testing.T) {
	original := requirements.Transition{
		Key:          "transition1",
		FromStateKey: "state1",
		EventKey:     "event1",
		GuardKey:     "guard1",
		ActionKey:    "action1",
		ToStateKey:   "state2",
		UmlComment:   "Transition comment",
	}

	inOut := FromRequirementsTransition(original)
	back := inOut.ToRequirements()
	assert.Equal(t, original, back)
}

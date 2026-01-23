package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"
	"github.com/stretchr/testify/assert"
)

func TestEventParameterInOutRoundTrip(t *testing.T) {
	original := model_state.EventParameter{
		Name:   "username",
		Source: "user_input",
	}

	inOut := FromRequirementsEventParameter(original)
	back := inOut.ToRequirements()
	assert.Equal(t, original, back)
}

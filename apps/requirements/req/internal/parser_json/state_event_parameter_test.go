package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/stretchr/testify/assert"
)

func TestEventParameterInOutRoundTrip(t *testing.T) {
	original := requirements.EventParameter{
		Name:   "username",
		Source: "user_input",
	}

	inOut := FromRequirementsEventParameter(original)
	back := inOut.ToRequirements()
	assert.Equal(t, original, back)
}

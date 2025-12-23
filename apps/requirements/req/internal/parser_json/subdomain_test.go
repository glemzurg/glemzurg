package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/stretchr/testify/assert"
)

func TestSubdomainInOutRoundTrip(t *testing.T) {
	original := requirements.Subdomain{
		Key:        "sub1",
		Name:       "Sub1",
		Details:    "Details",
		UmlComment: "comment",
	}

	inOut := FromRequirementsSubdomain(original)
	back := inOut.ToRequirements()

	assert.Equal(t, original, back)
}

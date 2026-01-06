package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/stretchr/testify/assert"
)

func TestDomainInOutRoundTrip(t *testing.T) {
	original := requirements.Domain{
		Key:        "domain1",
		Name:       "Domain1",
		Details:    "Details",
		Realized:   true,
		UmlComment: "comment",
	}

	inOut := FromRequirementsDomain(original)
	back := inOut.ToRequirements()
	assert.Equal(t, original, back)
}

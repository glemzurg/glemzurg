package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
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

	if back != original {
		t.Errorf("Round trip failed: got %+v, want %+v", back, original)
	}
}

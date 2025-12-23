package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
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

	if back.Key != original.Key || back.Name != original.Name || back.Details != original.Details ||
		back.Realized != original.Realized || back.UmlComment != original.UmlComment {
		t.Errorf("Round trip failed: got %+v, want %+v", back, original)
	}
}

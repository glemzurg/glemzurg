package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
)

func TestModelInOutRoundTrip(t *testing.T) {
	original := requirements.Requirements{
		Model: requirements.Model{
			Key:     "model1",
			Name:    "Test Model",
			Details: "Details",
		},
		Actors: []requirements.Actor{
			{Key: "actor1", Name: "User", Type: "person", UmlComment: "comment"},
		},
		Domains: []requirements.Domain{
			{Key: "domain1", Name: "Domain1", Realized: true, UmlComment: "comment"},
		},
		Subdomains: map[string][]requirements.Subdomain{
			"domain1": {{Key: "sub1", Name: "Sub1", UmlComment: "comment"}},
		},
		DomainAssociations: []requirements.DomainAssociation{
			{Key: "da1", ProblemDomainKey: "domain1", SolutionDomainKey: "domain2", UmlComment: "comment"},
		},
		Associations: []requirements.Association{
			{Key: "assoc1", Name: "Assoc1", FromClassKey: "class1", ToClassKey: "class2", UmlComment: "comment"},
		},
	}

	inOut := FromRequirementsModel(original)
	back := inOut.ToRequirements()

	// Check model
	if back.Model != original.Model {
		t.Errorf("Model round trip failed: got %+v, want %+v", back.Model, original.Model)
	}

	// Check actors
	if len(back.Actors) != len(original.Actors) {
		t.Errorf("Actors length: got %d, want %d", len(back.Actors), len(original.Actors))
	}

	// Check domains
	if len(back.Domains) != len(original.Domains) {
		t.Errorf("Domains length: got %d, want %d", len(back.Domains), len(original.Domains))
	}

	// Check subdomains
	if len(back.Subdomains) != len(original.Subdomains) {
		t.Errorf("Subdomains length: got %d, want %d", len(back.Subdomains), len(original.Subdomains))
	}

	// Check domain associations
	if len(back.DomainAssociations) != len(original.DomainAssociations) {
		t.Errorf("DomainAssociations length: got %d, want %d", len(back.DomainAssociations), len(original.DomainAssociations))
	}

	// Check associations
	if len(back.Associations) != len(original.Associations) {
		t.Errorf("Associations length: got %d, want %d", len(back.Associations), len(original.Associations))
	}
}

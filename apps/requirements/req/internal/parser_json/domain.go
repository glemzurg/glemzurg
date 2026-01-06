package parser_json

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"

// domainInOut is a root category of the model.
type domainInOut struct {
	Key        string `json:"key"`
	Name       string `json:"name"`
	Details    string `json:"details"` // Markdown.
	Realized   bool   `json:"realized"`
	UmlComment string `json:"uml_comment"`
	// Nested.
	Subdomains []subdomainInOut `json:"subdomains"`
}

// ToRequirements converts the domainInOut to requirements.Domain.
func (d domainInOut) ToRequirements() requirements.Domain {
	domain := requirements.Domain{
		Key:        d.Key,
		Name:       d.Name,
		Details:    d.Details,
		Realized:   d.Realized,
		UmlComment: d.UmlComment,
		Subdomains: nil,
	}

	for _, s := range d.Subdomains {
		domain.Subdomains = append(domain.Subdomains, s.ToRequirements())
	}
	return domain
}

// FromRequirements creates a domainInOut from requirements.Domain.
func FromRequirementsDomain(d requirements.Domain) domainInOut {
	domain := domainInOut{
		Key:        d.Key,
		Name:       d.Name,
		Details:    d.Details,
		Realized:   d.Realized,
		UmlComment: d.UmlComment,
		Subdomains: nil,
	}
	for _, s := range d.Subdomains {
		domain.Subdomains = append(domain.Subdomains, FromRequirementsSubdomain(s))
	}
	return domain
}

package parser_json

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"
)

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

// ToRequirements converts the domainInOut to model_domain.Domain.
func (d domainInOut) ToRequirements() (model_domain.Domain, error) {
	key, err := identity.ParseKey(d.Key)
	if err != nil {
		return model_domain.Domain{}, err
	}

	domain := model_domain.Domain{
		Key:        key,
		Name:       d.Name,
		Details:    d.Details,
		Realized:   d.Realized,
		UmlComment: d.UmlComment,
	}

	for _, s := range d.Subdomains {
		subdomain, err := s.ToRequirements()
		if err != nil {
			return model_domain.Domain{}, err
		}
		if domain.Subdomains == nil {
			domain.Subdomains = make(map[identity.Key]model_domain.Subdomain)
		}
		domain.Subdomains[subdomain.Key] = subdomain
	}
	return domain, nil
}

// FromRequirements creates a domainInOut from model_domain.Domain.
func FromRequirementsDomain(d model_domain.Domain) domainInOut {
	domain := domainInOut{
		Key:        d.Key.String(),
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

package parser_json

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"

// modelInOut is the documentation summary of a set of requirements.
type modelInOut struct {
	Key     string `json:"key"`
	Name    string `json:"name"`
	Details string `json:"details"` // Markdown.
	// Nested structure.
	Actors             []actorInOut             `json:"actors"`
	Domains            []domainInOut            `json:"domains"`
	DomainAssociations []domainAssociationInOut `json:"domain_associations"`
	Associations       []associationInOut       `json:"associations"` // Associations between classes that span domains.
}

// ToRequirements converts the modelInOut to requirements.Requirements.
func (m modelInOut) ToRequirements() requirements.Requirements {
	r := requirements.Requirements{
		Model: requirements.Model{
			Key:     m.Key,
			Name:    m.Name,
			Details: m.Details,
		},
		Subdomains: make(map[string][]requirements.Subdomain),
	}

	// Convert actors
	for _, a := range m.Actors {
		r.Actors = append(r.Actors, a.ToRequirements())
	}

	// Convert domains and subdomains
	for _, d := range m.Domains {
		domain := d.ToRequirements()
		r.Domains = append(r.Domains, domain)
		// Subdomains are handled in domain.ToRequirements, but since Requirements has Subdomains map,
		// we need to populate it here
		for _, s := range d.Subdomains {
			r.Subdomains[d.Key] = append(r.Subdomains[d.Key], s.ToRequirements())
		}
	}

	// Domain associations
	for _, da := range m.DomainAssociations {
		r.DomainAssociations = append(r.DomainAssociations, da.ToRequirements())
	}

	// Associations
	for _, a := range m.Associations {
		r.Associations = append(r.Associations, a.ToRequirements())
	}

	// Note: Other maps like Classes, Attributes, etc. are not populated here as they are nested in subdomains
	// This is a simplified version; in a full implementation, we would need to traverse all nested structures

	return r
}

// FromRequirements creates a modelInOut from requirements.Requirements.
func FromRequirementsModel(r requirements.Requirements) modelInOut {
	m := modelInOut{
		Key:     r.Model.Key,
		Name:    r.Model.Name,
		Details: r.Model.Details,
	}

	// Convert actors
	for _, a := range r.Actors {
		m.Actors = append(m.Actors, FromRequirementsActor(a))
	}

	// Convert domains
	for _, d := range r.Domains {
		domain := FromRequirementsDomain(d)
		// Add subdomains
		if subs, ok := r.Subdomains[d.Key]; ok {
			for _, s := range subs {
				domain.Subdomains = append(domain.Subdomains, FromRequirementsSubdomain(s))
			}
		}
		m.Domains = append(m.Domains, domain)
	}

	// Domain associations
	for _, da := range r.DomainAssociations {
		m.DomainAssociations = append(m.DomainAssociations, FromRequirementsDomainAssociation(da))
	}

	// Associations
	for _, a := range r.Associations {
		m.Associations = append(m.Associations, FromRequirementsAssociation(a))
	}

	return m
}

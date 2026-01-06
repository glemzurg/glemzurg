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
func (m modelInOut) ToRequirements() requirements.Model {
	model := requirements.Model{
		Key:                m.Key,
		Name:               m.Name,
		Details:            m.Details,
		Actors:             nil,
		Domains:            nil,
		DomainAssociations: nil,
		Associations:       nil,
	}

	// Convert actors
	for _, a := range m.Actors {
		model.Actors = append(model.Actors, a.ToRequirements())
	}

	// Convert domains and subdomains
	for _, d := range m.Domains {
		model.Domains = append(model.Domains, d.ToRequirements())
	}

	// Domain associations
	for _, da := range m.DomainAssociations {
		model.DomainAssociations = append(model.DomainAssociations, da.ToRequirements())
	}

	// Associations
	for _, a := range m.Associations {
		model.Associations = append(model.Associations, a.ToRequirements())
	}

	return model
}

// FromRequirements creates a modelInOut from requirements.Model.
func FromRequirementsModel(r requirements.Model) modelInOut {
	m := modelInOut{
		Key:                r.Key,
		Name:               r.Name,
		Details:            r.Details,
		Actors:             nil,
		Domains:            nil,
		DomainAssociations: nil,
		Associations:       nil,
	}

	// Convert actors
	for _, a := range r.Actors {
		m.Actors = append(m.Actors, FromRequirementsActor(a))
	}

	// Convert domains
	for _, d := range r.Domains {
		m.Domains = append(m.Domains, FromRequirementsDomain(d))
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

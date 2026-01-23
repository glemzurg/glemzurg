package parser_json

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"
)

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
func (m modelInOut) ToRequirements() (req_model.Model, error) {
	model := req_model.Model{
		Key:     m.Key,
		Name:    m.Name,
		Details: m.Details,
	}

	// Convert actors
	for _, a := range m.Actors {
		actor, err := a.ToRequirements()
		if err != nil {
			return req_model.Model{}, err
		}
		if model.Actors == nil {
			model.Actors = make(map[identity.Key]model_actor.Actor)
		}
		model.Actors[actor.Key] = actor
	}

	// Convert domains and subdomains
	for _, d := range m.Domains {
		domain, err := d.ToRequirements()
		if err != nil {
			return req_model.Model{}, err
		}
		if model.Domains == nil {
			model.Domains = make(map[identity.Key]model_domain.Domain)
		}
		model.Domains[domain.Key] = domain
	}

	// Domain associations
	for _, da := range m.DomainAssociations {
		domainAssoc, err := da.ToRequirements()
		if err != nil {
			return req_model.Model{}, err
		}
		if model.DomainAssociations == nil {
			model.DomainAssociations = make(map[identity.Key]model_domain.Association)
		}
		model.DomainAssociations[domainAssoc.Key] = domainAssoc
	}

	// Associations (model-level class associations)
	for _, a := range m.Associations {
		assoc, err := a.ToRequirements()
		if err != nil {
			return req_model.Model{}, err
		}
		if model.ClassAssociations == nil {
			model.ClassAssociations = make(map[identity.Key]model_class.Association)
		}
		model.ClassAssociations[assoc.Key] = assoc
	}

	return model, nil
}

// FromRequirements creates a modelInOut from req_model.Model.
func FromRequirementsModel(r req_model.Model) modelInOut {
	m := modelInOut{
		Key:     r.Key,
		Name:    r.Name,
		Details: r.Details,
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

	// Associations (model-level class associations)
	for _, a := range r.ClassAssociations {
		m.Associations = append(m.Associations, FromRequirementsAssociation(a))
	}

	return m
}

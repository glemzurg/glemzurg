package parser_json

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"
)

// UnpackRequirements converts a req_model.Model into a tree of parser_json objects.
func UnpackRequirements(reqs req_model.Model) modelInOut {
	tree := modelInOut{
		Key:     reqs.Key,
		Name:    reqs.Name,
		Details: reqs.Details,
	}

	// Actors
	for _, a := range reqs.Actors {
		tree.Actors = append(tree.Actors, FromRequirementsActor(a))
	}

	// Domains (subdomains are nested inside)
	for _, domain := range reqs.Domains {
		tree.Domains = append(tree.Domains, FromRequirementsDomain(domain))
	}

	// Domain associations
	for _, da := range reqs.DomainAssociations {
		tree.DomainAssociations = append(tree.DomainAssociations, FromRequirementsDomainAssociation(da))
	}

	// Model-level class associations
	for _, a := range reqs.ClassAssociations {
		tree.Associations = append(tree.Associations, FromRequirementsAssociation(a))
	}

	return tree
}

// PackRequirements converts a tree of parser_json objects back into req_model.Model.
func PackRequirements(tree modelInOut) (req_model.Model, error) {
	reqs := req_model.Model{
		Key:     tree.Key,
		Name:    tree.Name,
		Details: tree.Details,
	}

	// Actors
	for _, a := range tree.Actors {
		actor, err := a.ToRequirements()
		if err != nil {
			return req_model.Model{}, err
		}
		if reqs.Actors == nil {
			reqs.Actors = make(map[identity.Key]model_actor.Actor)
		}
		reqs.Actors[actor.Key] = actor
	}

	// Domains (subdomains, classes, etc. are nested inside)
	for _, d := range tree.Domains {
		domain, err := d.ToRequirements()
		if err != nil {
			return req_model.Model{}, err
		}
		if reqs.Domains == nil {
			reqs.Domains = make(map[identity.Key]model_domain.Domain)
		}
		reqs.Domains[domain.Key] = domain
	}

	// Domain Associations
	for _, da := range tree.DomainAssociations {
		domainAssoc, err := da.ToRequirements()
		if err != nil {
			return req_model.Model{}, err
		}
		if reqs.DomainAssociations == nil {
			reqs.DomainAssociations = make(map[identity.Key]model_domain.Association)
		}
		reqs.DomainAssociations[domainAssoc.Key] = domainAssoc
	}

	// Model-level class associations
	for _, a := range tree.Associations {
		assoc, err := a.ToRequirements()
		if err != nil {
			return req_model.Model{}, err
		}
		if reqs.ClassAssociations == nil {
			reqs.ClassAssociations = make(map[identity.Key]model_class.Association)
		}
		reqs.ClassAssociations[assoc.Key] = assoc
	}

	return reqs, nil
}

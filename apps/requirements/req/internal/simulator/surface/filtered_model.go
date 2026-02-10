package surface

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"
)

// BuildFilteredModel creates a new Model containing only the classes,
// associations, and invariants from the resolved surface. The original
// model is not modified.
func BuildFilteredModel(original *req_model.Model, resolved *ResolvedSurface) *req_model.Model {
	filtered := &req_model.Model{
		Key:            original.Key,
		Name:           original.Name,
		Details:        original.Details,
		TlaInvariants:  resolved.ModelInvariants,
		GlobalFunctions: original.GlobalFunctions,
		Actors:         original.Actors,
	}

	// Rebuild domain/subdomain/class tree with only included classes.
	filteredDomains := make(map[identity.Key]model_domain.Domain)
	for domainKey, domain := range original.Domains {
		filteredSubdomains := make(map[identity.Key]model_domain.Subdomain)
		for subdomainKey, subdomain := range domain.Subdomains {
			filteredClasses := make(map[identity.Key]model_class.Class)
			for classKey, class := range subdomain.Classes {
				if _, inScope := resolved.Classes[classKey]; inScope {
					filteredClasses[classKey] = class
				}
			}
			if len(filteredClasses) > 0 {
				filteredSubdomains[subdomainKey] = model_domain.Subdomain{
					Key:               subdomainKey,
					Name:              subdomain.Name,
					Details:           subdomain.Details,
					UmlComment:        subdomain.UmlComment,
					Generalizations:   subdomain.Generalizations,
					Classes:           filteredClasses,
					UseCases:          subdomain.UseCases,
					ClassAssociations: filterAssociations(subdomain.ClassAssociations, resolved.Associations),
					UseCaseShares:     subdomain.UseCaseShares,
				}
			}
		}
		if len(filteredSubdomains) > 0 {
			filteredDomains[domainKey] = model_domain.Domain{
				Key:               domainKey,
				Name:              domain.Name,
				Details:           domain.Details,
				Realized:          domain.Realized,
				UmlComment:        domain.UmlComment,
				Subdomains:        filteredSubdomains,
				ClassAssociations: filterAssociations(domain.ClassAssociations, resolved.Associations),
			}
		}
	}
	filtered.Domains = filteredDomains

	// Filter model-level associations.
	filtered.ClassAssociations = filterAssociations(original.ClassAssociations, resolved.Associations)

	// Preserve domain associations.
	filtered.DomainAssociations = original.DomainAssociations

	return filtered
}

// filterAssociations keeps only associations that are in the resolved set.
func filterAssociations(
	source map[identity.Key]model_class.Association,
	resolved map[identity.Key]model_class.Association,
) map[identity.Key]model_class.Association {
	if len(source) == 0 {
		return nil
	}
	result := make(map[identity.Key]model_class.Association)
	for k, v := range source {
		if _, inScope := resolved[k]; inScope {
			result[k] = v
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

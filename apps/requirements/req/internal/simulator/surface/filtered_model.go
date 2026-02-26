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
func BuildFilteredModel(original *req_model.Model, resolved *ResolvedSurface) (*req_model.Model, error) {
	filtered, err := req_model.NewModel(original.Key, original.Name, original.Details, resolved.ModelInvariants, original.GlobalFunctions)
	if err != nil {
		return nil, err
	}
	filtered.Actors = original.Actors

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
				filteredSub, err := model_domain.NewSubdomain(subdomainKey, subdomain.Name, subdomain.Details, subdomain.UmlComment)
				if err != nil {
					return nil, err
				}
				filteredSub.Generalizations = subdomain.Generalizations
				filteredSub.Classes = filteredClasses
				filteredSub.UseCases = subdomain.UseCases
				filteredSub.ClassAssociations = filterAssociations(subdomain.ClassAssociations, resolved.Associations)
				filteredSub.UseCaseShares = subdomain.UseCaseShares
				filteredSubdomains[subdomainKey] = filteredSub
			}
		}
		if len(filteredSubdomains) > 0 {
			filteredDom, err := model_domain.NewDomain(domainKey, domain.Name, domain.Details, domain.Realized, domain.UmlComment)
			if err != nil {
				return nil, err
			}
			filteredDom.Subdomains = filteredSubdomains
			filteredDom.ClassAssociations = filterAssociations(domain.ClassAssociations, resolved.Associations)
			filteredDomains[domainKey] = filteredDom
		}
	}
	filtered.Domains = filteredDomains

	// Filter model-level associations.
	filtered.ClassAssociations = filterAssociations(original.ClassAssociations, resolved.Associations)

	// Preserve domain associations.
	filtered.DomainAssociations = original.DomainAssociations

	return &filtered, nil
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

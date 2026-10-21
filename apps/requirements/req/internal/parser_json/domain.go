package parser_json

import (
	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

type pipeDomain struct {
	Key                   string                     `json:"key"`
	Name                  string                     `json:"name"`
	Details               string                     `json:"details,omitempty"`
	UnfinishedNotes       string                     `json:"unfinished_notes,omitempty"`
	Realized              bool                       `json:"realized,omitempty"`
	UmlComment            string                     `json:"uml_comment,omitempty"`
	Subdomains            []pipeSubdomain            `json:"subdomains,omitempty"`
	SubdomainAssociations []pipeSubdomainAssociation `json:"subdomain_associations,omitempty"`
	ClassAssociations     []pipeClassAssociation     `json:"class_associations,omitempty"`
}

type pipeSubdomain struct {
	Key                    string                      `json:"key"`
	Name                   string                      `json:"name"`
	Details                string                      `json:"details,omitempty"`
	UnfinishedNotes        string                      `json:"unfinished_notes,omitempty"`
	UmlComment             string                      `json:"uml_comment,omitempty"`
	Generalizations        []pipeClassGeneralization   `json:"generalizations,omitempty"`
	UseCaseGeneralizations []pipeUseCaseGeneralization `json:"use_case_generalizations,omitempty"`
	Classes                []pipeClass                 `json:"classes,omitempty"`
	UseCases               []pipeUseCase               `json:"use_cases,omitempty"`
	ClassAssociations      []pipeClassAssociation      `json:"class_associations,omitempty"`
	UseCaseShares          []pipeUseCaseShare          `json:"use_case_shares,omitempty"`
}

type pipeDomainAssociation struct {
	Key               string `json:"key"`
	ProblemDomainKey  string `json:"problem_domain_key"`
	SolutionDomainKey string `json:"solution_domain_key"`
	UmlComment        string `json:"uml_comment,omitempty"`
}

type pipeSubdomainAssociation struct {
	Key                  string `json:"key"`
	ProblemSubdomainKey  string `json:"problem_subdomain_key"`
	SolutionSubdomainKey string `json:"solution_subdomain_key"`
	UmlComment           string `json:"uml_comment,omitempty"`
}

func domainFromCore(domain model_domain.Domain) (pipeDomain, error) {
	item := pipeDomain{
		Key:             keyString(domain.Key),
		Name:            domain.Name,
		Details:         domain.Details,
		UnfinishedNotes: domain.UnfinishedNotes,
		Realized:        domain.Realized,
		UmlComment:      domain.UmlComment,
	}
	for _, subdomain := range domain.Subdomains {
		encoded, err := subdomainFromCore(subdomain)
		if err != nil {
			return pipeDomain{}, errors.Wrapf(err, "subdomain %s", subdomain.Key.String())
		}
		item.Subdomains = append(item.Subdomains, encoded)
	}
	sortByKey(item.Subdomains, func(sub pipeSubdomain) string { return sub.Key })
	for _, assoc := range domain.SubdomainAssociations {
		item.SubdomainAssociations = append(item.SubdomainAssociations, subdomainAssociationFromCore(assoc))
	}
	sortByKey(item.SubdomainAssociations, func(assoc pipeSubdomainAssociation) string { return assoc.Key })
	classAssociations, err := classAssociationsFromCore(domain.ClassAssociations)
	if err != nil {
		return pipeDomain{}, err
	}
	item.ClassAssociations = classAssociations
	return item, nil
}

func domainsToCore(items []pipeDomain) (map[identity.Key]model_domain.Domain, error) {
	if len(items) == 0 {
		return nil, nil //nolint:nilnil // omitted collection is the zero value, not an error
	}
	out := make(map[identity.Key]model_domain.Domain, len(items))
	for i, item := range items {
		domain, err := domainToCore(item)
		if err != nil {
			return nil, errors.Wrapf(err, "[%d]", i)
		}
		if err := putUnique(out, domain.Key, domain, i); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func domainToCore(item pipeDomain) (model_domain.Domain, error) {
	key, err := parseKey("key", item.Key)
	if err != nil {
		return model_domain.Domain{}, err
	}
	domain := model_domain.NewDomain(key, item.Name, item.Details, item.UnfinishedNotes, item.Realized, item.UmlComment)

	subdomains := make(map[identity.Key]model_domain.Subdomain, len(item.Subdomains))
	classAssociations, err := classAssociationsToCore(item.ClassAssociations)
	if err != nil {
		return model_domain.Domain{}, errors.Wrap(err, "class_associations")
	}
	for i, subItem := range item.Subdomains {
		subdomain, subAssocs, err := subdomainToCore(subItem)
		if err != nil {
			return model_domain.Domain{}, errors.Wrapf(err, "subdomains[%d]", i)
		}
		if _, exists := subdomains[subdomain.Key]; exists {
			return model_domain.Domain{}, errors.Errorf("subdomains[%d]: duplicate key %s", i, subdomain.Key.String())
		}
		subdomains[subdomain.Key] = subdomain
		for assocKey, assoc := range subAssocs {
			if _, exists := classAssociations[assocKey]; exists {
				return model_domain.Domain{}, errors.Errorf("duplicate class association key %s", assocKey.String())
			}
			classAssociations[assocKey] = assoc
		}
	}
	if len(subdomains) > 0 {
		domain.Subdomains = subdomains
	}

	subdomainAssociations, err := subdomainAssociationsToCore(item.SubdomainAssociations)
	if err != nil {
		return model_domain.Domain{}, errors.Wrap(err, "subdomain_associations")
	}
	domain.SubdomainAssociations = subdomainAssociations

	if len(classAssociations) > 0 {
		if err := domain.SetClassAssociations(classAssociations); err != nil {
			return model_domain.Domain{}, errors.Wrap(err, "class_associations")
		}
		nilEmptyClassAssociations(&domain)
	}
	return domain, nil
}

func nilEmptyClassAssociations(domain *model_domain.Domain) {
	if len(domain.ClassAssociations) == 0 {
		domain.ClassAssociations = nil
	}
	for key, subdomain := range domain.Subdomains {
		if len(subdomain.ClassAssociations) == 0 {
			subdomain.ClassAssociations = nil
			domain.Subdomains[key] = subdomain
		}
	}
}

func subdomainFromCore(subdomain model_domain.Subdomain) (pipeSubdomain, error) {
	classes := classesFromCore(subdomain.Classes)
	classAssociations, err := classAssociationsFromCore(subdomain.ClassAssociations)
	if err != nil {
		return pipeSubdomain{}, err
	}
	return pipeSubdomain{
		Key:                    keyString(subdomain.Key),
		Name:                   subdomain.Name,
		Details:                subdomain.Details,
		UnfinishedNotes:        subdomain.UnfinishedNotes,
		UmlComment:             subdomain.UmlComment,
		Generalizations:        classGeneralizationsFromCore(subdomain.Generalizations),
		UseCaseGeneralizations: useCaseGeneralizationsFromCore(subdomain.UseCaseGeneralizations),
		Classes:                classes,
		UseCases:               useCasesFromCore(subdomain.UseCases),
		ClassAssociations:      classAssociations,
		UseCaseShares:          useCaseSharesFromCore(subdomain.UseCaseShares),
	}, nil
}

func subdomainToCore(item pipeSubdomain) (model_domain.Subdomain, map[identity.Key]model_class.Association, error) {
	key, err := parseKey("key", item.Key)
	if err != nil {
		return model_domain.Subdomain{}, nil, err
	}
	generalizations, err := classGeneralizationsToCore(item.Generalizations)
	if err != nil {
		return model_domain.Subdomain{}, nil, err
	}
	classes, err := classesToCore(item.Classes)
	if err != nil {
		return model_domain.Subdomain{}, nil, err
	}
	classAssociations, err := classAssociationsToCore(item.ClassAssociations)
	if err != nil {
		return model_domain.Subdomain{}, nil, errors.Wrap(err, "class_associations")
	}
	useCaseGeneralizations, err := useCaseGeneralizationsToCore(item.UseCaseGeneralizations)
	if err != nil {
		return model_domain.Subdomain{}, nil, err
	}
	useCases, err := useCasesToCore(item.UseCases)
	if err != nil {
		return model_domain.Subdomain{}, nil, err
	}
	useCaseShares, err := useCaseSharesToCore(item.UseCaseShares)
	if err != nil {
		return model_domain.Subdomain{}, nil, err
	}
	subdomain := model_domain.NewSubdomain(key, item.Name, item.Details, item.UnfinishedNotes, item.UmlComment)
	subdomain.Generalizations = generalizations
	subdomain.UseCaseGeneralizations = useCaseGeneralizations
	subdomain.Classes = classes
	subdomain.UseCases = useCases
	subdomain.UseCaseShares = useCaseShares
	return subdomain, classAssociations, nil
}

func domainAssociationFromCore(assoc model_domain.Association) pipeDomainAssociation {
	return pipeDomainAssociation{
		Key:               keyString(assoc.Key),
		ProblemDomainKey:  keyString(assoc.ProblemDomainKey),
		SolutionDomainKey: keyString(assoc.SolutionDomainKey),
		UmlComment:        assoc.UmlComment,
	}
}

func domainAssociationsToCore(items []pipeDomainAssociation) (map[identity.Key]model_domain.Association, error) {
	if len(items) == 0 {
		return nil, nil //nolint:nilnil // omitted collection is the zero value, not an error
	}
	out := make(map[identity.Key]model_domain.Association, len(items))
	for i, item := range items {
		key, err := parseKey("key", item.Key)
		if err != nil {
			return nil, errors.Wrapf(err, "[%d]", i)
		}
		problem, err := parseKey("problem_domain_key", item.ProblemDomainKey)
		if err != nil {
			return nil, errors.Wrapf(err, "[%d]", i)
		}
		solution, err := parseKey("solution_domain_key", item.SolutionDomainKey)
		if err != nil {
			return nil, errors.Wrapf(err, "[%d]", i)
		}
		assoc := model_domain.NewAssociation(key, problem, solution, item.UmlComment)
		if err := putUnique(out, assoc.Key, assoc, i); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func subdomainAssociationFromCore(assoc model_domain.SubdomainAssociation) pipeSubdomainAssociation {
	return pipeSubdomainAssociation{
		Key:                  keyString(assoc.Key),
		ProblemSubdomainKey:  keyString(assoc.ProblemSubdomainKey),
		SolutionSubdomainKey: keyString(assoc.SolutionSubdomainKey),
		UmlComment:           assoc.UmlComment,
	}
}

func subdomainAssociationsToCore(items []pipeSubdomainAssociation) (map[identity.Key]model_domain.SubdomainAssociation, error) {
	if len(items) == 0 {
		return nil, nil //nolint:nilnil // omitted collection is the zero value, not an error
	}
	out := make(map[identity.Key]model_domain.SubdomainAssociation, len(items))
	for i, item := range items {
		key, err := parseKey("key", item.Key)
		if err != nil {
			return nil, errors.Wrapf(err, "[%d]", i)
		}
		problem, err := parseKey("problem_subdomain_key", item.ProblemSubdomainKey)
		if err != nil {
			return nil, errors.Wrapf(err, "[%d]", i)
		}
		solution, err := parseKey("solution_subdomain_key", item.SolutionSubdomainKey)
		if err != nil {
			return nil, errors.Wrapf(err, "[%d]", i)
		}
		assoc := model_domain.NewSubdomainAssociation(key, problem, solution, item.UmlComment)
		if err := putUnique(out, assoc.Key, assoc, i); err != nil {
			return nil, err
		}
	}
	return out, nil
}

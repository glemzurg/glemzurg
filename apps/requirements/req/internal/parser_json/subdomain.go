package parser_json

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_use_case"
)

// subdomainInOut is a nested category of the model.
type subdomainInOut struct {
	Key        string `json:"key"`
	Name       string `json:"name"`
	Details    string `json:"details"` // Markdown.
	UmlComment string `json:"uml_comment"`
	// Nested.
	Generalizations []generalizationInOut `json:"generalizations"` // Generalizations for the classes and use cases in this subdomain.
	Classes         []classInOut          `json:"classes"`
	UseCases        []useCaseInOut        `json:"use_cases"`
	Associations    []associationInOut    `json:"associations"` // Associations between classes in this subdomain.
}

// ToRequirements converts the subdomainInOut to model_domain.Subdomain.
func (s subdomainInOut) ToRequirements() (model_domain.Subdomain, error) {
	key, err := identity.ParseKey(s.Key)
	if err != nil {
		return model_domain.Subdomain{}, err
	}

	subdomain := model_domain.Subdomain{
		Key:        key,
		Name:       s.Name,
		Details:    s.Details,
		UmlComment: s.UmlComment,
	}

	for _, g := range s.Generalizations {
		gen, err := g.ToRequirements()
		if err != nil {
			return model_domain.Subdomain{}, err
		}
		if subdomain.Generalizations == nil {
			subdomain.Generalizations = make(map[identity.Key]model_class.Generalization)
		}
		subdomain.Generalizations[gen.Key] = gen
	}
	for _, c := range s.Classes {
		class, err := c.ToRequirements()
		if err != nil {
			return model_domain.Subdomain{}, err
		}
		if subdomain.Classes == nil {
			subdomain.Classes = make(map[identity.Key]model_class.Class)
		}
		subdomain.Classes[class.Key] = class
	}
	for _, u := range s.UseCases {
		useCase, err := u.ToRequirements()
		if err != nil {
			return model_domain.Subdomain{}, err
		}
		if subdomain.UseCases == nil {
			subdomain.UseCases = make(map[identity.Key]model_use_case.UseCase)
		}
		subdomain.UseCases[useCase.Key] = useCase
	}
	for _, a := range s.Associations {
		assoc, err := a.ToRequirements()
		if err != nil {
			return model_domain.Subdomain{}, err
		}
		if subdomain.ClassAssociations == nil {
			subdomain.ClassAssociations = make(map[identity.Key]model_class.Association)
		}
		subdomain.ClassAssociations[assoc.Key] = assoc
	}
	return subdomain, nil
}

// FromRequirements creates a subdomainInOut from model_domain.Subdomain.
func FromRequirementsSubdomain(s model_domain.Subdomain) subdomainInOut {
	subdomain := subdomainInOut{
		Key:        s.Key.String(),
		Name:       s.Name,
		Details:    s.Details,
		UmlComment: s.UmlComment,
	}
	for _, g := range s.Generalizations {
		subdomain.Generalizations = append(subdomain.Generalizations, FromRequirementsGeneralization(g))
	}
	for _, c := range s.Classes {
		subdomain.Classes = append(subdomain.Classes, FromRequirementsClass(c))
	}
	for _, u := range s.UseCases {
		subdomain.UseCases = append(subdomain.UseCases, FromRequirementsUseCase(u))
	}
	for _, a := range s.ClassAssociations {
		subdomain.Associations = append(subdomain.Associations, FromRequirementsAssociation(a))
	}
	return subdomain
}

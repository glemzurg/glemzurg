package parser_json

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"

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

// ToRequirements converts the subdomainInOut to requirements.Subdomain.
func (s subdomainInOut) ToRequirements() requirements.Subdomain {
	subdomain := requirements.Subdomain{
		Key:             s.Key,
		Name:            s.Name,
		Details:         s.Details,
		UmlComment:      s.UmlComment,
		Generalizations: nil, // Not handled here
		Classes:         nil,
		UseCases:        nil,
		Associations:    nil,
	}

	for _, g := range s.Generalizations {
		subdomain.Generalizations = append(subdomain.Generalizations, g.ToRequirements())
	}
	for _, c := range s.Classes {
		subdomain.Classes = append(subdomain.Classes, c.ToRequirements())
	}
	for _, u := range s.UseCases {
		subdomain.UseCases = append(subdomain.UseCases, u.ToRequirements())
	}
	for _, a := range s.Associations {
		subdomain.Associations = append(subdomain.Associations, a.ToRequirements())
	}
	return subdomain
}

// FromRequirements creates a subdomainInOut from requirements.Subdomain.
func FromRequirementsSubdomain(s requirements.Subdomain) subdomainInOut {
	subdomain := subdomainInOut{
		Key:             s.Key,
		Name:            s.Name,
		Details:         s.Details,
		UmlComment:      s.UmlComment,
		Generalizations: nil, // Not handled here
		Classes:         nil,
		UseCases:        nil,
		Associations:    nil,
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
	for _, a := range s.Associations {
		subdomain.Associations = append(subdomain.Associations, FromRequirementsAssociation(a))
	}
	return subdomain
}

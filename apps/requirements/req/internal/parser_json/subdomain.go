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
	return requirements.Subdomain{
		Key:        s.Key,
		Name:       s.Name,
		Details:    s.Details,
		UmlComment: s.UmlComment,
	}
}

// FromRequirements creates a subdomainInOut from requirements.Subdomain.
func FromRequirementsSubdomain(s requirements.Subdomain) subdomainInOut {
	return subdomainInOut{
		Key:             s.Key,
		Name:            s.Name,
		Details:         s.Details,
		UmlComment:      s.UmlComment,
		Generalizations: nil, // Not handled here
		Classes:         nil,
		UseCases:        nil,
		Associations:    nil,
	}
}

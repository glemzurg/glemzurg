package parser_json

// subdomainInOut is a nested category of the model.
type subdomainInOut struct {
	Key        string `json:"key"`
	Name       string `json:"name"`
	Details    string `json:"details,omitempty"` // Markdown.
	UmlComment string `json:"uml_comment,omitempty"`
	// Nested.
	Generalizations []generalizationInOut `json:"generalizations,omitempty"` // Generalizations for the classes and use cases in this subdomain.
	Classes         []classInOut          `json:"classes,omitempty"`
	UseCases        []useCaseInOut        `json:"use_cases,omitempty"`
	Associations    []associationInOut    `json:"associations,omitempty"` // Associations between classes in this subdomain.
}

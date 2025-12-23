package parser_json

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

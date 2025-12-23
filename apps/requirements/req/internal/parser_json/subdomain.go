package parser_json

// subdomainInOut is a nested category of the model.
type subdomainInOut struct {
	Key        string
	Name       string
	Details    string // Markdown.
	UmlComment string
	// Nested.
	Generalizations []generalizationInOut // Generalizations for the classes and use cases in this subdomain.
	Classes         []classInOut
	UseCases        []useCaseInOut
	Associations    []associationInOut // Associations between classes in this subdomain.
}

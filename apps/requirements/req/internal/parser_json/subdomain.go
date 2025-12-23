package parser_json

// subdomain is a nested category of the model.
type subdomain struct {
	Key        string
	Name       string
	Details    string // Markdown.
	UmlComment string
	// Nested.
	Generalizations []generalization // Generalizations for the classes and use cases in this subdomain.
	Classes         []class
	UseCases        []useCase
	Associations    []association // Associations between classes in this subdomain.
}

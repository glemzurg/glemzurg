package parser_json

// model is the documentation summary of a set of requirements.
type model struct {
	Key     string
	Name    string
	Details string // Markdown.
	// Nested structure.
	Actors             []actor
	Domains            []domain
	DomainAssociations []domainAssociation
	Associations       []association // Associations between classes that span domains.
}

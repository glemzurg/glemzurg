package parser_json

// modelInOut is the documentation summary of a set of requirements.
type modelInOut struct {
	Key     string
	Name    string
	Details string // Markdown.
	// Nested structure.
	Actors             []actorInOut
	Domains            []domainInOut
	DomainAssociations []domainAssociationInOut
	Associations       []associationInOut // Associations between classes that span domains.
}

package parser_json

// modelInOut is the documentation summary of a set of requirements.
type modelInOut struct {
	Key     string `json:"key"`
	Name    string `json:"name"`
	Details string `json:"details"` // Markdown.
	// Nested structure.
	Actors             []actorInOut             `json:"actors"`
	Domains            []domainInOut            `json:"domains"`
	DomainAssociations []domainAssociationInOut `json:"domain_associations"`
	Associations       []associationInOut       `json:"associations"` // Associations between classes that span domains.
}

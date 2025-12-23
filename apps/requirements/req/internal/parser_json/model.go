package parser_json

// modelInOut is the documentation summary of a set of requirements.
type modelInOut struct {
	Key     string `json:"key"`
	Name    string `json:"name"`
	Details string `json:"details,omitempty"` // Markdown.
	// Nested structure.
	Actors             []actorInOut             `json:"actors,omitempty"`
	Domains            []domainInOut            `json:"domains,omitempty"`
	DomainAssociations []domainAssociationInOut `json:"domain_associations,omitempty"`
	Associations       []associationInOut       `json:"associations,omitempty"` // Associations between classes that span domains.
}

package parser_json

// associationInOut is how two classes relate to each other.
type associationInOut struct {
	Key                 string            `json:"key"`
	Name                string            `json:"name"`
	Details             string            `json:"details,omitempty"` // Markdown.
	FromClassKey        string            `json:"from_class_key"`
	FromMultiplicity    multiplicityInOut `json:"from_multiplicity"`
	ToClassKey          string            `json:"to_class_key"`
	ToMultiplicity      multiplicityInOut `json:"to_multiplicity"`
	AssociationClassKey string            `json:"association_class_key,omitempty"`
	UmlComment          string            `json:"uml_comment,omitempty"`
}

package parser_ai_input

// inputAssociation represents an association JSON file.
type inputAssociation struct {
	Name                string  `json:"name"`
	Details             string  `json:"details,omitempty"`
	FromClassKey        string  `json:"from_class_key"`
	FromMultiplicity    string  `json:"from_multiplicity"`
	ToClassKey          string  `json:"to_class_key"`
	ToMultiplicity      string  `json:"to_multiplicity"`
	AssociationClassKey *string `json:"association_class_key,omitempty"`
	UmlComment          string  `json:"uml_comment,omitempty"`
}

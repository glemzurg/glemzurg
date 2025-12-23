package parser_json

// associationInOut is how two classes relate to each other.
type associationInOut struct {
	Key                 string
	Name                string
	Details             string            // Markdown.
	FromClassKey        string            // The class on one end of the association.
	FromMultiplicity    multiplicityInOut // The multiplicity from one end of the association.
	ToClassKey          string            // The class on the other end of the association.
	ToMultiplicity      multiplicityInOut // The multiplicity on the other end of the association.
	AssociationClassKey string            // Any class that points to this association.
	UmlComment          string
}

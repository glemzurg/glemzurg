package parser_json

// association is how two classes relate to each other.
type association struct {
	Key                 string
	Name                string
	Details             string       // Markdown.
	FromClassKey        string       // The class on one end of the association.
	FromMultiplicity    multiplicity // The multiplicity from one end of the association.
	ToClassKey          string       // The class on the other end of the association.
	ToMultiplicity      multiplicity // The multiplicity on the other end of the association.
	AssociationClassKey string       // Any class that points to this association.
	UmlComment          string
}

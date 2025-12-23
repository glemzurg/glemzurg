package parser_json

// multiplicity is how two classes relate to each other.
type multiplicity struct {
	LowerBound  uint // Zero is "any".
	HigherBound uint // Zero is "any".
}

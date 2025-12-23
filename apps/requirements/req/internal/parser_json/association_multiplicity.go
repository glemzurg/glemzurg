package parser_json

// multiplicityInOut is how two classes relate to each other.
type multiplicityInOut struct {
	LowerBound  uint // Zero is "any".
	HigherBound uint // Zero is "any".
}

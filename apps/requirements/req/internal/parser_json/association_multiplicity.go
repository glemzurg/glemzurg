package parser_json

// multiplicityInOut is how two classes relate to each other.
type multiplicityInOut struct {
	LowerBound  uint `json:"lower_bound"`  // Zero is "any".
	HigherBound uint `json:"higher_bound"` // Zero is "any".
}

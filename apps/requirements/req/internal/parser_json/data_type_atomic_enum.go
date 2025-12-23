package parser_json

// atomicEnumInOut represents an allowed value in an enumeration.
type atomicEnumInOut struct {
	Value     string `json:"value"`
	SortOrder int    `json:"sort_order"`
}

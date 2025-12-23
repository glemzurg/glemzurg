package parser_json

// dataType represents the main data type structure.
type dataType struct {
	Key              string
	CollectionType   string
	CollectionUnique *bool
	CollectionMin    *int
	CollectionMax    *int
	Atomic           *atomic
	RecordFields     []field
}

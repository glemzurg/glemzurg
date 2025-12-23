package parser_json

// dataTypeInOut represents the main data type structure.
type dataTypeInOut struct {
	Key              string
	CollectionType   string
	CollectionUnique *bool
	CollectionMin    *int
	CollectionMax    *int
	Atomic           *atomicInOut
	RecordFields     []fieldInOut
}

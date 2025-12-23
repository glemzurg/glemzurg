package parser_json

// dataTypeInOut represents the main data type structure.
type dataTypeInOut struct {
	Key              string       `json:"key"`
	CollectionType   string       `json:"collection_type"`
	CollectionUnique *bool        `json:"collection_unique"`
	CollectionMin    *int         `json:"collection_min"`
	CollectionMax    *int         `json:"collection_max"`
	Atomic           *atomicInOut `json:"atomic"`
	RecordFields     []fieldInOut `json:"record_fields"`
}

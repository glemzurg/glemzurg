package parser_json

// dataTypeInOut represents the main data type structure.
type dataTypeInOut struct {
	Key              string       `json:"key"`
	CollectionType   string       `json:"collection_type,omitempty"`
	CollectionUnique *bool        `json:"collection_unique,omitempty"`
	CollectionMin    *int         `json:"collection_min,omitempty"`
	CollectionMax    *int         `json:"collection_max,omitempty"`
	Atomic           *atomicInOut `json:"atomic,omitempty"`
	RecordFields     []fieldInOut `json:"record_fields,omitempty"`
}

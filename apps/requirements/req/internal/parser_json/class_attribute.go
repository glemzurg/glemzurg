package parser_json

// attributeInOut is a member of a class.
type attributeInOut struct {
	Key              string `json:"key"`
	Name             string `json:"name"`
	Details          string `json:"details,omitempty"` // Markdown.
	DataTypeRules    string `json:"data_type_rules,omitempty"`
	DerivationPolicy string `json:"derivation_policy,omitempty"`
	Nullable         bool   `json:"nullable"`
	UmlComment       string `json:"uml_comment,omitempty"`
	// Part of the data in a parsed file.
	IndexNums []uint         `json:"index_nums,omitempty"` // The indexes this attribute is part of.
	DataType  *dataTypeInOut `json:"data_type,omitempty"`  // If the DataTypeRules can be parsed, this is the resulting data type.
}

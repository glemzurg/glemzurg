package parser_json

// attributeInOut is a member of a class.
type attributeInOut struct {
	Key              string `json:"key"`
	Name             string `json:"name"`
	Details          string `json:"details"` // Markdown.
	DataTypeRules    string `json:"data_type_rules"`
	DerivationPolicy string `json:"derivation_policy"`
	Nullable         bool   `json:"nullable"`
	UmlComment       string `json:"uml_comment"`
	// Part of the data in a parsed file.
	IndexNums []uint         `json:"index_nums"` // The indexes this attribute is part of.
	DataType  *dataTypeInOut `json:"data_type"`  // If the DataTypeRules can be parsed, this is the resulting data type.
}

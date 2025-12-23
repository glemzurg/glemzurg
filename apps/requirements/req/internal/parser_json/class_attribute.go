package parser_json

// attributeInOut is a member of a class.
type attributeInOut struct {
	Key              string
	Name             string
	Details          string // Markdown.
	DataTypeRules    string // What are the bounds of this data type.
	DerivationPolicy string // If this is a derived attribute, how is it derived.
	Nullable         bool   // Is this attribute optional.
	UmlComment       string
	// Part of the data in a parsed file.
	IndexNums []uint         // The indexes this attribute is part of.
	DataType  *dataTypeInOut // If the DataTypeRules can be parsed, this is the resulting data type.
}

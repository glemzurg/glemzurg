package parser_json

// attribute is a member of a class.
type attribute struct {
	Key              string
	Name             string
	Details          string // Markdown.
	DataTypeRules    string // What are the bounds of this data type.
	DerivationPolicy string // If this is a derived attribute, how is it derived.
	Nullable         bool   // Is this attribute optional.
	UmlComment       string
	// Part of the data in a parsed file.
	IndexNums []uint    // The indexes this attribute is part of.
	DataType  *dataType // If the DataTypeRules can be parsed, this is the resulting data type.
}

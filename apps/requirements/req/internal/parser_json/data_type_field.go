package parser_json

// fieldInOut represents a single field of a record datatype.
type fieldInOut struct {
	Name          string         // The name of the field.
	FieldDataType *dataTypeInOut // The data type of this field.
}

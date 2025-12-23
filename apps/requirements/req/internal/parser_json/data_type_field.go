package parser_json

// field represents a single field of a record datatype.
type field struct {
	Name          string    // The name of the field.
	FieldDataType *dataType // The data type of this field.
}

package parser_json

// fieldInOut represents a single field of a record datatype.
type fieldInOut struct {
	Name          string         `json:"name"`                      // The name of the field.
	FieldDataType *dataTypeInOut `json:"field_data_type,omitempty"` // The data type of this field.
}

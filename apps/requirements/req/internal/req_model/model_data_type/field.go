package model_data_type

// Field represents a single field of a record datatype.
type Field struct {
	Name          string    `validate:"required"` // The name of the field.
	FieldDataType *DataType `validate:"required"` // The data type of this field.
}

// Validate validates the Field struct.
func (f Field) Validate() error {
	return _validate.Struct(f)
}

// String returns a string representation of the Field.
func (f Field) String() string {
	return f.Name + ": " + f.FieldDataType.String() + ";"
}

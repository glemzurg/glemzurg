package data_type

import validation "github.com/go-ozzo/ozzo-validation/v4"

// Field represents a single field of a record datatype.
type Field struct {
	Name             string // The name of the field.
	FieldDataTypeKey string // The data type key of this field.
}

// Validate validates the Field struct.
func (f Field) Validate() error {
	return validation.ValidateStruct(&f,
		validation.Field(&f.Name, validation.Required),
		validation.Field(&f.FieldDataTypeKey, validation.Required),
	)
}

// String returns a string representation of the Field.
func (f Field) String() string {
	return f.Name + ": " + f.FieldDataTypeKey + ";"
}

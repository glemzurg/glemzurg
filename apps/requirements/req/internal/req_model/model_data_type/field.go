package model_data_type

import (
	"fmt"
	"regexp"
)

// _fieldNameRegexp enforces that field names are lowercase identifiers.
// Field names become part of data type keys via UnpackNested() (parentKey + "/" + field.Name).
var _fieldNameRegexp = regexp.MustCompile(`^[a-z_][a-z0-9_]*$`)

// Field represents a single field of a record datatype.
type Field struct {
	Name          string    `validate:"required"` // The name of the field.
	FieldDataType *DataType `validate:"required"` // The data type of this field.
}

// Validate validates the Field struct.
func (f Field) Validate() error {
	if err := _validate.Struct(f); err != nil {
		return err
	}
	if !_fieldNameRegexp.MatchString(f.Name) {
		return fmt.Errorf("Name: '%s' must be a lowercase identifier matching [a-z_][a-z0-9_]*", f.Name)
	}
	return nil
}

// String returns a string representation of the Field.
func (f Field) String() string {
	return f.Name + ": " + f.FieldDataType.String() + ";"
}

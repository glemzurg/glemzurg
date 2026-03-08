package model_data_type

import (
	"fmt"
	"regexp"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
)

// _fieldNameRegexp enforces that field names are lowercase identifiers.
// Field names become part of data type keys via UnpackNested() (parentKey + "/" + field.Name).
var _fieldNameRegexp = regexp.MustCompile(`^[a-z_][a-z0-9_]*$`)

// Field represents a single field of a record datatype.
type Field struct {
	Name          string    // The name of the field.
	FieldDataType *DataType // The data type of this field.
}

// Validate validates the Field struct.
func (f Field) Validate() error {
	if f.Name == "" {
		return &coreerr.ValidationError{
			Code:    coreerr.DtypeFieldNameRequired,
			Message: "Name is required",
			Field:   "Name",
		}
	}
	if f.FieldDataType == nil {
		return &coreerr.ValidationError{
			Code:    coreerr.DtypeFieldDatatypeRequired,
			Message: "FieldDataType is required",
			Field:   "FieldDataType",
		}
	}
	if !_fieldNameRegexp.MatchString(f.Name) {
		return fmt.Errorf("name: '%s' must be a lowercase identifier matching [a-z_][a-z0-9_]*", f.Name)
	}
	return nil
}

// String returns a string representation of the Field.
func (f Field) String() string {
	return f.Name + ": " + f.FieldDataType.String() + ";"
}

package parser_json

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/data_type"

// fieldInOut represents a single field of a record datatype.
type fieldInOut struct {
	Name          string         `json:"name"`            // The name of the field.
	FieldDataType *dataTypeInOut `json:"field_data_type"` // The data type of this field.
}

// ToRequirements converts the fieldInOut to data_type.Field.
func (f fieldInOut) ToRequirements() data_type.Field {
	field := data_type.Field{
		Name:          f.Name,
		FieldDataType: nil, // TODO: convert
	}
	if f.FieldDataType != nil {
		dt := f.FieldDataType.ToRequirements()
		field.FieldDataType = &dt
	}
	return field
}

// FromRequirements creates a fieldInOut from data_type.Field.
func FromRequirementsField(f data_type.Field) fieldInOut {
	field := fieldInOut{
		Name:          f.Name,
		FieldDataType: nil,
	}
	if f.FieldDataType != nil {
		dt := FromRequirementsDataType(*f.FieldDataType)
		field.FieldDataType = &dt
	}
	return field
}

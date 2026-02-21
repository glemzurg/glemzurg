package model_data_type

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestFieldSuite(t *testing.T) {
	suite.Run(t, new(FieldSuite))
}

type FieldSuite struct {
	suite.Suite
}

func (suite *FieldSuite) TestValidate() {
	validDataType := &DataType{
		Key:            "key",
		CollectionType: "atomic",
		Atomic:         &Atomic{ConstraintType: "unconstrained"},
	}

	tests := []struct {
		name   string
		field  Field
		errstr string
	}{
		// OK: lowercase name.
		{
			name: "lowercase name",
			field: Field{
				Name:          "field_name",
				FieldDataType: validDataType,
			},
		},
		// OK: single letter.
		{
			name: "single letter",
			field: Field{
				Name:          "x",
				FieldDataType: validDataType,
			},
		},
		// OK: underscore prefix.
		{
			name: "underscore prefix",
			field: Field{
				Name:          "_private",
				FieldDataType: validDataType,
			},
		},
		// OK: letters and digits.
		{
			name: "letters and digits",
			field: Field{
				Name:          "field2name3",
				FieldDataType: validDataType,
			},
		},

		// Error: empty name.
		{
			name: "empty name",
			field: Field{
				Name:          "",
				FieldDataType: validDataType,
			},
			errstr: "Name",
		},
		// Error: uppercase letter.
		{
			name: "uppercase letter",
			field: Field{
				Name:          "FieldName",
				FieldDataType: validDataType,
			},
			errstr: "must be a lowercase identifier",
		},
		// Error: starts with digit.
		{
			name: "starts with digit",
			field: Field{
				Name:          "1field",
				FieldDataType: validDataType,
			},
			errstr: "must be a lowercase identifier",
		},
		// Error: contains space.
		{
			name: "contains space",
			field: Field{
				Name:          "field name",
				FieldDataType: validDataType,
			},
			errstr: "must be a lowercase identifier",
		},
		// Error: contains hyphen.
		{
			name: "contains hyphen",
			field: Field{
				Name:          "field-name",
				FieldDataType: validDataType,
			},
			errstr: "must be a lowercase identifier",
		},
		// Error: nil data type.
		{
			name: "nil data type",
			field: Field{
				Name:          "valid_name",
				FieldDataType: nil,
			},
			errstr: "FieldDataType",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := tt.field.Validate()
			if tt.errstr == "" {
				assert.Nil(suite.T(), err, "expected no error for %+v", tt.field)
			} else {
				assert.NotNil(suite.T(), err, "expected error for %+v", tt.field)
				assert.ErrorContains(suite.T(), err, tt.errstr, "error message mismatch for %+v", tt.field)
			}
		})
	}
}

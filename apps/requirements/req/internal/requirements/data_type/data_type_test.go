package data_type

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestDataTypeSuite(t *testing.T) {
	suite.Run(t, new(DataTypeSuite))
}

type DataTypeSuite struct {
	suite.Suite
}

func (suite *DataTypeSuite) TestValidate() {
	atomic := &Atomic{
		ConstraintType: "unconstrained",
		Details:        "string",
	}

	tests := []struct {
		key            string
		name           string
		details        string
		collectionType string
		atomic         *Atomic
		errstr         string
	}{
		// OK.
		{
			key:            "Key",
			name:           "Name",
			details:        "Details",
			collectionType: "atomic",
			atomic:         atomic,
		},
		{
			key:            "Key",
			name:           "Name",
			details:        "",
			collectionType: "atomic",
			atomic:         atomic,
		},

		// Error states.
		{
			key:            "",
			name:           "Name",
			details:        "Details",
			collectionType: "atomic",
			atomic:         atomic,
			errstr:         `Key: cannot be blank.`,
		},
		{
			key:            "Key",
			name:           "",
			details:        "Details",
			collectionType: "atomic",
			atomic:         atomic,
			errstr:         `Name: cannot be blank.`,
		},
		{
			key:            "Key",
			name:           "Name",
			details:        "Details",
			collectionType: "",
			atomic:         atomic,
			errstr:         `CollectionType: cannot be blank.`,
		},
		{
			key:            "Key",
			name:           "Name",
			details:        "Details",
			collectionType: "unknown",
			atomic:         atomic,
			errstr:         `CollectionType: must be a valid value.`,
		},
		{
			key:            "Key",
			name:           "Name",
			details:        "Details",
			collectionType: "atomic",
			atomic:         nil,
			errstr:         `Atomic: cannot be blank.`,
		},
	}

	for _, tt := range tests {
		dt := DataType{
			Key:            tt.key,
			Name:           tt.name,
			Details:        tt.details,
			CollectionType: tt.collectionType,
			Atomic:         tt.atomic,
		}
		err := dt.Validate()
		if tt.errstr == "" {
			assert.Nil(suite.T(), err, "expected no error for %+v", dt)
		} else {
			assert.NotNil(suite.T(), err, "expected error for %+v", dt)
			assert.Equal(suite.T(), tt.errstr, err.Error(), "error message mismatch for %+v", dt)
		}
	}
}

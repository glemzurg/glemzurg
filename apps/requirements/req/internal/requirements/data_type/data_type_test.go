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
	}
	atomicInvalid := &Atomic{
		ConstraintType: "unknown",
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
		{
			key:            "Key",
			name:           "Name",
			details:        "Details",
			collectionType: "atomic",
			atomic:         atomicInvalid,
			errstr:         `Atomic: (ConstraintType: must be a valid value.).`,
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

func TestNewUnparsable(t *testing.T) {

	// If we cannot parse the text, no error but instead just a nil result.
	result, err := New("key", "this cannot be parsed so it is just an unparsable blob")
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestNewInvalid(t *testing.T) {
	// Key is required.
	result, err := New("", "")
	assert.ErrorContains(t, err, "Key: cannot be blank.")
	assert.Nil(t, result)
}

func TestDataTypeString(t *testing.T) {
	tests := []struct {
		name         string
		dataType     DataType
		expected     string
		panicMessage string
	}{
		{
			name: "atomic unconstrained",
			dataType: DataType{
				CollectionType: "atomic",
				Atomic: &Atomic{
					ConstraintType: "unconstrained",
				},
			},
			expected: "unconstrained",
		},
		{
			name: "atomic reference",
			dataType: DataType{
				CollectionType: "atomic",
				Atomic: &Atomic{
					ConstraintType: "reference",
					Reference:      "some ref",
				},
			},
			expected: "ref: some ref",
		},
		{
			name: "non-atomic",
			dataType: DataType{
				CollectionType: "unknown",
			},
			panicMessage: "unsupported collection type: 'unknown'",
		},
		{
			name: "panic case: atomic nil",
			dataType: DataType{
				CollectionType: "atomic",
				// Atomic is nil to force panic.
			},
			panicMessage: "atomic is nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.panicMessage != "" {
				assert.PanicsWithValue(t, tt.panicMessage, func() { tt.dataType.String() })
			} else {
				assert.NotPanics(t, func() {
					result := tt.dataType.String()
					assert.Equal(t, tt.expected, result)
				})
			}
		})
	}
}

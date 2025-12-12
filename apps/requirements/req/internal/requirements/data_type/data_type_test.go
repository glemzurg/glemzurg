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

func TestParseBlank(t *testing.T) {
	key := "key"

	tests := []struct {
		name         string
		input        string
		expected     *DataType
		errorMessage string
	}{
		// Basic collections without multiplicity
		{
			name:  "blank",
			input: "",
			expected: &DataType{
				Key:            key,
				Name:           "unconstrained",
				CollectionType: "atomic",
				Atomic: &Atomic{
					ConstraintType: "unconstrained",
				},
			},
			errorMessage: "",
		},
		{
			name:  "only whitespace",
			input: " \t\n\r",
			expected: &DataType{
				Key:            key,
				Name:           "unconstrained",
				CollectionType: "atomic",
				Atomic: &Atomic{
					ConstraintType: "unconstrained",
				},
			},
			errorMessage: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := New(key, tt.input)
			if tt.errorMessage != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestParseCollections(t *testing.T) {
	trueValue := true
	falseValue := false

	tests := []struct {
		name         string
		input        string
		expected     *DataType
		errorMessage string
	}{
		// Basic collections without multiplicity
		{
			name:  "stack of unconstrained",
			input: "stack of unconstrained",
			expected: &DataType{
				CollectionType: "stack",
				CollectionMin:  intPtr(0),
				Atomic: &Atomic{
					ConstraintType: "unconstrained",
				},
			},
			errorMessage: "",
		},
		{
			name:  "unordered of ref from something",
			input: "unordered of ref from something",
			expected: &DataType{
				CollectionType: "unordered",
				CollectionMin:  intPtr(0),
				Atomic: &Atomic{
					ConstraintType: "reference",
					Reference:      "something",
				},
			},
			errorMessage: "",
		},
		{
			name:  "ordered of obj of class_key",
			input: "ordered of obj of class_key",
			expected: &DataType{
				CollectionType: "ordered",
				CollectionMin:  intPtr(0),
				Atomic: &Atomic{
					ConstraintType: "object",
					ObjectClassKey: "class_key",
				},
			},
			errorMessage: "",
		},
		{
			name:  "queue of enum of value_a, value_b",
			input: "queue of enum of value_a, value_b",
			expected: &DataType{
				CollectionType: "queue",
				CollectionMin:  intPtr(0),
				Atomic: &Atomic{
					ConstraintType: "enumeration",
					EnumOrdered:    &falseValue,
					Enums: []AtomicEnum{
						{Value: "value_a"},
						{Value: "value_b"},
					},
				},
			},
			errorMessage: "",
		},

		// Collections with multiplicity
		{
			name:  "3+ unordered of unconstrained",
			input: "3+ unordered of unconstrained",
			expected: &DataType{
				CollectionType: "unordered",
				CollectionMin:  intPtr(3),
				Atomic: &Atomic{
					ConstraintType: "unconstrained",
				},
			},
			errorMessage: "",
		},
		{
			name:  "2-5 ordered of ref from something",
			input: "2-5 ordered of ref from something",
			expected: &DataType{
				CollectionType: "ordered",
				CollectionMin:  intPtr(2),
				CollectionMax:  intPtr(5),
				Atomic: &Atomic{
					ConstraintType: "reference",
					Reference:      "something",
				},
			},
			errorMessage: "",
		},
		{
			name:  "0-7 queue of obj of class_key",
			input: "0-7 queue of obj of class_key",
			expected: &DataType{
				CollectionType: "queue",
				CollectionMin:  intPtr(0),
				CollectionMax:  intPtr(7),
				Atomic: &Atomic{
					ConstraintType: "object",
					ObjectClassKey: "class_key",
				},
			},
			errorMessage: "",
		},

		// Collections with unique
		{
			name:  "unique stack of unconstrained",
			input: "unique stack of unconstrained",
			expected: &DataType{
				CollectionType:   "stack",
				CollectionUnique: &trueValue,
				CollectionMin:    intPtr(0),
				Atomic: &Atomic{
					ConstraintType: "unconstrained",
				},
			},
			errorMessage: "",
		},
		{
			name:  "unique 3+ unordered of ref from something",
			input: "unique 3+ unordered of ref from something",
			expected: &DataType{
				CollectionType:   "unordered",
				CollectionUnique: &trueValue,
				CollectionMin:    intPtr(3),
				Atomic: &Atomic{
					ConstraintType: "reference",
					Reference:      "something",
				},
			},
			errorMessage: "",
		},
		{
			name:  "unique 2-5 ordered of obj of class_key",
			input: "unique 2-5 ordered of obj of class_key",
			expected: &DataType{
				CollectionType:   "ordered",
				CollectionUnique: &trueValue,
				CollectionMin:    intPtr(2),
				CollectionMax:    intPtr(5),
				Atomic: &Atomic{
					ConstraintType: "object",
					ObjectClassKey: "class_key",
				},
			},
			errorMessage: "",
		},
		{
			name:  "unique 0-7 queue of enum of value_a, value_b",
			input: "unique 0-7 queue of enum of value_a, value_b",
			expected: &DataType{
				CollectionType:   "queue",
				CollectionUnique: &trueValue,
				CollectionMin:    intPtr(0),
				CollectionMax:    intPtr(7),
				Atomic: &Atomic{
					ConstraintType: "enumeration",
					EnumOrdered:    &falseValue,
					Enums: []AtomicEnum{
						{Value: "value_a"},
						{Value: "value_b"},
					},
				},
			},
			errorMessage: "",
		},
	}

	for _, tt := range tests {
		pass := t.Run(tt.name, func(t *testing.T) {

			// Test calling directly into the parser.
			dataTypeAny, err := Parse("", []byte(tt.input), Entrypoint("CollectionType"))
			if tt.errorMessage == "" {
				assert.NoError(t, err, tt.input)

				dataType, ok := dataTypeAny.(*DataType)
				assert.Equal(t, true, ok, "cannot type cast to *DataType: '%s'", tt.input)

				assert.Equal(t, tt.expected, dataType, tt.input)
			} else {

				assert.ErrorContains(t, err, tt.errorMessage, tt.input)
				assert.Empty(t, dataTypeAny, tt.input)
			}
		})
		if !pass {
			// The earlier test set the basics for later tests, stop as soon as we have an error.
			break
		}
	}
}

func TestParseRecords(t *testing.T) {
	key := "key"

	tests := []struct {
		name         string
		input        string
		expected     *DataType
		errorMessage string
	}{

		// Records
		{
			name: "simple record",
			input: `{
ham: unconstrained;
radio: ref from something;
}`,
			expected: &DataType{
				Key: key,
				Name: `{
ham: unconstrained;
radio: ref from something;
}`,
				CollectionType: "record",
				RecordFields: []Field{
					{
						Name: "ham",
						FieldDataType: &DataType{
							Name:           "unconstrained",
							CollectionType: "atomic",
							Atomic: &Atomic{
								ConstraintType: "unconstrained",
							},
						},
					},
					{
						Name: "radio",
						FieldDataType: &DataType{
							Name:           "ref from something",
							CollectionType: "atomic",
							Atomic: &Atomic{
								ConstraintType: "reference",
								Reference:      "something",
							},
						},
					},
				},
			},
			errorMessage: "",
		},
		{
			name: "nested record",
			input: `{
outer: {
inner: unconstrained;
};
}`,
			expected: &DataType{
				Key: key,
				Name: `{
outer: {
inner: unconstrained;
};
}`,
				CollectionType: "record",
				RecordFields: []Field{
					{
						Name: "outer",
						FieldDataType: &DataType{
							Name: `{
inner: unconstrained;
}`,
							CollectionType: "record",
							RecordFields: []Field{
								{
									Name: "inner",
									FieldDataType: &DataType{
										Name:           "unconstrained",
										CollectionType: "atomic",
										Atomic: &Atomic{
											ConstraintType: "unconstrained",
										},
									},
								},
							},
						},
					},
				},
			},
			errorMessage: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := New(key, tt.input)
			if tt.errorMessage != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
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
	trueValue := true

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
			expected: "ref from some ref",
		},
		{
			name: "collection stack",
			dataType: DataType{
				CollectionType: "stack",
				CollectionMin:  intPtr(0),
				Atomic: &Atomic{
					ConstraintType: "unconstrained",
				},
			},
			expected: "stack of unconstrained",
		},
		{
			name: "collection ordered",
			dataType: DataType{
				CollectionType: "ordered",
				CollectionMin:  intPtr(0),
				Atomic: &Atomic{
					ConstraintType: "unconstrained",
				},
			},
			expected: "ordered collection of unconstrained",
		},
		{
			name: "collection unordered",
			dataType: DataType{
				CollectionType: "unordered",
				CollectionMin:  intPtr(0),
				Atomic: &Atomic{
					ConstraintType: "unconstrained",
				},
			},
			expected: "unordered collection of unconstrained",
		},
		{
			name: "collection queue",
			dataType: DataType{
				CollectionType: "queue",
				CollectionMin:  intPtr(0),
				Atomic: &Atomic{
					ConstraintType: "unconstrained",
				},
			},
			expected: "queue of unconstrained",
		},
		{
			name: "collection with multiplicity",
			dataType: DataType{
				CollectionType: "unordered",
				CollectionMin:  intPtr(3),
				Atomic: &Atomic{
					ConstraintType: "reference",
					Reference:      "something",
				},
			},
			expected: "3+ unordered collection of ref from something",
		},
		{
			name: "collection with unique",
			dataType: DataType{
				CollectionType:   "ordered",
				CollectionUnique: &trueValue,
				CollectionMin:    intPtr(2),
				CollectionMax:    intPtr(5),
				Atomic: &Atomic{
					ConstraintType: "object",
					ObjectClassKey: "class_key",
				},
			},
			expected: "unique 2-5 ordered collection of obj of class_key",
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

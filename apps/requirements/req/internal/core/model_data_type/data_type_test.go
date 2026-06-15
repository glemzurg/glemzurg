package model_data_type

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	validKey := t_dtKey("k")
	tests := []struct {
		key            identity.Key
		collectionType string
		atomic         *Atomic
		errstr         string
	}{
		// OK.
		{
			key:            validKey,
			collectionType: "atomic",
			atomic:         atomic,
		},

		// Error states.
		{
			key:            identity.Key{}, // zero-value: unallocated
			collectionType: "atomic",
			atomic:         atomic,
			errstr:         `Key`,
		},
		{
			key:            validKey,
			collectionType: "",
			atomic:         atomic,
			errstr:         `CollectionType`,
		},
		{
			key:            validKey,
			collectionType: "unknown",
			atomic:         atomic,
			errstr:         `CollectionType`,
		},
		{
			key:            validKey,
			collectionType: "atomic",
			atomic:         nil,
			errstr:         `atomic is required for atomic collection type`,
		},
		{
			key:            validKey,
			collectionType: "atomic",
			atomic:         atomicInvalid,
			errstr:         `ConstraintType`,
		},
	}

	ctx := coreerr.NewContext("test", "")
	for _, tt := range tests {
		dataType := &DataType{
			Key:            tt.key,
			CollectionType: tt.collectionType,
			Atomic:         tt.atomic,
		}
		err := dataType.Validate(ctx)
		if tt.errstr == "" {
			suite.Require().NoError(err, "expected no error for %+v", dataType)
		} else {
			suite.Require().Error(err, "expected error for %+v", dataType)
			suite.Require().ErrorContains(err, tt.errstr, "error message mismatch for %+v", dataType)
		}
	}

	// Collection field rules.
	falseValue := false
	trueValue := true

	collectionTests := []struct {
		name   string
		dt     DataType
		errstr string
	}{
		// Valid collections.
		{
			name: "valid collection no multiplicity",
			dt: DataType{
				Key:              t_dtKey("k"),
				CollectionType:   "stack",
				CollectionUnique: &falseValue,
				Atomic:           atomic,
			},
		},
		{
			name: "valid collection with min and max",
			dt: DataType{
				Key:              t_dtKey("k"),
				CollectionType:   "ordered",
				CollectionUnique: &trueValue,
				CollectionMin:    intPtr(2),
				CollectionMax:    intPtr(5),
				Atomic:           atomic,
			},
		},
		{
			name: "valid collection with min only",
			dt: DataType{
				Key:              t_dtKey("k"),
				CollectionType:   "unordered",
				CollectionUnique: &falseValue,
				CollectionMin:    intPtr(3),
				Atomic:           atomic,
			},
		},
		{
			name: "valid collection with max only",
			dt: DataType{
				Key:              t_dtKey("k"),
				CollectionType:   "queue",
				CollectionUnique: &falseValue,
				CollectionMax:    intPtr(7),
				Atomic:           atomic,
			},
		},

		// Invalid collections.
		{
			name: "collection missing CollectionUnique",
			dt: DataType{
				Key:            t_dtKey("k"),
				CollectionType: "stack",
				Atomic:         atomic,
			},
			errstr: "collection unique is required for collection types",
		},
		{
			name: "collection CollectionMin less than 1",
			dt: DataType{
				Key:              t_dtKey("k"),
				CollectionType:   "stack",
				CollectionUnique: &falseValue,
				CollectionMin:    intPtr(0),
				Atomic:           atomic,
			},
			errstr: "collection min must be at least 1",
		},
		{
			name: "collection CollectionMax less than 1",
			dt: DataType{
				Key:              t_dtKey("k"),
				CollectionType:   "stack",
				CollectionUnique: &falseValue,
				CollectionMax:    intPtr(0),
				Atomic:           atomic,
			},
			errstr: "collection max must be at least 1",
		},
		{
			name: "collection max less than min",
			dt: DataType{
				Key:              t_dtKey("k"),
				CollectionType:   "stack",
				CollectionUnique: &falseValue,
				CollectionMin:    intPtr(5),
				CollectionMax:    intPtr(2),
				Atomic:           atomic,
			},
			errstr: "collection max must be at least collection min",
		},

		// Non-collections must not have collection fields.
		{
			name: "atomic with CollectionUnique",
			dt: DataType{
				Key:              t_dtKey("k"),
				CollectionType:   "atomic",
				CollectionUnique: &falseValue,
				Atomic:           atomic,
			},
			errstr: "collection unique must be nil for non-collection types",
		},
		{
			name: "atomic with CollectionMin",
			dt: DataType{
				Key:            t_dtKey("k"),
				CollectionType: "atomic",
				CollectionMin:  intPtr(1),
				Atomic:         atomic,
			},
			errstr: "collection min must be nil for non-collection types",
		},
		{
			name: "record with CollectionMax",
			dt: DataType{
				Key:            t_dtKey("k"),
				CollectionType: "record",
				CollectionMax:  intPtr(1),
				RecordFields: []Field{
					{Name: "f", FieldDataType: &DataType{Key: t_dtKey("f"), CollectionType: "atomic", Atomic: atomic}},
				},
			},
			errstr: "collection max must be nil for non-collection types",
		},
	}

	for _, tt := range collectionTests {
		err := tt.dt.Validate(ctx)
		if tt.errstr == "" {
			suite.Require().NoError(err, "expected no error for %s", tt.name)
		} else {
			suite.Require().Error(err, "expected error for %s", tt.name)
			suite.Require().ErrorContains(err, tt.errstr, "error message mismatch for %s", tt.name)
		}
	}
}

func TestNewBlank(t *testing.T) {
	key := t_dtKey("k")

	tests := []struct {
		name         string
		input        string
		expected     *DataType
		errorMessage string
	}{
		{
			name:  "blank",
			input: "",
			expected: &DataType{
				Key:            key,
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
			result, err := New(key, tt.input, nil)
			if tt.errorMessage != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestNew(t *testing.T) {
	key := t_dtKey("k")

	tests := []struct {
		name         string
		input        string
		expected     *DataType
		errorMessage string
	}{
		{
			name:  "atomic unconstrained",
			input: "unconstrained",
			expected: &DataType{
				Key:            key,
				CollectionType: "atomic",
				Atomic: &Atomic{
					ConstraintType: "unconstrained",
				},
			},
			errorMessage: "",
		},
		{
			name:  "atomic unconstrained whitespace",
			input: "  \t\n\runconstrained \t\n\r",
			expected: &DataType{
				Key:            key,
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
			result, err := New(key, tt.input, nil)
			if tt.errorMessage != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
			} else {
				require.NoError(t, err)
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
				CollectionType:   "stack",
				CollectionUnique: &falseValue,
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
				CollectionType:   "unordered",
				CollectionUnique: &falseValue,
				Atomic: &Atomic{
					ConstraintType: "reference",
					Reference:      t_StrPtr("something"),
				},
			},
			errorMessage: "",
		},
		{
			name:  "ordered of obj of class_key",
			input: "ordered of obj of class_key",
			expected: &DataType{
				CollectionType:   "ordered",
				CollectionUnique: &falseValue,
				Atomic: &Atomic{
					ConstraintType: "object",
					ObjectClassKey: t_StrPtr("class_key"),
				},
			},
			errorMessage: "",
		},
		{
			name:  "queue of enum of value_a, value_b",
			input: "queue of enum of value_a, value_b",
			expected: &DataType{
				CollectionType:   "queue",
				CollectionUnique: &falseValue,
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
				CollectionType:   "unordered",
				CollectionUnique: &falseValue,
				CollectionMin:    intPtr(3),
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
				CollectionType:   "ordered",
				CollectionUnique: &falseValue,
				CollectionMin:    intPtr(2),
				CollectionMax:    intPtr(5),
				Atomic: &Atomic{
					ConstraintType: "reference",
					Reference:      t_StrPtr("something"),
				},
			},
			errorMessage: "",
		},
		{
			name:  "0-7 queue of obj of class_key",
			input: "0-7 queue of obj of class_key",
			expected: &DataType{
				CollectionType:   "queue",
				CollectionUnique: &falseValue,
				CollectionMax:    intPtr(7),
				Atomic: &Atomic{
					ConstraintType: "object",
					ObjectClassKey: t_StrPtr("class_key"),
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
				Atomic: &Atomic{
					ConstraintType: "unconstrained",
				},
			},
			errorMessage: "",
		},
		{
			name:  "unique unordered of ref from something",
			input: "unique unordered of ref from something",
			expected: &DataType{
				CollectionType:   "unordered",
				CollectionUnique: &trueValue,
				Atomic: &Atomic{
					ConstraintType: "reference",
					Reference:      t_StrPtr("something"),
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
					ObjectClassKey: t_StrPtr("class_key"),
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
			dataTypeAny, err := Parse("", []byte(tt.input), Entrypoint("CollectionDataType"))
			if tt.errorMessage == "" {
				require.NoError(t, err, tt.input)

				dataType, ok := dataTypeAny.(*DataType)
				assert.True(t, ok, "cannot type cast to *DataType: '%s'", tt.input)

				assert.Equal(t, tt.expected, dataType, tt.input)
			} else {
				require.ErrorContains(t, err, tt.errorMessage, tt.input)
				assert.Empty(t, dataTypeAny, tt.input)
			}
		})
		if !pass {
			// The earlier test set the basics for later tests, stop as soon as we have an error.
			break
		}
	}
}

func TestParseRecordFields(t *testing.T) {
	falseValue := false

	tests := []struct {
		name         string
		input        string
		expected     Field
		errorMessage string
	}{

		// Records

		{
			name:  "minimal field",
			input: `ham:unconstrained`,
			expected: Field{
				Name: "ham",
				FieldDataType: &DataType{
					CollectionType: "atomic",
					Atomic: &Atomic{
						ConstraintType: "unconstrained",
					},
				},
			},
			errorMessage: "",
		},
		{
			name:  "minimal field with space",
			input: `ham  :  unconstrained`,
			expected: Field{
				Name: "ham",
				FieldDataType: &DataType{
					CollectionType: "atomic",
					Atomic: &Atomic{
						ConstraintType: "unconstrained",
					},
				},
			},
			errorMessage: "",
		},

		{
			name:  "field with a collection",
			input: `ham  :  unordered of ref from something`,
			expected: Field{
				Name: "ham",
				FieldDataType: &DataType{
					CollectionType:   "unordered",
					CollectionUnique: &falseValue,
					Atomic: &Atomic{
						ConstraintType: "reference",
						Reference:      t_StrPtr("something"),
					},
				},
			},
			errorMessage: "",
		},
	}

	for _, tt := range tests {
		pass := t.Run(tt.name, func(t *testing.T) {
			// Test calling directly into the parser.
			dataTypeAny, err := Parse("", []byte(tt.input), Entrypoint("Field"))
			if tt.errorMessage == "" {
				require.NoError(t, err, tt.input)

				dataType, ok := dataTypeAny.(Field)
				assert.True(t, ok, "cannot type cast to Field: '%s'", tt.input)

				assert.Equal(t, tt.expected, dataType, tt.input)
			} else {
				require.ErrorContains(t, err, tt.errorMessage, tt.input)
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
	tests := []struct {
		name         string
		input        string
		expected     *DataType
		errorMessage string
	}{

		// Records

		{
			name:  "minimal record",
			input: `{ ham : unconstrained }`,
			expected: &DataType{
				CollectionType: "record",
				RecordFields: []Field{
					{
						Name: "ham",
						FieldDataType: &DataType{
							CollectionType: "atomic",
							Atomic: &Atomic{
								ConstraintType: "unconstrained",
							},
						},
					},
				},
			},
			errorMessage: "",
		},

		{
			name:  "simple record",
			input: `{ ham : unconstrained ; sandwich : unconstrained }`,
			expected: &DataType{
				CollectionType: "record",
				RecordFields: []Field{
					{
						Name: "ham",
						FieldDataType: &DataType{
							CollectionType: "atomic",
							Atomic: &Atomic{
								ConstraintType: "unconstrained",
							},
						},
					},
					{
						Name: "sandwich",
						FieldDataType: &DataType{
							CollectionType: "atomic",
							Atomic: &Atomic{
								ConstraintType: "unconstrained",
							},
						},
					},
				},
			},
			errorMessage: "",
		},

		{
			name:  "simple record trailing semicolon",
			input: `{ ham : unconstrained ; sandwich : unconstrained ; }`,
			expected: &DataType{
				CollectionType: "record",
				RecordFields: []Field{
					{
						Name: "ham",
						FieldDataType: &DataType{
							CollectionType: "atomic",
							Atomic: &Atomic{
								ConstraintType: "unconstrained",
							},
						},
					},
					{
						Name: "sandwich",
						FieldDataType: &DataType{
							CollectionType: "atomic",
							Atomic: &Atomic{
								ConstraintType: "unconstrained",
							},
						},
					},
				},
			},
			errorMessage: "",
		},

		{
			name:  "simple record compact",
			input: `{ham:unconstrained;sandwich:unconstrained}`,
			expected: &DataType{
				CollectionType: "record",
				RecordFields: []Field{
					{
						Name: "ham",
						FieldDataType: &DataType{
							CollectionType: "atomic",
							Atomic: &Atomic{
								ConstraintType: "unconstrained",
							},
						},
					},
					{
						Name: "sandwich",
						FieldDataType: &DataType{
							CollectionType: "atomic",
							Atomic: &Atomic{
								ConstraintType: "unconstrained",
							},
						},
					},
				},
			},
			errorMessage: "",
		},

		{
			name:  "nested record",
			input: `{ ham : unconstrained ; sandwich : { grilled : unconstrained ; cheese : unconstrained } }`,
			expected: &DataType{
				CollectionType: "record",
				RecordFields: []Field{
					{
						Name: "ham",
						FieldDataType: &DataType{
							CollectionType: "atomic",
							Atomic: &Atomic{
								ConstraintType: "unconstrained",
							},
						},
					},
					{
						Name: "sandwich",
						FieldDataType: &DataType{
							CollectionType: "record",
							RecordFields: []Field{
								{
									Name: "grilled",
									FieldDataType: &DataType{
										CollectionType: "atomic",
										Atomic: &Atomic{
											ConstraintType: "unconstrained",
										},
									},
								},
								{
									Name: "cheese",
									FieldDataType: &DataType{
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
		pass := t.Run(tt.name, func(t *testing.T) {
			// Test calling directly into the parser.
			dataTypeAny, err := Parse("", []byte(tt.input), Entrypoint("RecordDataType"))
			if tt.errorMessage == "" {
				require.NoError(t, err, tt.input)

				dataType, ok := dataTypeAny.(*DataType)
				assert.True(t, ok, "cannot type cast to *DataType: '%s'", tt.input)

				assert.Equal(t, tt.expected, dataType, tt.input)
			} else {
				require.ErrorContains(t, err, tt.errorMessage, tt.input)
				assert.Empty(t, dataTypeAny, tt.input)
			}
		})
		if !pass {
			// The earlier test set the basics for later tests, stop as soon as we have an error.
			break
		}
	}
}

func TestNewUnparsable(t *testing.T) {
	// If we cannot parse the text, no error but instead just a nil result.
	result, err := New(t_dtKey("k"), "this cannot be parsed so it is just an unparsable blob", nil)
	var targetType *CannotParseError
	require.ErrorAs(t, err, &targetType)
	require.ErrorContains(t, err, "failed to parse")
	assert.Nil(t, result)
}

func TestIsAtomicUnconstrained(t *testing.T) {
	tests := []struct {
		name     string
		dataType *DataType
		want     bool
	}{
		{
			name: "atomic unconstrained",
			dataType: &DataType{
				CollectionType: COLLECTION_TYPE_ATOMIC,
				Atomic:         &Atomic{ConstraintType: CONSTRAINT_TYPE_UNCONSTRAINED},
			},
			want: true,
		},
		{
			name: "atomic span",
			dataType: &DataType{
				CollectionType: COLLECTION_TYPE_ATOMIC,
				Atomic:         &Atomic{ConstraintType: CONSTRAINT_TYPE_SPAN},
			},
			want: false,
		},
		{
			name:     "nil",
			dataType: nil,
			want:     false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, IsAtomicUnconstrained(tc.dataType))
		})
	}
}

func TestNewInvalid(t *testing.T) {
	// Key is required.
	result, err := New(identity.Key{}, "", nil)
	require.ErrorContains(t, err, "Key")
	assert.Nil(t, result)
}

func TestDataTypeString(t *testing.T) {
	trueValue := true
	falseValue := false

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
					Reference:      t_StrPtr("some ref"),
				},
			},
			expected: "ref from some ref",
		},
		{
			name: "collection stack",
			dataType: DataType{
				CollectionType:   "stack",
				CollectionUnique: &falseValue,
				Atomic: &Atomic{
					ConstraintType: "unconstrained",
				},
			},
			expected: "stack of unconstrained",
		},
		{
			name: "collection ordered",
			dataType: DataType{
				CollectionType:   "ordered",
				CollectionUnique: &falseValue,
				Atomic: &Atomic{
					ConstraintType: "unconstrained",
				},
			},
			expected: "ordered collection of unconstrained",
		},
		{
			name: "collection unordered",
			dataType: DataType{
				CollectionType:   "unordered",
				CollectionUnique: &falseValue,
				Atomic: &Atomic{
					ConstraintType: "unconstrained",
				},
			},
			expected: "unordered collection of unconstrained",
		},
		{
			name: "collection queue",
			dataType: DataType{
				CollectionType:   "queue",
				CollectionUnique: &falseValue,
				Atomic: &Atomic{
					ConstraintType: "unconstrained",
				},
			},
			expected: "queue of unconstrained",
		},
		{
			name: "collection with multiplicity",
			dataType: DataType{
				CollectionType:   "unordered",
				CollectionUnique: &falseValue,
				CollectionMin:    intPtr(3),
				Atomic: &Atomic{
					ConstraintType: "reference",
					Reference:      t_StrPtr("something"),
				},
			},
			expected: "3+ unordered collection of ref from something",
		},
		{
			name: "collection with max only",
			dataType: DataType{
				CollectionType:   "queue",
				CollectionUnique: &falseValue,
				CollectionMax:    intPtr(7),
				Atomic: &Atomic{
					ConstraintType: "unconstrained",
				},
			},
			expected: "0-7 queue of unconstrained",
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
					ObjectClassKey: t_StrPtr("class_key"),
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
				assert.PanicsWithValue(t, tt.panicMessage, func() { _ = tt.dataType.String() })
			} else {
				assert.NotPanics(t, func() {
					result := tt.dataType.String()
					assert.Equal(t, tt.expected, result)
				})
			}
		})
	}
}

func (suite *DataTypeSuite) TestUnpackNested() {
	// Create a grandchild DataType
	grandchild := DataType{
		CollectionType: "atomic",
		Atomic:         &Atomic{ConstraintType: "unconstrained"},
	}

	// Create a child DataType with the grandchild as a field
	child := DataType{
		CollectionType: "record",
		RecordFields: []Field{
			{
				Name:          "grandchild",
				FieldDataType: &grandchild,
			},
		},
	}

	// Create the root DataType with the child as a field
	root := DataType{
		Key:            t_dtKey("root"),
		CollectionType: "record",
		RecordFields: []Field{
			{
				Name:          "child",
				FieldDataType: &child,
			},
		},
	}

	// Flatten the nested structure.
	result := root.UnpackNested()

	// Verify the result.
	suite.Len(result, 3)

	rootKey := t_dtKey("root")
	childKey := t_nestedDtKey(rootKey, "child")
	grandchildKey := t_nestedDtKey(childKey, "grandchild")

	// Deepest first: grandchild, child, root.
	suite.Equal(grandchildKey, result[0].Key)
	suite.Equal("atomic", result[0].CollectionType)

	suite.Equal(childKey, result[1].Key)
	suite.Equal("record", result[1].CollectionType)

	suite.Equal(rootKey, result[2].Key)
	suite.Equal("record", result[2].CollectionType)
}

func (suite *DataTypeSuite) TestSortDataTypesByKeyLengthDesc() {
	dataTypes := []DataType{
		{Key: t_dtKey("a")},
		{Key: t_dtKey("bb")},
		{Key: t_dtKey("ccc")},
		{Key: t_dtKey("dddd")},
	}

	SortDataTypesByKeyLengthDesc(dataTypes)

	suite.Equal(t_dtKey("dddd"), dataTypes[0].Key)
	suite.Equal(t_dtKey("ccc"), dataTypes[1].Key)
	suite.Equal(t_dtKey("bb"), dataTypes[2].Key)
	suite.Equal(t_dtKey("a"), dataTypes[3].Key)
}

func (suite *DataTypeSuite) TestExtractDatabaseObjects() {
	// Create test DataTypes
	atomic := DataType{
		Key:            t_dtKey("atomic_key"),
		CollectionType: "atomic",
		Atomic:         &Atomic{ConstraintType: "unconstrained"},
	}

	atomicSpan := DataType{
		Key:            t_dtKey("atomic_span_key"),
		CollectionType: "atomic",
		Atomic: &Atomic{
			ConstraintType: "span",
			Span: &AtomicSpan{
				LowerType:  "unconstrained",
				HigherType: "unconstrained",
			},
		},
	}

	atomicEnum := DataType{
		Key:            t_dtKey("atomic_enum_key"),
		CollectionType: "atomic",
		Atomic: &Atomic{
			ConstraintType: "enumeration",
			Enums: []AtomicEnum{
				{Value: "ValueA"},
				{Value: "ValueB"},
			},
		},
	}

	record := DataType{
		Key:            t_dtKey("record_key"),
		CollectionType: "record",
		RecordFields: []Field{
			{
				Name:          "name",
				FieldDataType: &DataType{Key: t_dtKey("field_type")},
			},
		},
	}

	dataTypes := []DataType{record, atomic, atomicSpan, atomicEnum}

	fieldMap, atomicMap, atomicSpanMap, atomicEnumMap := ExtractDatabaseObjects(dataTypes)

	recordKeyStr := t_dtKey("record_key").String()
	atomicKeyStr := t_dtKey("atomic_key").String()
	atomicSpanKeyStr := t_dtKey("atomic_span_key").String()
	atomicEnumKeyStr := t_dtKey("atomic_enum_key").String()

	suite.Equal(map[string][]Field{
		recordKeyStr: {
			{
				Name:          "name",
				FieldDataType: &DataType{Key: t_dtKey("field_type")},
			},
		},
	}, fieldMap)

	suite.Equal(map[string]Atomic{
		atomicKeyStr: {ConstraintType: "unconstrained"},
		atomicSpanKeyStr: Atomic{
			ConstraintType: "span",
			Span: &AtomicSpan{
				LowerType:  "unconstrained",
				HigherType: "unconstrained",
			},
		},
		atomicEnumKeyStr: Atomic{
			ConstraintType: "enumeration",
			Enums: []AtomicEnum{
				{Value: "ValueA"},
				{Value: "ValueB"},
			},
		},
	}, atomicMap)

	suite.Equal(map[string]AtomicSpan{
		atomicSpanKeyStr: {LowerType: "unconstrained", HigherType: "unconstrained"},
	}, atomicSpanMap)

	suite.Equal(map[string][]AtomicEnum{
		atomicEnumKeyStr: {
			{Value: "ValueA"},
			{Value: "ValueB"},
		},
	}, atomicEnumMap)
}

func (suite *DataTypeSuite) TestReconstituteDataTypes() {
	// Create base DataTypes (only Key and CollectionType)
	baseDataTypes := []DataType{
		{Key: t_dtKey("atomic_key"), CollectionType: "atomic"},
		{Key: t_dtKey("atomic_span_key"), CollectionType: "atomic"},
		{Key: t_dtKey("atomic_enum_key"), CollectionType: "atomic"},
		{Key: t_dtKey("record_key"), CollectionType: "record"},
	}

	// Create the maps with the same data as TestExtractDatabaseObjects.
	// Map keys are the .String() form of identity.Key (same as what ExtractDatabaseObjects produces).
	recordKeyStr := t_dtKey("record_key").String()
	atomicKeyStr := t_dtKey("atomic_key").String()
	atomicSpanKeyStr := t_dtKey("atomic_span_key").String()
	atomicEnumKeyStr := t_dtKey("atomic_enum_key").String()
	fieldMap := map[string][]Field{
		recordKeyStr: {
			{
				Name:          "name",
				FieldDataType: &DataType{Key: t_dtKey("field_type")},
			},
		},
	}

	atomicMap := map[string]Atomic{
		atomicKeyStr: {ConstraintType: "unconstrained"},
		atomicSpanKeyStr: Atomic{
			ConstraintType: "span",
			Span: &AtomicSpan{
				LowerType:  "unconstrained",
				HigherType: "unconstrained",
			},
		},
		atomicEnumKeyStr: Atomic{
			ConstraintType: "enumeration",
			Enums: []AtomicEnum{
				{Value: "ValueA"},
				{Value: "ValueB"},
			},
		},
	}

	atomicSpanMap := map[string]AtomicSpan{
		atomicSpanKeyStr: {LowerType: "unconstrained", HigherType: "unconstrained"},
	}

	atomicEnumMap := map[string][]AtomicEnum{
		atomicEnumKeyStr: {
			{Value: "ValueA"},
			{Value: "ValueB"},
		},
	}

	// Call ReconstituteDataTypes
	result := ReconstituteDataTypes(baseDataTypes, fieldMap, atomicMap, atomicSpanMap, atomicEnumMap)

	// Verify the result is sorted by key length descending and components are attached
	suite.Equal([]DataType{
		{
			Key:            t_dtKey("atomic_enum_key"),
			CollectionType: "atomic",
			Atomic: &Atomic{
				ConstraintType: "enumeration",
				Enums:          []AtomicEnum{{Value: "ValueA"}, {Value: "ValueB"}},
			},
		},
		{
			Key:            t_dtKey("atomic_span_key"),
			CollectionType: "atomic",
			Atomic: &Atomic{
				ConstraintType: "span",
				Span:           &AtomicSpan{LowerType: "unconstrained", HigherType: "unconstrained"},
			},
		},
		{
			Key:            t_dtKey("atomic_key"),
			CollectionType: "atomic",
			Atomic:         &Atomic{ConstraintType: "unconstrained"},
		},
		{
			Key:            t_dtKey("record_key"),
			CollectionType: "record",
			RecordFields: []Field{
				{
					Name:          "name",
					FieldDataType: &DataType{Key: t_dtKey("field_type")},
				},
			},
		},
	}, result)
}

func (suite *DataTypeSuite) TestFlattenAndReconstructNested() {
	// Create a three-deep nested structure
	grandchild := DataType{
		Key:            t_dtKey("grandchild"),
		CollectionType: "atomic",
		Atomic:         &Atomic{ConstraintType: "unconstrained"},
	}

	child := DataType{
		Key:            t_dtKey("child"),
		CollectionType: "record",
		RecordFields: []Field{
			{
				Name:          "grandchild_field",
				FieldDataType: &grandchild,
			},
		},
	}

	root := DataType{
		Key:            t_dtKey("root1"),
		CollectionType: "record",
		RecordFields: []Field{
			{
				Name:          "child_field",
				FieldDataType: &child,
			},
		},
	}

	// Create another root with the same structure

	// Create a three-deep nested structure
	grandchild2 := DataType{
		Key:            t_dtKey("grandchild2"),
		CollectionType: "atomic",
		Atomic:         &Atomic{ConstraintType: "unconstrained"},
	}

	child2 := DataType{
		Key:            t_dtKey("child2"),
		CollectionType: "record",
		RecordFields: []Field{
			{
				Name:          "grandchild_field",
				FieldDataType: &grandchild2,
			},
		},
	}

	root2 := DataType{
		Key:            t_dtKey("root2"),
		CollectionType: "record",
		RecordFields: []Field{
			{
				Name:          "other_child_field",
				FieldDataType: &child2,
			},
		},
	}

	original := []DataType{root, root2}

	// Flatten
	flat := FlattenDataTypes(original)

	// Reconstruct
	reconstructed := ReconstructNestedDataTypes(flat)

	// Verify that reconstructed matches original
	suite.Equal(original, reconstructed)
}

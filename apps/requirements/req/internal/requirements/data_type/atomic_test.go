package data_type

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseAtomic(t *testing.T) {

	key := "key"
	trueValue := true
	falseValue := false

	tests := []struct {
		name         string
		input        string
		expected     *DataType
		errorMessage string
	}{

		// Unconstrained atomics.
		{
			name:  "empty string",
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
			name:  "whitespace",
			input: "   \t\n",
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

		// References.
		{
			name:  "ref",
			input: "ref: listed somewhere else",
			expected: &DataType{
				Key:            key,
				Name:           "ref: listed somewhere else",
				CollectionType: "atomic",
				Atomic: &Atomic{
					ConstraintType: "reference",
					Reference:      "listed somewhere else",
				},
			},
			errorMessage: "",
		},
		{
			name:  "reference",
			input: "reference: listed somewhere else",
			expected: &DataType{
				Key:            key,
				Name:           "ref: listed somewhere else",
				CollectionType: "atomic",
				Atomic: &Atomic{
					ConstraintType: "reference",
					Reference:      "listed somewhere else",
				},
			},
			errorMessage: "",
		},
		{
			name:  "ref with whitespace",
			input: "   \t\nref   \t\n:    \t\nlisted somewhere else    \t\n",
			expected: &DataType{
				Key:            key,
				Name:           "ref: listed somewhere else",
				CollectionType: "atomic",
				Atomic: &Atomic{
					ConstraintType: "reference",
					Reference:      "listed somewhere else",
				},
			},
			errorMessage: "",
		},
		{
			name:  "reference with whitespace",
			input: "   \t\nreference   \t\n:    \t\nlisted somewhere else    \t\n",
			expected: &DataType{
				Key:            key,
				Name:           "ref: listed somewhere else",
				CollectionType: "atomic",
				Atomic: &Atomic{
					ConstraintType: "reference",
					Reference:      "listed somewhere else",
				},
			},
			errorMessage: "",
		},

		// Objects.
		{
			name:  "obj",
			input: "obj: class_key",
			expected: &DataType{
				Key:            key,
				Name:           "obj: class_key",
				CollectionType: "atomic",
				Atomic: &Atomic{
					ConstraintType: "object",
					ObjectClassKey: "class_key",
				},
			},
			errorMessage: "",
		},
		{
			name:  "object",
			input: "object: class_key",
			expected: &DataType{
				Key:            key,
				Name:           "obj: class_key",
				CollectionType: "atomic",
				Atomic: &Atomic{
					ConstraintType: "object",
					ObjectClassKey: "class_key",
				},
			},
			errorMessage: "",
		},
		{
			name:  "obj with whitespace",
			input: "   \t\nobj   \t\n:    \t\nclass_key    \t\n",
			expected: &DataType{
				Key:            key,
				Name:           "obj: class_key",
				CollectionType: "atomic",
				Atomic: &Atomic{
					ConstraintType: "object",
					ObjectClassKey: "class_key",
				},
			},
			errorMessage: "",
		},
		{
			name:  "object with whitespace",
			input: "   \t\nobject   \t\n:    \t\nclass_key    \t\n",
			expected: &DataType{
				Key:            key,
				Name:           "obj: class_key",
				CollectionType: "atomic",
				Atomic: &Atomic{
					ConstraintType: "object",
					ObjectClassKey: "class_key",
				},
			},
			errorMessage: "",
		},

		// Enumeration.
		{
			name:  "enum on value",
			input: "enum: value_a",
			expected: &DataType{
				Key:            key,
				Name:           "enum: value_a",
				CollectionType: "atomic",
				Atomic: &Atomic{
					ConstraintType: "enumeration",
					EnumOrdered:    &falseValue,
					Enums: []AtomicEnum{
						{Value: "value_a"},
					},
				},
			},
			errorMessage: "",
		},
		{
			name:  "enum",
			input: "enum: value_a, value_b, value_c",
			expected: &DataType{
				Key:            key,
				Name:           "enum: value_a, value_b, value_c",
				CollectionType: "atomic",
				Atomic: &Atomic{
					ConstraintType: "enumeration",
					EnumOrdered:    &falseValue,
					Enums: []AtomicEnum{
						{Value: "value_a"},
						{Value: "value_b"},
						{Value: "value_c"},
					},
				},
			},
			errorMessage: "",
		},
		{
			name:  "enumeration",
			input: "enumeration: value_a, value_b, value_c",
			expected: &DataType{
				Key:            key,
				Name:           "enum: value_a, value_b, value_c",
				CollectionType: "atomic",
				Atomic: &Atomic{
					ConstraintType: "enumeration",
					EnumOrdered:    &falseValue,
					Enums: []AtomicEnum{
						{Value: "value_a"},
						{Value: "value_b"},
						{Value: "value_c"},
					},
				},
			},
			errorMessage: "",
		},
		{
			name:  "enum with whitespace",
			input: "   \t\nenum   \t\n:    \t\n  value_a  \t\n ,  \t\n  value_b  \t\n ,   \t\n value_c    \t\n",
			expected: &DataType{
				Key:            key,
				Name:           "enum: value_a, value_b, value_c",
				CollectionType: "atomic",
				Atomic: &Atomic{
					ConstraintType: "enumeration",
					EnumOrdered:    &falseValue,
					Enums: []AtomicEnum{
						{Value: "value_a"},
						{Value: "value_b"},
						{Value: "value_c"},
					},
				},
			},
			errorMessage: "",
		},
		{
			name:  "enumeration with whitespace",
			input: "   \t\nenumeration   \t\n:    \t\n  value_a  \t\n ,  \t\n  value_b  \t\n ,   \t\n value_c    \t\n",
			expected: &DataType{
				Key:            key,
				Name:           "enum: value_a, value_b, value_c",
				CollectionType: "atomic",
				Atomic: &Atomic{
					ConstraintType: "enumeration",
					EnumOrdered:    &falseValue,
					Enums: []AtomicEnum{
						{Value: "value_a"},
						{Value: "value_b"},
						{Value: "value_c"},
					},
				},
			},
			errorMessage: "",
		},

		{
			name:  "ord enum with whitespace",
			input: "   \t\nord-enum   \t\n:    \t\n  value_a  \t\n ,  \t\n  value_b  \t\n ,   \t\n value_c    \t\n",
			expected: &DataType{
				Key:            key,
				Name:           "ord-enum: value_a, value_b, value_c",
				CollectionType: "atomic",
				Atomic: &Atomic{
					ConstraintType: "enumeration",
					EnumOrdered:    &trueValue,
					Enums: []AtomicEnum{
						{Value: "value_a"},
						{Value: "value_b"},
						{Value: "value_c"},
					},
				},
			},
			errorMessage: "",
		},
		{
			name:  "ordered enumeration with whitespace",
			input: "   \t\nordered-enumeration   \t\n:    \t\n  value_a  \t\n ,  \t\n  value_b  \t\n ,   \t\n value_c    \t\n",
			expected: &DataType{
				Key:            key,
				Name:           "ord-enum: value_a, value_b, value_c",
				CollectionType: "atomic",
				Atomic: &Atomic{
					ConstraintType: "enumeration",
					EnumOrdered:    &trueValue,
					Enums: []AtomicEnum{
						{Value: "value_a"},
						{Value: "value_b"},
						{Value: "value_c"},
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

func TestAtomicString(t *testing.T) {
	trueValue := true
	falseValue := false
	tests := []struct {
		name         string
		atomic       Atomic
		expected     string
		panicMessage string
	}{
		{
			name: "unconstrained",
			atomic: Atomic{
				ConstraintType: "unconstrained",
			},
			expected: "unconstrained",
		},
		{
			name: "reference",
			atomic: Atomic{
				ConstraintType: "reference",
				Reference:      "listed somewhere else",
			},
			expected: "ref: listed somewhere else",
		},
		{
			name: "reference empty",
			atomic: Atomic{
				ConstraintType: "reference",
				Reference:      "",
			},
			expected: "ref: ",
		},
		{
			name: "object",
			atomic: Atomic{
				ConstraintType: "object",
				ObjectClassKey: "some_class",
			},
			expected: "obj: some_class",
		},
		{
			name: "enumeration",
			atomic: Atomic{
				ConstraintType: "enumeration",
				EnumOrdered:    &falseValue,
				Enums: []AtomicEnum{
					{Value: "value_a"},
					{Value: "value_b"},
					{Value: "value_c"},
				},
			},
			expected: "enum: value_a, value_b, value_c",
		},
		{
			name: "ordered enumeration",
			atomic: Atomic{
				ConstraintType: "enumeration",
				EnumOrdered:    &trueValue,
				Enums: []AtomicEnum{
					{Value: "value_a"},
					{Value: "value_b"},
					{Value: "value_c"},
				},
			},
			expected: "ord-enum: value_a, value_b, value_c",
		},
		{
			name: "unknown type",
			atomic: Atomic{
				ConstraintType: "unknown",
			},
			panicMessage: "invalid constraint type: 'unknown'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.panicMessage != "" {
				assert.PanicsWithValue(t, tt.panicMessage, func() { tt.atomic.String() })
			} else {
				result := tt.atomic.String()
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

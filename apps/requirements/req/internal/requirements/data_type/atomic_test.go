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
			input: "ref from listed somewhere else",
			expected: &DataType{
				Key:            key,
				Name:           "ref from listed somewhere else",
				CollectionType: "atomic",
				Atomic: &Atomic{
					ConstraintType: "reference",
					Reference:      "listed somewhere else",
				},
			},
			errorMessage: "",
		},
		{
			name:  "ref of",
			input: "ref of listed somewhere else",
			expected: &DataType{
				Key:            key,
				Name:           "ref from listed somewhere else",
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
			input: "reference from listed somewhere else",
			expected: &DataType{
				Key:            key,
				Name:           "ref from listed somewhere else",
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
			input: "   \t\nref   \t\nfrom    \t\nlisted somewhere else    \t\n",
			expected: &DataType{
				Key:            key,
				Name:           "ref from listed somewhere else",
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
			input: "   \t\nreference   \t\nfrom    \t\nlisted somewhere else    \t\n",
			expected: &DataType{
				Key:            key,
				Name:           "ref from listed somewhere else",
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
			input: "obj of class_key",
			expected: &DataType{
				Key:            key,
				Name:           "obj of class_key",
				CollectionType: "atomic",
				Atomic: &Atomic{
					ConstraintType: "object",
					ObjectClassKey: "class_key",
				},
			},
			errorMessage: "",
		},
		{
			name:  "obj from",
			input: "obj from class_key",
			expected: &DataType{
				Key:            key,
				Name:           "obj of class_key",
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
			input: "object of class_key",
			expected: &DataType{
				Key:            key,
				Name:           "obj of class_key",
				CollectionType: "atomic",
				Atomic: &Atomic{
					ConstraintType: "object",
					ObjectClassKey: "class_key",
				},
			},
			errorMessage: "",
		},
		{
			name:  "object from",
			input: "object from class_key",
			expected: &DataType{
				Key:            key,
				Name:           "obj of class_key",
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
			input: "   \t\nobj   \t\nof    \t\nclass_key    \t\n",
			expected: &DataType{
				Key:            key,
				Name:           "obj of class_key",
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
			input: "   \t\nobject   \t\nof    \t\nclass_key    \t\n",
			expected: &DataType{
				Key:            key,
				Name:           "obj of class_key",
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
			name:  "enum of value",
			input: "enum of value_a",
			expected: &DataType{
				Key:            key,
				Name:           "enum of value_a",
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
			name:  "enum from value",
			input: "enum from value_a",
			expected: &DataType{
				Key:            key,
				Name:           "enum of value_a",
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
			input: "enum of value_a, value_b, value_c",
			expected: &DataType{
				Key:            key,
				Name:           "enum of value_a, value_b, value_c",
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
			name:  "ordered enumeration",
			input: "ordered enumeration of value_a, value_b, value_c",
			expected: &DataType{
				Key:            key,
				Name:           "ord-enum of value_a, value_b, value_c",
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
			name:  "ordered enum with whitespace",
			input: "   \t\nordered \t\n enum   \t\nof    \t\n  value_a  \t\n ,  \t\n  value_b  \t\n ,   \t\n value_c    \t\n",
			expected: &DataType{
				Key:            key,
				Name:           "enum of value_a, value_b, value_c",
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
			input: "   \t\nenumeration   \t\nof    \t\n  value_a  \t\n ,  \t\n  value_b  \t\n ,   \t\n value_c    \t\n",
			expected: &DataType{
				Key:            key,
				Name:           "enum of value_a, value_b, value_c",
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
			input: "   \t\nord-enum   \t\nof    \t\n  value_a  \t\n ,  \t\n  value_b  \t\n ,   \t\n value_c    \t\n",
			expected: &DataType{
				Key:            key,
				Name:           "ord-enum of value_a, value_b, value_c",
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
			input: "   \t\nordered-enumeration   \t\nof    \t\n  value_a  \t\n ,  \t\n  value_b  \t\n ,   \t\n value_c    \t\n",
			expected: &DataType{
				Key:            key,
				Name:           "ord-enum of value_a, value_b, value_c",
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

		// Spans.
		{
			name:  "span simple",
			input: "(3/4 .. 5/6] at 0.01 meter",
			expected: &DataType{
				Key:            key,
				Name:           "(3/4 .. 5/6] at 0.01 meter",
				CollectionType: "atomic",
				Atomic: &Atomic{
					ConstraintType: "span",
					Span: &AtomicSpan{
						LowerType:         "open",
						LowerValue:        intPtr(3),
						LowerDenominator:  intPtr(4),
						HigherType:        "closed",
						HigherValue:       intPtr(5),
						HigherDenominator: intPtr(6),
						Units:             "meter",
						Precision:         0.01,
					},
				},
			},
			errorMessage: "",
		},
		{
			name:  "span simple with whitespace",
			input: "(  \t\n3/4 \t\n ..  \t\n5/6 \t\n] \t\n at \t\n 0.01 \t\n meter",
			expected: &DataType{
				Key:            key,
				Name:           "(3/4 .. 5/6] at 0.01 meter",
				CollectionType: "atomic",
				Atomic: &Atomic{
					ConstraintType: "span",
					Span: &AtomicSpan{
						LowerType:         "open",
						LowerValue:        intPtr(3),
						LowerDenominator:  intPtr(4),
						HigherType:        "closed",
						HigherValue:       intPtr(5),
						HigherDenominator: intPtr(6),
						Units:             "meter",
						Precision:         0.01,
					},
				},
			},
			errorMessage: "",
		},
		{
			name:  "span simple, minimal whitespace",
			input: "(3/4..5/6]at 0.01 meter",
			expected: &DataType{
				Key:            key,
				Name:           "(3/4 .. 5/6] at 0.01 meter",
				CollectionType: "atomic",
				Atomic: &Atomic{
					ConstraintType: "span",
					Span: &AtomicSpan{
						LowerType:         "open",
						LowerValue:        intPtr(3),
						LowerDenominator:  intPtr(4),
						HigherType:        "closed",
						HigherValue:       intPtr(5),
						HigherDenominator: intPtr(6),
						Units:             "meter",
						Precision:         0.01,
					},
				},
			},
			errorMessage: "",
		},
		{
			name:  "span closed lower open higher",
			input: "[3/4 .. 5/6) at 1 meter",
			expected: &DataType{
				Key:            key,
				Name:           "[3/4 .. 5/6) at 1 meter",
				CollectionType: "atomic",
				Atomic: &Atomic{
					ConstraintType: "span",
					Span: &AtomicSpan{
						LowerType:         "closed",
						LowerValue:        intPtr(3),
						LowerDenominator:  intPtr(4),
						HigherType:        "open",
						HigherValue:       intPtr(5),
						HigherDenominator: intPtr(6),
						Units:             "meter",
						Precision:         1.0,
					},
				},
			},
			errorMessage: "",
		},
		{
			name:  "span both closed",
			input: "[3/4 .. 5/6] at 0.01 meter",
			expected: &DataType{
				Key:            key,
				Name:           "[3/4 .. 5/6] at 0.01 meter",
				CollectionType: "atomic",
				Atomic: &Atomic{
					ConstraintType: "span",
					Span: &AtomicSpan{
						LowerType:         "closed",
						LowerValue:        intPtr(3),
						LowerDenominator:  intPtr(4),
						HigherType:        "closed",
						HigherValue:       intPtr(5),
						HigherDenominator: intPtr(6),
						Units:             "meter",
						Precision:         0.01,
					},
				},
			},
			errorMessage: "",
		},
		{
			name:  "span both open",
			input: "(3/4 .. 5/6) at 0.01 meter",
			expected: &DataType{
				Key:            key,
				Name:           "(3/4 .. 5/6) at 0.01 meter",
				CollectionType: "atomic",
				Atomic: &Atomic{
					ConstraintType: "span",
					Span: &AtomicSpan{
						LowerType:         "open",
						LowerValue:        intPtr(3),
						LowerDenominator:  intPtr(4),
						HigherType:        "open",
						HigherValue:       intPtr(5),
						HigherDenominator: intPtr(6),
						Units:             "meter",
						Precision:         0.01,
					},
				},
			},
			errorMessage: "",
		},
		{
			name:  "span unconstrained lower",
			input: "(unconstrained .. 5/6] at 0.01 meter",
			expected: &DataType{
				Key:            key,
				Name:           "(unconstrained .. 5/6] at 0.01 meter",
				CollectionType: "atomic",
				Atomic: &Atomic{
					ConstraintType: "span",
					Span: &AtomicSpan{
						LowerType:         "unconstrained",
						HigherType:        "closed",
						HigherValue:       intPtr(5),
						HigherDenominator: intPtr(6),
						Units:             "meter",
						Precision:         0.01,
					},
				},
			},
			errorMessage: "",
		},
		{
			name:  "span unconstrained higher",
			input: "[3/4 .. unconstrained) at 0.01 meter",
			expected: &DataType{
				Key:            key,
				Name:           "[3/4 .. unconstrained) at 0.01 meter",
				CollectionType: "atomic",
				Atomic: &Atomic{
					ConstraintType: "span",
					Span: &AtomicSpan{
						LowerType:        "closed",
						LowerValue:       intPtr(3),
						LowerDenominator: intPtr(4),
						HigherType:       "unconstrained",
						Units:            "meter",
						Precision:        0.01,
					},
				},
			},
			errorMessage: "",
		},
		{
			name:  "span without denominators",
			input: "(3 .. 5] at 0.01 meter",
			expected: &DataType{
				Key:            key,
				Name:           "(3 .. 5] at 0.01 meter",
				CollectionType: "atomic",
				Atomic: &Atomic{
					ConstraintType: "span",
					Span: &AtomicSpan{
						LowerType:         "open",
						LowerValue:        intPtr(3),
						LowerDenominator:  intPtr(1),
						HigherType:        "closed",
						HigherValue:       intPtr(5),
						HigherDenominator: intPtr(1),
						Units:             "meter",
						Precision:         0.01,
					},
				},
			},
			errorMessage: "",
		},
		{
			name:  "span different units",
			input: "(1/2 .. 3/4] at 0.001 kilogram",
			expected: &DataType{
				Key:            key,
				Name:           "(1/2 .. 3/4] at 0.001 kilogram",
				CollectionType: "atomic",
				Atomic: &Atomic{
					ConstraintType: "span",
					Span: &AtomicSpan{
						LowerType:         "open",
						LowerValue:        intPtr(1),
						LowerDenominator:  intPtr(2),
						HigherType:        "closed",
						HigherValue:       intPtr(3),
						HigherDenominator: intPtr(4),
						Units:             "kilogram",
						Precision:         0.001,
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
			expected: "ref from listed somewhere else",
		},
		{
			name: "reference empty",
			atomic: Atomic{
				ConstraintType: "reference",
				Reference:      "",
			},
			expected: "ref from ",
		},
		{
			name: "object",
			atomic: Atomic{
				ConstraintType: "object",
				ObjectClassKey: "some_class",
			},
			expected: "obj of some_class",
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
			expected: "enum of value_a, value_b, value_c",
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
			expected: "ord-enum of value_a, value_b, value_c",
		},
		{
			name: "span",
			atomic: Atomic{
				ConstraintType: "span",
				Span: &AtomicSpan{
					LowerType:         "open",
					LowerValue:        intPtr(3),
					LowerDenominator:  intPtr(4),
					HigherType:        "closed",
					HigherValue:       intPtr(5),
					HigherDenominator: intPtr(6),
					Units:             "meter",
					Precision:         0.01,
				},
			},
			expected: "(3/4 .. 5/6] at 0.01 meter",
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

package data_type

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseAtomic(t *testing.T) {

	key := "key"

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

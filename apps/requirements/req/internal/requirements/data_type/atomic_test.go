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

		{
			name:         "non-whitespace",
			input:        "x",
			expected:     nil,
			errorMessage: "no match found",
		},

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

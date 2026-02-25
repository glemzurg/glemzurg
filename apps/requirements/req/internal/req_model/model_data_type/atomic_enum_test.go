package model_data_type

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAtomicEnumValidate(t *testing.T) {
	tests := []struct {
		name         string
		atomicEnum   AtomicEnum
		errorMessage string
	}{
		{
			name: "valid",
			atomicEnum: AtomicEnum{
				Value:     "some value",
				SortOrder: -1, // Allow negative sort orders.
			},
			errorMessage: "",
		},
		{
			name: "empty value",
			atomicEnum: AtomicEnum{
				Value:     "",
				SortOrder: 1,
			},
			errorMessage: "Value",
		},
		{
			name: "zero sort order",
			atomicEnum: AtomicEnum{
				Value:     "value",
				SortOrder: 0,
			},
			errorMessage: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.atomicEnum.Validate()
			if tt.errorMessage != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

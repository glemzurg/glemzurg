package data_type

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAtomicSpanValidate(t *testing.T) {
	tests := []struct {
		name         string
		atomicSpan   AtomicSpan
		errorMessage string
	}{
		{
			name: "valid closed bounds",
			atomicSpan: AtomicSpan{
				LowerType:         "closed",
				LowerValue:        intPtr(10),
				LowerDenominator:  intPtr(1),
				HigherType:        "closed",
				HigherValue:       intPtr(20),
				HigherDenominator: intPtr(1),
				Units:             "meters",
				Precision:         2,
			},
			errorMessage: "",
		},
		{
			name: "valid with unconstrained lower",
			atomicSpan: AtomicSpan{
				LowerType:         "unconstrained",
				HigherType:        "open",
				HigherValue:       intPtr(100),
				HigherDenominator: intPtr(2),
				Units:             "seconds",
			},
			errorMessage: "",
		},
		{
			name: "valid with unconstrained higher",
			atomicSpan: AtomicSpan{
				LowerType:        "open",
				LowerValue:       intPtr(5),
				LowerDenominator: intPtr(1),
				HigherType:       "unconstrained",
				Units:            "kg",
			},
			errorMessage: "",
		},
		{
			name: "missing LowerType",
			atomicSpan: AtomicSpan{
				LowerValue:       intPtr(1),
				LowerDenominator: intPtr(1),
				HigherType:       "unconstrained",
				Units:            "m",
			},
			errorMessage: "LowerType: cannot be blank.",
		},
		{
			name: "invalid LowerType",
			atomicSpan: AtomicSpan{
				LowerType:        "invalid",
				LowerValue:       intPtr(1),
				LowerDenominator: intPtr(1),
				HigherType:       "unconstrained",
				Units:            "m",
			},
			errorMessage: "LowerType: must be a valid value.",
		},
		{
			name: "missing HigherType",
			atomicSpan: AtomicSpan{
				LowerType:         "unconstrained",
				HigherValue:       intPtr(1),
				HigherDenominator: intPtr(1),
				Units:             "m",
			},
			errorMessage: "HigherType: cannot be blank.",
		},
		{
			name: "invalid HigherType",
			atomicSpan: AtomicSpan{
				LowerType:         "unconstrained",
				HigherType:        "invalid",
				HigherValue:       intPtr(1),
				HigherDenominator: intPtr(1),
				Units:             "m",
			},
			errorMessage: "HigherType: must be a valid value.",
		},
		{
			name: "missing Units",
			atomicSpan: AtomicSpan{
				LowerType:  "unconstrained",
				HigherType: "unconstrained",
			},
			errorMessage: "Units: cannot be blank.",
		},
		{
			name: "LowerType closed but LowerValue nil",
			atomicSpan: AtomicSpan{
				LowerType:        "closed",
				LowerDenominator: intPtr(1),
				HigherType:       "unconstrained",
				Units:            "m",
			},
			errorMessage: "LowerValue: cannot be blank.",
		},
		{
			name: "LowerType closed but LowerDenominator nil",
			atomicSpan: AtomicSpan{
				LowerType:  "closed",
				LowerValue: intPtr(10),
				HigherType: "unconstrained",
				Units:      "m",
			},
			errorMessage: "LowerDenominator: cannot be blank.",
		},
		{
			name: "LowerDenominator < 1",
			atomicSpan: AtomicSpan{
				LowerType:        "closed",
				LowerValue:       intPtr(10),
				LowerDenominator: intPtr(0),
				HigherType:       "closed",
				HigherValue:      intPtr(20),
				Units:            "m",
			},
			errorMessage: "LowerDenominator: must be no less than 1.",
		},
		{
			name: "HigherType open but HigherValue nil",
			atomicSpan: AtomicSpan{
				LowerType:         "unconstrained",
				HigherType:        "open",
				HigherDenominator: intPtr(1),
				Units:             "m",
			},
			errorMessage: "HigherValue: cannot be blank.",
		},
		{
			name: "HigherType open but HigherDenominator nil",
			atomicSpan: AtomicSpan{
				LowerType:   "unconstrained",
				HigherType:  "open",
				HigherValue: intPtr(1),
				Units:       "m",
			},
			errorMessage: "HigherDenominator: cannot be blank.",
		},
		{
			name: "HigherDenominator < 1",
			atomicSpan: AtomicSpan{
				LowerType:         "unconstrained",
				HigherType:        "closed",
				HigherValue:       intPtr(20),
				HigherDenominator: intPtr(0),
				Units:             "m",
			},
			errorMessage: "HigherDenominator: must be no less than 1.",
		},
		{
			name: "LowerDenominator < 1",
			atomicSpan: AtomicSpan{
				LowerType:        "closed",
				LowerValue:       intPtr(10),
				LowerDenominator: intPtr(0),
				HigherType:       "unconstrained",
				Units:            "m",
			},
			errorMessage: "LowerDenominator: must be no less than 1.",
		},
		{
			name: "LowerDenominator nil when unconstrained",
			atomicSpan: AtomicSpan{
				LowerType:  "unconstrained",
				HigherType: "unconstrained",
				Units:      "m",
			},
			errorMessage: "",
		},
		{
			name: "HigherDenominator nil when unconstrained",
			atomicSpan: AtomicSpan{
				LowerType:  "unconstrained",
				HigherType: "unconstrained",
				Units:      "m",
			},
			errorMessage: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.atomicSpan.Validate()
			if tt.errorMessage != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func intPtr(i int) *int {
	return &i
}

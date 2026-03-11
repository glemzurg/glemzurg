package model_data_type

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
				Precision:         1.0,
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
				Precision:         1.0,
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
				Precision:        0.01,
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
				Precision:        1.0,
			},
			errorMessage: "LowerType",
		},
		{
			name: "invalid LowerType",
			atomicSpan: AtomicSpan{
				LowerType:        "invalid",
				LowerValue:       intPtr(1),
				LowerDenominator: intPtr(1),
				HigherType:       "unconstrained",
				Units:            "m",
				Precision:        1.0,
			},
			errorMessage: "LowerType",
		},
		{
			name: "missing HigherType",
			atomicSpan: AtomicSpan{
				LowerType:         "unconstrained",
				HigherValue:       intPtr(1),
				HigherDenominator: intPtr(1),
				Units:             "m",
				Precision:         1.0,
			},
			errorMessage: "HigherType",
		},
		{
			name: "invalid HigherType",
			atomicSpan: AtomicSpan{
				LowerType:         "unconstrained",
				HigherType:        "invalid",
				HigherValue:       intPtr(1),
				HigherDenominator: intPtr(1),
				Units:             "m",
				Precision:         1.0,
			},
			errorMessage: "HigherType",
		},
		{
			name: "missing Units",
			atomicSpan: AtomicSpan{
				LowerType:  "unconstrained",
				HigherType: "unconstrained",
				Precision:  1.0,
			},
			errorMessage: "Units",
		},
		{
			name: "LowerType closed but LowerValue nil",
			atomicSpan: AtomicSpan{
				LowerType:        "closed",
				LowerDenominator: intPtr(1),
				HigherType:       "unconstrained",
				Units:            "m",
				Precision:        1.0,
			},
			errorMessage: "DTYPE_SPAN_LOWERVAL_REQUIRED",
		},
		{
			name: "LowerType closed but LowerDenominator nil",
			atomicSpan: AtomicSpan{
				LowerType:  "closed",
				LowerValue: intPtr(10),
				HigherType: "unconstrained",
				Units:      "m",
				Precision:  1.0,
			},
			errorMessage: "DTYPE_SPAN_LOWERDENOM_REQUIRED",
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
				Precision:        1.0,
			},
			errorMessage: "DTYPE_SPAN_LOWERDENOM_INVALID",
		},
		{
			name: "HigherType open but HigherValue nil",
			atomicSpan: AtomicSpan{
				LowerType:         "unconstrained",
				HigherType:        "open",
				HigherDenominator: intPtr(1),
				Units:             "m",
				Precision:         1.0,
			},
			errorMessage: "DTYPE_SPAN_HIGHERVAL_REQUIRED",
		},
		{
			name: "HigherType open but HigherDenominator nil",
			atomicSpan: AtomicSpan{
				LowerType:   "unconstrained",
				HigherType:  "open",
				HigherValue: intPtr(1),
				Units:       "m",
				Precision:   1.0,
			},
			errorMessage: "DTYPE_SPAN_HIGHERDENOM_REQUIRED",
		},
		{
			name: "HigherDenominator < 1",
			atomicSpan: AtomicSpan{
				LowerType:         "unconstrained",
				HigherType:        "closed",
				HigherValue:       intPtr(20),
				HigherDenominator: intPtr(0),
				Units:             "m",
				Precision:         1.0,
			},
			errorMessage: "DTYPE_SPAN_HIGHERDENOM_INVALID",
		},
		{
			name: "LowerDenominator < 1",
			atomicSpan: AtomicSpan{
				LowerType:        "closed",
				LowerValue:       intPtr(10),
				LowerDenominator: intPtr(0),
				HigherType:       "unconstrained",
				Units:            "m",
				Precision:        1.0,
			},
			errorMessage: "DTYPE_SPAN_LOWERDENOM_INVALID",
		},
		{
			name: "LowerDenominator nil when unconstrained",
			atomicSpan: AtomicSpan{
				LowerType:  "unconstrained",
				HigherType: "unconstrained",
				Units:      "m",
				Precision:  1.0,
			},
			errorMessage: "",
		},
		{
			name: "HigherDenominator nil when unconstrained",
			atomicSpan: AtomicSpan{
				LowerType:  "unconstrained",
				HigherType: "unconstrained",
				Units:      "m",
				Precision:  1.0,
			},
			errorMessage: "",
		},
		{
			name: "missing Precision",
			atomicSpan: AtomicSpan{
				LowerType:  "unconstrained",
				HigherType: "unconstrained",
				Units:      "m",
			},
			errorMessage: "Precision",
		},
		{
			name: "Precision <= 0",
			atomicSpan: AtomicSpan{
				LowerType:  "unconstrained",
				HigherType: "unconstrained",
				Units:      "m",
				Precision:  -1.0,
			},
			errorMessage: "DTYPE_SPAN_PRECISION_INVALID",
		},
		{
			name: "Precision > 1",
			atomicSpan: AtomicSpan{
				LowerType:  "unconstrained",
				HigherType: "unconstrained",
				Units:      "m",
				Precision:  2.0,
			},
			errorMessage: "DTYPE_SPAN_PRECISION_INVALID",
		},
		{
			name: "Precision not power of 10",
			atomicSpan: AtomicSpan{
				LowerType:  "unconstrained",
				HigherType: "unconstrained",
				Units:      "m",
				Precision:  0.5,
			},
			errorMessage: "DTYPE_SPAN_PRECISION_NOT_POW10",
		},
		{
			name: "valid Precision 0.1",
			atomicSpan: AtomicSpan{
				LowerType:  "unconstrained",
				HigherType: "unconstrained",
				Units:      "m",
				Precision:  1.0,
			},
			errorMessage: "",
		},
		{
			name: "valid Precision 0.01",
			atomicSpan: AtomicSpan{
				LowerType:  "unconstrained",
				HigherType: "unconstrained",
				Units:      "m",
				Precision:  1.0,
			},
			errorMessage: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := coreerr.NewContext("test", "")
			err := tt.atomicSpan.Validate(ctx)
			if tt.errorMessage != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func intPtr(i int) *int {
	return &i
}

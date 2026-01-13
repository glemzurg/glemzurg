package model_class

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestMultiplicitySuite(t *testing.T) {
	suite.Run(t, new(MultiplicitySuite))
}

type MultiplicitySuite struct {
	suite.Suite
}

func (suite *MultiplicitySuite) TestNew() {
	tests := []struct {
		testName string
		value    string
		obj      Multiplicity
		errstr   string
	}{
		// OK.
		{
			testName: "ok with range",
			value:    "2..3",
			obj: Multiplicity{
				LowerBound:  2,
				HigherBound: 3,
			},
		},
		{
			testName: "ok with any",
			value:    "any",
			obj: Multiplicity{
				LowerBound:  0,
				HigherBound: 0,
			},
		},

		// Error states.
		{
			testName: "error with unknown value",
			value:    "unknown",
			errstr:   `invalid multiplicity: 'unknown'`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			obj, err := NewMultiplicity(tt.value)
			if tt.errstr == "" {
				assert.NoError(t, err)
				assert.Equal(t, tt.obj, obj)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
				assert.Empty(t, obj)
			}
		})
	}
}

func (suite *MultiplicitySuite) TestParseMultiplicity() {

	tests := []struct {
		testName     string
		multiplicity string
		lowerBound   uint
		higherBound  uint
		errstr       string
	}{
		// OK.
		{
			testName:     "ok with 0..1",
			multiplicity: MULTIPLICITY_0_1,
			lowerBound:   0,
			higherBound:  1,
		},
		{
			testName:     "ok with any",
			multiplicity: MULTIPLICITY_ANY,
			lowerBound:   0,
			higherBound:  0,
		},
		{
			testName:     "ok with 1",
			multiplicity: MULTIPLICITY_1,
			lowerBound:   1,
			higherBound:  1,
		},
		{
			testName:     "ok with 1..many",
			multiplicity: MULTIPLICITY_1_MANY,
			lowerBound:   1,
			higherBound:  0,
		},

		// Free-form parse.
		{
			testName:     "ok with 0..3",
			multiplicity: "0..3",
			lowerBound:   0,
			higherBound:  3,
		},
		{
			testName:     "ok with 1..3",
			multiplicity: "1..3",
			lowerBound:   1,
			higherBound:  3,
		},
		{
			testName:     "ok with 2..3",
			multiplicity: "2..3",
			lowerBound:   2,
			higherBound:  3,
		},
		{
			testName:     "ok with 3..3",
			multiplicity: "3..3",
			lowerBound:   3,
			higherBound:  3,
		},
		{
			testName:     "ok with 3",
			multiplicity: "3",
			lowerBound:   3,
			higherBound:  3,
		},
		{
			testName:     "ok with any..3",
			multiplicity: "any..3",
			lowerBound:   0,
			higherBound:  3,
		},
		{
			testName:     "ok with 3..many",
			multiplicity: "3..many",
			lowerBound:   3,
			higherBound:  0,
		},
		{
			testName:     "ok with 3..0",
			multiplicity: "3..0",
			lowerBound:   3,
			higherBound:  0,
		},
		{
			testName:     "ok with any..many",
			multiplicity: "any..many",
			lowerBound:   0,
			higherBound:  0,
		},
		{
			testName:     "ok with many..any",
			multiplicity: "many..any",
			lowerBound:   0,
			higherBound:  0,
		},
		{
			testName:     "ok with any..any",
			multiplicity: "any..any",
			lowerBound:   0,
			higherBound:  0,
		},
		{
			testName:     "ok with many..many",
			multiplicity: "many..many",
			lowerBound:   0,
			higherBound:  0,
		},

		// Errors.
		{
			testName:     "error with any---3",
			multiplicity: "any---3",
			errstr:       `invalid multiplicity: 'any---3'`,
		},
		{
			testName:     "error with ham..3",
			multiplicity: "ham..3",
			errstr:       `invalid multiplicity: 'ham..3'`,
		},
		{
			testName:     "error with 3..sandwitch",
			multiplicity: "3..sandwitch",
			errstr:       `invalid multiplicity: '3..sandwitch'`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			lowerBound, higherBound, err := parseMultiplicity(tt.multiplicity)
			if tt.errstr == "" {
				assert.NoError(t, err)
				assert.Equal(t, tt.lowerBound, lowerBound)
				assert.Equal(t, tt.higherBound, higherBound)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
				assert.Empty(t, lowerBound)
				assert.Empty(t, higherBound)
			}
		})
	}
}

func (suite *MultiplicitySuite) TestString() {

	tests := []struct {
		testName     string
		multiplicity Multiplicity
		value        string
	}{
		// Single value.
		{
			testName:     "ok with 0..0 as *",
			multiplicity: Multiplicity{LowerBound: 0, HigherBound: 0},
			value:        "*",
		},
		{
			testName:     "ok with 1..1 as 1",
			multiplicity: Multiplicity{LowerBound: 1, HigherBound: 1},
			value:        "1",
		},
		{
			testName:     "ok with 2..2 as 2",
			multiplicity: Multiplicity{LowerBound: 2, HigherBound: 2},
			value:        "2",
		},
		{
			testName:     "ok with 3..3 as 3",
			multiplicity: Multiplicity{LowerBound: 3, HigherBound: 3},
			value:        "3",
		},

		// Range.
		{
			testName:     "ok with 0..3",
			multiplicity: Multiplicity{LowerBound: 0, HigherBound: 3},
			value:        "0..3",
		},
		{
			testName:     "ok with 1..3",
			multiplicity: Multiplicity{LowerBound: 1, HigherBound: 3},
			value:        "1..3",
		},
		{
			testName:     "ok with 2..3",
			multiplicity: Multiplicity{LowerBound: 2, HigherBound: 3},
			value:        "2..3",
		},

		// No top.
		{
			testName:     "ok with 1..0 as 1..*",
			multiplicity: Multiplicity{LowerBound: 1, HigherBound: 0},
			value:        "1..*",
		},
		{
			testName:     "ok with 2..0 as 2..*",
			multiplicity: Multiplicity{LowerBound: 2, HigherBound: 0},
			value:        "2..*",
		},
		{
			testName:     "ok with 3..0 as 3..*",
			multiplicity: Multiplicity{LowerBound: 3, HigherBound: 0},
			value:        "3..*",
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			value := tt.multiplicity.String()
			assert.Equal(t, tt.value, value)
		})
	}
}

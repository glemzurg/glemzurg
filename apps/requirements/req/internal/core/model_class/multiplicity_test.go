package model_class

import (
	"fmt"
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

// TestValidate tests the Validate method directly.
func (suite *MultiplicitySuite) TestValidate() {
	tests := []struct {
		name   string
		obj    Multiplicity
		errstr string
	}{
		// Valid cases.
		{
			name: "both zero (any)",
			obj:  Multiplicity{LowerBound: 0, HigherBound: 0},
		},
		{
			name: "equal bounds",
			obj:  Multiplicity{LowerBound: 2, HigherBound: 2},
		},
		{
			name: "higher > lower",
			obj:  Multiplicity{LowerBound: 1, HigherBound: 3},
		},
		{
			name: "lower zero (any), higher set",
			obj:  Multiplicity{LowerBound: 0, HigherBound: 5},
		},
		{
			name: "higher zero (unlimited), lower set",
			obj:  Multiplicity{LowerBound: 3, HigherBound: 0},
		},
		// Invalid cases.
		{
			name:   "higher < lower",
			obj:    Multiplicity{LowerBound: 5, HigherBound: 2},
			errstr: "higher bound (2) must be >= lower bound (5)",
		},
	}
	for _, test := range tests {
		suite.T().Run(test.name, func(t *testing.T) {
			err := test.obj.Validate()
			if test.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, test.errstr)
			}
		})
	}
}

// TestNew tests that NewMultiplicity populates the struct and calls Validate.
func (suite *MultiplicitySuite) TestNew() {
	// Test struct population.
	obj, err := NewMultiplicity("2..3")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), Multiplicity{LowerBound: 2, HigherBound: 3}, obj, "struct should be populated")

	// Test that Validate is called (parsing error).
	_, err = NewMultiplicity("unknown")
	assert.ErrorContains(suite.T(), err, "invalid multiplicity", "NewMultiplicity should fail on invalid input")
}

func (suite *MultiplicitySuite) TestParseMultiplicity() {

	tests := []struct {
		multiplicity string
		lowerBound   uint
		higherBound  uint
		errstr       string
	}{
		// OK.
		{
			multiplicity: MULTIPLICITY_0_1,
			lowerBound:   0,
			higherBound:  1,
		},
		{
			multiplicity: MULTIPLICITY_ANY,
			lowerBound:   0,
			higherBound:  0,
		},
		{
			multiplicity: MULTIPLICITY_1,
			lowerBound:   1,
			higherBound:  1,
		},
		{
			multiplicity: MULTIPLICITY_1_MANY,
			lowerBound:   1,
			higherBound:  0,
		},

		// Free-form parse.
		{
			multiplicity: "0..3",
			lowerBound:   0,
			higherBound:  3,
		},
		{
			multiplicity: "1..3",
			lowerBound:   1,
			higherBound:  3,
		},
		{
			multiplicity: "2..3",
			lowerBound:   2,
			higherBound:  3,
		},
		{
			multiplicity: "3..3",
			lowerBound:   3,
			higherBound:  3,
		},
		{
			multiplicity: "3",
			lowerBound:   3,
			higherBound:  3,
		},
		{
			multiplicity: "any..3",
			lowerBound:   0,
			higherBound:  3,
		},
		{
			multiplicity: "3..many",
			lowerBound:   3,
			higherBound:  0,
		},
		{
			multiplicity: "3..0",
			lowerBound:   3,
			higherBound:  0,
		},
		{
			multiplicity: "any..many",
			lowerBound:   0,
			higherBound:  0,
		},
		{
			multiplicity: "many..any",
			lowerBound:   0,
			higherBound:  0,
		},
		{
			multiplicity: "any..any",
			lowerBound:   0,
			higherBound:  0,
		},
		{
			multiplicity: "many..many",
			lowerBound:   0,
			higherBound:  0,
		},

		// Errors.
		{
			multiplicity: "any---3",
			errstr:       `invalid multiplicity: 'any---3'`,
		},
		{
			multiplicity: "ham..3",
			errstr:       `invalid multiplicity: 'ham..3'`,
		},
		{
			multiplicity: "3..sandwitch",
			errstr:       `invalid multiplicity: '3..sandwitch'`,
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		lowerBound, higherBound, err := parseMultiplicity(test.multiplicity)
		if test.errstr == "" {
			assert.Nil(suite.T(), err, testName)
			assert.Equal(suite.T(), test.lowerBound, lowerBound, testName)
			assert.Equal(suite.T(), test.higherBound, higherBound, testName)
		} else {
			assert.ErrorContains(suite.T(), err, test.errstr, testName)
			assert.Empty(suite.T(), lowerBound, testName)
			assert.Empty(suite.T(), higherBound, testName)
		}
	}
}

// TestParsedString tests that ParsedString converts "*" to "any" and passes other values through.
func (suite *MultiplicitySuite) TestParsedString() {
	tests := []struct {
		name         string
		multiplicity Multiplicity
		expected     string
	}{
		{
			name:         "star becomes any",
			multiplicity: Multiplicity{LowerBound: 0, HigherBound: 0},
			expected:     "any",
		},
		{
			name:         "single value unchanged",
			multiplicity: Multiplicity{LowerBound: 1, HigherBound: 1},
			expected:     "1",
		},
		{
			name:         "range unchanged",
			multiplicity: Multiplicity{LowerBound: 1, HigherBound: 3},
			expected:     "1..3",
		},
		{
			name:         "no upper bound unchanged",
			multiplicity: Multiplicity{LowerBound: 2, HigherBound: 0},
			expected:     "2..*",
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.multiplicity.ParsedString())
		})
	}
}

func (suite *MultiplicitySuite) TestString() {

	tests := []struct {
		multiplicity Multiplicity
		value        string
	}{
		// Single value.
		{
			multiplicity: Multiplicity{LowerBound: 0, HigherBound: 0},
			value:        "*",
		},
		{
			multiplicity: Multiplicity{LowerBound: 1, HigherBound: 1},
			value:        "1",
		},
		{
			multiplicity: Multiplicity{LowerBound: 2, HigherBound: 2},
			value:        "2",
		},
		{
			multiplicity: Multiplicity{LowerBound: 3, HigherBound: 3},
			value:        "3",
		},

		// Range.
		{
			multiplicity: Multiplicity{LowerBound: 0, HigherBound: 3},
			value:        "0..3",
		},
		{
			multiplicity: Multiplicity{LowerBound: 1, HigherBound: 3},
			value:        "1..3",
		},
		{
			multiplicity: Multiplicity{LowerBound: 2, HigherBound: 3},
			value:        "2..3",
		},

		// No top.
		{
			multiplicity: Multiplicity{LowerBound: 1, HigherBound: 0},
			value:        "1..*",
		},
		{
			multiplicity: Multiplicity{LowerBound: 2, HigherBound: 0},
			value:        "2..*",
		},
		{
			multiplicity: Multiplicity{LowerBound: 3, HigherBound: 0},
			value:        "3..*",
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		value := test.multiplicity.String()
		assert.Equal(suite.T(), test.value, value, testName)
	}
}

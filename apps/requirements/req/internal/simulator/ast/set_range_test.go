package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestSetRangeSuite(t *testing.T) {
	suite.Run(t, new(SetRangeSuite))
}

type SetRangeSuite struct {
	suite.Suite
}

func (suite *SetRangeSuite) TestString() {
	tests := []struct {
		testName string
		start    int
		end      int
		expected string
	}{
		{
			testName: `1 to 12`,
			start:    1,
			end:      12,
			expected: `1 .. 12`,
		},
		{
			testName: `0 to 100`,
			start:    0,
			end:      100,
			expected: `0 .. 100`,
		},
		{
			testName: `negative range`,
			start:    -5,
			end:      5,
			expected: `-5 .. 5`,
		},
		{
			testName: `single element`,
			start:    42,
			end:      42,
			expected: `42 .. 42`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			s := &SetRange{
				Start: tt.start,
				End:   tt.end,
			}
			assert.Equal(t, tt.expected, s.String())
		})
	}
}

func (suite *SetRangeSuite) TestAscii() {
	tests := []struct {
		testName string
		start    int
		end      int
		expected string
	}{
		{
			testName: `1 to 12`,
			start:    1,
			end:      12,
			expected: `1 .. 12`,
		},
		{
			testName: `0 to 100`,
			start:    0,
			end:      100,
			expected: `0 .. 100`,
		},
		{
			testName: `negative range`,
			start:    -5,
			end:      5,
			expected: `-5 .. 5`,
		},
		{
			testName: `single element`,
			start:    42,
			end:      42,
			expected: `42 .. 42`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			s := &SetRange{
				Start: tt.start,
				End:   tt.end,
			}
			assert.Equal(t, tt.expected, s.Ascii())
		})
	}
}

func (suite *SetRangeSuite) TestValidate() {
	tests := []struct {
		testName string
		s        *SetRange
		errstr   string
	}{
		// OK.
		{
			testName: `valid range`,
			s: &SetRange{
				Start: 1,
				End:   12,
			},
		},
		{
			testName: `valid single element`,
			s: &SetRange{
				Start: 5,
				End:   5,
			},
		},
		{
			testName: `valid zero start`,
			s: &SetRange{
				Start: 0,
				End:   10,
			},
		},
		{
			testName: `valid negative range`,
			s: &SetRange{
				Start: -5,
				End:   5,
			},
		},

		// Errors.
		{
			testName: `error start greater than end`,
			s: &SetRange{
				Start: 10,
				End:   1,
			},
			errstr: `Start`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.s.Validate()
			if tt.errstr == `` {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

func (suite *SetRangeSuite) TestExpressionNode() {
	// Verify that SetRange implements the expressionNode interface method.
	s := &SetRange{
		Start: 1,
		End:   12,
	}
	// This should compile and not panic.
	s.expressionNode()
}

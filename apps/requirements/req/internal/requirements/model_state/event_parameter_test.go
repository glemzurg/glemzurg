package model_state

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestEventParameterSuite(t *testing.T) {
	suite.Run(t, new(EventParameterSuite))
}

type EventParameterSuite struct {
	suite.Suite
}

func (suite *EventParameterSuite) TestNew() {
	tests := []struct {
		testName string
		name     string
		source   string
		obj      EventParameter
		errstr   string
	}{
		// OK.
		{
			testName: "ok with all fields",
			name:     "Name",
			source:   "Source",
			obj: EventParameter{
				Name:   "Name",
				Source: "Source",
			},
		},

		// Error states.
		{
			testName: "error with blank name",
			name:     "",
			source:   "Source",
			errstr:   `Name: cannot be blank`,
		},
		{
			testName: "error with blank source",
			name:     "Name",
			source:   "",
			errstr:   `Source: cannot be blank`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			obj, err := NewEventParameter(tt.name, tt.source)
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

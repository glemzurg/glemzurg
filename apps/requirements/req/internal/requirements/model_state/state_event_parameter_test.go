package model_state

import (
	"fmt"
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
		name   string
		source string
		obj    EventParameter
		errstr string
	}{
		// OK.
		{
			name:   "Name",
			source: "Source",
			obj: EventParameter{
				Name:   "Name",
				Source: "Source",
			},
		},

		// Error states.
		{
			name:   "",
			source: "Source",
			errstr: `Name: cannot be blank`,
		},
		{
			name:   "Name",
			source: "",
			errstr: `Source: cannot be blank`,
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		obj, err := NewEventParameter(test.name, test.source)
		if test.errstr == "" {
			assert.Nil(suite.T(), err, testName)
			assert.Equal(suite.T(), test.obj, obj, testName)
		} else {
			assert.ErrorContains(suite.T(), err, test.errstr, testName)
			assert.Empty(suite.T(), obj, testName)
		}
	}
}

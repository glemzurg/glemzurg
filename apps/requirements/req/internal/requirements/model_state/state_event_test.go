package model_state

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestEventSuite(t *testing.T) {
	suite.Run(t, new(EventSuite))
}

type EventSuite struct {
	suite.Suite
}

func (suite *EventSuite) TestNew() {
	tests := []struct {
		key        string
		name       string
		details    string
		parameters []EventParameter
		obj        Event
		errstr     string
	}{
		// OK.
		{
			key:        "Key",
			name:       "Name",
			details:    "Details",
			parameters: []EventParameter{{Name: "ParamA", Source: "SourceA"}, {Name: "ParamB", Source: "SourceB"}},
			obj: Event{
				Key:        "Key",
				Name:       "Name",
				Details:    "Details",
				Parameters: []EventParameter{{Name: "ParamA", Source: "SourceA"}, {Name: "ParamB", Source: "SourceB"}},
			},
		},
		{
			key:        "Key",
			name:       "Name",
			details:    "",
			parameters: nil,
			obj: Event{
				Key:        "Key",
				Name:       "Name",
				Details:    "",
				Parameters: nil,
			},
		},

		// Error states.
		{
			key:        "",
			name:       "Name",
			details:    "Details",
			parameters: []EventParameter{{Name: "ParamA", Source: "SourceA"}, {Name: "ParamB", Source: "SourceB"}},
			errstr:     `Key: cannot be blank`,
		},
		{
			key:        "Key",
			name:       "",
			details:    "Details",
			parameters: []EventParameter{{Name: "ParamA", Source: "SourceA"}, {Name: "ParamB", Source: "SourceB"}},
			errstr:     `Name: cannot be blank`,
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		obj, err := NewEvent(test.key, test.name, test.details, test.parameters)
		if test.errstr == "" {
			assert.Nil(suite.T(), err, testName)
			assert.Equal(suite.T(), test.obj, obj, testName)
		} else {
			assert.ErrorContains(suite.T(), err, test.errstr, testName)
			assert.Empty(suite.T(), obj, testName)
		}
	}
}

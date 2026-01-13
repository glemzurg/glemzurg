package model_state

import (
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
		testName   string
		key        string
		name       string
		details    string
		parameters []EventParameter
		obj        Event
		errstr     string
	}{
		// OK.
		{
			testName:   "ok with all fields",
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
			testName:   "ok with minimal fields",
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
			testName:   "error with blank key",
			key:        "",
			name:       "Name",
			details:    "Details",
			parameters: []EventParameter{{Name: "ParamA", Source: "SourceA"}, {Name: "ParamB", Source: "SourceB"}},
			errstr:     `Key: cannot be blank`,
		},
		{
			testName:   "error with blank name",
			key:        "Key",
			name:       "",
			details:    "Details",
			parameters: []EventParameter{{Name: "ParamA", Source: "SourceA"}, {Name: "ParamB", Source: "SourceB"}},
			errstr:     `Name: cannot be blank`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			obj, err := NewEvent(tt.key, tt.name, tt.details, tt.parameters)
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

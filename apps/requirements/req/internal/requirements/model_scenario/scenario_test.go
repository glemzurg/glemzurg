package model_scenario

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestScenarioSuite(t *testing.T) {
	suite.Run(t, new(ScenarioSuite))
}

type ScenarioSuite struct {
	suite.Suite
}

func (suite *ScenarioSuite) TestNew() {
	tests := []struct {
		testName string
		key      string
		name     string
		details  string
		obj      Scenario
		errstr   string
	}{
		// OK.
		{
			testName: "ok with all fields",
			key:      "Key",
			name:     "Name",
			details:  "Details",
			obj: Scenario{
				Key:     "Key",
				Name:    "Name",
				Details: "Details",
			},
		},
		{
			testName: "ok with minimal fields",
			key:      "Key",
			name:     "Name",
			details:  "",
			obj: Scenario{
				Key:     "Key",
				Name:    "Name",
				Details: "",
			},
		},

		// Error states.
		{
			testName: "error with blank key",
			key:      "",
			name:     "Name",
			details:  "Details",
			errstr:   `Key: cannot be blank`,
		},
		{
			testName: "error with blank name",
			key:      "Key",
			name:     "",
			details:  "Details",
			errstr:   `Name: cannot be blank`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			obj, err := NewScenario(tt.key, tt.name, tt.details)
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

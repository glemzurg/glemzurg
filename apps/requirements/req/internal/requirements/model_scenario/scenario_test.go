package model_scenario

import (
	"fmt"
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
		key     string
		name    string
		details string
		obj     Scenario
		errstr  string
	}{
		// OK.
		{
			key:     "Key",
			name:    "Name",
			details: "Details",
			obj: Scenario{
				Key:     "Key",
				Name:    "Name",
				Details: "Details",
			},
		},
		{
			key:     "Key",
			name:    "Name",
			details: "",
			obj: Scenario{
				Key:     "Key",
				Name:    "Name",
				Details: "",
			},
		},

		// Error states.
		{
			key:     "",
			name:    "Name",
			details: "Details",
			errstr:  `Key: cannot be blank`,
		},
		{
			key:     "Key",
			name:    "",
			details: "Details",
			errstr:  `Name: cannot be blank`,
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		obj, err := NewScenario(test.key, test.name, test.details)
		if test.errstr == "" {
			assert.Nil(suite.T(), err, testName)
			assert.Equal(suite.T(), test.obj, obj, testName)
		} else {
			assert.ErrorContains(suite.T(), err, test.errstr, testName)
			assert.Empty(suite.T(), obj, testName)
		}
	}
}

package model_scenario

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
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

	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	useCaseKey := helper.Must(identity.NewUseCaseKey(subdomainKey, "usecase1"))

	tests := []struct {
		testName string
		key      identity.Key
		name     string
		details  string
		obj      Scenario
		errstr   string
	}{
		// OK.
		{
			testName: "ok with all fields",
			key:      helper.Must(identity.NewScenarioKey(useCaseKey, "scenario1")),
			name:     "Name",
			details:  "Details",
			obj: Scenario{
				Key:     helper.Must(identity.NewScenarioKey(useCaseKey, "scenario1")),
				Name:    "Name",
				Details: "Details",
			},
		},
		{
			testName: "ok with minimal fields",
			key:      helper.Must(identity.NewScenarioKey(useCaseKey, "scenario2")),
			name:     "Name",
			details:  "",
			obj: Scenario{
				Key:     helper.Must(identity.NewScenarioKey(useCaseKey, "scenario2")),
				Name:    "Name",
				Details: "",
			},
		},

		// Error states.
		{
			testName: "error empty key",
			key:      identity.Key{},
			name:     "Name",
			details:  "Details",
			errstr:   "keyType: cannot be blank",
		},
		{
			testName: "error wrong key type",
			key:      helper.Must(identity.NewDomainKey("domain1")),
			name:     "Name",
			details:  "Details",
			errstr:   "Key: invalid key type 'domain' for scenario.",
		},
		{
			testName: "error with blank name",
			key:      helper.Must(identity.NewScenarioKey(useCaseKey, "scenario3")),
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

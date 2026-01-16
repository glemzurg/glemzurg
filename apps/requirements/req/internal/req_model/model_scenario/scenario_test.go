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

// TestValidate tests all validation rules for Scenario.
func (suite *ScenarioSuite) TestValidate() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	useCaseKey := helper.Must(identity.NewUseCaseKey(subdomainKey, "usecase1"))
	validKey := helper.Must(identity.NewScenarioKey(useCaseKey, "scenario1"))

	tests := []struct {
		testName string
		scenario Scenario
		errstr   string
	}{
		{
			testName: "valid scenario",
			scenario: Scenario{
				Key:  validKey,
				Name: "Name",
			},
		},
		{
			testName: "error empty key",
			scenario: Scenario{
				Key:  identity.Key{},
				Name: "Name",
			},
			errstr: "keyType: cannot be blank",
		},
		{
			testName: "error wrong key type",
			scenario: Scenario{
				Key:  domainKey,
				Name: "Name",
			},
			errstr: "Key: invalid key type 'domain' for scenario.",
		},
		{
			testName: "error blank name",
			scenario: Scenario{
				Key:  validKey,
				Name: "",
			},
			errstr: "Name: cannot be blank",
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.scenario.Validate()
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

// TestNew tests that NewScenario maps parameters correctly and calls Validate.
func (suite *ScenarioSuite) TestNew() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	useCaseKey := helper.Must(identity.NewUseCaseKey(subdomainKey, "usecase1"))
	key := helper.Must(identity.NewScenarioKey(useCaseKey, "scenario1"))

	// Test parameters are mapped correctly.
	scenario, err := NewScenario(key, "Name", "Details")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), key, scenario.Key)
	assert.Equal(suite.T(), "Name", scenario.Name)
	assert.Equal(suite.T(), "Details", scenario.Details)

	// Test that Validate is called (invalid data should fail).
	_, err = NewScenario(key, "", "Details")
	assert.ErrorContains(suite.T(), err, "Name: cannot be blank")
}

// TestValidateWithParent tests that ValidateWithParent calls Validate and ValidateParent.
func (suite *ScenarioSuite) TestValidateWithParent() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	useCaseKey := helper.Must(identity.NewUseCaseKey(subdomainKey, "usecase1"))
	validKey := helper.Must(identity.NewScenarioKey(useCaseKey, "scenario1"))
	otherUseCaseKey := helper.Must(identity.NewUseCaseKey(subdomainKey, "other_usecase"))

	// Test that Validate is called.
	scenario := Scenario{
		Key:  validKey,
		Name: "", // Invalid
	}
	err := scenario.ValidateWithParent(&useCaseKey)
	assert.ErrorContains(suite.T(), err, "Name: cannot be blank", "ValidateWithParent should call Validate()")

	// Test that ValidateParent is called - scenario key has usecase1 as parent, but we pass other_usecase.
	scenario = Scenario{
		Key:  validKey,
		Name: "Name",
	}
	err = scenario.ValidateWithParent(&otherUseCaseKey)
	assert.ErrorContains(suite.T(), err, "does not match expected parent", "ValidateWithParent should call ValidateParent()")

	// Test valid case.
	err = scenario.ValidateWithParent(&useCaseKey)
	assert.NoError(suite.T(), err)
}

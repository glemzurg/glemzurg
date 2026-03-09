package model_scenario

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
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
			errstr: "key type is required",
		},
		{
			testName: "error wrong key type",
			scenario: Scenario{
				Key:  domainKey,
				Name: "Name",
			},
			errstr: "key: invalid key type 'domain' for scenario",
		},
		{
			testName: "error blank name",
			scenario: Scenario{
				Key:  validKey,
				Name: "",
			},
			errstr: "Name",
		},
	}
	for _, tt := range tests {
		suite.Run(tt.testName, func() {
			err := tt.scenario.Validate()
			if tt.errstr == "" {
				suite.Require().NoError(err)
			} else {
				suite.Require().ErrorContains(err, tt.errstr)
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

	scenario := NewScenario(key, "Name", "Details")
	suite.Equal(Scenario{
		Key:     key,
		Name:    "Name",
		Details: "Details",
	}, scenario)
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
	suite.Require().ErrorContains(err, "Name", "ValidateWithParent should call Validate()")

	// Test that ValidateParent is called - scenario key has usecase1 as parent, but we pass other_usecase.
	scenario = Scenario{
		Key:  validKey,
		Name: "Name",
	}
	err = scenario.ValidateWithParent(&otherUseCaseKey)
	suite.Require().ErrorContains(err, "does not match expected parent", "ValidateWithParent should call ValidateParent()")

	// Test valid case.
	err = scenario.ValidateWithParent(&useCaseKey)
	suite.Require().NoError(err)
}

// TestValidateWithParentAndClasses tests that ValidateWithParentAndClasses validates child Objects.
func (suite *ScenarioSuite) TestValidateWithParentAndClasses() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	useCaseKey := helper.Must(identity.NewUseCaseKey(subdomainKey, "usecase1"))
	scenarioKey := helper.Must(identity.NewScenarioKey(useCaseKey, "scenario1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	objectKey := helper.Must(identity.NewScenarioObjectKey(scenarioKey, "obj1"))
	nonExistentClassKey := helper.Must(identity.NewClassKey(subdomainKey, "nonexistent"))

	classes := map[identity.Key]bool{
		classKey: true,
	}

	// Test valid scenario with valid Object child.
	scenario := Scenario{
		Key:  scenarioKey,
		Name: "Name",
		Objects: map[identity.Key]Object{
			objectKey: {Key: objectKey, ObjectNumber: 1, Name: "Obj", NameStyle: "name", ClassKey: classKey},
		},
	}
	err := scenario.ValidateWithParentAndClasses(&useCaseKey, classes)
	suite.Require().NoError(err)

	// Test invalid child Object (blank name with name style) propagates error.
	scenario = Scenario{
		Key:  scenarioKey,
		Name: "Name",
		Objects: map[identity.Key]Object{
			objectKey: {Key: objectKey, ObjectNumber: 1, Name: "", NameStyle: "name", ClassKey: classKey}, // Invalid: name required for "name" style
		},
	}
	err = scenario.ValidateWithParentAndClasses(&useCaseKey, classes)
	suite.Require().ErrorContains(err, "Name", "Should validate child Objects")

	// Test Object references non-existent class.
	scenario = Scenario{
		Key:  scenarioKey,
		Name: "Name",
		Objects: map[identity.Key]Object{
			objectKey: {Key: objectKey, ObjectNumber: 1, Name: "Obj", NameStyle: "name", ClassKey: nonExistentClassKey},
		},
	}
	err = scenario.ValidateWithParentAndClasses(&useCaseKey, classes)
	suite.Require().ErrorContains(err, "references non-existent class", "Should validate Object class references")
}

// TestSetObjects tests that SetObjects correctly sets objects.
func (suite *ScenarioSuite) TestSetObjects() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	useCaseKey := helper.Must(identity.NewUseCaseKey(subdomainKey, "usecase1"))
	scenarioKey := helper.Must(identity.NewScenarioKey(useCaseKey, "scenario1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	objectKey := helper.Must(identity.NewScenarioObjectKey(scenarioKey, "obj1"))

	scenario := Scenario{Key: scenarioKey, Name: "Name"}
	objects := map[identity.Key]Object{
		objectKey: {Key: objectKey, ObjectNumber: 1, Name: "Obj", NameStyle: "name", ClassKey: classKey},
	}
	scenario.SetObjects(objects)
	suite.Equal(objects, scenario.Objects)
}

package model_scenario

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestObjectSuite(t *testing.T) {
	suite.Run(t, new(ObjectSuite))
}

type ObjectSuite struct {
	suite.Suite
}

// TestValidate tests all validation rules for Object.
func (suite *ObjectSuite) TestValidate() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	useCaseKey := helper.Must(identity.NewUseCaseKey(subdomainKey, "usecase1"))
	scenarioKey := helper.Must(identity.NewScenarioKey(useCaseKey, "scenario1"))
	validKey := helper.Must(identity.NewScenarioObjectKey(scenarioKey, "obj1"))

	tests := []struct {
		testName string
		object   Object
		errstr   string
	}{
		{
			testName: "valid object with name style",
			object: Object{
				Key:       validKey,
				Name:      "Name",
				NameStyle: _NAME_STYLE_NAME,
				ClassKey:  classKey,
			},
		},
		{
			testName: "valid object with id style",
			object: Object{
				Key:       validKey,
				Name:      "Name",
				NameStyle: _NAME_STYLE_ID,
				ClassKey:  classKey,
			},
		},
		{
			testName: "valid object with unnamed style",
			object: Object{
				Key:       validKey,
				Name:      "",
				NameStyle: _NAME_STYLE_UNNAMED,
				ClassKey:  classKey,
			},
		},
		{
			testName: "error empty key",
			object: Object{
				Key:       identity.Key{},
				Name:      "Name",
				NameStyle: _NAME_STYLE_NAME,
				ClassKey:  classKey,
			},
			errstr: "keyType: cannot be blank",
		},
		{
			testName: "error wrong key type",
			object: Object{
				Key:       domainKey,
				Name:      "Name",
				NameStyle: _NAME_STYLE_NAME,
				ClassKey:  classKey,
			},
			errstr: "Key: invalid key type 'domain' for scenario object.",
		},
		{
			testName: "error blank name for name style",
			object: Object{
				Key:       validKey,
				Name:      "",
				NameStyle: _NAME_STYLE_NAME,
				ClassKey:  classKey,
			},
			errstr: "Name: Name cannot be blank",
		},
		{
			testName: "error name for unnamed style",
			object: Object{
				Key:       validKey,
				Name:      "Name",
				NameStyle: _NAME_STYLE_UNNAMED,
				ClassKey:  classKey,
			},
			errstr: "Name: Name must be blank for unnamed style",
		},
		{
			testName: "error empty class key",
			object: Object{
				Key:       validKey,
				Name:      "Name",
				NameStyle: _NAME_STYLE_NAME,
				ClassKey:  identity.Key{},
			},
			errstr: "keyType: cannot be blank",
		},
		{
			testName: "error wrong class key type",
			object: Object{
				Key:       validKey,
				Name:      "Name",
				NameStyle: _NAME_STYLE_NAME,
				ClassKey:  domainKey,
			},
			errstr: "ClassKey: invalid key type 'domain' for class.",
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.object.Validate()
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

// TestNew tests that NewObject maps parameters correctly and calls Validate.
func (suite *ObjectSuite) TestNew() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	useCaseKey := helper.Must(identity.NewUseCaseKey(subdomainKey, "usecase1"))
	scenarioKey := helper.Must(identity.NewScenarioKey(useCaseKey, "scenario1"))
	key := helper.Must(identity.NewScenarioObjectKey(scenarioKey, "obj1"))

	// Test parameters are mapped correctly.
	obj, err := NewObject(key, 1, "Name", _NAME_STYLE_NAME, classKey, true, "UmlComment")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), key, obj.Key)
	assert.Equal(suite.T(), uint(1), obj.ObjectNumber)
	assert.Equal(suite.T(), "Name", obj.Name)
	assert.Equal(suite.T(), _NAME_STYLE_NAME, obj.NameStyle)
	assert.Equal(suite.T(), classKey, obj.ClassKey)
	assert.Equal(suite.T(), true, obj.Multi)
	assert.Equal(suite.T(), "UmlComment", obj.UmlComment)

	// Test that Validate is called (invalid data should fail).
	_, err = NewObject(key, 1, "", _NAME_STYLE_NAME, classKey, true, "UmlComment")
	assert.ErrorContains(suite.T(), err, "Name: Name cannot be blank")
}

// TestValidateWithParent tests that ValidateWithParent calls Validate and ValidateParent.
func (suite *ObjectSuite) TestValidateWithParent() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	useCaseKey := helper.Must(identity.NewUseCaseKey(subdomainKey, "usecase1"))
	scenarioKey := helper.Must(identity.NewScenarioKey(useCaseKey, "scenario1"))
	validKey := helper.Must(identity.NewScenarioObjectKey(scenarioKey, "obj1"))
	otherScenarioKey := helper.Must(identity.NewScenarioKey(useCaseKey, "other_scenario"))

	// Test that Validate is called.
	obj := Object{
		Key:       validKey,
		Name:      "", // Invalid for name style
		NameStyle: _NAME_STYLE_NAME,
		ClassKey:  classKey,
	}
	err := obj.ValidateWithParent(&scenarioKey)
	assert.ErrorContains(suite.T(), err, "Name: Name cannot be blank", "ValidateWithParent should call Validate()")

	// Test that ValidateParent is called - object key has scenario1 as parent, but we pass other_scenario.
	obj = Object{
		Key:       validKey,
		Name:      "Name",
		NameStyle: _NAME_STYLE_NAME,
		ClassKey:  classKey,
	}
	err = obj.ValidateWithParent(&otherScenarioKey)
	assert.ErrorContains(suite.T(), err, "does not match expected parent", "ValidateWithParent should call ValidateParent()")

	// Test valid case.
	err = obj.ValidateWithParent(&scenarioKey)
	assert.NoError(suite.T(), err)
}

package model_scenario

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
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
	assert.Equal(suite.T(), Object{
		Key:          key,
		ObjectNumber: 1,
		Name:         "Name",
		NameStyle:    _NAME_STYLE_NAME,
		ClassKey:     classKey,
		Multi:        true,
		UmlComment:   "UmlComment",
	}, obj)

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

// TestValidateReferences tests that ValidateReferences validates class references correctly.
func (suite *ObjectSuite) TestValidateReferences() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	nonExistentClassKey := helper.Must(identity.NewClassKey(subdomainKey, "nonexistent"))
	useCaseKey := helper.Must(identity.NewUseCaseKey(subdomainKey, "usecase1"))
	scenarioKey := helper.Must(identity.NewScenarioKey(useCaseKey, "scenario1"))
	validKey := helper.Must(identity.NewScenarioObjectKey(scenarioKey, "obj1"))

	// Build lookup map with valid classes.
	classes := map[identity.Key]bool{
		classKey: true,
	}

	tests := []struct {
		testName string
		object   Object
		classes  map[identity.Key]bool
		errstr   string
	}{
		{
			testName: "valid object with existing class",
			object: Object{
				Key:       validKey,
				Name:      "Name",
				NameStyle: _NAME_STYLE_NAME,
				ClassKey:  classKey,
			},
			classes: classes,
		},
		{
			testName: "error ClassKey references non-existent class",
			object: Object{
				Key:       validKey,
				Name:      "Name",
				NameStyle: _NAME_STYLE_NAME,
				ClassKey:  nonExistentClassKey,
			},
			classes: classes,
			errstr:  "references non-existent class",
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.object.ValidateReferences(tt.classes)
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

// TestGetName tests that GetName formats the object name correctly based on NameStyle and Multi.
func (suite *ObjectSuite) TestGetName() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))

	class := model_class.Class{
		Key:  classKey,
		Name: "ClassName",
	}

	tests := []struct {
		testName string
		object   Object
		expected string
	}{
		{
			testName: "name style",
			object:   Object{Name: "objName", NameStyle: _NAME_STYLE_NAME},
			expected: "objName:ClassName",
		},
		{
			testName: "name style with multi",
			object:   Object{Name: "objName", NameStyle: _NAME_STYLE_NAME, Multi: true},
			expected: "*objName:ClassName",
		},
		{
			testName: "id style",
			object:   Object{Name: "123", NameStyle: _NAME_STYLE_ID},
			expected: "ClassName 123",
		},
		{
			testName: "id style with multi",
			object:   Object{Name: "123", NameStyle: _NAME_STYLE_ID, Multi: true},
			expected: "*ClassName 123",
		},
		{
			testName: "unnamed style",
			object:   Object{Name: "", NameStyle: _NAME_STYLE_UNNAMED},
			expected: ":ClassName",
		},
		{
			testName: "unnamed style with multi",
			object:   Object{Name: "", NameStyle: _NAME_STYLE_UNNAMED, Multi: true},
			expected: "*:ClassName",
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.testName, func(t *testing.T) {
			result := tt.object.GetName(class)
			assert.Equal(t, tt.expected, result)
		})
	}
}

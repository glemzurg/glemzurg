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

func (suite *ObjectSuite) TestNew() {

	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	useCaseKey := helper.Must(identity.NewUseCaseKey(subdomainKey, "usecase1"))
	scenarioKey := helper.Must(identity.NewScenarioKey(useCaseKey, "scenario1"))

	tests := []struct {
		testName     string
		key          identity.Key
		objectNumber uint
		name         string
		nameStyle    string
		classKey     identity.Key
		multi        bool
		umlComment   string
		obj          Object
		errstr       string
	}{
		// OK.
		{
			testName:     "ok with name style",
			key:          helper.Must(identity.NewScenarioObjectKey(scenarioKey, "obj1")),
			objectNumber: 1,
			name:         "Name",
			nameStyle:    "name",
			classKey:     classKey,
			multi:        true,
			umlComment:   "UmlComment",
			obj: Object{
				Key:          helper.Must(identity.NewScenarioObjectKey(scenarioKey, "obj1")),
				ObjectNumber: 1,
				Name:         "Name",
				NameStyle:    "name",
				ClassKey:     classKey,
				Multi:        true,
				UmlComment:   "UmlComment",
			},
		},
		{
			testName:     "ok with id style",
			key:          helper.Must(identity.NewScenarioObjectKey(scenarioKey, "obj2")),
			objectNumber: 1,
			name:         "Name",
			nameStyle:    "id",
			classKey:     classKey,
			multi:        true,
			umlComment:   "UmlComment",
			obj: Object{
				Key:          helper.Must(identity.NewScenarioObjectKey(scenarioKey, "obj2")),
				ObjectNumber: 1,
				Name:         "Name",
				NameStyle:    "id",
				ClassKey:     classKey,
				Multi:        true,
				UmlComment:   "UmlComment",
			},
		},
		{
			testName:     "ok with unnamed style",
			key:          helper.Must(identity.NewScenarioObjectKey(scenarioKey, "obj3")),
			objectNumber: 0,
			name:         "",
			nameStyle:    "unnamed",
			classKey:     classKey,
			multi:        false,
			umlComment:   "",
			obj: Object{
				Key:          helper.Must(identity.NewScenarioObjectKey(scenarioKey, "obj3")),
				ObjectNumber: 0,
				Name:         "",
				NameStyle:    "unnamed",
				ClassKey:     classKey,
				Multi:        false,
				UmlComment:   "",
			},
		},

		// Error states.
		{
			testName:     "error empty key",
			key:          identity.Key{},
			objectNumber: 1,
			name:         "Name",
			nameStyle:    "name",
			classKey:     classKey,
			multi:        false,
			umlComment:   "UmlComment",
			errstr:       "keyType: cannot be blank",
		},
		{
			testName:     "error wrong key type",
			key:          helper.Must(identity.NewDomainKey("domain1")),
			objectNumber: 1,
			name:         "Name",
			nameStyle:    "name",
			classKey:     classKey,
			multi:        false,
			umlComment:   "UmlComment",
			errstr:       "Key: invalid key type 'domain' for scenario object.",
		},
		{
			testName:     "error with blank name for name style",
			key:          helper.Must(identity.NewScenarioObjectKey(scenarioKey, "obj4")),
			objectNumber: 1,
			name:         "",
			nameStyle:    "name",
			classKey:     classKey,
			multi:        false,
			umlComment:   "UmlComment",
			errstr:       `Name: Name cannot be blank`,
		},
		{
			testName:     "error with blank name for id style",
			key:          helper.Must(identity.NewScenarioObjectKey(scenarioKey, "obj5")),
			objectNumber: 1,
			name:         "",
			nameStyle:    "id",
			classKey:     classKey,
			multi:        false,
			umlComment:   "UmlComment",
			errstr:       `Name: Name cannot be blank`,
		},
		{
			testName:     "error with name for unnamed style",
			key:          helper.Must(identity.NewScenarioObjectKey(scenarioKey, "obj6")),
			objectNumber: 1,
			name:         "Name",
			nameStyle:    "unnamed",
			classKey:     classKey,
			multi:        false,
			umlComment:   "UmlComment",
			errstr:       `Name: Name must be blank for unnamed style`,
		},
		{
			testName:     "error empty class key",
			key:          helper.Must(identity.NewScenarioObjectKey(scenarioKey, "obj7")),
			objectNumber: 1,
			name:         "Name",
			nameStyle:    "name",
			classKey:     identity.Key{},
			multi:        false,
			umlComment:   "UmlComment",
			errstr:       "keyType: cannot be blank",
		},
		{
			testName:     "error wrong class key type",
			key:          helper.Must(identity.NewScenarioObjectKey(scenarioKey, "obj8")),
			objectNumber: 1,
			name:         "Name",
			nameStyle:    "name",
			classKey:     helper.Must(identity.NewDomainKey("domain1")),
			multi:        false,
			umlComment:   "UmlComment",
			errstr:       "ClassKey: invalid key type 'domain' for class.",
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			obj, err := NewObject(tt.key, tt.objectNumber, tt.name, tt.nameStyle, tt.classKey, tt.multi, tt.umlComment)
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

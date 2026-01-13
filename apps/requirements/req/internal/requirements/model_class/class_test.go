package model_class

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestClassSuite(t *testing.T) {
	suite.Run(t, new(ClassSuite))
}

type ClassSuite struct {
	suite.Suite
}

func (suite *ClassSuite) TestNew() {

	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	actorKey := helper.Must(identity.NewActorKey("actor1"))
	generalizationKey := helper.Must(identity.NewGeneralizationKey(subdomainKey, "gen1"))

	tests := []struct {
		testName        string
		key             identity.Key
		name            string
		details         string
		actorKey        *identity.Key
		superclassOfKey *identity.Key
		subclassOfKey   *identity.Key
		umlComment      string
		obj             Class
		errstr          string
	}{
		// OK.
		{
			testName:        "ok with all fields",
			key:             helper.Must(identity.NewClassKey(subdomainKey, "class1")),
			name:            "Name",
			details:         "Details",
			actorKey:        &actorKey,
			superclassOfKey: &generalizationKey,
			subclassOfKey:   &generalizationKey,
			umlComment:      "UmlComment",
			obj: Class{
				Key:             helper.Must(identity.NewClassKey(subdomainKey, "class1")),
				Name:            "Name",
				Details:         "Details",
				ActorKey:        &actorKey,
				SuperclassOfKey: &generalizationKey,
				SubclassOfKey:   &generalizationKey,
				UmlComment:      "UmlComment",
			},
		},
		{
			testName:        "ok with minimal fields",
			key:             helper.Must(identity.NewClassKey(subdomainKey, "class2")),
			name:            "Name",
			details:         "",
			actorKey:        nil,
			superclassOfKey: nil,
			subclassOfKey:   nil,
			umlComment:      "",
			obj: Class{
				Key:             helper.Must(identity.NewClassKey(subdomainKey, "class2")),
				Name:            "Name",
				Details:         "",
				ActorKey:        nil,
				SuperclassOfKey: nil,
				SubclassOfKey:   nil,
				UmlComment:      "",
			},
		},

		// Error states.
		{
			testName:        "error empty key",
			key:             identity.Key{},
			name:            "Name",
			details:         "Details",
			actorKey:        &actorKey,
			superclassOfKey: &generalizationKey,
			subclassOfKey:   &generalizationKey,
			umlComment:      "UmlComment",
			errstr:          "keyType: cannot be blank",
		},
		{
			testName:        "error wrong key type",
			key:             helper.Must(identity.NewDomainKey("domain1")),
			name:            "Name",
			details:         "Details",
			actorKey:        &actorKey,
			superclassOfKey: &generalizationKey,
			subclassOfKey:   &generalizationKey,
			umlComment:      "UmlComment",
			errstr:          "Key: invalid key type 'domain' for class.",
		},
		{
			testName:        "error with blank name",
			key:             helper.Must(identity.NewClassKey(subdomainKey, "class3")),
			name:            "",
			details:         "Details",
			actorKey:        &actorKey,
			superclassOfKey: &generalizationKey,
			subclassOfKey:   &generalizationKey,
			umlComment:      "UmlComment",
			errstr:          `Name: cannot be blank`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			obj, err := NewClass(tt.key, tt.name, tt.details, tt.actorKey, tt.superclassOfKey, tt.subclassOfKey, tt.umlComment)
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

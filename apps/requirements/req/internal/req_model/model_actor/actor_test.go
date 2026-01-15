package model_actor

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestActorSuite(t *testing.T) {
	suite.Run(t, new(ActorSuite))
}

type ActorSuite struct {
	suite.Suite
}

func (suite *ActorSuite) TestNew() {
	tests := []struct {
		testName   string
		key        identity.Key
		name       string
		details    string
		userType   string
		umlComment string
		obj        Actor
		errstr     string
	}{
		// OK.
		{
			testName:   "ok with person type",
			key:        helper.Must(identity.NewActorKey("actor1")),
			name:       "Name",
			details:    "Details",
			userType:   _USER_TYPE_PERSON,
			umlComment: "UmlComment",
			obj: Actor{
				Key:        helper.Must(identity.NewActorKey("actor1")),
				Name:       "Name",
				Details:    "Details",
				Type:       "person",
				UmlComment: "UmlComment",
			},
		},
		{
			testName:   "ok with system type",
			key:        helper.Must(identity.NewActorKey("actor2")),
			name:       "Name",
			details:    "",
			userType:   _USER_TYPE_SYSTEM,
			umlComment: "",
			obj: Actor{
				Key:        helper.Must(identity.NewActorKey("actor2")),
				Name:       "Name",
				Details:    "",
				Type:       "system",
				UmlComment: "",
			},
		},

		// Error states.
		{
			testName:   "error empty key",
			key:        identity.Key{},
			name:       "Name",
			details:    "Details",
			userType:   _USER_TYPE_PERSON,
			umlComment: "UmlComment",
			errstr:     "keyType: cannot be blank",
		},
		{
			testName:   "error wrong key type",
			key:        helper.Must(identity.NewDomainKey("domain1")),
			name:       "Name",
			details:    "Details",
			userType:   _USER_TYPE_PERSON,
			umlComment: "UmlComment",
			errstr:     "Key: invalid key type 'domain' for actor.",
		},
		{
			testName:   "error with blank name",
			key:        helper.Must(identity.NewActorKey("actor3")),
			name:       "",
			details:    "Details",
			userType:   _USER_TYPE_PERSON,
			umlComment: "UmlComment",
			errstr:     `Name: cannot be blank.`,
		},
		{
			testName:   "error with blank type",
			key:        helper.Must(identity.NewActorKey("actor4")),
			name:       "Name",
			details:    "Details",
			userType:   "",
			umlComment: "UmlComment",
			errstr:     `Type: cannot be blank.`,
		},
		{
			testName:   "error with invalid type",
			key:        helper.Must(identity.NewActorKey("actor5")),
			name:       "Name",
			details:    "Details",
			userType:   "unknown",
			umlComment: "UmlComment",
			errstr:     `Type: must be a valid value.`,
		},
	}
	for _, tt := range tests {
		pass := suite.T().Run(tt.testName, func(t *testing.T) {
			obj, err := NewActor(tt.key, tt.name, tt.details, tt.userType, tt.umlComment)
			if tt.errstr == "" {
				assert.Nil(t, err)
				assert.Equal(t, tt.obj, obj)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
				assert.Empty(t, obj)
			}
		})
		if !pass {
			break
		}
	}
}

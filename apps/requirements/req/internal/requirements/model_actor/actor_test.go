package model_actor

import (
	"fmt"
	"testing"

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
		key        string
		name       string
		details    string
		userType   string
		umlComment string
		obj        Actor
		errstr     string
	}{
		// OK.
		{
			key:        "Key",
			name:       "Name",
			details:    "Details",
			userType:   _USER_TYPE_PERSON,
			umlComment: "UmlComment",
			obj: Actor{
				Key:        "Key",
				Name:       "Name",
				Details:    "Details",
				Type:       "person",
				UmlComment: "UmlComment",
			},
		},
		{
			key:        "Key",
			name:       "Name",
			details:    "",
			userType:   _USER_TYPE_SYSTEM,
			umlComment: "",
			obj: Actor{
				Key:        "Key",
				Name:       "Name",
				Details:    "",
				Type:       "system",
				UmlComment: "",
			},
		},

		// Error states.
		{
			key:        "",
			name:       "Name",
			details:    "Details",
			userType:   _USER_TYPE_PERSON,
			umlComment: "UmlComment",
			errstr:     `Key: cannot be blank`,
		},
		{
			key:        "Key",
			name:       "",
			details:    "Details",
			userType:   _USER_TYPE_PERSON,
			umlComment: "UmlComment",
			errstr:     `Name: cannot be blank.`,
		},
		{
			key:        "Key",
			name:       "Name",
			details:    "Details",
			userType:   "",
			umlComment: "UmlComment",
			errstr:     `Type: cannot be blank.`,
		},
		{
			key:        "Key",
			name:       "Name",
			details:    "Details",
			userType:   "unknown",
			umlComment: "UmlComment",
			errstr:     `Type: must be a valid value.`,
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		obj, err := NewActor(test.key, test.name, test.details, test.userType, test.umlComment)
		if test.errstr == "" {
			assert.Nil(suite.T(), err, testName)
			assert.Equal(suite.T(), test.obj, obj, testName)
		} else {
			assert.ErrorContains(suite.T(), err, test.errstr, testName)
			assert.Empty(suite.T(), obj, testName)
		}
	}
}

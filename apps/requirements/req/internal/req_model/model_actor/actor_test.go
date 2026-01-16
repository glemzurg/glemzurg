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

// TestValidate tests all validation rules for Actor.
func (suite *ActorSuite) TestValidate() {
	validKey := helper.Must(identity.NewActorKey("actor1"))

	tests := []struct {
		testName string
		actor    Actor
		errstr   string
	}{
		{
			testName: "valid actor with person type",
			actor: Actor{
				Key:  validKey,
				Name: "Name",
				Type: _USER_TYPE_PERSON,
			},
		},
		{
			testName: "valid actor with system type",
			actor: Actor{
				Key:  validKey,
				Name: "Name",
				Type: _USER_TYPE_SYSTEM,
			},
		},
		{
			testName: "error empty key",
			actor: Actor{
				Key:  identity.Key{},
				Name: "Name",
				Type: _USER_TYPE_PERSON,
			},
			errstr: "keyType: cannot be blank",
		},
		{
			testName: "error wrong key type",
			actor: Actor{
				Key:  helper.Must(identity.NewDomainKey("domain1")),
				Name: "Name",
				Type: _USER_TYPE_PERSON,
			},
			errstr: "Key: invalid key type 'domain' for actor.",
		},
		{
			testName: "error blank name",
			actor: Actor{
				Key:  validKey,
				Name: "",
				Type: _USER_TYPE_PERSON,
			},
			errstr: "Name: cannot be blank",
		},
		{
			testName: "error blank type",
			actor: Actor{
				Key:  validKey,
				Name: "Name",
				Type: "",
			},
			errstr: "Type: cannot be blank",
		},
		{
			testName: "error invalid type",
			actor: Actor{
				Key:  validKey,
				Name: "Name",
				Type: "unknown",
			},
			errstr: "Type: must be a valid value",
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.actor.Validate()
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

// TestNew tests that NewActor maps parameters correctly and calls Validate.
func (suite *ActorSuite) TestNew() {
	key := helper.Must(identity.NewActorKey("actor1"))

	// Test parameters are mapped correctly.
	actor, err := NewActor(key, "Name", "Details", _USER_TYPE_PERSON, "UmlComment")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), Actor{
		Key:        key,
		Name:       "Name",
		Details:    "Details",
		Type:       _USER_TYPE_PERSON,
		UmlComment: "UmlComment",
	}, actor)

	// Test that Validate is called (invalid data should fail).
	_, err = NewActor(key, "", "Details", _USER_TYPE_PERSON, "UmlComment")
	assert.ErrorContains(suite.T(), err, "Name: cannot be blank")
}

// TestValidateWithParent tests that ValidateWithParent calls Validate and ValidateParent.
func (suite *ActorSuite) TestValidateWithParent() {
	validKey := helper.Must(identity.NewActorKey("actor1"))

	// Test that Validate is called.
	actor := Actor{
		Key:  validKey,
		Name: "", // Invalid
		Type: _USER_TYPE_PERSON,
	}
	err := actor.ValidateWithParent(nil)
	assert.ErrorContains(suite.T(), err, "Name: cannot be blank", "ValidateWithParent should call Validate()")

	// Test that ValidateParent is called - actors should have nil parent.
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	actor = Actor{
		Key:  validKey,
		Name: "Name",
		Type: _USER_TYPE_PERSON,
	}
	err = actor.ValidateWithParent(&domainKey)
	assert.ErrorContains(suite.T(), err, "should not have a parent", "ValidateWithParent should call ValidateParent()")

	// Test valid case.
	err = actor.ValidateWithParent(nil)
	assert.NoError(suite.T(), err)
}

package model_actor

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
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
	genKeyA := helper.Must(identity.NewActorGeneralizationKey("gen_a"))
	genKeyB := helper.Must(identity.NewActorGeneralizationKey("gen_b"))

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
			testName: "valid actor with superclass and subclass",
			actor: Actor{
				Key:             validKey,
				Name:            "Name",
				Type:            _USER_TYPE_PERSON,
				SuperclassOfKey: &genKeyA,
				SubclassOfKey:   &genKeyB,
			},
		},
		{
			testName: "error empty key",
			actor: Actor{
				Key:  identity.Key{},
				Name: "Name",
				Type: _USER_TYPE_PERSON,
			},
			errstr: "ACTOR_KEY_INVALID",
		},
		{
			testName: "error wrong key type",
			actor: Actor{
				Key:  helper.Must(identity.NewDomainKey("domain1")),
				Name: "Name",
				Type: _USER_TYPE_PERSON,
			},
			errstr: "invalid key type 'domain' for actor",
		},
		{
			testName: "error blank name",
			actor: Actor{
				Key:  validKey,
				Name: "",
				Type: _USER_TYPE_PERSON,
			},
			errstr: "Name is required",
		},
		{
			testName: "error blank type",
			actor: Actor{
				Key:  validKey,
				Name: "Name",
				Type: "",
			},
			errstr: "Type is required",
		},
		{
			testName: "error invalid type",
			actor: Actor{
				Key:  validKey,
				Name: "Name",
				Type: "unknown",
			},
			errstr: "Type must be one of",
		},
		{
			testName: "error superclass and subclass same key",
			actor: Actor{
				Key:             validKey,
				Name:            "Name",
				Type:            _USER_TYPE_PERSON,
				SuperclassOfKey: &genKeyA,
				SubclassOfKey:   &genKeyA,
			},
			errstr: "SuperclassOfKey and SubclassOfKey cannot be the same",
		},
		{
			testName: "error SuperclassOfKey wrong key type",
			actor: func() Actor {
				wrongKey := helper.Must(identity.NewDomainKey("domain1"))
				return Actor{
					Key:             validKey,
					Name:            "Name",
					Type:            _USER_TYPE_PERSON,
					SuperclassOfKey: &wrongKey,
				}
			}(),
			errstr: "SuperclassOfKey: invalid key type 'domain' for actor generalization",
		},
		{
			testName: "error SubclassOfKey wrong key type",
			actor: func() Actor {
				wrongKey := helper.Must(identity.NewDomainKey("domain1"))
				return Actor{
					Key:           validKey,
					Name:          "Name",
					Type:          _USER_TYPE_PERSON,
					SubclassOfKey: &wrongKey,
				}
			}(),
			errstr: "SubclassOfKey: invalid key type 'domain' for actor generalization",
		},
	}
	for _, tt := range tests {
		suite.Run(tt.testName, func() {
			err := tt.actor.Validate()
			if tt.errstr == "" {
				suite.Require().NoError(err)
			} else {
				suite.Require().ErrorContains(err, tt.errstr)
			}
		})
	}
}

// TestNew tests that NewActor maps parameters correctly and calls Validate.
func (suite *ActorSuite) TestNew() {
	key := helper.Must(identity.NewActorKey("actor1"))
	genKeyA := helper.Must(identity.NewActorGeneralizationKey("gen_a"))
	genKeyB := helper.Must(identity.NewActorGeneralizationKey("gen_b"))

	// Test parameters are mapped correctly.
	actor, err := NewActor(key, "Name", "Details", _USER_TYPE_PERSON, &genKeyA, &genKeyB, "UmlComment")
	suite.Require().NoError(err)
	suite.Equal(Actor{
		Key:             key,
		Name:            "Name",
		Details:         "Details",
		Type:            _USER_TYPE_PERSON,
		SuperclassOfKey: &genKeyA,
		SubclassOfKey:   &genKeyB,
		UmlComment:      "UmlComment",
	}, actor)

	// Test with nil superclass/subclass.
	actor, err = NewActor(key, "Name", "Details", _USER_TYPE_PERSON, nil, nil, "UmlComment")
	suite.Require().NoError(err)
	suite.Equal(Actor{
		Key:        key,
		Name:       "Name",
		Details:    "Details",
		Type:       _USER_TYPE_PERSON,
		UmlComment: "UmlComment",
	}, actor)

	// Test that Validate is called (invalid data should fail).
	_, err = NewActor(key, "", "Details", _USER_TYPE_PERSON, nil, nil, "UmlComment")
	suite.Require().ErrorContains(err, "Name is required")
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
	suite.Require().ErrorContains(err, "Name is required", "ValidateWithParent should call Validate()")

	// Test that ValidateParent is called - actors should have nil parent.
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	actor = Actor{
		Key:  validKey,
		Name: "Name",
		Type: _USER_TYPE_PERSON,
	}
	err = actor.ValidateWithParent(&domainKey)
	suite.Require().ErrorContains(err, "should not have a parent", "ValidateWithParent should call ValidateParent()")

	// Test valid case.
	err = actor.ValidateWithParent(nil)
	suite.Require().NoError(err)
}

// TestValidateReferences tests that ValidateReferences validates actor references.
func (suite *ActorSuite) TestValidateReferences() {
	validKey := helper.Must(identity.NewActorKey("actor1"))
	genKeyA := helper.Must(identity.NewActorGeneralizationKey("gen_a"))
	genKeyB := helper.Must(identity.NewActorGeneralizationKey("gen_b"))
	genKeyC := helper.Must(identity.NewActorGeneralizationKey("gen_c"))

	generalizations := map[identity.Key]bool{
		genKeyA: true,
		genKeyB: true,
	}

	// Valid: references existing generalizations.
	actor := Actor{
		Key:             validKey,
		Name:            "Name",
		Type:            _USER_TYPE_PERSON,
		SuperclassOfKey: &genKeyA,
		SubclassOfKey:   &genKeyB,
	}
	err := actor.ValidateReferences(generalizations)
	suite.Require().NoError(err)

	// Valid: no references.
	actor = Actor{
		Key:  validKey,
		Name: "Name",
		Type: _USER_TYPE_PERSON,
	}
	err = actor.ValidateReferences(generalizations)
	suite.Require().NoError(err)

	// Error: superclass references non-existent generalization.
	actor = Actor{
		Key:             validKey,
		Name:            "Name",
		Type:            _USER_TYPE_PERSON,
		SuperclassOfKey: &genKeyC,
	}
	err = actor.ValidateReferences(generalizations)
	suite.Require().ErrorContains(err, "non-existent generalization")

	// Error: subclass references non-existent generalization.
	actor = Actor{
		Key:           validKey,
		Name:          "Name",
		Type:          _USER_TYPE_PERSON,
		SubclassOfKey: &genKeyC,
	}
	err = actor.ValidateReferences(generalizations)
	suite.Require().ErrorContains(err, "non-existent generalization")
}
